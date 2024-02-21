package events

import "bytes"

type RegisterEvent struct {
	Login    string
	Password string
}

func (e *RegisterEvent) Serialize() *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteByte(1)

	buf.WriteByte(byte(len(e.Login)))
	buf.WriteString(e.Login)

	buf.WriteByte(byte(len(e.Password)))
	buf.WriteString(e.Password)

	return buf
}
