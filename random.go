package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"

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

	current, err := client.CurrentTheme(config.GetSocketPath())
	if err != nil {
		logger.Debugw("could not get the current theme, so it may be picked again", "error", err)
	}

	choice := pickRandom(otherThemes(candidates, current))
	logger.Debugw("random pick", "theme", choice)
	fmt.Println(choice)

	return errors.Wrapf(
		client.ChangerClient{Theme: choice}.Start(ctx, config.GetSocketPath()),
		"failed to set theme %s", choice,
	)
}

// otherThemes drops the current theme so random always changes something. A
// lone candidate stays, since there's nothing else to switch to.
func otherThemes(candidates []string, current string) []string {
	others := slices.DeleteFunc(slices.Clone(candidates), func(name string) bool {
		return name == current
	})
	if len(others) == 0 {
		return candidates
	}
	return others
}

func pickRandom(candidates []string) string {
	return candidates[rand.IntN(len(candidates))]
}
