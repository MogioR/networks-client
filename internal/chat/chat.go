package chat

import (
	"bufio"
	"client/internal/config"
	"client/internal/utills"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type Chat struct {
	Chats  map[int64]*ChatModel
	conn   net.Conn
	client fasthttp.Client
	cfg    *config.Config

	mxChats sync.RWMutex
	mxConn  sync.RWMutex

	state       string
	currentChat int64
}

func NewChat(cfg *config.Config) *Chat {
	return &Chat{
		Chats: make(map[int64]*ChatModel, 0),
		client: fasthttp.Client{
			MaxResponseBodySize: 90 * 1024 * 1024,
			ReadTimeout:         time.Duration(30 * time.Second),
			WriteTimeout:        time.Duration(30 * time.Second),
			MaxConnWaitTimeout:  time.Duration(30 * time.Second),
		},
		cfg:         cfg,
		state:       "start",
		currentChat: -1,
	}
}

func (c *Chat) processEvent(eventMsg []byte) error {
	c.mxChats.Lock()
	defer c.mxChats.Unlock()

	var err error
	switch eventMsg[0] {
	case 1:
		var chats []ChatModel
		chats, err = deserializeChats(eventMsg)
		if err != nil {
			return err
		}

		for _, chat := range chats {
			c.Chats[chat.Id] = &chat
		}
		fmt.Print("\nChats recived\n>")
	case 2:
		msg, err := deserializeMessage(eventMsg)
		if err != nil {
			return err
		}
		if _, ok := c.Chats[msg.ChatId]; ok {
			c.Chats[msg.ChatId].Messages = append(c.Chats[msg.ChatId].Messages, msg)
		} else {
			return fmt.Errorf("chat are unknown")
		}
		if c.state != "inChat" || c.currentChat != msg.ChatId {
			fmt.Print("\nSome message recived\n>")
		} else {
			fmt.Printf("%d:%s>%s\n", msg.UserId, msg.PostedAt.Format("2 Jan 2006 15:04"), msg.Message)
		}
	case 3:
		msg, err := deserializeMessageReadAt(eventMsg)
		if err != nil {
			return err
		}
		if _, ok := c.Chats[msg.ChatId]; ok {
			for i := range c.Chats[msg.ChatId].Messages {
				if c.Chats[msg.ChatId].Messages[i].Id == msg.Id {
					c.Chats[msg.ChatId].Messages[i].ReadedAt = msg.ReadedAt
				}
			}
		} else {
			return fmt.Errorf("chat are unknown")
		}
		fmt.Print("\nSome message readed\n>")

	case 4:
		msg, err := deserializeChat(eventMsg)
		if err != nil {
			return err
		}
		if _, ok := c.Chats[msg.Id]; !ok {
			c.Chats[msg.Id] = msg
		} else {
			return fmt.Errorf("chat are known")
		}
		fmt.Print("\nNew chat invited\n>")
	case 5:
		msg, err := deserializeChangeChatMembers(eventMsg)
		if err != nil {
			return err
		}
		if _, ok := c.Chats[msg.Id]; ok {
			c.Chats[msg.Id].Users = msg.Users
		} else {
			return fmt.Errorf("chat are unknown")
		}
		fmt.Print("\nChat changes members\n>")
	default:
		return fmt.Errorf("unknown event type")
	}

	return err
}

func (c *Chat) Update(ctx context.Context) {
	scanner := bufio.NewScanner(os.Stdin)
	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
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
	return

}

func (c *Chat) prosessCommand(input string, ctx context.Context) error {
	command := strings.Split(input, " ")

	switch command[0] {
	case "auth":
		c.mxConn.RLock()

		if c.conn != nil {
			c.mxConn.RUnlock()
			return fmt.Errorf("alrady auth")
		}
		c.mxConn.RUnlock()

		if len(command) != 3 {
			return fmt.Errorf("broken command")
		}

		if c.state != "start" {
			return fmt.Errorf("alrady login")
		}

		return c.auth(command[1], command[2], ctx)
	case "register":
		c.mxConn.RLock()

		if c.conn != nil {
			c.mxConn.RUnlock()
			return fmt.Errorf("alrady auth")
		}
		c.mxConn.RUnlock()

		if len(command) != 3 {
			return fmt.Errorf("broken command")
		}

		if c.state != "start" {
			return fmt.Errorf("alrady auth")
		}

		return c.register(command[1], command[2], ctx)
	case "chatlist":
		if c.state == "start" {
			return fmt.Errorf("do not auth")
		}

		c.mxChats.RLock()
		fmt.Println("№\tName\tUsersCount\tLast Message")
		for i, chat := range c.Chats {
			fmt.Printf("%d\t%s\t%d\t\t%s\n", i, chat.Name, len(chat.Users), chat.LastMessage.Format("2 Jan 2006 15:04"))
		}
		c.mxChats.RUnlock()

	case "opencchat":
		if c.state == "start" {
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

		var chat *ChatModel
		var ok bool
		if chat, ok = c.Chats[charId]; !ok {
			return fmt.Errorf("unknown chat")
		}
		c.state = "inChat"
		c.currentChat = chat.Id

		c.mxChats.RUnlock()
		for _, msg := range chat.Messages {
			fmt.Printf("%d:%s>%s\n", msg.UserId, msg.PostedAt.Format("2 Jan 2006 15:04"), msg.Message)
		}
	case "send":
		c.mxConn.RLock()

		if c.conn != nil {
			c.mxConn.RUnlock()
			return fmt.Errorf("alrady auth")
		}
		c.mxConn.RUnlock()

		if len(command) != 3 {
			return fmt.Errorf("broken command")
		}

		if c.state != "start" {
			return fmt.Errorf("alrady auth")
		}

		return c.register(command[1], command[2], ctx)
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
	}

	conn, err := net.Dial("tcp", c.cfg.Api.Host+c.cfg.Api.TCPPort)
	if err != nil {
		return err
	}

	_, err = conn.Write(token)
	if err != nil {
		return err
	}

	c.mxConn.Lock()
	c.conn = conn
	c.mxConn.Unlock()

	go func() { // Получение ответа от сервера
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			default:
			}
			response := make([]byte, 1024)
			n, err := c.conn.Read(response)
			if err != nil {
				continue
			}
			c.processEvent(response[:n])
		}
	}()

	return nil
}

func (c *Chat) register(login, pass string, ctx context.Context) error {
	url := `http://` + c.cfg.Api.Host + c.cfg.Api.HTTPPort + `/api/v1/register`
	params := []byte(fmt.Sprintf(`{"login": "%s","pass": "%s"}`, login, pass))

	token, err := utills.Post(&c.client, url, params, map[string]string{})
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", c.cfg.Api.Host+c.cfg.Api.TCPPort)
	if err != nil {
		return err
	}

	_, err = conn.Write(token)
	if err != nil {
		return err
	}

	c.mxConn.Lock()
	c.conn = conn
	c.mxConn.Unlock()

	go func() { // Получение ответа от сервера
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			default:
			}
			response := make([]byte, 1024)
			n, err := c.conn.Read(response)
			if err != nil {
				continue
			}
			c.processEvent(response[:n])
		}
	}()

	return nil
}
