package events

import (
	"bytes"
	"encoding/binary"
)

type ReadMessageEvent struct {
	chatId    int64
	messageId int64
}

func (e *ReadMessageEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(14)

	binary.Write(buf, binary.BigEndian, e.chatId)
	binary.Write(buf, binary.BigEndian, e.messageId)

	return buf
}
