package server

import (
	"errors"
	"github.com/power/internal/message"
)

var (
	ErrIncorrectMessageFormat     = errors.New("incorrect message format")
	ErrUnknownMessageType         = errors.New("unknown message type")
	ErrClientNotFound             = errors.New("client not found in task requester list")
	ErrHashcashHeaderNotCorrect   = errors.New("hashcash header not correct")
	ErrHashcashExpirationExceeded = errors.New("hashcash expiration exceeded")
	ErrInternalError              = errors.New("internal error")
)

func errorMessage(err error) message.Message {
	return message.Message{
		Type:    message.MessageError,
		Payload: err.Error(),
	}
}
