package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"reflect"
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

	slog.Info("server running", "listenAddr", s.ListenAddr)

	return s.acceptLoop()
}

func (s *Server) handleMessage(msg Message) error {
	slog.Info("go message from client", "type", reflect.TypeOf(msg.cmd))

	switch v := msg.cmd.(type) {

	case SetCommand:
		return s.kv.Set(v.key, v.val)

	case GetCommand:
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}

		_, err := msg.peer.Send(val)
		if err != nil {
			return fmt.Errorf("peer send error: %s", err)
		}

	case HelloCommand:
		spec := map[string]string{
			"server":  "redis",
			"version": "6.0.0",
			"proto":   "3",
			"mode":    "standalone",
			"role":    "master",
		}

		_, err := msg.peer.Send(respWriteMap(spec))
		if err != nil {
			return err
		}

		fmt.Println("client hello command")
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
				slog.Error("raw message error", "err", err)
			}

		case peer := <-s.addCh:
			slog.Info("new peer connected", "remoteAddr", peer.conn.RemoteAddr())
			s.peers[peer] = true

		case peer := <-s.delCh:
			slog.Info("peer disconnected", "remoteAddr", peer.conn.RemoteAddr())
			delete(s.peers, peer)
		}
	}
}

// can merge with start handler
func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
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
		slog.Error("peer read error", "err", err)
	}
}

// Main Server Run Code
func main() {
	listenAddr := flag.String("listenAddr", defaultListenAddr, "listen address of goredis server")
	flag.Parse()
	server := NewServer(ServerConfig{
		ListenAddr: *listenAddr,
	})
	log.Fatal(server.Start())
}
