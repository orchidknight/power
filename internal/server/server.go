package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/power/internal/config"
	"github.com/power/internal/message"
	"github.com/power/pkg/hashcash"
	"github.com/power/pkg/logger"
)

const (
	srv = "server"
)

type WisdomCache interface {
	GetRandom() (string, bool)
}

type Server struct {
	address         string
	listener        net.Listener
	wg              sync.WaitGroup
	shutdown        chan struct{}
	shutdownTimeout time.Duration
	wisdomCache     WisdomCache
	ZeroBits        int
	hashCashTTL     time.Duration
	connTTL         time.Duration
	hashCashPool    *hashcash.HashCashPool

	log logger.Logger
}

func NewServer(cfg *config.ServerConfig, wc WisdomCache, log logger.Logger) *Server {
	return &Server{
		wisdomCache:     wc,
		log:             log,
		address:         cfg.Address,
		connTTL:         time.Minute,
		hashCashTTL:     time.Duration(cfg.HashCashTTL) * time.Millisecond,
		ZeroBits:        cfg.ZeroBits,
		shutdown:        make(chan struct{}),
		shutdownTimeout: time.Duration(cfg.ShutdownTimeout) * time.Millisecond,
		hashCashPool:    hashcash.NewHashCashPool(),
	}
}

func (s *Server) Listen() error {
	var err error
	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

func (s *Server) Shutdown() {
	close(s.shutdown)

	s.listener.Close()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.log.Debug(srv, "Grace shutdown successful")
		return
	case <-time.After(s.shutdownTimeout):
		s.log.Debug(srv, "Timed out waiting for connections to finish.")
		return
	}
}

func (s *Server) acceptConnections() {

	defer s.wg.Done()

	for {
		select {
		case <-s.shutdown:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				continue
			}

			go s.handleConnection(conn)
		}
	}

}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	err := conn.SetReadDeadline(time.Now().Add(s.connTTL))
	if err != nil {
		s.log.Error(srv, "SetReadDeadline err: %v", err)
	}

	clientID := conn.RemoteAddr().String()

	err = s.handleMessages(clientID, conn)
	if err != nil {
		s.log.Error(srv, "handleMessages: %v", err)
	}
}

func (s *Server) handleMessages(clientID string, conn io.ReadWriter) error {
	s.log.Info(srv, "New client %s", clientID)

	for {
		rawMessage, err := bufio.NewReader(conn).ReadString(message.DelimiterEndOfMessage)
		if err != nil {
			s.writeError(ErrInternalError, conn)
			return fmt.Errorf("bufio.NewReader: %v", err)
		}

		msg, err := message.NewMessageFromString(rawMessage)
		if err != nil {
			s.writeError(ErrIncorrectMessageFormat, conn)
			return fmt.Errorf("NewMessageFromString: %v", err)
		}

		switch msg.Type {
		case message.MessageTypeRequestTask:
			err = s.responseTask(clientID, conn)
			if err != nil {
				s.log.Error(srv, "responseTask: %v", err)
			}
		case message.MessageTypeRequestWisdom:
			return s.responseWisdom(clientID, msg.Payload, conn)
		default:
			s.writeError(ErrUnknownMessageType, conn)
			return fmt.Errorf("%s %d", ErrUnknownMessageType.Error(), msg.Type)
		}
	}
}

func (s *Server) responseTask(clientID string, conn io.Writer) error {
	s.log.Info(srv, "Requested new task for client %s", clientID)

	hc, err := hashcash.NewHashCash(s.ZeroBits, clientID)
	if err != nil {
		s.writeError(ErrInternalError, conn)
		return err
	}

	msg := message.Message{
		Type:    message.MessageTypeResponseTask,
		Payload: hc.String(),
	}

	s.hashCashPool.AddHashCash(hc)

	s.writeMessage(msg, conn)
	s.log.Info(srv, "Sent task %s", msg.Payload)

	return nil
}

func (s *Server) responseWisdom(clientID string, payload string, conn io.Writer) error {
	s.log.Info(srv, "Requested wisdom for client %s, task solution %s", clientID, payload)

	hc, err := hashcash.NewHashCashFromString(payload)
	if err != nil {
		s.writeError(ErrHashcashHeaderNotCorrect, conn)
		return fmt.Errorf("NewHashCashFromString: %v", err)
	}

	if !hc.CheckSender(clientID) {
		s.writeError(ErrClientNotFound, conn)
		return nil
	}

	internalHC, ok := s.hashCashPool.GetHashCash(hc.Key())
	if !ok {
		s.writeError(ErrClientNotFound, conn)
	} else {
		s.hashCashPool.RemoveHashCash(internalHC)
	}

	if hc.Expired(s.hashCashTTL) {
		s.writeError(ErrHashcashExpirationExceeded, conn)
		return nil
	}

	isHashCorrect, err := hc.Validate()
	if err != nil {
		s.writeError(ErrInternalError, conn)
		return fmt.Errorf("hc.Validate: %v", err)
	}

	if !isHashCorrect {
		s.writeError(ErrHashcashHeaderNotCorrect, conn)
		return nil
	}

	resource, ok := s.wisdomCache.GetRandom()
	if !ok {
		s.writeError(ErrInternalError, conn)
		return fmt.Errorf("wisdomCache.GetRandom: no records")
	}

	msg := message.Message{
		Type:    message.MessageTypeResponseWisdom,
		Payload: resource,
	}

	s.writeMessage(msg, conn)
	s.log.Info(srv, "Sent to %s message: %s", clientID, msg.Payload)
	return nil
}

func (s *Server) writeMessage(msg message.Message, conn io.Writer) {
	if _, err := conn.Write(msg.Bytes()); err != nil {
		s.log.Error(srv, "writeMessage: %v", err)
	}
}

func (s *Server) writeError(sendErr error, conn io.Writer) {
	_, err := conn.Write(message.Message{Type: message.MessageError, Payload: sendErr.Error()}.Bytes())
	if err != nil {
		s.log.Error(srv, "writeError: %v", err)
	}
}
