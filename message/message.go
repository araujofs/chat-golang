package message

import (
	c "chat/consts"
	"encoding/binary"
	"errors"
	"fmt"
)

type MessageType uint8

const (
	Send MessageType = iota + 1
	Reply
	CreateGroup
	EnterGroup
	ExitGroup
	Error
)

var msgTypeName = map[MessageType]string{
	Send:        "send",
	Reply:       "reply",
	CreateGroup: "create group",
	EnterGroup:  "enter group",
	ExitGroup:   "exit group",
	Error:       "error",
}

func (mt MessageType) String() string {
	return msgTypeName[mt]
}

func NewMessageType(n uint8) (MessageType, error) {
	switch n {
	case uint8(Send):
		return Send, nil
	case uint8(Reply):
		return Reply, nil
	case uint8(CreateGroup):
		return CreateGroup, nil
	case uint8(EnterGroup):
		return EnterGroup, nil
	case uint8(ExitGroup):
		return ExitGroup, nil
	case uint8(Error):
		return Error, nil
	}

	return 0, fmt.Errorf("\"%d\" não corresponde a um tipo de mensagem válido", n)
}

type Message struct {
	Type    MessageType // uint8
	Group   uint16
	Payload []byte
}

func NewMessage(msgType MessageType, group uint16, payload []byte) (*Message, error) {
	return &Message{
		Type:    msgType,
		Group:   group,
		Payload: payload,
	}, nil
}

func (m Message) Encode() []byte {
	msg := make([]byte, len(m.Payload)+c.HEADER_SIZE)

	msg[0] = uint8(m.Type)

	binary.BigEndian.PutUint16(msg[1:3], m.Group)

	copy(msg[3:], m.Payload)

	return msg
}

func Decode(nbytes int, msg []byte) (*Message, error) {
	if nbytes < c.HEADER_SIZE {
		return nil, errors.New("Mensagem inválida!")
	}

	mt, err := NewMessageType(msg[0])
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:    mt,
		Group:   binary.BigEndian.Uint16(msg[1:3]),
		Payload: msg[3:nbytes],
	}, nil
}

func (m Message) String() string {
	return fmt.Sprintf("{Type: %q, Group: %d, Payload: %q}", m.Type.String(), m.Group, m.Payload)
}
