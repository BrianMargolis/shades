package main

import (
	"fmt"

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
	filter := client.Filter{}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-l", "--light":
			filter.OnlyLight = true
		case "-d", "--dark":
			filter.OnlyDark = true
		case "-f", "--favorites":
			filter.OnlyFavorites = true
		default:
			fatalUsage("unknown flag %q for gallery\n\nGALLERY FLAGS\n%s", args[i], galleryFlags)
		}
	}
	if filter.OnlyLight && filter.OnlyDark {
		fatalUsage("cannot specify both --light and --dark")
	}

	entries := galleryEntries(config, filter)
	if len(entries) == 0 {
		scope := "themes"
		if filter.OnlyDark {
			scope = "dark variants"
		} else if filter.OnlyLight {
			scope = "light variants"
		}
		if filter.OnlyFavorites {
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

func galleryEntries(config client.ConfigModel, filter client.Filter) []galleryEntry {
	entries := []galleryEntry{}
	for _, name := range config.Themes.Names(filter) {
		variant, err := config.Themes.GetVariant(name)
		if err != nil {
			continue
		}
		entries = append(entries, galleryEntry{label: name, variant: variant})
	}

	return entries
}
