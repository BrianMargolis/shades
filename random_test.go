package main

import (
	"slices"
	"testing"
)

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
