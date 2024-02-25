package client

import "fmt"

type MacWallpaperClient struct {
}

func (m MacWallpaperClient) Start(socketName string, config map[string]string) error {
	return SubscribeToSocket(m.makeSetter(config))(socketName)
}

func (m MacWallpaperClient) makeSetter(config map[string]string) func(string) error {
	return func(theme string) error {
		wallpaperPath := config["dark"]
		if theme == "light" {
			wallpaperPath = config["light"]
		}

		script := `tell application "Finder" to set desktop picture to POSIX file "` + wallpaperPath + `"`
		fmt.Println(script)
		_, err := RunApplescript(script)
		return err
	}
}
