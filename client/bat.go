package client

import (
	"fmt"

	"go.uber.org/zap"
)

type BatClient struct{}

func NewBatClient() Client {
	return BatClient{}
}

func (b BatClient) Start(socketName string) error {
	return SubscribeToSocket(SetterWithContext(b.set, "bat"))(socketName)
}

func (b BatClient) set(theme ThemeVariant) error {
	zap.S().Debugw("applying theme", "client", "bat", "theme", theme.ThemeName, "variant", theme.VariantName)
	_, err := Run("fish", "-c", fmt.Sprintf("set -Ux BAT_THEME %s-%s", theme.ThemeName, theme.VariantName))
	return err
}
