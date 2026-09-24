package main

import (
	"slices"
	"testing"

	"github.com/brianmargolis/shades/client"
)

func testConfig() client.ConfigModel {
	return client.ConfigModel{
		Themes: client.Themes{
			"everforest": {
				Variants: map[string]client.ThemeVariant{
					"dark-medium": {Light: false, Favorite: true},
					"light-soft":  {Light: true, Favorite: true},
				},
			},
			"gruvbox": {
				Variants: map[string]client.ThemeVariant{
					"dark": {Light: false},
				},
			},
		},
	}
}

func TestRandomCandidates(t *testing.T) {
	tests := []struct {
		name          string
		onlyDark      bool
		onlyLight     bool
		onlyFavorites bool
		want          []string
	}{
		{
			name: "no filter returns every variant, sorted",
			want: []string{"everforest;dark-medium", "everforest;light-soft", "gruvbox;dark"},
		},
		{
			name:     "dark filter drops light variants",
			onlyDark: true,
			want:     []string{"everforest;dark-medium", "gruvbox;dark"},
		},
		{
			name:      "light filter drops dark variants",
			onlyLight: true,
			want:      []string{"everforest;light-soft"},
		},
		{
			name:          "favorites filter drops non-favorites",
			onlyFavorites: true,
			want:          []string{"everforest;dark-medium", "everforest;light-soft"},
		},
		{
			name:          "favorites filter combines with dark filter",
			onlyDark:      true,
			onlyFavorites: true,
			want:          []string{"everforest;dark-medium"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := randomCandidates(testConfig(), test.onlyDark, test.onlyLight, test.onlyFavorites)
			if !slices.Equal(got, test.want) {
				t.Errorf("randomCandidates() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRandomCandidatesEmptyConfig(t *testing.T) {
	got := randomCandidates(client.ConfigModel{}, false, false, false)
	if len(got) != 0 {
		t.Errorf("randomCandidates() = %v, want no candidates", got)
	}
}

func TestPickRandomStaysInRange(t *testing.T) {
	candidates := randomCandidates(testConfig(), false, false, false)

	seen := map[string]bool{}
	for range 200 {
		choice := pickRandom(candidates)
		if !slices.Contains(candidates, choice) {
			t.Fatalf("pickRandom() = %q, not one of %v", choice, candidates)
		}
		seen[choice] = true
	}

	if len(seen) != len(candidates) {
		t.Errorf("pickRandom() only ever returned %d of %d candidates", len(seen), len(candidates))
	}
}
