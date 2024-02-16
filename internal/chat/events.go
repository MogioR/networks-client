package chat

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
)

func deserializeChats(data []byte) ([]ChatModel, error) {
	var chats []ChatModel

	buf := bytes.NewBuffer(data)

	// Чтение номера команды
	command, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	// Проверка номера команды
	if command != 1 {
		return nil, errors.New("unexpected command number")
	}

	// Чтение количества чатов
	numChats, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(numChats); i++ {
		var chat ChatModel

		// Чтение чат айди
		if err := binary.Read(buf, binary.BigEndian, &chat.Id); err != nil {
			return nil, err
		}

		// Чтение размера названия чата и самого названия
		nameLen, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		nameBytes := make([]byte, nameLen)
		if _, err := buf.Read(nameBytes); err != nil {
			return nil, err
		}
		chat.Name = string(nameBytes)

		// Чтение количества пользователей в чате
		numUsers, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		for j := 0; j < int(numUsers); j++ {
			var user UserModel

			// Чтение пользователя айди
			if err := binary.Read(buf, binary.BigEndian, &user.Id); err != nil {
				return nil, err
			}

			// Чтение размера имени пользователя и самого имени
			loginLen, err := buf.ReadByte()
			if err != nil {
				return nil, err
			}
			loginBytes := make([]byte, loginLen)
			if _, err := buf.Read(loginBytes); err != nil {
				return nil, err
			}
			user.Login = string(loginBytes)

			// Добавление пользователя к чату
			chat.Users = append(chat.Users, user)
		}

		// Чтение количества сообщений в чате
		var numMessages int8
		if err := binary.Read(buf, binary.BigEndian, &numMessages); err != nil {
			return nil, err
		}

		// Десериализация сообщений
		for k := 0; k < int(numMessages); k++ {
			var msg MessageModel

			// Чтение сообщения айди, пользователя айди и флага наличия файла
			if err := binary.Read(buf, binary.BigEndian, &msg.Id); err != nil {
				return nil, err
			}
			if err := binary.Read(buf, binary.BigEndian, &msg.UserId); err != nil {
				return nil, err
			}
			isFile, err := buf.ReadByte()
			if err != nil {
				return nil, err
			}
			msg.IsFile = isFile == 1

			// Чтение длины сообщения и самого сообщения
			var msgLen int16
			if err := binary.Read(buf, binary.BigEndian, &msgLen); err != nil {
				return nil, err
			}

			b := make([]byte, msgLen)
			if _, err := buf.Read(b); err != nil {
				return nil, err
			}
			msg.Message = string(b)

			// Чтение даты публикации и прочтения сообщения
			var postedAtUnix, readedAtUnix int64
			if err := binary.Read(buf, binary.BigEndian, &postedAtUnix); err != nil {
				return nil, err
			}
			if err := binary.Read(buf, binary.BigEndian, &readedAtUnix); err != nil {
				return nil, err
			}
			msg.PostedAt = time.Unix(postedAtUnix, 0)
			msg.ReadedAt = time.Unix(readedAtUnix, 0)

			// Добавление сообщения к чату
			chat.Messages = append(chat.Messages, msg)
		}

		var lastMessage int64
		if err := binary.Read(buf, binary.BigEndian, &lastMessage); err != nil {
			return nil, err
		}
		chat.LastMessage = time.Unix(lastMessage, 0)

		// Добавление чата и сообщений к спискам
		chats = append(chats, chat)
	}

	return chats, nil
}

func deserializeMessage(data []byte) (MessageModel, error) {
	var message MessageModel

	buf := bytes.NewReader(data)

	// Пропускаем тип события
	if _, err := buf.ReadByte(); err != nil {
		return message, err
	}

	// Чтение Id чата
	if err := binary.Read(buf, binary.BigEndian, &message.ChatId); err != nil {
		return message, err
	}

	// Чтение Id сообщения
	if err := binary.Read(buf, binary.BigEndian, &message.Id); err != nil {
		return message, err
	}

	// Чтение Id отправителя
	if err := binary.Read(buf, binary.BigEndian, &message.UserId); err != nil {
		return message, err
	}

	// Чтение типа сообщения
	messageTypeByte, err := buf.ReadByte()
	if err != nil {
		return message, err
	}
	message.IsFile = messageTypeByte == 1

	// Чтение длины сообщения
	var messageLength int16
	if err := binary.Read(buf, binary.BigEndian, &messageLength); err != nil {
		return message, err
	}

	// Чтение сообщения
	messageBytes := make([]byte, messageLength)
	if _, err := buf.Read(messageBytes); err != nil {
		return message, err
	}
	message.Message = string(messageBytes)

	// Чтение времени отправки
	var postedAtUnix int64
	if err := binary.Read(buf, binary.BigEndian, &postedAtUnix); err != nil {
		return message, err
	}
	message.PostedAt = time.Unix(postedAtUnix, 0)

	// Чтение времени прочтения
	var readedAtUnix int64
	if err := binary.Read(buf, binary.BigEndian, &readedAtUnix); err != nil {
		return message, err
	}
	message.ReadedAt = time.Unix(readedAtUnix, 0)

	return message, nil
}

