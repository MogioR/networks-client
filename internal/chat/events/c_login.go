package events

import "bytes"

type LoginEvent struct {
	Login    string
	Password string
}

func (e *LoginEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(2)

	buf.WriteByte(byte(len(e.Login)))
	buf.WriteString(e.Login)

	buf.WriteByte(byte(len(e.Password)))
	buf.WriteString(e.Password)

	return buf
}
