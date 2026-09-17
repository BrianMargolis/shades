package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/brianmargolis/shades/client"

	"go.uber.org/zap"
)

const randomFlags = `  -d, --dark    Only dark variants
  -l, --light   Only light variants`

// runRandom picks one theme variant at random and switches to it, printing the
// pick so the user can name the theme they just landed on.
func runRandom(ctx context.Context, config client.ConfigModel, args []string) {
	logger := zap.S()

	onlyLight := false
	onlyDark := false

	for _, arg := range args {
		switch arg {
		case "-l", "--light":
			onlyLight = true
		case "-d", "--dark":
			onlyDark = true
		default:
			fatalUsage("unknown flag %q for random\n\nRANDOM FLAGS\n%s", arg, randomFlags)
		}
	}
	if onlyLight && onlyDark {
		fatalUsage("cannot specify both --light and --dark")
	}

	candidates := randomCandidates(config, onlyDark, onlyLight)
	logger.Debugw("random candidates", "count", len(candidates), "onlyDark", onlyDark, "onlyLight", onlyLight)

	if len(candidates) == 0 {
		scope := "theme variants"
		if onlyDark {
			scope = "dark theme variants"
		} else if onlyLight {
			scope = "light theme variants"
		}
		fatalUsage("no %s in your config\n\nRun 'shades -l' to list the themes in your config.", scope)
	}

	choice := pickRandom(candidates)
	logger.Debugw("random pick", "theme", choice)
	fmt.Println(choice)

	if err := (client.ChangerClient{Theme: choice}).Start(ctx, socketPath); err != nil {
		logger.Fatalw("failed to set theme", "theme", choice, "error", err)
	}
}

// randomCandidates is a deliberate copy of picker.getOptions rather than a
// shared helper, to keep this command from conflicting with the gallery
// command, which filters the same way. Consolidating the copies is a follow-up.
//
// The result is sorted so the index rand.IntN produces maps to a stable
// variant, independent of Go's randomized map iteration order.
func randomCandidates(config client.ConfigModel, onlyDark, onlyLight bool) []string {
	candidates := []string{}
	for themeName, theme := range config.Themes {
		for variantName, variant := range theme.Variants {
			if onlyLight && !variant.Light {
				continue
			}
			if onlyDark && variant.Light {
				continue
			}
			candidates = append(candidates, fmt.Sprintf("%s;%s", themeName, variantName))
		}
	}
	slices.Sort(candidates)

	return candidates
}

func pickRandom(candidates []string) string {
	return candidates[rand.IntN(len(candidates))]
}
