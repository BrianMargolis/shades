package client

import (
	"os/exec"
	"strconv"

	"go.uber.org/zap"
)

type ClaudeClient struct {
}

func NewClaudeClient() Client {
	return ClaudeClient{}
}

func (b ClaudeClient) Start(socketName string) error {
	return SubscribeToSocket(b.set)(socketName)
}

func (b ClaudeClient) set(theme ThemeVariant) error {
	logger := zap.S()
	config, err := GetConfig()
	if err != nil {
		return err
	}

	themeStr := "dark"
	if theme.Light {
		themeStr = "light"
	}

	useANSIStr := config.Client["claude"]["use-ansi"]
	useANSI, err := strconv.ParseBool(useANSIStr)
	if err != nil {
		logger.With("useANSIStr", useANSIStr).Error(err)
	}

	if useANSI {
		themeStr += "-ansi"
	}

	c := exec.Command(
		"sh", "-c",
		`jq '.theme = "`+themeStr+`"' ~/.claude.json > /tmp/.claude.tmp && mv /tmp/.claude.tmp ~/.claude.json`,
	)
	return c.Run()
}
