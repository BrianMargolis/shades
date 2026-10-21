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
		for _, order := range Orders {
			gotFilter, gotOrder := parsePrompt(prompt(filter, order))
			if gotFilter != filter || gotOrder != order {
				t.Errorf(
					"parsePrompt(%q) = %+v, %s, want %+v, %s",
					prompt(filter, order), gotFilter, gotOrder, filter, order,
				)
			}
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
		{
			name:   "order goes from name to dark first",
			prompt: "dark > ",
			action: FilterOrder,
			want:   "unbind(one)+change-prompt(dark dark→light > )+reload(shades _picker-list --dark --order dark-first)",
		},
		{
			name:   "order goes from dark first to light first",
			prompt: "dark→light > ",
			action: FilterOrder,
			want:   "unbind(one)+change-prompt(light→dark > )+reload(shades _picker-list --order light-first)",
		},
		{
			name:   "order goes from light first back to name",
			prompt: "★ light→dark > ",
			action: FilterOrder,
			want:   "unbind(one)+change-prompt(★ > )+reload(shades _picker-list --favorites)",
		},
		{
			name:   "all clears the filters but keeps the order",
			prompt: "light ★ dark→light > ",
			action: FilterAll,
			want:   "unbind(one)+change-prompt(dark→light > )+reload(shades _picker-list --order dark-first)",
		},
		{
			name:   "refresh keeps the order",
			prompt: "dark light→dark > ",
			action: FilterRefresh,
			want:   "unbind(one)+change-prompt(dark light→dark > )+reload(shades _picker-list --dark --order light-first)",
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
