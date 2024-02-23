package main

import (
	client "brianmargolis/theme-daemon/clients"
	"fmt"
	"os"
	"sync"
)

var CLIENTS = map[string]client.Client{
	"tmux":      client.TMUXClient{},
	"debug":     client.DebugClient{},
	"alacritty": client.AlacrittyClient{},
	"mac":       client.MacClient{},
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "-c" {
		clientNames := args[1:]
		fmt.Println("Starting clients: ", clientNames)
		wg := sync.WaitGroup{}

		for _, clientName := range clientNames {
			wg.Add(1)

			go func(clientName string) {
				defer wg.Done()

				fmt.Println("Starting client: " + clientName)

				client, ok := CLIENTS[clientName]
				if ok {
					client.Start("/tmp/theme-change.sock")
				}
			}(clientName)
		}

		wg.Wait()
	} else {
		fmt.Println("Starting server")
		startServer()
	}
}
