package client

import (
	"fmt"
)

type BatClient struct{}

func NewBatClient() Client {
	return BatClient{}
}

func (b BatClient) Start(socketName string) error {
	return SubscribeToSocket(b.set)(socketName)
}

func (b BatClient) set(theme ThemeVariant) error {
	fishCommand := fmt.Sprintf(
		"set -Ux BAT_THEME %s-%s",
		theme.ThemeName,
		theme.VariantName,
	)
	_, err := Run("fish", "-c", fishCommand)
	return err
}
