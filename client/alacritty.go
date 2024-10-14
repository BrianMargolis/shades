package client

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

type AlacrittyClient struct{}

func NewAlacrittyClient() Client {
	return AlacrittyClient{}
}

func (a AlacrittyClient) Start(socket string) error {
	return SubscribeToSocket(a.set)(socket)
}

func (a AlacrittyClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	alacrittyConfigPath := ExpandTilde(config.Client["alacritty"]["alacritty-config-path"])
	if alacrittyConfigPath == "" {
		alacrittyConfigPath = os.Getenv("HOME") + "/.config/alacritty/alacritty.toml"
	}
	alacrittyMainConfigPath := ExpandTilde(config.Client["alacritty"]["alacritty-main-config-path"])
	fmt.Println(alacrittyMainConfigPath)
	fmt.Println(config.Client["alacritty"])

	themePath := fmt.Sprintf("%s/%s-%s.toml", config.Client["alacritty"]["theme-path"], theme.ThemeName, theme.VariantName)
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

	// touch the file to trigger a reload
	err = os.Chtimes(alacrittyMainConfigPath, time.Now(), time.Now())
	if err != nil {
		return errors.Wrap(err, "touching alacritty config")
	}

	return nil
}
