package main

import (
	"strings"
	"testing"
)

// Fails when a cobra upgrade changes the line patchFishCompletion rewrites, so
// the fix can't silently stop applying.
func TestFishScriptEscapesEarlierWords(t *testing.T) {
	script, err := fishScript(newRootCommand(), true)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(script, fishEscapedArgsLine) {
		t.Errorf("fish script is missing %q", fishEscapedArgsLine)
	}
	if strings.Contains(script, fishArgsLine) {
		t.Errorf("fish script still has the unescaped %q", fishArgsLine)
	}
}
