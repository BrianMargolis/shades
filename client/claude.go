package client

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ClaudeClient struct{}

func NewClaudeClient() Client {
	return ClaudeClient{}
}

func (b ClaudeClient) Start(socketName string) error {
	return SubscribeToSocket(b.set)(socketName)
}

func (b ClaudeClient) set(theme ThemeVariant) error {
	// TODO: this does not work anymore, they removed `config` from the CLI. think this needs to invoke claude, send /config, navigate the TUI, etc... awful
	return nil
	themeStr := "dark"
	if theme.Light {
		themeStr = "light"
	}

	claudePath, err := b.findClaudeBinary()
	if err != nil {
		return errors.Wrap(err, "finding claude binary")
	}

	_, err = Run(claudePath, "config", "set", "-g", "theme", themeStr)
	return err
}

func (b ClaudeClient) findClaudeBinary() (string, error) {
	logger := zap.S()

	claudePath, err := exec.LookPath("claude")
	if err == nil {
		logger.Debugf("Found claude binary at %s", claudePath)
		return claudePath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "getting user home directory")
	}

	localClaudePath := filepath.Join(homeDir, ".claude", "local", "claude")
	if _, err := os.Stat(localClaudePath); err == nil {
		logger.Debugf("Found local claude binary at %s", localClaudePath)
		return localClaudePath, nil
	}

	return "", errors.New("claude binary not found in PATH or ~/.claude/local")
}
