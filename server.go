package main

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/tidwall/resp"
)

const defaultListenAddr = ":5001"

// Server Config struct
type ServerConfig struct {
	ListenAddr string
}

// Server struct
type Server struct {
	ServerConfig
	peers  map[*Peer]bool
	ln     net.Listener
	addCh  chan *Peer
	delCh  chan *Peer
	quitCh chan *struct{}
	msgCh  chan Message

	kv *KV
}

// Message struct
type Message struct {
	cmd  Command
	peer *Peer
}

// Server Constructor
func NewServer(cfg ServerConfig) *Server {
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}
	return &Server{
		ServerConfig: cfg,
		peers:        make(map[*Peer]bool),
		addCh:        make(chan *Peer),
		delCh:        make(chan *Peer),
		quitCh:       make(chan *struct{}),
		msgCh:        make(chan Message),
		kv:           NewKV(),
	}
}

// Server Start Handler
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()

	logMessage(slog.LevelInfo, "Server Running", "listenAddr", s.ListenAddr)

	return s.acceptLoop()
}

// Server Shutdown Handler
func (s *Server) Shutdown() {
	close(s.quitCh)
	s.ln.Close()

	for peer := range s.peers {
		peer.conn.Close()
	}

	logMessage(slog.LevelInfo, "Server Shutdown Completed")
}

// Chan Message Handler
func (s *Server) handleMessage(msg Message) error {
	// logMessage(slog.LevelInfo, "Received Message from Client", "type", reflect.TypeOf(msg.cmd))

	switch v := msg.cmd.(type) {

	case HelloCommand:
		spec := map[string]string{
			"server": "redis",
		}

		if _, err := msg.peer.Send(respWriteMap(spec)); err != nil {
			logMessage(slog.LevelError, "Hello Command Failed", "err", err)
			return err
		}

	case ClientCommand:
		if err := resp.NewWriter(msg.peer.conn).WriteString("OK"); err != nil {
			logMessage(slog.LevelError, "Client Command Response Failed", "err", err)
			return err
		}

	case SetCommand:
		if err := s.kv.Set(v.key, v.val); err != nil {
			logMessage(slog.LevelError, "SET Command Failed", "err", err, "key", string(v.key))
			return resp.NewWriter(msg.peer.conn).WriteString("(null)")
		}

		if err := resp.NewWriter(msg.peer.conn).WriteString("OK"); err != nil {
			logMessage(slog.LevelError, "Response Write Failed for SET Command", "err", err, "key", string(v.key))
			return err
		}

		logMessage(slog.LevelInfo, "Client Executed SET Command", "key", v.key, "val", v.val, "peer", msg.peer.conn.RemoteAddr())

	case GetCommand:
		val, ok := s.kv.Get(v.key)
		if !ok {
			err := fmt.Errorf("key not found")
			logMessage(slog.LevelError, "GET Command Failed", "err", err, "key", string(v.key))
			return resp.NewWriter(msg.peer.conn).WriteString("(nil)")
		}

		if err := resp.NewWriter(msg.peer.conn).WriteString(string(val)); err != nil {
			logMessage(slog.LevelError, "Response Write Failed for GET Command", "err", err, "key", string(v.key))
			return err
		}

		logMessage(slog.LevelInfo, "Client Executed GET Command", "key", v.key, "peer", msg.peer.conn.RemoteAddr())

	case DelCommand:
		ok := s.kv.Del(v.key)
		if !ok {
			err := fmt.Errorf("key not found")
			logMessage(slog.LevelError, "DEL Command Failed", "err", err, "key", string(v.key))
			return resp.NewWriter(msg.peer.conn).WriteString("0")
		}

		if err := resp.NewWriter(msg.peer.conn).WriteString("1"); err != nil {
			logMessage(slog.LevelError, "Response Write Failed for DEL Command", "err", err, "key", string(v.key))
			return err
		}

	}

	return nil
}

// Channel Manager Loop
func (s *Server) loop() {
	for {
		select {

		case <-s.quitCh:
			return

		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				logMessage(slog.LevelError, "Message Handling Error", "err", err)
			}

		case peer := <-s.addCh:
			logMessage(slog.LevelInfo, "New Peer Connected", "remoteAddr", peer.conn.RemoteAddr())
			s.peers[peer] = true

		case peer := <-s.delCh:
			logMessage(slog.LevelInfo, "Peer Disconnected", "remoteAddr", peer.conn.RemoteAddr())
			delete(s.peers, peer)
		}
	}
}

// Connection Accept Loop
func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			logMessage(slog.LevelError, "Connection Accept Error", "err", err)
			continue
		}

		go s.handleConn(conn)
	}
}

// Handle Peer Connections
func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh, s.delCh)
	s.addCh <- peer

	if err := peer.readLoop(); err != nil {
		logMessage(slog.LevelError, "Peer Read Error", "err", err, "remoteAddr", conn.RemoteAddr())
	}
}
