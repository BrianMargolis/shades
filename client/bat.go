package client

import (
	"fmt"
	"os/exec"
)

type BatClient struct{}

func NewBatClient() Client {
	return BatClient{}
}

func (b BatClient) Start(socketName string) error {
	return SubscribeToSocket(b.set)(socketName)
}

func (b BatClient) set(theme ThemeVariant) error {
	return exec.Command("fish", "-c", fmt.Sprintf("set -Ux BAT_THEME %s-%s", theme.ThemeName, theme.VariantName)).Run()
}
