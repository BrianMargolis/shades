package main

import (
	"fmt"
	"slices"

	"github.com/brianmargolis/shades/client"
	"github.com/brianmargolis/shades/preview"
)

const galleryFlags = `  -d, --dark        Only dark variants
  -l, --light       Only light variants
  -f, --favorites   Only variants marked favorite: true`

type galleryEntry struct {
	label   string
	variant client.ThemeVariant
}

// runGallery prints every theme in the config as one row of swatches each.
func runGallery(config client.ConfigModel, args []string) {
	onlyLight := false
	onlyDark := false
	onlyFavorites := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-l", "--light":
			onlyLight = true
		case "-d", "--dark":
			onlyDark = true
		case "-f", "--favorites":
			onlyFavorites = true
		default:
			fatalUsage("unknown flag %q for gallery\n\nGALLERY FLAGS\n%s", args[i], galleryFlags)
		}
	}
	if onlyLight && onlyDark {
		fatalUsage("cannot specify both --light and --dark")
	}

	entries := galleryEntries(config, onlyDark, onlyLight, onlyFavorites)
	if len(entries) == 0 {
		scope := "themes"
		if onlyDark {
			scope = "dark variants"
		} else if onlyLight {
			scope = "light variants"
		}
		if onlyFavorites {
			fmt.Printf("No favorite %s in your config. Mark a variant with 'favorite: true' to add it.\n", scope)
		} else {
			fmt.Printf("No %s in your config.\n", scope)
		}
		return
	}

	labelWidth := 0
	for _, entry := range entries {
		labelWidth = max(labelWidth, len(entry.label))
	}

	previewer := preview.NewPreviewer()
	for _, entry := range entries {
		fmt.Printf("%-*s  %s\n", labelWidth, entry.label, previewer.Swatches(entry.variant))
	}
}

// galleryEntries collects the variants to render, sorted by theme and then by
// variant. Sorting isn't cosmetic: ranging over the config's maps directly
// would reshuffle the gallery on every run.
func galleryEntries(config client.ConfigModel, onlyDark, onlyLight, onlyFavorites bool) []galleryEntry {
	themeNames := make([]string, 0, len(config.Themes))
	for themeName := range config.Themes {
		themeNames = append(themeNames, themeName)
	}
	slices.Sort(themeNames)

	entries := []galleryEntry{}
	for _, themeName := range themeNames {
		theme := config.Themes[themeName]

		variantNames := make([]string, 0, len(theme.Variants))
		for variantName := range theme.Variants {
			variantNames = append(variantNames, variantName)
		}
		slices.Sort(variantNames)

		for _, variantName := range variantNames {
			variant := theme.Variants[variantName]
			if onlyLight && !variant.Light {
				continue
			}
			if onlyDark && variant.Light {
				continue
			}
			if onlyFavorites && !variant.Favorite {
				continue
			}

			variant.ThemeName = themeName
			variant.VariantName = variantName
			entries = append(entries, galleryEntry{
				label:   fmt.Sprintf("%s;%s", themeName, variantName),
				variant: variant,
			})
		}
	}

	return entries
}
