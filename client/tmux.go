package client

import (
	"fmt"
	"os/exec"
)

type TMUXClient struct{}

func (t TMUXClient) Start(socket string, config map[string]string) error {
	return SubscribeToSocket(t.set)(socket)
}

func (t TMUXClient) set(theme string) error {
	statusBackground := "#2d353b"
	statusForeground := "#d3c6aa"
	windowStatusBackground := "#2d353b"
	windowStatusForeground := "#a7c080"
	if theme == "light" {
		statusBackground = "#fdf6e3"
		statusForeground = "#5c6a72"
		windowStatusBackground = "#fdf6e3"
		windowStatusForeground = "#8da101"
	}

	t.setTMUXOption("status-bg", statusBackground)
	t.setTMUXOption("status-fg", statusForeground)
	t.setTMUXOption("window-status-format",
		"#{?#{==:#{session_windows},1},,"+
			"#[fg="+windowStatusBackground+",bg="+windowStatusForeground+"]"+
			" "+
			"#W }",
	)
	t.setTMUXOption("window-status-current-format",
		"#{?#{==:#{session_windows},1},,"+
			"#[fg="+windowStatusBackground+",bg="+windowStatusForeground+"]"+
			" *"+
			"#W* }",
	)
	t.setTMUXOption("status-left",
		"#[fg="+statusBackground+",bg="+windowStatusForeground+"]"+
			" #(cd #{pane_current_path}; pwd)"+
			" "+
			"#[fg="+windowStatusForeground+",bg="+statusBackground+"]",
	)
	t.setTMUXOption("status-right",
		"#[fg="+statusBackground+",bg="+windowStatusForeground+"]"+
			" "+
			" #(cd #{pane_current_path}; git branch --show-current) ",
	)

	return nil
}

func (t TMUXClient) setTMUXOption(optionName, value string) error {
	fmt.Printf("%s: %s\n", optionName, value)
	_, err := exec.Command("/usr/local/bin/tmux", "set-option", "-g", optionName, value).Output()
	if err != nil {
		fmt.Printf("ERROR setting %s: %s", optionName, err.Error())
	}
	return err
}
