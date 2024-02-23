package client

import (
	"brianmargolis/theme-daemon/protocol"
	"fmt"
)

type DebugClient struct{}

func (d DebugClient) Start(socket string) error {
	fmt.Println("Starting debug client")
	read, write, err := SocketAsChannel(socket)
	if err != nil {
		return err
	}

	write <- string(protocol.Subscribe("debug"))

	for result := range read {
		fmt.Printf("DEBUG: %s\n", result)
	}

	write <- string(protocol.Unsubscribe())
	return nil
}
