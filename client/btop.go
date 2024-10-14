package client

import (
	"fmt"

	"github.com/pkg/errors"
)

type BtopClient struct{}

func NewBtopClient() Client {
	return BtopClient{}
}

func (b BtopClient) Start(
	socket string,
) error {
	return SubscribeToSocket(b.set)(socket)
}

func (b BtopClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	btopConfigPath := ExpandTilde(config.Client["btop"]["btop-config-path"])
	themePath := ExpandTilde(config.Client["btop"]["theme-path"])

	path := fmt.Sprintf("%s/%s-%s.theme", themePath, theme.ThemeName, theme.VariantName)
	configLine := "color_theme = " + path

	n, err := ReplaceAtTag(btopConfigPath, configLine, "color_theme = ")
	if err != nil {
		return errors.Wrap(err, "replacing theme in btop config")
	}

	if n == 0 {
		fmt.Println("WARNING: couldn't find a color_theme in the btop config: " + btopConfigPath)
	}

	return nil
}
