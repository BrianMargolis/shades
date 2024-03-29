package main

import (
	"brianmargolis/shades/client"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

var CLIENTS = map[string]client.Client{
	"alacritty":     client.AlacrittyClient{},
	"bat":           client.BatClient{},
	"btop":          client.BtopClient{},
	"debug":         client.DebugClient{},
	"fzf":           client.FZFClient{},
	"mac":           client.MacClient{},
	"mac-wallpaper": client.MacWallpaperClient{},
	"tmux":          client.TMUXClient{},
}

const usage = `shades usage: 
	Start server mode:
	shades -s

	Start clients:
	shades -c client1 client2

	List available clients;
	shades -l

	Change the theme:
	shades dark 
	shades d
	shades light
	shades l

	Toggle the theme:
	shades toggle
	shades t`

// TODO make this configurable
const socketPath = "/tmp/theme-change.sock"

func main() {
	args := os.Args[1:]
	mode := "toggle"
	if len(args) > 0 {
		mode = args[0]
	}

	config, err := getConfig()
	if err != nil {
		panic(err)
	}

	switch mode {
	case "-h", "--help":
		fmt.Println(usage)
	case "-c":
		wg := sync.WaitGroup{}

		clientNames := args[1:]
		for _, clientName := range clientNames {
			wg.Add(1)

			go func(clientName string) {
				defer wg.Done()

				client, ok := CLIENTS[clientName]
				if ok {
					err := client.Start(socketPath, config.Client[clientName])
					panic(err)
				} else {
					fmt.Printf("no such client %s, ignoring\n", clientName)
				}
			}(clientName)
		}

		wg.Wait()
	case "-l":
		fmt.Println("available clients:")
		for client := range CLIENTS {
			fmt.Printf("\t%s", client)
		}
	case "-s":
		startServer(socketPath)
	case "dark", "d", "light", "l":
		// expando
		theme := args[0]
		if theme == "d" {
			theme = "dark"
		} else if theme == "l" {
			theme = "light"
		}

		changer := client.ChangerClient{Theme: theme}
		changer.Start(socketPath)
	case "toggle", "t":
		toggler := client.TogglerClient{}
		toggler.Start(socketPath)
	}
}

type configModel struct {
	SocketPath string                       `yaml:"socket-path"`
	Client     map[string]map[string]string `yaml:"client"`
}

func getConfig() (configModel, error) {
	f, err := os.Open(getConfigPath())
	if err != nil {
		return configModel{}, err
	}

	defer f.Close()

	config := configModel{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		return configModel{}, err
	}

	return config, nil
}

func getConfigPath() string {
	// if SHADES_CONFIG is defined, use that
	envValue := os.Getenv("SHADES_CONFIG")
	if envValue != "" {
		return envValue
	}

	return os.Getenv("HOME") + "/.config/shades/shades.yaml"
}
