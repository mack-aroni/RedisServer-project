package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/tidwall/resp"
)

// Holds RESP Connection/Server Interaction Chans
type Peer struct {
	conn  net.Conn
	msgCh chan Message
	delCh chan *Peer
}

// Peer Constructor
func NewPeer(conn net.Conn, msgCh chan Message, delCh chan *Peer) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
		delCh: delCh,
	}
}

// Send to RESP/Peer
func (p *Peer) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)
}

// Attempt to Parse RESP Command
func parseCommand(v resp.Value) (Command, error) {
	// Check for valid RESP array
	if v.Type() != resp.Array {
		return nil, fmt.Errorf("invalid command format: expected array, got %s", v.Type())
	}

	// Error check empty array
	rawArray := v.Array()
	if len(rawArray) == 0 {
		return nil, fmt.Errorf("empty command array")
	}

	// Extract CMD from first index
	rawCMD := rawArray[0].String()
	switch rawCMD {

	case CommandHELLO:
		if len(rawArray) < 2 {
			return nil, fmt.Errorf("missing argument for HELLO command")
		}
		return HelloCommand{val: rawArray[1].String()}, nil

	case CommandCLIENT:
		if len(rawArray) < 2 {
			return nil, fmt.Errorf("missing argument for CLIENT command")
		}
		return ClientCommand{val: rawArray[1].String()}, nil

	case CommandGET:
		if len(rawArray) < 2 {
			return nil, fmt.Errorf("missing argument for GET command")
		}
		return GetCommand{key: rawArray[1].Bytes()}, nil

	case CommandSET:
		if len(rawArray) < 3 {
			return nil, fmt.Errorf("missing arguments for SET command")
		}
		return SetCommand{key: rawArray[1].Bytes(), val: rawArray[2].Bytes()}, nil

	case CommandDEL:
		if len(rawArray) < 2 {
			return nil, fmt.Errorf("missing argument for DEL command")
		}
		return DelCommand{key: rawArray[1].Bytes()}, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", rawCMD)
	}
}

// Peer Readloop from RESP->Server
func (p *Peer) readLoop() error {
	rd := resp.NewReader(p.conn)

	for {
		// Read a RESP value from the connection
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			logMessage(slog.LevelInfo, "Disconnecting", "remoteAddr", p.conn.RemoteAddr())
			p.delCh <- p
			break
		}
		if err != nil {
			logMessage(slog.LevelError, "Error reading value", "err", err, "remoteAddr", p.conn.RemoteAddr())
			continue
		}

		// Parse the command using the parseCommand function
		logMessage(slog.LevelInfo, "Peer received RESP value", "value", v)
		cmd, err := parseCommand(v)
		if err != nil {
			logMessage(slog.LevelError, "Command parsing error", "err", err, "rawValue", v, "remoteAddr", p.conn.RemoteAddr())
			continue
		}

		// Send parsed command to the server
		p.msgCh <- Message{cmd: cmd, peer: p}
	}
	return nil
}
