package events

import (
	"bytes"
	"encoding/binary"
)

type GetFileFromChatEvent struct {
	ChatId    int64
	MessageId int64
}

func (e *GetFileFromChatEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(13)

	binary.Write(buf, binary.BigEndian, e.ChatId)
	binary.Write(buf, binary.BigEndian, e.MessageId)

	return buf
}