func deserializeMessageReadAt(data []byte) (MessageModel, error) {
	var message MessageModel

	buf := bytes.NewReader(data)

	// Пропускаем тип события
	if _, err := buf.ReadByte(); err != nil {
		return message, err
	}

	// Чтение Id чата
	if err := binary.Read(buf, binary.BigEndian, &message.ChatId); err != nil {
		return message, err
	}

	// Чтение Id сообщения
	if err := binary.Read(buf, binary.BigEndian, &message.Id); err != nil {
		return message, err
	}

	// Чтение времени первого прочтения
	var firstReadUnix int64
	if err := binary.Read(buf, binary.BigEndian, &firstReadUnix); err != nil {
		return message, err
	}
	message.ReadedAt = time.Unix(firstReadUnix, 0)

	return message, nil
}

func deserializeChangeChatMembers(data []byte) (ChatModel, error) {
	var chat ChatModel

	buf := bytes.NewReader(data)

	// Пропускаем тип события
	if _, err := buf.ReadByte(); err != nil {
		return chat, err
	}

	// Чтение Id чата
	if err := binary.Read(buf, binary.BigEndian, &chat.Id); err != nil {
		return chat, err
	}

	// Чтение количества пользователей в чате
	numUsers, err := buf.ReadByte()
	if err != nil {
		return chat, err
	}

	// Пользователи в чате
	for i := 0; i < int(numUsers); i++ {
		var user UserModel

		// Чтение Id пользователя
		if err := binary.Read(buf, binary.BigEndian, &user.Id); err != nil {
			return chat, err
		}

		// Чтение длины логина пользователя
		loginLen, err := buf.ReadByte()
		if err != nil {
			return chat, err
		}

		// Чтение логина пользователя
		loginBytes := make([]byte, loginLen)
		if _, err := buf.Read(loginBytes); err != nil {
			return chat, err
		}
		user.Login = string(loginBytes)

		// Добавляем пользователя к чату
		chat.Users = append(chat.Users, user)
	}

	var lastMessage int64
	if err := binary.Read(buf, binary.BigEndian, &lastMessage); err != nil {
		return chat, err
	}
	chat.LastMessage = time.Unix(lastMessage, 0)

	return chat, nil
}

func SerializeChatCreate(chat ChatModel) ([]byte, error) {
	var buf bytes.Buffer

	// Тип события
	buf.WriteByte(7)

	// Длина имени чата
	buf.WriteByte(byte(len(chat.Name)))

	// Название чата
	buf.WriteString(chat.Name)

	// Количество пользователей в чате
	buf.WriteByte(byte(len(chat.Users)))

	// Пользователи в чате
	for _, user := range chat.Users {
		// Длина логина пользователя
		buf.WriteByte(byte(len(user.Login)))

		// Логин пользователя
		buf.WriteString(user.Login)
	}

	return buf.Bytes(), nil
}

func serializeTextMessage(message MessageModel) ([]byte, error) {
	var buf bytes.Buffer

	// Тип события
	buf.WriteByte(8)

	// Id чата
	if err := binary.Write(&buf, binary.BigEndian, message.ChatId); err != nil {
		return nil, err
	}

	// Длина сообщения
	messageLength := int16(len(message.Message))
	if err := binary.Write(&buf, binary.BigEndian, messageLength); err != nil {
		return nil, err
	}

	// Сообщение
	buf.WriteString(message.Message)

	return buf.Bytes(), nil
}

// func serializeFile(file FileModel) ([]byte, error) {
// 	var buf bytes.Buffer

// 	// Тип события
// 	buf.WriteByte(9)

// 	// Id чата
// 	if err := binary.Write(&buf, binary.BigEndian, file.ChatId); err != nil {
// 		return nil, err
// 	}

// 	// Длина названия файла
// 	fileNameLength := uint16(len(file.FileName))
// 	if err := binary.Write(&buf, binary.BigEndian, fileNameLength); err != nil {
// 		return nil, err
// 	}

// 	// Название файла
// 	buf.WriteString(file.FileName)

// 	// Размер файла
// 	if err := binary.Write(&buf, binary.BigEndian, file.FileSize); err != nil {
// 		return nil, err
// 	}

// 	// Файл
// 	buf.Write(file.FileContents)

// 	return buf.Bytes(), nil
// }

