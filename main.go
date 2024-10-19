package main

import (
	"brianmargolis/shades/client"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
const verbose = false

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

	level := zap.InfoLevel
	if verbose {
		level = zap.DebugLevel
	}
	logger, err := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"/var/log/shades/shades.log"},
		ErrorOutputPaths: []string{"/var/log/shades/shades.error.log"},
	}.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(err)
	}

	logger.Info("shades started", zap.Strings("args", args))

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
		for themeName, theme := range config.Themes {
			for variantName, _ := range theme.Variants {
				fmt.Printf("%s;%s\n", themeName, variantName)
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
	case "set":
		if len(args) < 2 {
			os.Exit(1)
		}
		client.ChangerClient{Theme: args[1]}.Start(socketPath)
	}
}
