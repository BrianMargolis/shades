package client

import (
	"fmt"

	"github.com/pkg/errors"
)

type BtopClient struct {
	btopConfigPath string
	lightThemePath string
	darkThemePath  string
}

func (b BtopClient) Start(socket string, config map[string]string) error {
	b.btopConfigPath = ExpandTilde(config["btop-config-path"])
	b.lightThemePath = ExpandTilde(config["light-theme-path"])
	b.darkThemePath = ExpandTilde(config["dark-theme-path"])

	return SubscribeToSocket(b.set)(socket)
}

func (b BtopClient) set(theme string) error {
	configLine := "color_theme = " + b.darkThemePath
	if theme == "light" {
		configLine = "color_theme = " + b.lightThemePath
	}

	n, err := ReplaceAtTag(b.btopConfigPath, configLine, "color_theme = ")
	if err != nil {
		return errors.Wrap(err, "replacing theme in btop config")
	}

	if n == 0 {
		fmt.Println("WARNING: couldn't find a color_theme in the btop config: " + b.btopConfigPath)
	}

	return nil
}

// color_theme = "/Users/brian/.config/btop/themes/everforest-light-medium.theme"
