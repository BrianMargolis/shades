package picker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/brianmargolis/shades/client"

	"go.uber.org/zap"
)

func testConfig() client.ConfigModel {
	return client.ConfigModel{
		Themes: client.Themes{
			"gruvbox": {
				Variants: map[string]client.ThemeVariant{
					"dark": {Light: false},
				},
			},
			"everforest": {
				Variants: map[string]client.ThemeVariant{
					"light-soft":  {Light: true},
					"dark-medium": {Light: false, Favorite: true},
				},
			},
		},
	}
}

// fakeFzf puts an fzf on PATH that prints output and exits with status, and
// points the config at a one-theme file.
func fakeFzf(t *testing.T, output string, status int) {
	t.Helper()
	directory := t.TempDir()

	script := fmt.Sprintf("#!/bin/sh\ncat >/dev/null\nprintf '%%s' '%s'\nexit %d\n", output, status)
	if err := os.WriteFile(filepath.Join(directory, "fzf"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))

	config := "themes:\n  gruvbox:\n    variants:\n      dark:\n        light: false\n"
	configPath := filepath.Join(directory, "shades.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SHADES_CONFIG", configPath)
	t.Setenv("SHADES_STATE", filepath.Join(directory, "state.yaml"))
}

func TestPickExitStatus(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		wantResult string
		wantErr    error
	}{
		{name: "a pick is returned without the trailing newline", status: 0, wantResult: "gruvbox;dark"},
		{name: "esc cancels", status: 130, wantErr: ErrCancelled},
		{name: "no match cancels", status: 1, wantErr: ErrCancelled},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeFzf(t, "gruvbox;dark\n", test.status)

			result, err := (&picker{}).pick(zap.S(), PickerOpts{})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("pick() error = %v, want %v", err, test.wantErr)
			}
			if result != test.wantResult {
				t.Errorf("pick() = %q, want %q", result, test.wantResult)
			}
		})
	}
}

func TestPickFailure(t *testing.T) {
	fakeFzf(t, "", 2)

	_, err := (&picker{}).pick(zap.S(), PickerOpts{})
	if err == nil || errors.Is(err, ErrCancelled) {
		t.Fatalf("pick() error = %v, want a failure that isn't a cancel", err)
	}
}

func TestLines(t *testing.T) {
	tests := []struct {
		name   string
		filter client.Filter
		want   []string
	}{
		{
			name: "every variant, sorted, with favorites starred",
			want: []string{
				"everforest;dark-medium\t" + favoriteMarker + " everforest;dark-medium",
				"everforest;light-soft\t  everforest;light-soft",
				"gruvbox;dark\t  gruvbox;dark",
			},
		},
		{
			name:   "filtered to favorites",
			filter: client.Filter{OnlyFavorites: true},
			want: []string{
				"everforest;dark-medium\t" + favoriteMarker + " everforest;dark-medium",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Lines(testConfig(), test.filter)
			if !slices.Equal(got, test.want) {
				t.Errorf("Lines() = %q, want %q", got, test.want)
			}
		})
	}
}
