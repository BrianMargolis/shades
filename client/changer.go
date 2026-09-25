package client

import (
	"context"
	"strings"
	"time"

	"github.com/brianmargolis/shades/protocol"

	"github.com/pkg/errors"
)

// ChangerClient is a special client that's just invoked from the CLI to change
// the theme, but it uses the same protocol all the other clients do.
type ChangerClient struct {
	Theme string
}

func (c ChangerClient) Start(ctx context.Context, socketName string) error {
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

// CurrentTheme asks the server which theme is set. Before anything has been
// proposed, the server answers with the default for the current appearance.
func CurrentTheme(socketName string) (string, error) {
	read, write, err := SocketAsChannel(socketName)
	if err != nil {
		return "", errors.Wrap(err, "connect to the shades server")
	}
	defer close(write)

	write <- string(protocol.Get())

	timeout := time.After(2 * time.Second)
	for {
		select {
		case message, ok := <-read:
			if !ok {
				return "", errors.New("server hung up before naming the current theme")
			}
			// the palette line arrives first and is of no use here
			verb, noun, err := protocol.Parse(message)
			if err == nil && verb == "set" {
				return strings.TrimSpace(noun), nil
			}
		case <-timeout:
			return "", errors.New("timed out waiting for the server to name the current theme")
		}
	}
}

// TogglerClient is built on top of a ChangerClient and just inverts the theme.
type TogglerClient struct {
	DarkTheme  string
	LightTheme string
	Themes     Themes
}

func (c TogglerClient) Start(ctx context.Context, socketName string) error {
	currentTheme, err := CurrentTheme(socketName)
	if err != nil {
		return errors.Wrap(err, "could not get current theme")
	}

	newTheme, err := c.opposite(currentTheme)
	if err != nil {
		return err
	}

	changerClient := ChangerClient{Theme: newTheme}
	return changerClient.Start(ctx, socketName)
}

// opposite picks the default theme on the other side of the current one, so
// toggling away from any light theme lands on the default dark and vice versa.
func (c TogglerClient) opposite(currentTheme string) (string, error) {
	variant, err := c.Themes.GetVariant(currentTheme)
	if err != nil {
		return "", errors.Wrapf(err, "the current theme %q isn't in your config", currentTheme)
	}

	if variant.Light {
		return c.DarkTheme, nil
	}
	return c.LightTheme, nil
}
