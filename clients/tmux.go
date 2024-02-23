package client

import (
	"brianmargolis/theme-daemon/protocol"
	"fmt"
	"os/exec"
	"strings"
)

type TMUXClient struct{}

func (t TMUXClient) Start(socket string) error {
	read, write, err := SocketAsChannel(socket)
	if err != nil {
		return err
	}

	write <- string(protocol.Subscribe("tmux"))

	for message := range read {
		if !strings.HasPrefix(message, "set") {
			continue
		}

		parts := strings.Split(message, ":")
		if len(parts) < 2 {
			fmt.Printf("ERROR: malformed message: %s", message)
		}

		theme := strings.TrimSpace(parts[1])

		status_bg := "#2d353b"
		status_fg := "#d3c6aa"
		window_status_bg := "#2d353b"
		window_status_fg := "#a7c080"
		if theme == "light" {
			status_bg = "#fdf6e3"
			status_fg = "#5c6a72"
			window_status_bg = "#fdf6e3"
			window_status_fg = "#8da101"
		}

		t.setTMUXOption("status-bg", status_bg)
		t.setTMUXOption("status-fg", status_fg)
		t.setTMUXOption("window-status-format",
			"#{?#{==:#{session_windows},1},,"+
				"#[fg="+window_status_bg+",bg="+window_status_fg+"]"+
				" "+
				"#W }",
		)
		t.setTMUXOption("window-status-current-format",
			"#{?#{==:#{session_windows},1},,"+
				"#[fg="+window_status_bg+",bg="+window_status_fg+"]"+
				" *"+
				"#W* }",
		)
		t.setTMUXOption("status-left",
			"#[fg="+status_bg+",bg="+window_status_fg+"]"+
				" #(cd #{pane_current_path}; pwd)"+
				" "+
				"#[fg="+window_status_fg+",bg="+status_bg+"]",
		)
		t.setTMUXOption("status-right",
			"#[fg="+status_bg+",bg="+window_status_fg+"]"+
				" "+
				" #(cd #{pane_current_path}; git branch --show-current) ",
		)
	}

	return nil
}

func (t TMUXClient) setTMUXOption(optionName, value string) error {
	_, err := exec.Command("/usr/local/bin/tmux", "set-option", "-g", optionName, value).Output()
	if err != nil {
		fmt.Printf("ERROR setting %s: %s", optionName, err.Error())
	}
	return err
}
