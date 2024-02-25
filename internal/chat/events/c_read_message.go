package events

import (
	"bytes"
	"encoding/binary"
)

type ReadMessageEvent struct {
	ChatId    int64
	MessageId int64
}

func (e *ReadMessageEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(14)

	binary.Write(buf, binary.BigEndian, e.ChatId)
	binary.Write(buf, binary.BigEndian, e.MessageId)

	return buf
}
