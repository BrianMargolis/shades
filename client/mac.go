package client

type MacClient struct{}

func (m MacClient) Start(socket string, config map[string]string) error {
	return SubscribeToSocket(m.set)(socket)
}

func (m MacClient) set(theme string) error {
	shouldBeDark := "true"
	if theme == "light" {
		shouldBeDark = "false"
	}

	// the single line of AppleScript that I know:
	script := `tell application "System Events" to tell appearance preferences to set dark mode to ` + shouldBeDark

	_, err := RunApplescript(script)
	return err
}
