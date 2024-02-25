package client

import (
	"fmt"
	"os/exec"
)

type BatClient struct {
	darkTheme  string
	lightTheme string
}

func (b BatClient) Start(socketName string, config map[string]string) error {
	b.darkTheme = config["dark-theme"]
	b.lightTheme = config["light-theme"]
	return SubscribeToSocket(b.set)(socketName)
}

func (b BatClient) set(theme string) error {
	t := b.lightTheme
	if theme == "dark" {
		t = b.darkTheme
	}
	return exec.Command("fish", "-c", fmt.Sprintf("set -Ux BAT_THEME %s", t)).Run()
}
