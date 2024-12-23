package main

import (
	"RedisServer-project/client"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"
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
	quitCh chan *struct{}
	msgCh  chan Message

	kv *KV
}

type Message struct {
	data []byte
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
	cmd, err := parseCommand(string(msg.data))
	if err != nil {
		return err
	}

	switch v := cmd.(type) {

	case SetCommand:
		return s.kv.Set(v.key, v.val)

	case GetCommand:
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}

		_, err := msg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
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
				slog.Error("raw message error", "err", err)
			}

		case peer := <-s.addCh:
			s.peers[peer] = true

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
	peer := NewPeer(conn, s.msgCh)
	s.addCh <- peer

	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err)
	}
}

func main() {
	server := NewServer(ServerConfig{})

	go func() {
		log.Fatal(server.Start())
	}()

	time.Sleep(time.Second)

	c := client.NewClient("localhost:5001")
	for i := 0; i < 10; i++ {
		if err := c.Set(context.TODO(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second)

		val, err := c.Get(context.TODO(), fmt.Sprintf("foo_%d", i))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("got back: ", val)
	}

	time.Sleep(time.Second)
	fmt.Println(server.kv.data)
}
