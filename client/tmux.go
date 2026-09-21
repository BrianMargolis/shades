package client

import (
	"strings"

	"go.uber.org/zap"
)

type TMUXClient struct{}

func NewTMUXClient() Client {
	return TMUXClient{}
}

func (t TMUXClient) Start(socket string) error {
	return SubscribeToSocket(SetterWithContext(t.set, "tmux"))(socket)
}

func (t TMUXClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	theme = lowercaseColors(theme)

	for _, optionName := range []string{
		"status-bg",
		"status-fg",
		"window-status-format",
		"window-status-current-format",
		"status-left",
		"status-right",
		"status-format[1]",
		"clock-mode-colour",
		"window-style",
		"cursor-colour",
	} {
		template, ok := config.Client["tmux"][optionName]
		if !ok {
			continue
		}
		value := DoTemplate(template, theme)
		zap.S().Debugw("setting tmux option", "option", optionName)

		if err := t.setTMUXOption(optionName, value); err != nil {
			return err
		}
	}

	return nil
}

// lowercaseColors downcases every hex value before it reaches a tmux format
// string. tmux expands "#X" as a legacy single-character alias, so an uppercase
// color like "#D47766" turns into pane_id followed by "47766" wherever the
// format actually gets expanded (inside a "#{?...}" branch, for instance). Hex
// digits are 0-9 and a-f, and "#h" is the only lowercase alias tmux defines, so
// downcasing can never collide.
func lowercaseColors(theme ThemeVariant) ThemeVariant {
	colors := make(map[Color]string, len(theme.Colors))
	for name, value := range theme.Colors {
		colors[name] = strings.ToLower(value)
	}
	theme.Colors = colors
	return theme
}

func (t TMUXClient) setTMUXOption(optionName, value string) error {
	tmuxPath, err := LookPath("tmux")
	if err != nil {
		zap.S().Errorw("tmux executable not found", "error", err)
		return err
	}

	_, err = Run(tmuxPath, "set-option", "-g", optionName, value)
	return err
}
