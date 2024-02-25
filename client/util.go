package client

import "os/exec"

// RunApplescript runs an AppleScript command and returns the output.
func RunApplescript(script string) ([]byte, error) {
	return exec.Command("osascript", "-e", script).Output()
}
