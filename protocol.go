package main

import (
	"bytes"
	"fmt"

	"github.com/tidwall/resp"
)

const (
	CommandHELLO  = "hello"
	CommandCLIENT = "client"
	CommandSET    = "set"
	CommandGET    = "get"
	CommandDEL    = "del"
)

type Command interface {
}

type HelloCommand struct {
	val string
}

type ClientCommand struct {
	val string
}

type SetCommand struct {
	key, val []byte
}

type GetCommand struct {
	key []byte
}

type DelCommand struct {
	key []byte
}

// Helper Function to Parse KV into RESP Format
func respWriteMap(m map[string]string) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("%" + fmt.Sprintf("%d\r\n", len(m)))
	rw := resp.NewWriter(buf)

	for k, v := range m {
		rw.WriteString(k)
		rw.WriteString(":" + v)
	}

	return buf.Bytes()
}
