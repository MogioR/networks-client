package events

import (
	"bytes"
	"encoding/binary"
	"time"
)

type ChatEvent struct {
	Chats []*Chat
}

func (c *ChatEvent) Deserialize(buf *bytes.Buffer) error {
	var chatCount int8
	if err := binary.Read(buf, binary.BigEndian, &chatCount); err != nil {
		return err
	}

	c.Chats = make([]*Chat, 0, chatCount)

	for i := int8(0); i < chatCount; i++ {
		chat, err := DeserializeChat(buf)
		if err != nil {
			return err
		}
		c.Chats = append(c.Chats, chat)
	}

	return nil
}

func DeserializeChat(buf *bytes.Buffer) (*Chat, error) {
	chat := &Chat{}

	// Считываем Id чата
	if err := binary.Read(buf, binary.BigEndian, &chat.ChatId); err != nil {
		return nil, err
	}

	// Считываем длину имени чата
	chatNameLength, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	// Считываем название чата
	chatNameBytes := make([]byte, chatNameLength)
	if _, err := buf.Read(chatNameBytes); err != nil {
		return nil, err
	}
	chat.ChatName = string(chatNameBytes)

	// Считываем количество пользователей в чате
	var numUsers byte
	if err := binary.Read(buf, binary.BigEndian, &numUsers); err != nil {
		return nil, err
	}

	// Считываем каждого пользователя
	for i := byte(0); i < numUsers; i++ {
		user := User{}
		if err := binary.Read(buf, binary.BigEndian, &user.Id); err != nil {
			return nil, err
		}

		// Считываем длину логина пользователя
		loginLength, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}

		// Считываем логин пользователя
		loginBytes := make([]byte, loginLength)
		if _, err := buf.Read(loginBytes); err != nil {
			return nil, err
		}
		user.Login = string(loginBytes)

		chat.Users = append(chat.Users, user)
	}

	// Считываем количество сообщений в чате
	var numMessages byte
	if err := binary.Read(buf, binary.BigEndian, &numMessages); err != nil {
		return nil, err
	}

	// Считываем каждое сообщение
	for i := byte(0); i < numMessages; i++ {
		message := Message{}
		if err := binary.Read(buf, binary.BigEndian, &message.Id); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &message.SenderId); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &message.MessageType); err != nil {
			return nil, err
		}

		// Считываем длину сообщения
		var messageLength uint16
		if err := binary.Read(buf, binary.BigEndian, &messageLength); err != nil {
			return nil, err
		}

		// Считываем сообщение
		messageBytes := make([]byte, messageLength)
		if _, err := buf.Read(messageBytes); err != nil {
			return nil, err
		}
		message.Message = string(messageBytes)

		var postedAtUnix, readedAtUnix int64
		if err := binary.Read(buf, binary.BigEndian, &postedAtUnix); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &readedAtUnix); err != nil {
			return nil, err
		}
		message.SendTime = time.Unix(postedAtUnix, 0)
		message.ReadTime = time.Unix(readedAtUnix, 0)

		chat.Messages = append(chat.Messages, message)
	}

	return chat, nil
}
