package events

import (
	"bytes"
	"encoding/binary"
	"time"
)

type NewMessageEvent struct {
	MessageId   int64
	ChatId      int64
	UserId      int64
	MessageType bool
	Message     string
	SendTime    time.Time
	ReadTime    time.Time
}

func (c *NewMessageEvent) Deserialize(buf *bytes.Buffer) error {
	if err := binary.Read(buf, binary.BigEndian, &c.MessageId); err != nil {
		return err
	}

	if err := binary.Read(buf, binary.BigEndian, &c.ChatId); err != nil {
		return err
	}

	if err := binary.Read(buf, binary.BigEndian, &c.UserId); err != nil {
		return err
	}

	var messageTypeByte byte
	if err := binary.Read(buf, binary.BigEndian, &messageTypeByte); err != nil {
		return err
	}
	c.MessageType = messageTypeByte != 0

	var messageLength uint16
	if err := binary.Read(buf, binary.BigEndian, &messageLength); err != nil {
		return err
	}

	messageBytes := make([]byte, messageLength)
	if _, err := buf.Read(messageBytes); err != nil {
		return err
	}
	c.Message = string(messageBytes)

	var postedAtUnix, readedAtUnix int64
	if err := binary.Read(buf, binary.BigEndian, &postedAtUnix); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &readedAtUnix); err != nil {
		return err
	}
	c.SendTime = time.Unix(postedAtUnix, 0)
	c.ReadTime = time.Unix(readedAtUnix, 0)

	return nil
}
