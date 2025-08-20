package client

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type AlacrittyClient struct {
	logger *zap.SugaredLogger
}

func NewAlacrittyClient(
	logger *zap.SugaredLogger,
) Client {
	return AlacrittyClient{
		logger: logger.With("client", "alacritty"),
	}
}

func (a AlacrittyClient) Start(socket string) error {
	return SubscribeToSocket(func(theme ThemeVariant) error {
		err := a.set(theme)
		if err != nil {
			a.logger.Error("Error setting theme:", err)
		}
		return errors.Wrap(err, "setting alacritty theme")
	})(socket)
}

func (a AlacrittyClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}
	logger := a.logger.With("theme", theme.ThemeName, "variant", theme.VariantName)

	themeConfigPath := ExpandTilde(config.Client["alacritty"]["alacritty-config-path"])
	if themeConfigPath == "" {
		themeConfigPath = os.Getenv("HOME") + "/.config/alacritty/alacritty.toml"
	}
	mainConfigPath := ExpandTilde(config.Client["alacritty"]["alacritty-main-config-path"])
	logger = logger.With(
		"themeConfigPath", themeConfigPath,
		"mainConfigPath", mainConfigPath,
	)

	themePath := ExpandTilde(fmt.Sprintf("%s/%s-%s.toml", config.Client["alacritty"]["theme-path"], theme.ThemeName, theme.VariantName))
	logger = logger.With("themePath", themePath)

	themeContent, err := os.ReadFile(themePath)
	if err != nil {
		return errors.Wrap(err, "reading theme file")
	}

	err = os.WriteFile(themeConfigPath, themeContent, 0644)
	if err != nil {
		return errors.Wrap(err, "overwriting alacritty config with theme")
	}

	// touch the file to trigger a reload
	err = os.Chtimes(mainConfigPath, time.Now(), time.Now())
	if err != nil {
		return errors.Wrap(err, "touching alacritty config")
	}

	return nil
}
