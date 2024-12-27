package main

import (
	"bytes"
	"fmt"

	"github.com/tidwall/resp"
)

const (
	CommandSET    = "SET"
	CommandGET    = "GET"
	CommandHELLO  = "hello"
	CommandCLIENT = "client"
)

type Command interface {
}

type SetCommand struct {
	key, val []byte
}

type GetCommand struct {
	key []byte
}

type HelloCommand struct {
	val string
}

type ClientCommand struct {
	val string
}

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
