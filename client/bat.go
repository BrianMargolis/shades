package client

import (
	"context"
	"fmt"
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
	fishCommand := fmt.Sprintf(
		"set -Ux BAT_THEME %s-%s",
		theme.ThemeName,
		theme.VariantName,
	)
	_, err := Run("fish", "-c", fishCommand)
	return err
}
