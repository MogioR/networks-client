package chat

import (
	"bufio"
	"bytes"
	"client/internal/chat/events"
	"client/internal/config"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Chat struct {
	Users map[int64]string
	Chats map[int64]*events.Chat
	conn  net.Conn

	cfg *config.Config

	mxChats sync.RWMutex
	mxConn  sync.RWMutex

	isAuth      bool
	currentChat int64
}

func NewChat(cfg *config.Config) *Chat {
	return &Chat{
		Chats: make(map[int64]*events.Chat, 0),
		Users: make(map[int64]string, 0),

		cfg:         cfg,
		isAuth:      false,
		currentChat: -1,
	}
}

func (c *Chat) processEvent(eventMsg *bytes.Buffer) error {
	c.mxChats.Lock()
	defer c.mxChats.Unlock()

	var command int8
	err := binary.Read(eventMsg, binary.BigEndian, &command)
	if err != nil {
		return err
	}
	switch command {
	case -1:
		event := events.SystemMessageEvent{}
		err := event.Deserialize(eventMsg)
		if err != nil {
			return err
		}
		fmt.Printf("Info: System message: %s\n> ", event.Message)
	case -2:
		event := events.ChatEvent{}
		err := event.Deserialize(eventMsg)
		if err != nil {
			return err
		}

		for _, chat := range event.Chats {
			c.Chats[chat.ChatId] = chat
			for _, user := range chat.Users {
				c.Users[user.Id] = user.Login
			}
		}
		fmt.Printf("Info: Chats info updated\n> ")
	case -3:
		fmt.Printf("Chat users updated\n> ")
	case -4:
		event := events.NewMessageEvent{}
		err := event.Deserialize(eventMsg)
		if err != nil {
			return err
		}

		if _, ok := c.Chats[event.ChatId]; !ok {
			return fmt.Errorf("unknown chat")
		}

		var isFileByte byte
		if event.MessageType {
			isFileByte = 1
		}

		c.Chats[event.ChatId].Messages = append(c.Chats[event.ChatId].Messages,
			events.Message{
				Id:          event.MessageId,
				SenderId:    event.UserId,
				MessageType: isFileByte,
				Message:     event.Message,
				SendTime:    event.SendTime,
				ReadTime:    event.ReadTime,
			})

		if c.currentChat == event.ChatId {
			fmt.Printf(
				"%s:%s> %s\n> ",
				c.Users[event.UserId],
				event.SendTime.Format("2006-01-02"),
				event.Message,
			)
		} else {
			fmt.Printf("New message in chat\n> ")
		}
	case -5:
		fmt.Printf("Some one read message in chat\n> ")
	case -6:
		event := events.SendFileToChatEvent{}
		err := event.Deserialize(eventMsg)
		if err != nil {
			return err
		}
		file := make([]byte, event.FileSize)

		for i := 0; i < len(file); i += 1000 {
			top := i + 1000
			if top > len(file) {
				top = len(file)
			}
			c.mxConn.RLock()
			_, err = c.conn.Read(file[i:top])
			c.mxConn.RUnlock()
			if err != nil {
				return err
			}
		}

		err = SaveBytesToFile(fmt.Sprintf("downloads/%s", event.FileName), file)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown event type")
	}
	return nil
}

func (c *Chat) ClientHeandler(ctx context.Context, i int) {
	err := c.prosessCommand(fmt.Sprintf("r user3_%d admin", i), ctx)
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			fmt.Print("> ")
			scanner.Scan()
			input := scanner.Text()

			err := c.prosessCommand(input, ctx)
			if err != nil {
				fmt.Println(err)
			}

		}
	}()

	<-ctx.Done()
}

func (c *Chat) serverHeandler(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		response := make([]byte, 1024)
		n, err := c.conn.Read(response)
		if err != nil {
			log.Fatal(err)
		}

		buf := bytes.NewBuffer(response[:n])
		c.processEvent(buf)
	}
}

