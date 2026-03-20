package client

import "go.uber.org/zap"

type MacWallpaperClient struct{}

func NewMacWallpaperClient() Client {
	return MacWallpaperClient{}
}

func (m MacWallpaperClient) Start(socketName string) error {
	return SubscribeToSocket(SetterWithContext(m.set, "mac-wallpaper"))(socketName)
}

func (m MacWallpaperClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	wallpaperPath := ExpandTilde(config.Client["mac-wallpaper"]["dark"])
	if theme.Light {
		wallpaperPath = ExpandTilde(config.Client["mac-wallpaper"]["light"])
	}

	zap.S().Debugw("applying theme", "client", "mac-wallpaper", "theme", theme.ThemeName, "variant", theme.VariantName, "wallpaperPath", wallpaperPath)

	script := `tell application "Finder" to set desktop picture to POSIX file "` + wallpaperPath + `"`
	_, err = RunApplescript(script)
	return err
}
