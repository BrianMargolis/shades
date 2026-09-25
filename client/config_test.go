package client

import "testing"

func TestGetSocketPath(t *testing.T) {
	if got := (ConfigModel{}).GetSocketPath(); got != defaultSocketPath {
		t.Errorf("GetSocketPath() with nothing set = %q, want %q", got, defaultSocketPath)
	}

	config := ConfigModel{SocketPath: "/tmp/elsewhere.sock"}
	if got := config.GetSocketPath(); got != "/tmp/elsewhere.sock" {
		t.Errorf("GetSocketPath() = %q, want the configured path", got)
	}
}
