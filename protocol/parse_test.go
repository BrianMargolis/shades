package protocol

import (
	"strings"
	"testing"
)

func TestFormattersAreNewlineTerminated(t *testing.T) {
	// A message missing its delimiter leaves the next reader's ReadString('\n')
	// blocked until some later message happens to arrive, and a message with two
	// makes the reader after that see an empty line.
	messages := map[string][]byte{
		"subscribe":   Subscribe("neovim"),
		"unsubscribe": Unsubscribe(),
		"propose":     Propose("everforest;dark-medium"),
		"set":         Set("everforest;dark-medium"),
		"palette":     Palette(`{"BG0":"#2D353B"}`),
		"get":         Get(),
	}

	for name, message := range messages {
		if !strings.HasSuffix(string(message), "\n") {
			t.Errorf("%s message %q is not newline terminated", name, message)
		}
		if count := strings.Count(string(message), "\n"); count != 1 {
			t.Errorf("%s message %q has %d newlines, want 1", name, message, count)
		}
	}
}

func TestParseRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		message []byte
		verb    string
		noun    string
	}{
		{"subscribe", Subscribe("neovim"), "subscribe", "neovim"},
		{"unsubscribe", Unsubscribe(), "unsubscribe", ""},
		{"propose", Propose("everforest;dark-medium"), "propose", "everforest;dark-medium"},
		{"set", Set("everforest;dark-medium"), "set", "everforest;dark-medium"},
		{"get", Get(), "get", ""},
		{
			"palette payload contains colons of its own",
			Palette(`{"BG0":"#2D353B","FG":"#D3C6AA"}`),
			"palette",
			`{"BG0":"#2D353B","FG":"#D3C6AA"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			verb, noun, err := Parse(string(test.message))
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", test.message, err)
			}
			if verb != test.verb {
				t.Errorf("verb = %q, want %q", verb, test.verb)
			}
			// readers strip the framing newline off the noun, same as the clients do
			if got := strings.TrimSpace(noun); got != test.noun {
				t.Errorf("noun = %q, want %q", got, test.noun)
			}
		})
	}
}

func TestParseRejectsMessageWithoutVerb(t *testing.T) {
	if _, _, err := Parse("everforest;dark-medium\n"); err == nil {
		t.Error("Parse accepted a message with no verb separator")
	}
}
