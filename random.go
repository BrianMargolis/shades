package main

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/brianmargolis/shades/client"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// runRandom picks one theme variant at random and switches to it, printing the
// pick so the user can name the theme they just landed on.
func runRandom(ctx context.Context, config client.ConfigModel, filter client.Filter) error {
	logger := zap.S()

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
			return errors.Errorf("no favorite %s in your config; mark a variant with 'favorite: true' to add it", scope)
		}
		return errors.Errorf("no %s in your config; run 'shades list' to see the themes in it", scope)
	}

	choice := pickRandom(candidates)
	logger.Debugw("random pick", "theme", choice)
	fmt.Println(choice)

	return errors.Wrapf(
		client.ChangerClient{Theme: choice}.Start(ctx, socketPath),
		"failed to set theme %s", choice,
	)
}

func pickRandom(candidates []string) string {
	return candidates[rand.IntN(len(candidates))]
}