func serializeFileMessage(message MessageModel) ([]byte, int32, error) {
	var buf bytes.Buffer

	// Тип события
	buf.WriteByte(9)

	// Id чата
	if err := binary.Write(&buf, binary.BigEndian, message.ChatId); err != nil {
		return nil, 0, err
	}

	// Id сообщения
	if err := binary.Write(&buf, binary.BigEndian, message.Id); err != nil {
		return nil, 0, err
	}

	// Длина названия файла
	fileNameLength := uint16(len(message.Message))
	if err := binary.Write(&buf, binary.BigEndian, fileNameLength); err != nil {
		return nil, 0, err
	}

	// Название файла
	buf.WriteString(message.Message)

	// Размер файла
	var fileSize int32
	if err := binary.Write(&buf, binary.BigEndian, fileSize); err != nil {
		return nil, 0, err
	}

	return buf.Bytes(), fileSize, nil
}

func serializeEventLeaveChat(chatID int64) ([]byte, error) {
	// Создание буфера для сериализации
	buffer := new(bytes.Buffer)

	// Запись типа события
	if err := binary.Write(buffer, binary.BigEndian, byte(11)); err != nil {
		return nil, err
	}

	// Запись ID чата
	if err := binary.Write(buffer, binary.BigEndian, chatID); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func serializeEventInviteChat(chatID int64, userLogin string) ([]byte, error) {
	// Создание буфера для сериализации
	buffer := new(bytes.Buffer)

	// Запись типа события
	if err := binary.Write(buffer, binary.BigEndian, byte(12)); err != nil {
		return nil, err
	}

	// Запись ID чата
	if err := binary.Write(buffer, binary.BigEndian, chatID); err != nil {
		return nil, err
	}

	// Запись длины логина пользователя
	userLoginLength := byte(len(userLogin))
	if err := binary.Write(buffer, binary.BigEndian, userLoginLength); err != nil {
		return nil, err
	}

	// Запись логина пользователя
	if _, err := buffer.WriteString(userLogin); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func deserializeChat(data []byte) (*ChatModel, error) {

	buf := bytes.NewBuffer(data)

	// Чтение номера команды
	command, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	// Проверка номера команды
	if command != 4 {
		return nil, errors.New("unexpected command number")
	}

	var chat ChatModel

	// Чтение чат айди
	if err := binary.Read(buf, binary.BigEndian, &chat.Id); err != nil {
		return nil, err
	}

	// Чтение размера названия чата и самого названия
	nameLen, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	nameBytes := make([]byte, nameLen)
	if _, err := buf.Read(nameBytes); err != nil {
		return nil, err
	}
	chat.Name = string(nameBytes)

	// Чтение количества пользователей в чате
	numUsers, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	for j := 0; j < int(numUsers); j++ {
		var user UserModel

		// Чтение пользователя айди
		if err := binary.Read(buf, binary.BigEndian, &user.Id); err != nil {
			return nil, err
		}

		// Чтение размера имени пользователя и самого имени
		loginLen, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		loginBytes := make([]byte, loginLen)
		if _, err := buf.Read(loginBytes); err != nil {
			return nil, err
		}
		user.Login = string(loginBytes)

		// Добавление пользователя к чату
		chat.Users = append(chat.Users, user)
	}

	// Чтение количества сообщений в чате
	var numMessages int8
	if err := binary.Read(buf, binary.BigEndian, &numMessages); err != nil {
		return nil, err
	}

	// Десериализация сообщений
	for k := 0; k < int(numMessages); k++ {
		var msg MessageModel

		// Чтение сообщения айди, пользователя айди и флага наличия файла
		if err := binary.Read(buf, binary.BigEndian, &msg.Id); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &msg.UserId); err != nil {
			return nil, err
		}
		isFile, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		msg.IsFile = isFile == 1

		// Чтение длины сообщения и самого сообщения
		var msgLen int16
		if err := binary.Read(buf, binary.BigEndian, &msgLen); err != nil {
			return nil, err
		}

		b := make([]byte, msgLen)
		if _, err := buf.Read(b); err != nil {
			return nil, err
		}
		msg.Message = string(b)

		// Чтение даты публикации и прочтения сообщения
		var postedAtUnix, readedAtUnix int64
		if err := binary.Read(buf, binary.BigEndian, &postedAtUnix); err != nil {
			return nil, err
		}
		if err := binary.Read(buf, binary.BigEndian, &readedAtUnix); err != nil {
			return nil, err
		}
		msg.PostedAt = time.Unix(postedAtUnix, 0)
		msg.ReadedAt = time.Unix(readedAtUnix, 0)

		// Добавление сообщения к чату
		chat.Messages = append(chat.Messages, msg)
	}

	return &chat, nil
}
