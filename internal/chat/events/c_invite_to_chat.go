package events

import (
	"bytes"
	"encoding/binary"
)

type InviteToChatEvent struct {
	ChatId int64
	User   string
}

func (e *InviteToChatEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(7)

	binary.Write(buf, binary.BigEndian, e.ChatId)

	buf.WriteByte(byte(len(e.User)))
	buf.WriteString(e.User)

	return buf
}
