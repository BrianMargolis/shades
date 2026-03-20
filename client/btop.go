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
	return SubscribeToSocket(SetterWithContext(b.set, "btop"))(socket)
}

func (b BtopClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	btopConfigPath := ExpandTilde(config.Client["btop"]["btop-config-path"])
	themePath := ExpandTilde(config.Client["btop"]["theme-path"])
	configLine := fmt.Sprintf("color_theme = %s/%s-%s.theme", themePath, theme.ThemeName, theme.VariantName)

	zap.S().Debugw("applying theme", "client", "btop", "theme", theme.ThemeName, "variant", theme.VariantName, "btopConfigPath", btopConfigPath)

	n, err := ReplaceAtTag(btopConfigPath, configLine, "color_theme = ")
	if err != nil {
		return errors.Wrap(err, "replacing theme in btop config")
	}
	if n == 0 {
		zap.S().Warnw("no replacements made in btop config", "btopConfigPath", btopConfigPath)
	}

	return nil
}
