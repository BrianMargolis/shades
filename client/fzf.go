package client

import (
	"fmt"
	"os/exec"
)

type FZFClient struct{}

func (b FZFClient) Start(socketName string, config map[string]string) error {
	return SubscribeToSocket(b.set)(socketName)
}

func (b FZFClient) set(theme string) error {
	fmt.Printf("FZF SET to %s\n", theme)
	// generate these with https://vitormv.github.io/fzf-themes/
	fzfTheme := `--color=fg:#d3c6aa,fg+:#d0d0d0,bg:#232a2e,bg+:#232a2e --color=hl:#425047,hl+:#5fd7ff,info:#dbbc7f,marker:#d699b6 --color=prompt:#7fbbb3,spinner:#a7c080,pointer:#a7c080,header:#e69875 --color=border:#232a2e,label:#9da9a0,query:#d3c6aa --border="rounded" --border-label="" --preview-window="border-rounded" --prompt="> " --marker=">" --pointer="◆" --separator="─" --scrollbar="│"`
	if theme == "light" {
		fzfTheme = `--color=fg:#5c6a72,fg+:#5c6a72,bg:#fdf6e3,bg+:#fdf6e3 --color=hl:#5f87af,hl+:#5fd7ff,info:#dfa000,marker:#df69ba --color=prompt:#3a94c5,spinner:#8da101,pointer:#8da101,header:#f57d25 --color=border:#fdf6e3,label:#aeaeae,query:#d9d9d9 --border="rounded" --border-label="" --preview-window="border-rounded" --prompt="> " --marker=">" --pointer="◆" --separator="─" --scrollbar="│"`
	}

	// do our best
	if err := b.destroyTheGlobe(); err != nil {
		fmt.Println("failed to destroy the globe: " + err.Error())
	}

	cmd := fmt.Sprintf("set -Ux FZF_DEFAULT_OPTS '%s'", fzfTheme)
	fmt.Println(cmd)
	err := exec.Command("fish", "-c", cmd).Run()
	if err != nil {
		fmt.Println("failed to set the theme: " + err.Error())
	}

	cmd = fmt.Sprintf("set -Ux _ZO_FZF_OPTS '%s'", fzfTheme)
	fmt.Println(cmd)
	err = exec.Command("fish", "-c", cmd).Run()
	if err != nil {
		fmt.Println("failed to set the zoxide theme: " + err.Error())
	}
	return nil
}

// something is trying to make a global FZF_DEFAULT_OPTS, and that will shadow our universal value
func (b FZFClient) destroyTheGlobe() error {
	cmd := "set --erase --global FZF_DEFAULT_OPTS"
	return exec.Command("fish", "-c", cmd).Run()
}
