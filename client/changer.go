package client

import (
	"brianmargolis/shades/protocol"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// ChangerClient is a special client that's just invoked from the CLI to change
// the theme, but it uses the same protocol all the other clients do.
type ChangerClient struct {
	Theme string
}

func (c ChangerClient) Start(socketName string) error {
	_, write, err := SocketAsChannel(socketName)
	if err != nil {
		return err
	}

	// hello
	write <- string(protocol.Subscribe("cli"))
	// please set the theme to this
	write <- string(protocol.Propose(c.Theme))
	// goodbye,
	write <- string(protocol.Unsubscribe())
	// forever
	close(write)

	return nil
}

// TogglerClient is built on top of a ChangerClient and just inverts the theme.
type TogglerClient struct {
	DarkTheme  string
	LightTheme string
}

func (c TogglerClient) Start(socketName string) error {
	currentTheme, err := c.getCurrentTheme()
	if err != nil {
		return errors.Wrap(err, "could not get current theme")
	}

	newTheme := c.LightTheme
	if currentTheme == c.LightTheme {
		newTheme = c.DarkTheme
	}

	changerClient := ChangerClient{Theme: newTheme}
	return changerClient.Start(socketName)
}

func (c TogglerClient) getCurrentTheme() (string, error) {
	script := `tell application "System Events" to tell appearance preferences to get dark mode`

	output, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(string(output)) == "true" {
		return c.DarkTheme, nil
	} else {
		return c.LightTheme, nil
	}
}
