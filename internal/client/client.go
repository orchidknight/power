package client

import (
	"bufio"
	"fmt"
	"net"

	"github.com/power/internal/message"
	"github.com/power/pkg/hashcash"
	"github.com/power/pkg/logger"
)

const cli = "client"

type Client struct {
	serverAddress string
	localAddress  string
	triesLimit    int
	conn          net.Conn

	log logger.Logger
}

func NewClient(serverAddress string, maxTries int, log logger.Logger) *Client {
	return &Client{
		serverAddress: serverAddress,
		triesLimit:    maxTries,
		log:           log,
	}
}

func (c *Client) Connect() error {
	var err error

	c.conn, err = net.Dial("tcp", c.serverAddress)
	if err != nil {
		return fmt.Errorf("net.Dial: %v", err)
	}

	return nil
}

func (c *Client) Close() {
	err := c.conn.Close()
	if err != nil {
		c.log.Error(cli, "conn.Close: %v", err)
	}
}

func (c *Client) GetWisdom() (string, error) {
	var err error

	taskResponse, err := c.requestTask()
	if err != nil {
		return "", fmt.Errorf("requestTask: %v", err)
	}

	c.log.Info(cli, "Task received: %s", taskResponse)

	hc, err := hashcash.NewHashCashFromString(taskResponse)
	if err != nil {
		return "", fmt.Errorf("NewHashCashFromString: %v", err)
	}

	if err = hc.Calculate(c.triesLimit); err != nil {
		return "", fmt.Errorf("calculate: %v", err)
	}

	c.log.Info(cli, "Task was solved in %d tries", hc.Counter())

	requestWisdomMessage := message.Message{
		Type:    message.MessageTypeRequestWisdom,
		Payload: hc.String(),
	}

	wisdom, err := c.requestWisdom(requestWisdomMessage)
	if err != nil {
		return "", fmt.Errorf("request wisdom: %v", err)
	}

	return wisdom, err
}

func (c *Client) requestTask() (string, error) {
	response, err := c.request(message.NewTaskRequest())
	if err != nil {
		return "", fmt.Errorf("request: %v", err)
	}

	return response, nil
}

func (c *Client) requestWisdom(msg message.Message) (string, error) {
	response, err := c.request(msg)
	if err != nil {
		return "", fmt.Errorf("request: %v", err)
	}

	return response, nil
}

func (c *Client) request(msg message.Message) (string, error) {
	err := c.send(msg)
	if err != nil {
		return "", fmt.Errorf("send: %v", err)
	}

	rawResponse, err := bufio.NewReader(c.conn).ReadString(message.DelimiterEndOfMessage)
	if err != nil {
		return "", fmt.Errorf("ReadString: %v", err)
	}

	response, err := message.NewMessageFromString(rawResponse)
	if err != nil {
		return "", fmt.Errorf("NewMessageFromString: %v", err)
	}

	err = c.validateResponse(msg.Type, response)
	if err != nil {
		return "", fmt.Errorf("validateResponse: %v", err)
	}

	return response.Payload, nil
}

func (c *Client) send(msg message.Message) error {
	_, err := c.conn.Write(msg.Bytes())
	if err != nil {
		return fmt.Errorf("write: %v", err)
	}

	return nil
}

func (c *Client) validateResponse(requestType message.MessageType, response message.Message) error {
	switch response.Type {
	case message.MessageError:
		return fmt.Errorf(response.Payload)
	case message.MessageTypeResponseTask:
		if requestType != message.MessageTypeRequestTask {
			return ErrIncorrectResponseMessageType
		}
	case message.MessageTypeResponseWisdom:
		if requestType != message.MessageTypeRequestWisdom {
			return ErrIncorrectResponseMessageType
		}
	default:
		return ErrIncorrectResponseMessageType
	}

	return nil
}
