package client

import (
	"context"
	"strconv"
)

type MacClient struct{}

func NewMacClient() Client {
	return MacClient{}
}

func (m MacClient) Start(ctx context.Context, socket string) error {
	return SubscribeToSocket(
		ctx,
		SetterWithContext(m.set, "mac"),
	)(socket)
}

func (m MacClient) set(ctx context.Context, theme ThemeVariant) error {
	// the single line of AppleScript that I know:
	script := `tell application "System Events" to tell appearance preferences to set dark mode to ` + strconv.FormatBool(!theme.Light)

	_, err := RunApplescript(script)
	return err
}
