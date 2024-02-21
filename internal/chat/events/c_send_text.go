package events

import (
	"bytes"
	"encoding/binary"
)

type SendTextEvent struct {
	ChatId  int64
	Message string
}

func (e *SendTextEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(8)

	binary.Write(buf, binary.BigEndian, e.ChatId)

	buf.WriteByte(byte(len(e.Message)))
	buf.WriteString(e.Message)

	return buf
}
