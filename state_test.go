package main

import (
	"testing"

	"github.com/brianmargolis/shades/client"
)

func TestDescribeTheme(t *testing.T) {
	config := client.ConfigModel{
		DefaultDarkTheme:  "everforest;dark-medium",
		DefaultLightTheme: "everforest;light-soft",
	}

	tests := []struct {
		name    string
		theme   string
		variant client.ThemeVariant
		want    string
	}{
		{
			name:  "plain theme has no notes",
			theme: "gruvbox;dark",
			want:  "gruvbox;dark",
		},
		{
			name:    "favorite",
			theme:   "gruvbox;dark",
			variant: client.ThemeVariant{Favorite: true},
			want:    "gruvbox;dark (favorite)",
		},
		{
			name:    "favorite default",
			theme:   "everforest;dark-medium",
			variant: client.ThemeVariant{Favorite: true},
			want:    "everforest;dark-medium (favorite, default dark)",
		},
		{
			name:    "light default",
			theme:   "everforest;light-soft",
			variant: client.ThemeVariant{Light: true},
			want:    "everforest;light-soft (default light)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := describeTheme(config, test.theme, test.variant)
			if got != test.want {
				t.Errorf("describeTheme() = %q, want %q", got, test.want)
			}
		})
	}
}