func (c *Chat) prosessCommand(input string, ctx context.Context) error {
	command := strings.Split(input, " ")

	switch command[0] {
	case "login", "l":
		c.mxConn.RLock()
		if c.conn != nil {
			c.mxConn.RUnlock()
			return fmt.Errorf("alrady auth")
		}
		c.mxConn.RUnlock()
		if len(command) != 3 {
			return fmt.Errorf("broken command")
		}

		return c.auth(ctx, command[1], command[2], false)
	case "register", "r":
		c.mxConn.RLock()
		if c.conn != nil {
			c.mxConn.RUnlock()
			return fmt.Errorf("alrady auth")
		}
		c.mxConn.RUnlock()
		if len(command) != 3 {
			return fmt.Errorf("broken command")
		}
		return c.auth(ctx, command[1], command[2], true)

	case "chatlist", "cl":
		if !c.isAuth {
			return fmt.Errorf("do not auth")
		}

		c.mxChats.RLock()
		c.currentChat = -1
		fmt.Println("№\tName\tUsersCount\tLast Message")
		for i, chat := range c.Chats {
			lastMessage := ""
			if len(chat.Messages) > 0 {
				lastMessage = chat.Messages[len(chat.Messages)-1].SendTime.Format("2 Jan 2006 15:04")
			}

			fmt.Printf("%d\t%s\t%d\t\t%s\n", i, chat.ChatName, len(chat.Users), lastMessage)
		}
		c.mxChats.RUnlock()
		return nil

	case "opencchat", "oc":
		if !c.isAuth {
			return fmt.Errorf("do not auth")
		}
		if len(command) != 2 {
			return fmt.Errorf("broken command")
		}
		var err error
		var charId int64
		if charId, err = strconv.ParseInt(command[1], 10, 64); err != nil {
			return err
		}
		c.mxChats.RLock()

		var chat *events.Chat
		var ok bool
		if chat, ok = c.Chats[charId]; !ok {
			return fmt.Errorf("unknown chat")
		}
		c.currentChat = chat.ChatId

		for _, msg := range chat.Messages {
			if msg.MessageType == 0 {
				fmt.Printf("%s:%s>%s\n", c.Users[msg.SenderId], msg.SendTime.Format("2 Jan 2006 15:04"), msg.Message)
			} else {
				fmt.Printf("%s:%s>%s [%d]\n", c.Users[msg.SenderId], msg.SendTime.Format("2 Jan 2006 15:04"), msg.Message, msg.Id)
			}

		}
		c.mxChats.RUnlock()
		return nil

	case "sendtext", "st":
		if !c.isAuth {
			return fmt.Errorf("not auth")
		}
		if c.currentChat == -1 {
			return fmt.Errorf("not in chat")
		}

		return c.sendMessage(strings.Join(command[1:], ""))
	default:
		return fmt.Errorf("unknown command")
	}
	return nil
}

func (c *Chat) auth(login, pass string, ctx context.Context) error {
	url := `http://` + c.cfg.Api.Host + c.cfg.Api.HTTPPort + `/api/v1/auth`
	params := []byte(fmt.Sprintf(`{"login": "%s","pass": "%s"}`, login, pass))

	token, err := utills.Post(&c.client, url, params, map[string]string{})
	if err != nil {
		return err
	case "sendfile", "sf":
		if !c.isAuth {
			return fmt.Errorf("not auth")
		}
		if c.currentChat == -1 {
			return fmt.Errorf("not in chat")
		}

		file, err := ReadFileToBytes(command[1])
		if err != nil {
			return err
		}

		event := events.SendFileInitEvent{
			ChatId:   c.currentChat,
			FileName: filepath.Base(command[1]),
			FileSize: int32(len(file)),
		}

		c.mxConn.RLock()
		_, err = c.conn.Write(event.Serialize().Bytes())
		c.mxConn.RUnlock()
		if err != nil {
			return err
		}

		for i := 0; i < len(file); i += 1000 {
			top := i + 1000
			if top > len(file) {
				top = len(file)
			}
			c.mxConn.RLock()
			_, err = c.conn.Write(file[i:top])
			c.mxConn.RUnlock()
			if err != nil {
				return err
			}
		}
		return err

	case "downloadfile", "df":
		if !c.isAuth {
			return fmt.Errorf("not auth")
		}
		if c.currentChat == -1 {
			return fmt.Errorf("not in chat")
		}
		var err error
		var messageId int64
		if messageId, err = strconv.ParseInt(command[1], 10, 64); err != nil {
			return err
		}

		event := events.GetFileFromChatEvent{
			MessageId: messageId,
			ChatId:    c.currentChat,
		}
		c.mxConn.RLock()
		_, err = c.conn.Write(event.Serialize().Bytes())
		c.mxConn.RUnlock()

		return err

	default:
		return fmt.Errorf("unknown command")
	}
}

func (c *Chat) auth(ctx context.Context, login, password string, isRegister bool) error {
	conn, err := net.Dial("tcp", c.cfg.Api.Host+c.cfg.Api.TCPPort)
	if err != nil {
		return err
	}

	var event events.ClientEvent
	if isRegister {
		event = &events.RegisterEvent{
			Login:    login,
			Password: password,
		}
	} else {
		event = &events.LoginEvent{
			Login:    login,
			Password: password,
		}
	}

	msg := event.Serialize().Bytes()
	_, err = conn.Write(msg)
	if err != nil {
		return err
	}

	c.mxConn.Lock()
	c.conn = conn
	c.mxConn.Unlock()

	c.isAuth = true

	go c.serverHeandler(ctx)

	return nil
}

func ReadFileToBytes(filePath string) ([]byte, error) {
	// Читаем содержимое файла в байтовый массив
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func SaveBytesToFile(filename string, data []byte) error {
	// Используем функцию ioutil.WriteFile для записи данных в файл
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
