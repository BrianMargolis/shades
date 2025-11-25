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
	logger := zap.S()
	logger.Debug("Starting debug client")
	read, write, err := SocketAsChannel(socket)
	if err != nil {
		return err
	}

	logger.Debug("subscribing to debug messages...")
	write <- string(protocol.Subscribe("debug"))
	logger.Debug("subscribed to debug messages")

	for message := range read {
		logger.With("message", message).Debug("protocol message received")
	}

	logger.Debug("unsubscribing from debug messages...")
	write <- string(protocol.Unsubscribe())
	logger.Debug("unsubscribed from debug messages")
	return nil
}
