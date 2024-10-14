package client

import (
	"brianmargolis/shades/protocol"
	"bufio"
	"net"
	"strings"
)

type Client interface {
	Start(socket string) error
}

type ClientConstructor func() Client

// SocketAsChannel takes a socket path and opens up a channel on top of it.
func SocketAsChannel(socket string) (chan string, chan string, error) {
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, nil, err
	}

	// set up a channel that pulls out new line-delimited messages from the
	// socket
	readChan := make(chan string)
	go func() {
		defer func() {
			conn.Close()
			close(readChan)
		}()

		r := bufio.NewReader(conn)
		for {
			result, err := r.ReadString('\n')
			if err != nil {
				return
			}
			readChan <- result
		}
	}()

	// set up a channel that dumps messages back into the socket (callers need to
	// handle delimitation)
	writeChan := make(chan string)
	go func() {
		for message := range writeChan {
			conn.Write([]byte(message))
		}
	}()

	return readChan, writeChan, nil
}

// SubscribeToSocket is a simple way to build a client if you don't need the
// propose or get functionalities.
func SubscribeToSocket(
	setter func(theme ThemeVariant) error,
) func(string) error {
	config, err := GetConfig()
	if err != nil {
		panic(err)
	}

	return func(socketName string) error {
		read, write, err := SocketAsChannel(socketName)
		if err != nil {
			return err
		}

		write <- string(protocol.Subscribe("mac"))

		for message := range read {
			verb, noun, err := protocol.Parse(message)
			if err != nil {
				return err
			}

			if verb == "set" {
				themeAndVariant := strings.TrimSpace(noun)
				themeVariant, err := config.Themes.GetVariant(themeAndVariant)
				if err != nil {
					return err
				}

				if err = setter(themeVariant); err != nil {
					return err
				}
			}
		}

		return nil
	}
}
