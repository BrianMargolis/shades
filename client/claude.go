package client

import (
	"strconv"

	"go.uber.org/zap"
)

type ClaudeClient struct {
}

func NewClaudeClient() Client {
	return ClaudeClient{}
}

func (b ClaudeClient) Start(socketName string) error {
	return SubscribeToSocket(SetterWithContext(b.set, "claude"))(socketName)
}

func (b ClaudeClient) set(theme ThemeVariant) error {
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
		zap.S().Warnw("could not parse use-ansi config, defaulting to false", "value", useANSIStr, "error", err)
	}
	if useANSI {
		themeStr += "-ansi"
	}

	zap.S().Debugw("applying theme", "client", "claude", "theme", theme.ThemeName, "variant", theme.VariantName, "claudeTheme", themeStr)

	_, err = Run(
		"sh", "-c",
		`jq '.theme = "`+themeStr+`"' ~/.claude.json > /tmp/.claude.tmp && mv /tmp/.claude.tmp ~/.claude.json`,
	)
	return err
}
