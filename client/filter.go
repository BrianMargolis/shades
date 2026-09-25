package client

import (
	"fmt"
	"slices"
)

// Filter narrows the theme variants a command works over. The zero value
// matches every variant.
type Filter struct {
	OnlyDark      bool
	OnlyLight     bool
	OnlyFavorites bool
}

func (f Filter) Matches(variant ThemeVariant) bool {
	if f.OnlyLight && !variant.Light {
		return false
	}
	if f.OnlyDark && variant.Light {
		return false
	}
	if f.OnlyFavorites && !variant.Favorite {
		return false
	}
	return true
}

// Flags turns the filter back into the command line flags that select it, for
// commands that shell back out to shades with the same filter.
func (f Filter) Flags() []string {
	flags := []string{}
	if f.OnlyDark {
		flags = append(flags, "--dark")
	}
	if f.OnlyLight {
		flags = append(flags, "--light")
	}
	if f.OnlyFavorites {
		flags = append(flags, "--favorites")
	}
	return flags
}

// Names lists the "theme;variant" names the filter matches. The result is
// sorted, since ranging over the config's maps directly would come out in a
// different order on every run.
func (t Themes) Names(filter Filter) []string {
	names := []string{}
	for themeName, theme := range t {
		for variantName, variant := range theme.Variants {
			if filter.Matches(variant) {
				names = append(names, fmt.Sprintf("%s;%s", themeName, variantName))
			}
		}
	}
	slices.Sort(names)
	return names
}
