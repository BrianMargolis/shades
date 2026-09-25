package picker

import (
	"slices"
	"testing"

	"github.com/brianmargolis/shades/client"
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
