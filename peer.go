package main

import (
	"fmt"
	"io"
	"log"
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

// Peer Readloop from RESP->Server
func (p *Peer) readLoop() error {
	rd := resp.NewReader(p.conn)

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			p.delCh <- p
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// Parse RESP Commands Into Server Commands
		var cmd Command
		if v.Type() == resp.Array {
			rawCMD := v.Array()[0]

			switch rawCMD.String() {

			case CommandGET:
				cmd = GetCommand{
					key: v.Array()[1].Bytes(),
				}

			case CommandSET:
				cmd = SetCommand{
					key: v.Array()[1].Bytes(),
					val: v.Array()[2].Bytes(),
				}

			case CommandHELLO:
				cmd = HelloCommand{
					val: v.Array()[1].String(),
				}

			case CommandCLIENT:
				cmd = ClientCommand{
					val: v.Array()[1].String(),
				}

			default:
				fmt.Printf("got unkown command => %+v\n", v.Array())

			}

			p.msgCh <- Message{
				cmd:  cmd,
				peer: p,
			}
		}
	}
	return nil
}
