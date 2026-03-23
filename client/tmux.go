package client

import (
	"go.uber.org/zap"
)

type TMUXClient struct{}

func NewTMUXClient() Client {
	return TMUXClient{}
}

func (t TMUXClient) Start(socket string) error {
	return SubscribeToSocket(SetterWithContext(t.set, "tmux"))(socket)
}

func (t TMUXClient) set(theme ThemeVariant) error {
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
		"clock-mode-colour",
	} {
		template, ok := config.Client["tmux"][optionName]
		if !ok {
			continue
		}
		value := DoTemplate(template, theme)
		zap.S().Debugw("setting tmux option", "option", optionName)

		if err := t.setTMUXOption(optionName, value); err != nil {
			return err
		}
	}

	return nil
}

func (t TMUXClient) setTMUXOption(optionName, value string) error {
	tmuxPath, err := LookPath("tmux")
	if err != nil {
		zap.S().Errorw("tmux executable not found", "error", err)
		return err
	}

	_, err = Run(tmuxPath, "set-option", "-g", optionName, value)
	return err
}
