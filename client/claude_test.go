package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLiveThemeFor(t *testing.T) {
	themesDirectory := t.TempDir()
	custom := `{"name": "light-ansi-custom", "base": "light-ansi", "overrides": {"promptBorder": "ansi:yellowBright"}}`
	if err := os.WriteFile(filepath.Join(themesDirectory, "light-ansi-custom.json"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		themeStr      string
		wantBase      string
		wantOverrides bool
	}{
		{themeStr: "custom:light-ansi-custom", wantBase: "light-ansi", wantOverrides: true},
		{themeStr: "dark", wantBase: "dark"},
	}
	for _, test := range tests {
		t.Run(test.themeStr, func(t *testing.T) {
			content, err := liveThemeFor(test.themeStr, themesDirectory)
			if err != nil {
				t.Fatal(err)
			}
			var theme map[string]any
			if err := json.Unmarshal(content, &theme); err != nil {
				t.Fatal(err)
			}
			if theme["name"] != liveThemeName {
				t.Errorf("name = %v, want %q", theme["name"], liveThemeName)
			}
			if theme["base"] != test.wantBase {
				t.Errorf("base = %v, want %q", theme["base"], test.wantBase)
			}
			if _, ok := theme["overrides"]; ok != test.wantOverrides {
				t.Errorf("has overrides = %v, want %v", ok, test.wantOverrides)
			}
		})
	}

	if _, err := liveThemeFor("custom:missing", themesDirectory); err == nil {
		t.Error("expected an error for a custom theme with no file")
	}
}
