package client

import (
	"brianmargolis/shades/protocol"

	"go.uber.org/zap"
)

type DebugClient struct{}

func NewDebugClient() Client {
	return DebugClient{}
}

func (d DebugClient) Start(socket string) error {
	logger := zap.S().With("client", "debug")
	read, write, err := SocketAsChannel(socket)
	if err != nil {
		return err
	}

	write <- string(protocol.Subscribe("debug"))
	logger.Debug("subscribed, waiting for messages")

	for message := range read {
		logger.Debugw("received message", "message", message)
	}

	logger.Debug("server closed connection")
	return nil
}
