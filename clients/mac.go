package client

import (
	"brianmargolis/theme-daemon/protocol"
	"os/exec"
	"strings"
)

type MacClient struct{}

func (m MacClient) Start(socket string) error {
	read, write, err := SocketAsChannel(socket)
	if err != nil {
		return err
	}

	write <- string(protocol.Subscribe("mac"))

	for message := range read {
		verb, noun, err := protocol.Parse(message)
		if err != nil {
			return err
		}

		if verb == "set" {
			theme := strings.TrimSpace(noun)
			err = m.set(theme)
			if err != nil {
				panic(err)
			}
		}
	}

	return nil
}

func (m MacClient) set(theme string) error {
	script := `osascript -e 'tell application "System Events" to tell appearance preferences to set dark mode to not dark mode'`
	_, err := exec.Command("osascript", "-e", script).Output()
	return err
}
