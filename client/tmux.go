package client

import (
	"context"
	"os/exec"
)

type TMUXClient struct{}

func NewTMUXClient() Client {
	return TMUXClient{}
}

func (t TMUXClient) Start(ctx context.Context, socket string) error {
	return SubscribeToSocket(
		ctx,
		SetterWithContext(t.set, "tmux"),
	)(socket)
}

func (t TMUXClient) set(ctx context.Context, theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}
	for _, optionName := range []string{
		"status-bg",
		"status-fg",
		"window-status-format",
		"window-status-current-format",
		"status-left",
		"status-right",
	} {
		template := config.Client["tmux"][optionName]
		value := DoTemplate(template, theme)

		err := t.setTMUXOption(optionName, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t TMUXClient) setTMUXOption(optionName, value string) error {
	logger := LoggerFromContext(context.Background())
	tmuxPath, err := LookPath("tmux")
	if err != nil {
		return err
	}

	cmd := []string{"set-option", "-g", optionName, value}
	logger = logger.With("tmuxPath", tmuxPath, "cmd", cmd)
	logger.Debug("setting tmux option...")
	return exec.Command(tmuxPath, cmd...).Run()
}
