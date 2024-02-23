package main

import (
	client "brianmargolis/theme-daemon/clients"
	"brianmargolis/theme-daemon/protocol"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		panic("No theme name provided")
	}

	theme := args[0]
	if theme != "dark" && theme != "light" {
		panic("Invalid theme name")
	}

	_, write, err := client.SocketAsChannel("/tmp/theme-change.sock")
	if err != nil {
		panic("Could not connect to socket: " + err.Error())
	}

	// hello
	write <- string(protocol.Subscribe("cli"))
	// please set the theme to this
	write <- string(protocol.Propose(theme))
	// goodbye,
	write <- string(protocol.Unsubscribe())
	// forever
	close(write)
}
