package client

import (
	"os/exec"
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
	// jq '.theme = "light"' ~/.claude.json > /tmp/.claude.tmp && mv /tmp/.claude.tmp ~/.claude.json
	themeStr := "dark"
	if theme.Light {
		themeStr = "light"
	}

	c := exec.Command(
		"sh", "-c",
		`jq '.theme = "`+themeStr+`"' ~/.claude.json > /tmp/.claude.tmp && mv /tmp/.claude.tmp ~/.claude.json`,
	)
	return c.Run()
}
