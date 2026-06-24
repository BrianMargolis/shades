package client

import (
	"fmt"
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

	return setClaudeTheme(themeStr)
}

// setClaudeTheme writes the theme name into Claude Code's settings.json. The
// theme used to live in ~/.claude.json, but Claude Code now reads it from
// ~/.claude/settings.json. jq (rather than encoding/json) is used so the rest
// of the user's hand-edited settings keep their original key order.
func setClaudeTheme(themeStr string) error {
	settingsPath := ExpandTilde("~/.claude/settings.json")

	// jq has no in-place flag, so emit to an adjacent temp file and atomically
	// rename it over the original; keeping the temp file on the same filesystem
	// is what makes the rename atomic. The theme is passed via --arg to avoid
	// quoting/injection issues.
	script := `jq --arg theme "$1" '.theme = $theme' "$2" > "$2.tmp" && mv "$2.tmp" "$2"`
	if _, err := Run("sh", "-c", script, "sh", themeStr, settingsPath); err != nil {
		return fmt.Errorf("failed to set claude theme in %s: %w", settingsPath, err)
	}
	return nil
}
