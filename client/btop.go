package client

import (
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type BtopClient struct{}

func NewBtopClient() Client {
	return BtopClient{}
}

func (b BtopClient) Start(socket string) error {
	return SubscribeToSocket(b.set)(socket)
}

func (b BtopClient) set(theme ThemeVariant) error {
	logger := zap.S()

	config, err := GetConfig()
	if err != nil {
		return err
	}

	btopConfigPath := ExpandTilde(config.Client["btop"]["btop-config-path"])
	themePath := ExpandTilde(config.Client["btop"]["theme-path"])
	logger = logger.With("btopConfigPath", btopConfigPath, "themePath", themePath)

	path := fmt.Sprintf("%s/%s-%s.theme", themePath, theme.ThemeName, theme.VariantName)
	configLine := "color_theme = " + path
	logger = logger.With("configLine", configLine)

	logger.Debug("setting btop theme...")
	n, err := ReplaceAtTag(btopConfigPath, configLine, "color_theme = ")
	if err != nil {
		return errors.Wrap(err, "replacing theme in btop config")
	}

	if n == 0 {
		logger.Warn("no replacements made in btop config")
	}

	return nil
}
