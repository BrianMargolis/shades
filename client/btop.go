package client

import (
	"fmt"

	"github.com/pkg/errors"
)

type BtopClient struct {
	btopConfigPath string
	themePath      string
}

func NewBtopClient(config map[string]string) Client {
	btopConfigPath := ExpandTilde(config["btop-config-path"])
	themePath := ExpandTilde(config["theme-path"])
	return BtopClient{
		btopConfigPath: btopConfigPath,
		themePath:      themePath,
	}
}

func (b BtopClient) Start(
	socket string,
) error {
	return SubscribeToSocket(b.set)(socket)
}

func (b BtopClient) set(theme ThemeVariant) error {
	path := fmt.Sprintf("%s/%s-%s.theme", b.themePath, theme.ThemeName, theme.VariantName)
	configLine := "color_theme = " + path

	n, err := ReplaceAtTag(b.btopConfigPath, configLine, "color_theme = ")
	if err != nil {
		return errors.Wrap(err, "replacing theme in btop config")
	}

	if n == 0 {
		fmt.Println("WARNING: couldn't find a color_theme in the btop config: " + b.btopConfigPath)
	}

	return nil
}
