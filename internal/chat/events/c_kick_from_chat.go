package events

import (
	"bytes"
	"encoding/binary"
)

type KickFromChatEvent struct {
	chatId int64
	user   string
}

func (e *KickFromChatEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(6)

	binary.Write(buf, binary.BigEndian, e.chatId)

	buf.WriteByte(byte(len(e.user)))
	buf.WriteString(e.user)

	return buf
}
