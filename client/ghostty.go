package client

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

type GhosttyClient struct{}

func NewGhosttyClient() Client {
	return GhosttyClient{}
}

func (a GhosttyClient) Start(socket string) error {
	return SubscribeToSocket(SetterWithContext(a.set, "ghostty"))(socket)
}

func (a GhosttyClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	path := ExpandTilde(config.Client["ghostty"]["path"])
	zap.S().Debugw("applying theme", "client", "ghostty", "theme", theme.ThemeName, "variant", theme.VariantName, "path", path)

	// clear out the file and replace it with `theme = <theme>`
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open ghostty config file: %w", err)
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("theme = \"%s-%s\"\n", theme.ThemeName, theme.VariantName)); err != nil {
		return fmt.Errorf("failed to write to ghostty config file: %w", err)
	}

	// send USR2 signal to ghostty process to reload config
	_, err = Run("pkill", "-USR2", "ghostty")
	return err
}
