package client

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

type ClaudeClient struct{}

func NewClaudeClient() Client {
	return ClaudeClient{}
}

func (b ClaudeClient) Start(ctx context.Context, socketName string) error {
	return SubscribeToSocket(
		ctx,
		SetterWithContext(b.set, "claude"),
	)(socketName)
}

func (b ClaudeClient) set(ctx context.Context, theme ThemeVariant) error {
	themeStr := "dark"
	if theme.Light {
		themeStr = "light"
	}

	claudePath, err := b.findClaudeBinary(ctx)
	if err != nil {
		return errors.Wrap(err, "finding claude binary")
	}

	return exec.Command(claudePath, "config", "set", "-g", "theme", themeStr).Run()
}

func (b ClaudeClient) findClaudeBinary(ctx context.Context) (string, error) {
	logger := LoggerFromContext(ctx)

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
