package main

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/brianmargolis/shades/client"

	"go.uber.org/zap"
)

const randomFlags = `  -d, --dark        Only dark variants
  -l, --light       Only light variants
  -f, --favorites   Only variants marked favorite: true`

// runRandom picks one theme variant at random and switches to it, printing the
// pick so the user can name the theme they just landed on.
func runRandom(ctx context.Context, config client.ConfigModel, args []string) {
	logger := zap.S()

	filter := client.Filter{}
	for _, arg := range args {
		switch arg {
		case "-l", "--light":
			filter.OnlyLight = true
		case "-d", "--dark":
			filter.OnlyDark = true
		case "-f", "--favorites":
			filter.OnlyFavorites = true
		default:
			fatalUsage("unknown flag %q for random\n\nRANDOM FLAGS\n%s", arg, randomFlags)
		}
	}
	if filter.OnlyLight && filter.OnlyDark {
		fatalUsage("cannot specify both --light and --dark")
	}

	candidates := config.Themes.Names(filter)
	logger.Debugw("random candidates", "count", len(candidates), "filter", filter)

	if len(candidates) == 0 {
		scope := "theme variants"
		if filter.OnlyDark {
			scope = "dark theme variants"
		} else if filter.OnlyLight {
			scope = "light theme variants"
		}
		if filter.OnlyFavorites {
			fatalUsage("no favorite %s in your config\n\nMark a variant with 'favorite: true' to add it.", scope)
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

func pickRandom(candidates []string) string {
	return candidates[rand.IntN(len(candidates))]
}
