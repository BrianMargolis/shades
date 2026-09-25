package picker

import (
	"testing"

	"github.com/brianmargolis/shades/client"
)

func TestPromptRoundTrips(t *testing.T) {
	filters := []client.Filter{
		{},
		{OnlyDark: true},
		{OnlyLight: true},
		{OnlyFavorites: true},
		{OnlyDark: true, OnlyFavorites: true},
		{OnlyLight: true, OnlyFavorites: true},
	}

	for _, filter := range filters {
		if got := parsePrompt(prompt(filter)); got != filter {
			t.Errorf("parsePrompt(%q) = %+v, want %+v", prompt(filter), got, filter)
		}
	}
}

func TestFilterTransform(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
		action string
		want   string
	}{
		{
			name:   "dark from everything",
			prompt: "> ",
			action: FilterDark,
			want:   "unbind(one)+change-prompt(dark > )+reload(shades _picker-list --dark)",
		},
		{
			name:   "light replaces dark and keeps favorites",
			prompt: "dark ★ > ",
			action: FilterLight,
			want:   "unbind(one)+change-prompt(light ★ > )+reload(shades _picker-list --light --favorites)",
		},
		{
			name:   "favorites toggles on",
			prompt: "dark > ",
			action: FilterFavorites,
			want:   "unbind(one)+change-prompt(dark ★ > )+reload(shades _picker-list --dark --favorites)",
		},
		{
			name:   "favorites toggles off",
			prompt: "★ > ",
			action: FilterFavorites,
			want:   "unbind(one)+change-prompt(> )+reload(shades _picker-list)",
		},
		{
			name:   "all clears everything",
			prompt: "light ★ > ",
			action: FilterAll,
			want:   "unbind(one)+change-prompt(> )+reload(shades _picker-list)",
		},
		{
			name:   "refresh keeps the filter",
			prompt: "dark ★ > ",
			action: FilterRefresh,
			want:   "unbind(one)+change-prompt(dark ★ > )+reload(shades _picker-list --dark --favorites)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := FilterTransform(test.prompt, test.action)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("FilterTransform() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestFilterTransformUnknownAction(t *testing.T) {
	if _, err := FilterTransform("> ", "sideways"); err == nil {
		t.Error("FilterTransform() with an unknown action returned no error")
	}
}
