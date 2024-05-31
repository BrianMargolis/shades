package client

import "fmt"

type MacWallpaperClient struct {
	light string
	dark  string
}

func NewMacWallpaperClient(config map[string]string) Client {
	return MacWallpaperClient{
		light: config["light"],
		dark:  config["dark"],
	}
}

func (m MacWallpaperClient) Start(socketName string) error {
	return SubscribeToSocket(m.set)(socketName)
}

func (m MacWallpaperClient) set(theme ThemeVariant) error {
	wallpaperPath := m.dark
	if theme.Light {
		wallpaperPath = m.light
	}

	wallpaperPath = ExpandTilde(wallpaperPath)

	script := `tell application "Finder" to set desktop picture to POSIX file "` + wallpaperPath + `"`
	fmt.Println(script)
	_, err := RunApplescript(script)
	return err
}
