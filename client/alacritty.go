package client

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type AlacrittyClient struct{}

func NewAlacrittyClient() Client {
	return AlacrittyClient{}
}

func (a AlacrittyClient) Start(socket string) error {
	return SubscribeToSocket(SetterWithContext(a.set, "alacritty"))(socket)
}

func (a AlacrittyClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	themeConfigPath := ExpandTilde(config.Client["alacritty"]["alacritty-config-path"])
	if themeConfigPath == "" {
		themeConfigPath = os.Getenv("HOME") + "/.config/alacritty/alacritty.toml"
	}
	mainConfigPath := ExpandTilde(config.Client["alacritty"]["alacritty-main-config-path"])
	themePath := ExpandTilde(fmt.Sprintf("%s/%s-%s.toml", config.Client["alacritty"]["theme-path"], theme.ThemeName, theme.VariantName))

	zap.S().Debugw("applying theme", "client", "alacritty", "theme", theme.ThemeName, "variant", theme.VariantName, "themePath", themePath, "themeConfigPath", themeConfigPath)

	themeContent, err := os.ReadFile(themePath)
	if err != nil {
		return errors.Wrap(err, "reading theme file")
	}

	if err = os.WriteFile(themeConfigPath, themeContent, 0644); err != nil {
		return errors.Wrap(err, "overwriting alacritty config with theme")
	}

	// touch the file to trigger a reload
	if err = os.Chtimes(mainConfigPath, time.Now(), time.Now()); err != nil {
		return errors.Wrap(err, "touching alacritty config")
	}

	return nil
}
