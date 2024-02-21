package events

import (
	"bytes"
	"encoding/binary"
)

type ChatUsersUpdateEvent struct {
	ChatId int64
	Users  []User
}

func (c *ChatUsersUpdateEvent) DeserializeChatUsersUpdateEvent(msg *bytes.Buffer) error {
	// Читаем chatId
	if err := binary.Read(msg, binary.BigEndian, &c.ChatId); err != nil {
		return err
	}

	// Читаем количество пользователей в чате
	var numUsers byte
	if err := binary.Read(msg, binary.BigEndian, &numUsers); err != nil {
		return err
	}

	// Читаем каждого пользователя
	c.Users = make([]User, numUsers)
	for i := 0; i < int(numUsers); i++ {
		user := User{}

		// Читаем Id пользователя
		if err := binary.Read(msg, binary.BigEndian, &user.Id); err != nil {
			return err
		}

		// Читаем длину логина пользователя
		var loginLength byte
		if err := binary.Read(msg, binary.BigEndian, &loginLength); err != nil {
			return err
		}

		// Читаем логин пользователя
		loginBytes := make([]byte, loginLength)
		if _, err := msg.Read(loginBytes); err != nil {
			return err
		}
		user.Login = string(loginBytes)

		c.Users[i] = user
	}

	return nil
}
