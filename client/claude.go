package client

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ClaudeClient struct {
	logger *zap.SugaredLogger
}

func NewClaudeClient(
	logger *zap.SugaredLogger,
) Client {
	return ClaudeClient{
		logger: logger.With("client", "claude"),
	}
}

func (b ClaudeClient) Start(socketName string) error {
	return SubscribeToSocket(func(theme ThemeVariant) error {
		err := b.set(theme)
		if err != nil {
			b.logger.Error("Error setting theme:", err)
		}
		return errors.Wrap(err, "setting claude theme")
	})(socketName)
}

func (b ClaudeClient) set(theme ThemeVariant) error {
	themeStr := "dark"
	if theme.Light {
		themeStr = "light"
	}

	claudePath, err := b.findClaudeBinary()
	if err != nil {
		return errors.Wrap(err, "finding claude binary")
	}

	return exec.Command(claudePath, "config", "set", "-g", "theme", themeStr).Run()
}

func (b ClaudeClient) findClaudeBinary() (string, error) {
	claudePath, err := exec.LookPath("claude")
	if err == nil {
		b.logger.Debugf("Found claude binary at %s", claudePath)
		return claudePath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "getting user home directory")
	}

	localClaudePath := filepath.Join(homeDir, ".claude", "local", "claude")
	if _, err := os.Stat(localClaudePath); err == nil {
		b.logger.Debugf("Found local claude binary at %s", localClaudePath)
		return localClaudePath, nil
	}

	return "", errors.New("claude binary not found in PATH or ~/.claude/local")
}
