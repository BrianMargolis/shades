package client

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// btopLiveTheme is the theme file this client rewrites on every change. btop.conf
// names it once (color_theme = "shades") and never changes, so the config can
// live in git without churning on each theme switch.
const btopLiveTheme = "shades.theme"

type BtopClient struct{}

func NewBtopClient() Client {
	return BtopClient{}
}

func (b BtopClient) Start(socket string) error {
	return SubscribeToSocket(SetterWithContext(b.set, "btop"))(socket)
}

func (b BtopClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	themePath := ExpandTilde(config.Client["btop"]["theme-path"])
	sourcePath := filepath.Join(themePath, fmt.Sprintf("%s-%s.theme", theme.ThemeName, theme.VariantName))
	livePath := filepath.Join(themePath, btopLiveTheme)

	zap.S().Debugw("applying theme", "client", "btop", "theme", theme.ThemeName, "variant", theme.VariantName, "sourcePath", sourcePath, "livePath", livePath)

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return errors.Wrapf(err, "reading btop theme %s", sourcePath)
	}
	if err := os.WriteFile(livePath, content, 0644); err != nil {
		return errors.Wrapf(err, "writing btop theme to %s", livePath)
	}

	return nil
}
