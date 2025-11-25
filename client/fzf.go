package client

import (
	"fmt"

	"github.com/pkg/errors"
)

type FZFClient struct{}

func NewFZFClient() Client {
	return FZFClient{}
}

func (b FZFClient) Start(socketName string) error {
	return SubscribeToSocket(b.set)(socketName)
}

func (b FZFClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}
	fzfTheme := DoTemplate(config.Client["fzf"]["theme"], theme)

	cmd := fmt.Sprintf("set -Ux FZF_DEFAULT_OPTS '%s'", fzfTheme)
	_, err = Run("fish", "-c", cmd)
	if err != nil {
		return errors.Wrap(err, "failed to set FZF_DEFAULT_OPTS")
	}

	// set theme for fzf within zoxide
	cmd = fmt.Sprintf("set -Ux _ZO_FZF_OPTS '%s'", fzfTheme)
	_, err = Run("fish", "-c", cmd)
	if err != nil {
		return errors.Wrap(err, "failed to set _ZO_FZF_OPTS")
	}
	return nil
}
