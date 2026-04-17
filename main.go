package main

import (
	"brianmargolis/shades/client"
	"brianmargolis/shades/picker"
	"brianmargolis/shades/preview"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"text/template"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const usage = `shades usage:
	Install as a launchd daemon:
	shades install
	shades uninstall

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

const serverLaunchdLabel = "com.brianmargolis.shades-server"
const clientsLaunchdLabel = "com.brianmargolis.shades-embedded-clients"

var serverPlistTemplate = template.Must(template.New("server-plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.BinaryPath}}</string>
        <string>-s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogDir}}/server-stdout.log</string>
    <key>StandardErrorPath</key>
    <string>{{.LogDir}}/server-stderr.log</string>
</dict>
</plist>
`))

var clientsPlistTemplate = template.Must(template.New("clients-plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.BinaryPath}}</string>
        <string>-c</string>{{range .Clients}}
        <string>{{.}}</string>{{end}}
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogDir}}/clients-stdout.log</string>
    <key>StandardErrorPath</key>
    <string>{{.LogDir}}/clients-stderr.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>{{.GoBin}}:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
</dict>
</plist>
`))

func main() {
	args := os.Args[1:]
	mode := "toggle"
	if len(args) > 0 {
		mode = args[0]
	}

	logger := initLogger().Sugar()
	logger.With("mode", mode, "args", args).Info("starting shades")

	switch mode {
	case "-h", "--help":
		fmt.Println(usage)
		return
	case "install":
		if err := runInstall(); err != nil {
			logger.Fatal("install failed", zap.Error(err))
		}
		return
	case "uninstall":
		if err := runUninstall(); err != nil {
			logger.Fatal("uninstall failed", zap.Error(err))
		}
		return
	}

	config, err := client.GetConfig()
	if err != nil {
		logger.Fatal("error loading config", zap.Error(err))
	}

	ctx := context.Background()

	var CLIENTS = map[string]client.Client{
		"alacritty":     client.NewAlacrittyClient(),
		"bat":           client.NewBatClient(),
		"btop":          client.NewBtopClient(),
		"claude":        client.NewClaudeClient(),
		"debug":         client.NewDebugClient(),
		"firefox":       client.NewFirefoxClient(),
		"fzf":           client.NewFZFClient(),
		"ghostty":       client.NewGhosttyClient(),
		"mac":           client.NewMacClient(),
		"mac-wallpaper": client.NewMacWallpaperClient(),
		"tmux":          client.NewTMUXClient(),
	}

	switch mode {
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

				err := client.Start(socketPath)
				if err != nil {
					logger.Fatal("error starting client", zap.String("client", clientName), zap.Error(err))
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
		NewServer().Start(socketPath)
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

		_, err := picker.NewPicker().Start(picker.PickerOpts{
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
		swatches, err := preview.NewPreviewer().Preview(theme)
		if err != nil {
			logger.Fatal(err.Error())
		}
		fmt.Println(swatches)
	}
}

func shadesLogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".shades", "logs"), nil
}

func initLogger() *zap.Logger {
	logDir, err := shadesLogDir()
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		panic(fmt.Sprintf("create log dir: %v", err))
	}

	logger, err := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{filepath.Join(logDir, "shades.log")},
		ErrorOutputPaths: []string{filepath.Join(logDir, "shades.error.log")},
	}.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		panic(err)
	}
	// set a global logger, it's a fine crutch for a package this small and
	// avoids considerable plumbing
	zap.ReplaceGlobals(logger)
	return logger
}

// runInstall installs shades as launchd user agents so the server and clients
// start automatically at login and restart if they crash.
//
// It writes two plists:
//   - com.brianmargolis.shades-server   (runs: shades -s)
//   - com.brianmargolis.shades-embedded-clients  (runs: shades -c <clients...>)
func runInstall() error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve binary path: %w", err)
	}
	// Follow symlinks so the plist points to the real binary even after rebuilds.
	binaryPath, err = filepath.EvalSymlinks(binaryPath)
	if err != nil {
		return fmt.Errorf("eval symlinks for binary path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	logDir := filepath.Join(home, ".shades", "logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	agentsDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		return fmt.Errorf("create LaunchAgents dir: %w", err)
	}

	// Write and load the server plist.
	serverPlistPath := filepath.Join(agentsDir, serverLaunchdLabel+".plist")
	var serverBuf bytes.Buffer
	if err := serverPlistTemplate.Execute(&serverBuf, struct {
		Label      string
		BinaryPath string
		LogDir     string
	}{serverLaunchdLabel, binaryPath, logDir}); err != nil {
		return fmt.Errorf("render server plist: %w", err)
	}
	if err := os.WriteFile(serverPlistPath, serverBuf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write server plist: %w", err)
	}
	fmt.Printf("Wrote %s\n", serverPlistPath)

	_ = exec.Command("launchctl", "unload", serverPlistPath).Run()
	if out, err := exec.Command("launchctl", "load", serverPlistPath).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load server: %w\n%s", err, out)
	}
	fmt.Println("Server daemon started.")

	// Determine which clients to enable.
	config, err := client.GetConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	clients := config.Daemon.EnabledComponents
	if len(clients) == 0 {
		clients = []string{"alacritty", "bat", "btop", "claude", "fzf", "ghostty", "mac", "tmux"}
	}

	// The clients plist needs a PATH that lets shades find auxiliary binaries
	// installed with `go install` (e.g. ghostty-shader-manager). GOBIN wins
	// if set; otherwise fall back to the default $HOME/go/bin.
	goBin := os.Getenv("GOBIN")
	if goBin == "" {
		goBin = filepath.Join(home, "go", "bin")
	}

	// Write and load the clients plist.
	clientsPlistPath := filepath.Join(agentsDir, clientsLaunchdLabel+".plist")
	var clientsBuf bytes.Buffer
	if err := clientsPlistTemplate.Execute(&clientsBuf, struct {
		Label      string
		BinaryPath string
		LogDir     string
		GoBin      string
		Clients    []string
	}{clientsLaunchdLabel, binaryPath, logDir, goBin, clients}); err != nil {
		return fmt.Errorf("render clients plist: %w", err)
	}
	if err := os.WriteFile(clientsPlistPath, clientsBuf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write clients plist: %w", err)
	}
	fmt.Printf("Wrote %s\n", clientsPlistPath)

	_ = exec.Command("launchctl", "unload", clientsPlistPath).Run()
	if out, err := exec.Command("launchctl", "load", clientsPlistPath).CombinedOutput(); err != nil {
		return fmt.Errorf("launchctl load clients: %w\n%s", err, out)
	}
	fmt.Println("Clients daemon started.")

	fmt.Printf("\nLogs: %s/{server,clients}-{stdout,stderr}.log\n", logDir)
	return nil
}

// runUninstall stops and removes both shades launchd agents.
func runUninstall() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	agentsDir := filepath.Join(home, "Library", "LaunchAgents")
	for _, label := range []string{serverLaunchdLabel, clientsLaunchdLabel} {
		plistPath := filepath.Join(agentsDir, label+".plist")
		if _, statErr := os.Stat(plistPath); statErr != nil {
			fmt.Printf("No plist found for %s (already uninstalled?)\n", label)
			continue
		}
		if out, err := exec.Command("launchctl", "unload", plistPath).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: launchctl unload %s: %v\n%s\n", label, err, out)
		}
		if err := os.Remove(plistPath); err != nil {
			return fmt.Errorf("remove plist %s: %w", plistPath, err)
		}
		fmt.Printf("Removed %s\n", plistPath)
	}

	fmt.Println("Daemons stopped and removed from login items.")
	return nil
}
