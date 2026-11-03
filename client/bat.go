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

// BatThemeName is the name the bat theme for a variant is installed under.
func BatThemeName(theme ThemeVariant) string {
	return fmt.Sprintf("%s-%s", theme.ThemeName, theme.VariantName)
}

func (b BatClient) set(theme ThemeVariant) error {
	zap.S().Debugw("applying theme", "client", "bat", "theme", theme.ThemeName, "variant", theme.VariantName)
	_, err := Run("fish", "-c", "set -Ux BAT_THEME "+BatThemeName(theme))
	return err
}
