package client

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

type FZFClient struct {
	darkTheme  string
	lightTheme string
}

func (b FZFClient) Start(socketName string, config map[string]string) error {
	b.darkTheme = config["dark-theme"]
	b.lightTheme = config["light-theme"]
	return SubscribeToSocket(b.set)(socketName)
}

func (b FZFClient) set(theme string) error {
	// generate these with https://vitormv.github.io/fzf-themes/
	fzfTheme := b.darkTheme
	if theme == "light" {
		fzfTheme = b.lightTheme
	}

	// set the theme for fzf
	cmd := fmt.Sprintf("set -Ux FZF_DEFAULT_OPTS '%s'", fzfTheme)
	err := exec.Command("fish", "-c", cmd).Run()
	if err != nil {
		return errors.Wrap(err, "failed to set FZF_DEFAULT_OPTS")
	}

	// set theme for fzf within zoxide
	cmd = fmt.Sprintf("set -Ux _ZO_FZF_OPTS '%s'", fzfTheme)
	err = exec.Command("fish", "-c", cmd).Run()
	if err != nil {
		return errors.Wrap(err, "failed to set _ZO_FZF_OPTS")
	}
	return nil
}
