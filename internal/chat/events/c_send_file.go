package events

import (
	"bytes"
	"encoding/binary"
)

type SendFileInitEvent struct {
	ChatId   int64
	FileName string
	FileSize int32
}

func (e *SendFileInitEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(9)

	binary.Write(buf, binary.BigEndian, e.ChatId)

	binary.Write(buf, binary.BigEndian, int16(len(e.FileName)))
	buf.WriteString(e.FileName)

	binary.Write(buf, binary.BigEndian, e.FileSize)

	return buf
}
