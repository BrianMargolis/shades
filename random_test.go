package main

import (
	"slices"
	"testing"
)

func TestOtherThemes(t *testing.T) {
	tests := []struct {
		name       string
		candidates []string
		current    string
		want       []string
	}{
		{
			name:       "drops the current theme",
			candidates: []string{"everforest;dark-medium", "gruvbox;dark"},
			current:    "gruvbox;dark",
			want:       []string{"everforest;dark-medium"},
		},
		{
			name:       "keeps everything when the current theme isn't a candidate",
			candidates: []string{"everforest;dark-medium", "gruvbox;dark"},
			current:    "nord;dark",
			want:       []string{"everforest;dark-medium", "gruvbox;dark"},
		},
		{
			name:       "keeps a lone candidate even if it's current",
			candidates: []string{"gruvbox;dark"},
			current:    "gruvbox;dark",
			want:       []string{"gruvbox;dark"},
		},
		{
			name:       "no current theme",
			candidates: []string{"gruvbox;dark"},
			want:       []string{"gruvbox;dark"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := otherThemes(test.candidates, test.current)
			if !slices.Equal(got, test.want) {
				t.Errorf("otherThemes() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestPickRandomStaysInRange(t *testing.T) {
	candidates := []string{"everforest;dark-medium", "everforest;light-soft", "gruvbox;dark"}

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
