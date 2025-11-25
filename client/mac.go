package client

import (
	"strconv"
)

type MacClient struct{}

func NewMacClient() Client {
	return MacClient{}
}

func (m MacClient) Start(socket string) error {
	return SubscribeToSocket(m.set)(socket)
}

func (m MacClient) set(theme ThemeVariant) error {
	// the single line of AppleScript that I know:
	script := `tell application "System Events" to tell appearance preferences to set dark mode to ` + strconv.FormatBool(!theme.Light)

	_, err := RunApplescript(script)
	return err
}
