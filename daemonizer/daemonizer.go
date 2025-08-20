package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"brianmargolis/shades/client"

	"github.com/pkg/errors"
)

func validateDependencies() error {
	// validate shades is installed
	err := exec.Command("which", "shades").Run()
	if err != nil {
		return errors.New("shades not found in GOPATH/bin")
	}

	// validate launchctl exists
	err = exec.Command("which", "launchctl").Run()
	if err != nil {
		return errors.New("launchctl not found")
	}

	return nil
}

func getEnabledComponents(config client.ConfigModel) []string {
	// if daemon config is present, use it
	if len(config.Daemon.EnabledComponents) > 0 {
		return config.Daemon.EnabledComponents
	}

	// fallback to all available components if not configured
	availableComponents := []string{
		"alacritty",
		"bat",
		"btop",
		"claude",
		"fzf",
		"mac",
		"tmux",
	}

	return availableComponents
}

// TODO support empty selectedChoices
func installConfigs(selectedChoices []string, verbose bool) error {
	serverName := "com.brianmargolis.shades-server"
	clientsName := "com.brianmargolis.shades-embedded-clients"
	serverPlistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", serverName+".plist")
	clientsPlistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", clientsName+".plist")

	shadesPath := os.Getenv("GOPATH") + "/bin/shades"

	// generate and write plist content for shades-server
	serverPlistContent := generatePlistContent(serverName, shadesPath, []string{"-s"}, true)
	err := os.WriteFile(serverPlistPath, []byte(serverPlistContent), 0644)
	if err != nil {
		return errors.Wrap(err, "error writing plist file")
	}

	// generate and write plist content for shades-embedded-clients with selected choices
	clientsPlistContent := generatePlistContent(clientsName, shadesPath, append([]string{"-c"}, selectedChoices...), true)
	err = writePlistFile(clientsPlistPath, clientsPlistContent)
	if err != nil {
		return err
	}

	// load the plist files using launchctl
	err = reloadPlist(serverPlistPath)
	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second) // wait for the server to start before we get the clients going
	err = reloadPlist(clientsPlistPath)
	if err != nil {
		return err
	}

	return nil
}

func generatePlistContent(label, command string, args []string, runAtLoad bool) string {
	// Generate the XML content for the plist file
	// TODO real templating
	plist := `<?xml version="1.0" encoding="UTF-8"?>
						<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
						<plist version="1.0">
							<dict>
								<key>Label</key>
								<string>%s</string>
								<key>Program</key>
								<string>%s</string>
								<key>ProgramArguments</key>
								<array>
									<string>%s</string>%s
								</array>
								<key>RunAtLoad</key>
								<%s/>
								<key>StandardOutPath</key>
								<string>/tmp/%s-out.log</string>
								<key>StandardErrorPath</key>
								<string>/tmp/%s-error.log</string>
								<key>EnvironmentVariables</key>
								<dict>
									<key>PATH</key>
										<string>/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:%s/bin:/usr/bin/osascript</string>
								</dict>
							</dict>
						</plist>`
	argsContent := ""
	for _, arg := range args {
		argsContent += fmt.Sprintf("\n\t\t<string>%s</string>", arg)
	}
	runAtLoadStr := "false"
	if runAtLoad {
		runAtLoadStr = "true"
	}
	return fmt.Sprintf(plist, label, command, command, argsContent, runAtLoadStr, label, label, os.Getenv("GOPATH"))
}

func writePlistFile(path, content string) error {
	err := os.WriteFile(path, []byte(content), 0644)
	return errors.Wrap(err, "error writing plist file")
}

func reloadPlist(path string) error {
	// fine if the unload fails
	exec.Command("launchctl", "unload", path).Run()
	// not fine if the load fails though
	err := exec.Command("launchctl", "load", path).Run()
	return errors.Wrap(err, "error loading plist file")
}

func main() {
	args := os.Args[1:]
	var verbose bool

	for _, arg := range args {
		switch arg {
		case "-u":
			log.Fatal("Uninstall functionality not yet implemented in config-driven mode")
		case "-v":
			verbose = true
		}
	}

	// validate dependencies
	err := validateDependencies()
	if err != nil {
		log.Fatal(err)
	}

	// load configuration
	config, err := client.GetConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// get enabled components from config
	enabledComponents := getEnabledComponents(config)

	if verbose {
		fmt.Printf("Configuring daemon with components: %v\n", enabledComponents)
	}

	// install/configure the daemon services
	err = installConfigs(enabledComponents, verbose)
	if err != nil {
		log.Fatal("Failed to configure daemon:", err)
	}

	fmt.Println("Daemon configuration completed successfully")
}
