package events

import (
	"bytes"
	"encoding/binary"
)

type SystemMessageEvent struct {
	Code    int8
	Message string
}

func (c *SystemMessageEvent) Deserialize(msg *bytes.Buffer) error {

	err := binary.Read(msg, binary.BigEndian, &c.Code)
	if err != nil {
		return err
	}

	messageLength, err := msg.ReadByte()
	if err != nil {
		return err
	}
	messageBytes := make([]byte, messageLength)
	msg.Read(messageBytes)
	c.Message = string(messageBytes)

	return nil
}
