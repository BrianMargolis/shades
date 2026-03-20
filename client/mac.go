package client

import (
	"strconv"

	"go.uber.org/zap"
)

type MacClient struct{}

func NewMacClient() Client {
	return MacClient{}
}

func (m MacClient) Start(socket string) error {
	return SubscribeToSocket(SetterWithContext(m.set, "mac"))(socket)
}

func (m MacClient) set(theme ThemeVariant) error {
	zap.S().Debugw("applying theme", "client", "mac", "theme", theme.ThemeName, "variant", theme.VariantName, "darkMode", !theme.Light)
	// the single line of AppleScript that I know:
	script := `tell application "System Events" to tell appearance preferences to set dark mode to ` + strconv.FormatBool(!theme.Light)
	_, err := RunApplescript(script)
	return err
}
