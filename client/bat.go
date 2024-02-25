package client

import (
	"fmt"
	"os/exec"
)

type BatClient struct{}

func (b BatClient) Start(socketName string, config map[string]string) error {
	return SubscribeToSocket(b.set)(socketName)
}

func (b BatClient) set(theme string) error {
	return exec.Command("fish", "-c", fmt.Sprintf("set -Ux BAT_THEME everforest-%s", theme)).Run()
}
