package events

import "bytes"

type CreateChatEvent struct {
	ChatName    string
	InviteUsers []string
}

func (e *CreateChatEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(4)

	buf.WriteByte(byte(len(e.ChatName)))
	buf.WriteString(e.ChatName)

	buf.WriteByte(byte(len(e.InviteUsers)))
	for _, user := range e.InviteUsers {
		buf.WriteByte(byte(len(user)))
		buf.WriteString(user)
	}

	return buf
}
