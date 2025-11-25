package client

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/creack/pty"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ClaudeClient struct {
	sleepMultiplier int
}

func NewClaudeClient() Client {
	return ClaudeClient{
		sleepMultiplier: 3, // crank this up to debug
	}
}

func (b ClaudeClient) Start(socketName string) error {
	return SubscribeToSocket(b.set)(socketName)
}

func (b ClaudeClient) set(theme ThemeVariant) error {
	logger := zap.S()
	claudePath, err := b.findClaudeBinary()
	if err != nil {
		logger.Error("Could not find claude binary", "error", err)
		return fmt.Errorf("finding claude binary: %w", err)
	}
	// Start `claude /config` attached to a PTY.
	cmd := exec.Command(claudePath, "/config")
	logger.Debug("starting claude command", "cmd", cmd.String())

	logger.With("path", claudePath).Infof("Starting claude binary to set theme to %s", theme.ThemeName)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("start pty: %w", err)
	}
	defer func() {
		err = ptmx.Close()
		if err != nil {
			logger.Error("failed to close pty", "error", err)
		}
	}()

	// Wire your terminal <-> PTY so you can see and type.
	// PTY -> stdout
	go func() {
		_, _ = io.Copy(os.Stdout, ptmx)
	}()

	// stdin -> PTY
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()

	// Give Claude a moment to boot and render.
	// TODO - can we detect when it's ready instead of sleeping?
	time.Sleep(500 * time.Millisecond * time.Duration(b.sleepMultiplier))

	logger.Debug("Sending keys to set theme...")

	// send 6 down arrows
	for range 6 {
		if _, err := fmt.Fprint(ptmx, "\x1b[B"); err != nil {
			err = errors.Wrap(err, "sending down arrow")
			return err
		}
	}
	time.Sleep(100 * time.Millisecond * time.Duration(b.sleepMultiplier))

	// send an enter
	if _, err := fmt.Fprint(ptmx, "\r"); err != nil {
		return errors.Wrap(err, "sending enter")
	}
	time.Sleep(100 * time.Millisecond * time.Duration(b.sleepMultiplier))

	// send a 5 for dark theme, 6 for light
	themeIndex := "5"
	if theme.Light {
		themeIndex = "6"
	}
	logger.Debugf("Setting theme index to %s", themeIndex)

	if _, err := fmt.Fprint(ptmx, themeIndex); err != nil {
		return errors.Wrap(err, "sending theme index")
	}
	time.Sleep(10 * time.Millisecond * time.Duration(b.sleepMultiplier))

	err = cmd.Process.Kill()
	if err != nil {
		return errors.Wrap(err, "killing claude process")
	}

	return nil
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
