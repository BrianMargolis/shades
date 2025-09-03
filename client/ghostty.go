package client

import (
	"context"
	"fmt"
	"os"
)

type GhosttyClient struct{}

func NewGhosttyClient() Client {
	return GhosttyClient{}
}

func (a GhosttyClient) Start(ctx context.Context, socket string) error {
	return SubscribeToSocket(
		ctx,
		SetterWithContext(a.set, "ghostty"),
	)(socket)
}

func (a GhosttyClient) set(ctx context.Context, theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}
	logger := LoggerFromContext(ctx)
	logger = logger.With("theme", theme.ThemeName, "variant", theme.VariantName)

	path := ExpandTilde(config.Client["ghostty"]["path"])
	logger = logger.With(
		"path", path,
	)

	// clear out the file and replace it with `theme = <theme>`
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("failed to open ghostty config file", "error", err)
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("theme = \"%s-%s\"\n", theme.ThemeName, theme.VariantName))
	if err != nil {
		logger.Error("failed to write to ghostty config file", "error", err)
		return err
	}

	// TODO: send USR2 signal to ghostty to reload config once that's landed:
	// https://github.com/ghostty-org/ghostty/discussions/3643#discussioncomment-13899379
	// in the meantime, this works, but requires the user to grant shades accessibility permissions:
	RunApplescript(`
		tell application "System Events"
			tell process "Ghostty"
					click menu item "Reload Configuration" of menu "Ghostty" of menu bar item "Ghostty" of menu bar 1
			end tell
		end tell`,
	)

	logger.Info("updated ghostty config file")
	return nil
}
