package events

import (
	"bytes"
	"encoding/binary"
)

type LeaveFromChatEvent struct {
	chatId int64
}

func (e *LeaveFromChatEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(5)

	binary.Write(buf, binary.BigEndian, e.chatId)

	return buf
}
