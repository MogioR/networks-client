package events

import (
	"bytes"
	"encoding/binary"
)

type InviteToChatEvent struct {
	chatId int64
	user   string
}

func (e *InviteToChatEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(7)

	binary.Write(buf, binary.BigEndian, e.chatId)

	buf.WriteByte(byte(len(e.user)))
	buf.WriteString(e.user)

	return buf
}
