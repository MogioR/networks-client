package events

import "bytes"

type CreateChatEvent struct {
	chatName    string
	inviteUsers []string
}

func (e *CreateChatEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(4)

	buf.WriteByte(byte(len(e.chatName)))
	buf.WriteString(e.chatName)

	buf.WriteByte(byte(len(e.inviteUsers)))
	for _, user := range e.inviteUsers {
		buf.WriteByte(byte(len(user)))
		buf.WriteString(user)
	}

	return buf
}
