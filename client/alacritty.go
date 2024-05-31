package client

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

type AlacrittyClient struct {
	config map[string]string
}

func NewAlacrittyClient(config map[string]string) Client {
	return AlacrittyClient{config: config}
}

func (a AlacrittyClient) Start(socket string) error {
	return SubscribeToSocket(a.set)(socket)
}

func (a AlacrittyClient) set(theme ThemeVariant) error {
	alacrittyConfigPath := ExpandTilde(a.config["alacritty-config-path"])
	if alacrittyConfigPath == "" {
		alacrittyConfigPath = os.Getenv("HOME") + "/.config/alacritty/alacritty.toml"
	}

	themePath := fmt.Sprintf("%s/%s-%s.toml", a.config["theme-path"], theme.ThemeName, theme.VariantName)
	n, err := ReplaceAtTag(
		alacrittyConfigPath,
		fmt.Sprintf("\"%s\", # shades-replace", themePath),
		"# shades-replace",
	)
	if err != nil {
		return errors.Wrap(err, "replacing alacritty theme path")
	}

	if n == 0 {
		fmt.Println("WARNING: no '# shades-replace' tag found in alacritty config: " + alacrittyConfigPath)
	}

	return nil
}
