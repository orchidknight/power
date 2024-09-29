package message

import (
	"fmt"
	"strconv"
	"strings"
)

type MessageType int8

const (
	MessageError MessageType = iota
	MessageTypeRequestTask
	MessageTypeResponseTask
	MessageTypeRequestWisdom
	MessageTypeResponseWisdom
)

const (
	DelimiterEndOfMessage = '\n'
	DelimiterMessage      = '#'
)

func NewMessageFromString(msg string) (Message, error) {
	m := Message{}
	msg = strings.TrimSpace(msg)

	parts := strings.Split(msg, string(DelimiterMessage))
	if len(parts) < 2 {
		return m, ErrIncorrectMessageFormat
	}

	msgType, err := strconv.Atoi(parts[0])
	if err != nil {
		return m, ErrIncorrectMessageFormat
	}

	switch MessageType(msgType) {
	case MessageError, MessageTypeRequestTask, MessageTypeResponseTask, MessageTypeRequestWisdom, MessageTypeResponseWisdom:
		m.Type = MessageType(msgType)
	default:
		return m, ErrIncorrectMessageFormat
	}

	m.Payload = parts[1]

	return m, nil
}

type Message struct {
	Type    MessageType
	Payload string
}

func NewTaskRequest() Message {
	return Message{Type: MessageTypeRequestTask}
}

func (m Message) String() string {
	return fmt.Sprintf("%d%c%s%c", m.Type, DelimiterMessage, m.Payload, DelimiterEndOfMessage)
}

func (m Message) Bytes() []byte {
	return []byte(m.String())
}
