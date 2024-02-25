package events

import (
	"bytes"
	"encoding/binary"
)

type KickFromChatEvent struct {
	ChatId int64
	User   string
}

func (e *KickFromChatEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(6)

	binary.Write(buf, binary.BigEndian, e.ChatId)

	buf.WriteByte(byte(len(e.User)))
	buf.WriteString(e.User)

	return buf
}
