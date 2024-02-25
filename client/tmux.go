package client

import (
	"fmt"
	"os/exec"
)

type TMUXClient struct {
	config map[string]string
}

func (t TMUXClient) Start(socket string, config map[string]string) error {
	t.config = config
	return SubscribeToSocket(t.set)(socket)
}

func (t TMUXClient) set(theme string) error {
	for _, optionName := range []string{
		"status-bg",
		"status-fg",
		"window-status-format",
		"window-status-current-format",
		"status-left",
		"status-right",
	} {
		err := t.setTMUXOption(optionName, t.config[fmt.Sprintf("%s-%s", theme, optionName)])
		if err != nil {
			return err
		}
	}

	return nil
}

func (t TMUXClient) setTMUXOption(optionName, value string) error {
	fmt.Printf("%s: %s\n", optionName, value)
	_, err := exec.Command("/usr/local/bin/tmux", "set-option", "-g", optionName, value).Output()
	if err != nil {
		fmt.Printf("ERROR setting %s: %s", optionName, err.Error())
	}
	return err
}
