package main

import (
	"brianmargolis/shades/client"
	"fmt"
	"os"
	"reflect"
	"sync"
)

var CLIENTS = map[string]client.ClientConstructor{
	"alacritty":     client.NewAlacrittyClient,
	"bat":           client.NewBatClient,
	"btop":          client.NewBtopClient,
	"debug":         client.NewDebugClient,
	"fzf":           client.NewFZFClient,
	"mac":           client.NewMacClient,
	"mac-wallpaper": client.NewMacWallpaperClient,
	"tmux":          client.NewTMUXClient,
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

	config, err := client.GetConfig()
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

				clientConstructor, ok := CLIENTS[clientName]
				if !ok {
					fmt.Printf("no such client %s, ignoring\n", clientName)
				}

				err := clientConstructor().Start(socketPath)
				if err != nil {
					panic(err)
				}
			}(clientName)
		}

		wg.Wait()
	case "-l":
		themesVal := reflect.ValueOf(config.Themes)
		nThemes := themesVal.NumField()
		for i := 0; i < nThemes; i++ {
			theme := themesVal.Field(i)
			themeName := theme.FieldByName("Name")
			variants := theme.FieldByName("Variants").MapKeys()
			for _, variant := range variants {
				fmt.Printf("%s;%s\n", themeName.String(), variant.String())
			}
		}
	case "-s":
		NewServer().Start(socketPath)
	case "dark", "d":
		changer := client.ChangerClient{Theme: config.DefaultDarkTheme}
		changer.Start(socketPath)
	case "light", "l":
		changer := client.ChangerClient{Theme: config.DefaultLightTheme}
		changer.Start(socketPath)
	case "toggle", "t":
		toggler := client.TogglerClient{
			DarkTheme:  config.DefaultDarkTheme,
			LightTheme: config.DefaultLightTheme,
		}
		toggler.Start(socketPath)
	}
}
