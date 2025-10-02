package client

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

	// send USR2 signal to ghostty process to reload config
	cmd := exec.Command("pkill", "-USR2", "ghostty")
	err = cmd.Run()
	if err != nil {
		stdout := cmd.Stdout
		stderr := cmd.Stderr
		logger.With("stdout", stdout, "stderr", stderr).Debug("failed pkill command output")
		logger.Error("failed to send USR2 signal to ghostty", "error", err)
		return err
	}

	logger.Info("updated ghostty config file")
	return nil
}
