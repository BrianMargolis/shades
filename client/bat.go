package client

import (
	"context"
	"fmt"
	"os/exec"
)

type BatClient struct{}

func NewBatClient() Client {
	return BatClient{}
}

func (b BatClient) Start(ctx context.Context, socketName string) error {
	return SubscribeToSocket(
		ctx,
		SetterWithContext(b.set, "bat"),
	)(socketName)
}

func (b BatClient) set(ctx context.Context, theme ThemeVariant) error {
	return exec.Command("fish", "-c", fmt.Sprintf("set -Ux BAT_THEME %s-%s", theme.ThemeName, theme.VariantName)).Run()
}
