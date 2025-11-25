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
	return SubscribeToSocket(a.set)(socket)
}

func (a GhosttyClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}
	logger := zap.S()
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

	// send USR2 signal to ghostty process to reload config
	_, err = Run("pkill", "-USR2", "ghostty")
	if err != nil {
		return err
	}

	logger.Info("updated ghostty config file")
	return nil
}
