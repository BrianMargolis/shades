package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
)

type model struct {
	choices     []string
	cursor      int
	selected    map[int]bool
	configuring bool
	uninstall   bool
	verbose     bool
}

func initialModel() model {
	return model{
		choices: []string{
			"alacritty",
			"bat",
			"btop",
			"fzf",
			"mac",
			"mac-wallpaper",
			"tmux",
		},
		cursor:   0,
		selected: make(map[int]bool),
	}
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type confirmMsg struct{}

func (m model) Init() tea.Cmd {
	// validate shades is installed
	err := exec.Command("which", "shades").Run()
	if err != nil {
		return func() tea.Msg {
			return errMsg{errors.New("shades not found in GOPATH/bin")}
		}
	}

	// validate launchctl exists
	err = exec.Command("which", "launchctl").Run()
	if err != nil {
		return func() tea.Msg {
			return errMsg{errors.New("launchctl not found")}
		}
	}

	// default to all because that's what I use
	for i := range m.choices {
		m.selected[i] = true
	}

	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case tea.KeySpace:
			m.selected[m.cursor] = !m.selected[m.cursor]
		case tea.KeyEnter:
			return m, createLaunchdConfigs(m.selectedChoices(), m.uninstall, m.verbose)

		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				return m, tea.Quit
			case "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			}
		}
	case errMsg:
		log.Fatal(msg)
		return m, tea.Quit
	case confirmMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	if m.configuring {
		return "Configuring...\n"
	}

	s := "Choose components for shades-embedded-clients (space to select, enter to confirm):\n\n"

	for i, choice := range m.choices {
		checked := " "
		if m.selected[i] {
			checked = "x"
		}

		pointer := " "
		if i == m.cursor {
			pointer = ">"
		}

		s += fmt.Sprintf("%s [%s] %s\n", pointer, checked, choice)
	}

	s += "\nPress q to quit.\n"
	return s
}

func (m model) selectedChoices() []string {
	var choices []string
	for i, choice := range m.choices {
		if m.selected[i] {
			choices = append(choices, choice)
		}
	}
	return choices
}

func createLaunchdConfigs(selectedChoices []string, uninstall bool, verbose bool) tea.Cmd {
	return func() tea.Msg {
		err := installConfigs(selectedChoices, verbose)
		if err != nil {
			return errMsg{err}
		}

		return confirmMsg{}
	}
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
								<string>/usr/local/var/log/%s-out.log</string>
								<key>StandardErrorPath</key>
								<string>/usr/local/var/log/%s-error.log</string>
								<key>EnvironmentVariables</key>
								<dict>
									<key>PATH</key>
									<string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:%s/bin:/usr/bin/osascript</string>
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
	m := initialModel()

	for _, arg := range args {
		switch arg {
		case "-u":
			m.uninstall = true
		case "-v":
			m.verbose = true
		}
	}

	program := tea.NewProgram(m)
	_, err := program.Run()
	if err != nil {
		log.Fatal(err)
	}
}
