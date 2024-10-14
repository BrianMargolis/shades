package client

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

type TMUXClient struct{}

func NewTMUXClient() Client {
	return TMUXClient{}
}

func (t TMUXClient) Start(socket string) error {
	return SubscribeToSocket(t.set)(socket)
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
	tmuxPath, err := t.getTMUXExecutablePath()
	if err != nil {
		return err
	}

	_, err = exec.Command(tmuxPath, "set-option", "-g", optionName, value).Output()
	if err != nil {
		fmt.Printf("ERROR setting %s: %s", optionName, err.Error())
	}
	return err
}

func (t TMUXClient) getTMUXExecutablePath() (string, error) {
	path, err := exec.LookPath("tmux")
	if err != nil {
		return "", errors.Wrap(err, "tmux executable not found")
	}
	return path, nil
}
