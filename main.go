package main

import (
	"brianmargolis/shades/client"
	"brianmargolis/shades/picker"
	"brianmargolis/shades/preview"
	"context"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
	shades t

  Interactively pick the theme:
  shades interactive
  shades i

  Preview a theme:
  shades preview
  shades p`

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

	logger, err := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
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

	ctx := context.Background()
	ctx = client.WithLogger(ctx, logger.Sugar())

	var CLIENTS = map[string]client.Client{
		"alacritty":     client.NewAlacrittyClient(),
		"bat":           client.NewBatClient(),
		"btop":          client.NewBtopClient(),
		"claude":        client.NewClaudeClient(),
		"debug":         client.NewDebugClient(),
		"fzf":           client.NewFZFClient(),
		"ghostty":       client.NewGhosttyClient(),
		"mac":           client.NewMacClient(),
		"mac-wallpaper": client.NewMacWallpaperClient(),
		"tmux":          client.NewTMUXClient(),
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
				if !ok {
					fmt.Printf("no such client %s, ignoring\n", clientName)
				}

				err := client.Start(ctx, socketPath)
				if err != nil {
					panic(err)
				}
			}(clientName)
		}

		wg.Wait()
	case "-l":
		for themeName, theme := range config.Themes {
			for variantName := range theme.Variants {
				fmt.Printf("%s;%s\n", themeName, variantName)
			}
		}
	case "-s":
		NewServer().Start(ctx, socketPath)
	case "dark", "d":
		changer := client.ChangerClient{Theme: config.DefaultDarkTheme}
		changer.Start(ctx, socketPath)
	case "light", "l":
		changer := client.ChangerClient{Theme: config.DefaultLightTheme}
		changer.Start(ctx, socketPath)
	case "toggle", "t":
		toggler := client.TogglerClient{
			DarkTheme:  config.DefaultDarkTheme,
			LightTheme: config.DefaultLightTheme,
		}
		toggler.Start(ctx, socketPath)
	case "set":
		if len(args) < 2 {
			os.Exit(1)
		}
		client.ChangerClient{Theme: args[1]}.Start(ctx, socketPath)
	case "i", "interactive":
		useTmux := false
		onlyLight := false
		onlyDark := false

		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "--tmux":
				useTmux = true
			case "-l", "--light":
				onlyLight = true
			case "-d", "--dark":
				onlyDark = true
			}
		}
		if onlyLight && onlyDark {
			logger.Fatal("cannot specify both only-light and only-dark")
		}

		_, err := picker.NewPicker(logger).Start(picker.PickerOpts{
			SocketPath: socketPath,
			UseTmux:    useTmux,
			OnlyDark:   onlyDark,
			OnlyLight:  onlyLight,
		})
		if err != nil {
			logger.Fatal(err.Error())
		}
	case "p", "preview":
		if len(args) < 1 {
			fmt.Println(args)
			os.Exit(1)
		}
		theme, err := config.Themes.GetVariant(args[1])
		if err != nil {
			logger.Fatal(err.Error())
		}
		swatches, err := preview.NewPreviewer(logger).Preview(theme)
		if err != nil {
			logger.Fatal(err.Error())
		}
		fmt.Println(swatches)
	}
}
