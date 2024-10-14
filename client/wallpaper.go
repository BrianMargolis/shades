package client

type MacWallpaperClient struct{}

func NewMacWallpaperClient() Client {
	return MacWallpaperClient{}
}

func (m MacWallpaperClient) Start(socketName string) error {
	return SubscribeToSocket(m.set)(socketName)
}

func (m MacWallpaperClient) set(theme ThemeVariant) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	dark := config.Client["mac-wallpaper"]["dark"]
	light := config.Client["mac-wallpaper"]["light"]
	wallpaperPath := dark
	if theme.Light {
		wallpaperPath = light
	}

	wallpaperPath = ExpandTilde(wallpaperPath)

	script := `tell application "Finder" to set desktop picture to POSIX file "` + wallpaperPath + `"`
	_, err = RunApplescript(script)
	return err
}
