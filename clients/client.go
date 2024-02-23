package client

import (
	"bufio"
	"net"
)

type Client interface {
	Start(socket string) error
}

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
