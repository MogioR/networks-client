package events

import (
	"bytes"
	"encoding/binary"
)

type ReadedMessageEvent struct {
	ChatId    int64
	MessageId int64
}

func (r *ReadedMessageEvent) DeserializeReadedMessageEvent(msg *bytes.Buffer) error {
	err := binary.Read(msg, binary.BigEndian, &r.ChatId)
	if err != nil {
		return err
	}

	err = binary.Read(msg, binary.BigEndian, &r.MessageId)
	if err != nil {
		return err
	}

	return nil
}
