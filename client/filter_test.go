package client

import (
	"slices"
	"testing"
)

func filterTestThemes() Themes {
	return Themes{
		"everforest": {
			Variants: map[string]ThemeVariant{
				"dark-medium": {Light: false, Favorite: true},
				"light-soft":  {Light: true, Favorite: true},
			},
		},
		"gruvbox": {
			Variants: map[string]ThemeVariant{
				"dark": {Light: false},
			},
		},
	}
}

func TestThemesNames(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		want   []string
	}{
		{
			name: "no filter returns every variant, sorted",
			want: []string{"everforest;dark-medium", "everforest;light-soft", "gruvbox;dark"},
		},
		{
			name:   "dark filter drops light variants",
			filter: Filter{OnlyDark: true},
			want:   []string{"everforest;dark-medium", "gruvbox;dark"},
		},
		{
			name:   "light filter drops dark variants",
			filter: Filter{OnlyLight: true},
			want:   []string{"everforest;light-soft"},
		},
		{
			name:   "favorites filter drops non-favorites",
			filter: Filter{OnlyFavorites: true},
			want:   []string{"everforest;dark-medium", "everforest;light-soft"},
		},
		{
			name:   "favorites filter combines with dark filter",
			filter: Filter{OnlyDark: true, OnlyFavorites: true},
			want:   []string{"everforest;dark-medium"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := filterTestThemes().Names(test.filter)
			if !slices.Equal(got, test.want) {
				t.Errorf("Names() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestThemesNamesEmpty(t *testing.T) {
	got := Themes{}.Names(Filter{})
	if len(got) != 0 {
		t.Errorf("Names() = %v, want no names", got)
	}
}

func TestFilterFlags(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		want   []string
	}{
		{name: "zero filter has no flags", want: []string{}},
		{name: "dark", filter: Filter{OnlyDark: true}, want: []string{"--dark"}},
		{
			name:   "light and favorites",
			filter: Filter{OnlyLight: true, OnlyFavorites: true},
			want:   []string{"--light", "--favorites"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.filter.Flags()
			if !slices.Equal(got, test.want) {
				t.Errorf("Flags() = %v, want %v", got, test.want)
			}
		})
	}
}
