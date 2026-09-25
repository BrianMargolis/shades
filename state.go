package main

import (
	"fmt"

	"github.com/brianmargolis/shades/client"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// currentVariant resolves the theme the server says is set. The lookup against
// the config catches a server that is still holding a theme since removed from
// shades.yaml, before that theme gets written into the state file.
func currentVariant(config client.ConfigModel) (string, client.ThemeVariant, error) {
	current, err := client.CurrentTheme(socketPath)
	if err != nil {
		return "", client.ThemeVariant{}, errors.Wrap(err, "could not get the current theme (is 'shades server' running?)")
	}

	variant, err := config.Themes.GetVariant(current)
	if err != nil {
		return "", client.ThemeVariant{}, errors.Wrapf(err, "the current theme %q isn't in your config", current)
	}

	return current, variant, nil
}

// runFavorite marks or unmarks the current theme as a favorite.
func runFavorite(config client.ConfigModel, favorite bool) error {
	current, _, err := currentVariant(config)
	if err != nil {
		return err
	}

	return setFavorite(current, favorite)
}

// runToggleFavorite flips a named theme's favorite, for the picker's keybind.
// It takes the theme from fzf rather than asking the server, because the
// server only hears about the focused theme through an async set that may not
// have landed yet.
func runToggleFavorite(config client.ConfigModel, theme string) error {
	variant, err := config.Themes.GetVariant(theme)
	if err != nil {
		return errors.Wrapf(err, "%q isn't in your config", theme)
	}

	return setFavorite(theme, !variant.Favorite)
}

func setFavorite(theme string, favorite bool) error {
	state, err := client.GetState()
	if err != nil {
		return errors.Wrap(err, "could not load state")
	}
	if state.Favorites == nil {
		state.Favorites = map[string]bool{}
	}
	state.Favorites[theme] = favorite
	if err := client.SaveState(state); err != nil {
		return errors.Wrap(err, "could not save state")
	}

	zap.S().Debugw("set favorite", "theme", theme, "favorite", favorite)
	if favorite {
		fmt.Printf("%s is now a favorite\n", theme)
	} else {
		fmt.Printf("%s is no longer a favorite\n", theme)
	}
	return nil
}

// runDefault makes the current theme the default for whichever side, dark or
// light, it belongs to.
func runDefault(config client.ConfigModel) error {
	current, variant, err := currentVariant(config)
	if err != nil {
		return err
	}

	state, err := client.GetState()
	if err != nil {
		return errors.Wrap(err, "could not load state")
	}
	side := "dark"
	if variant.Light {
		side = "light"
		state.DefaultLightTheme = current
	} else {
		state.DefaultDarkTheme = current
	}
	if err := client.SaveState(state); err != nil {
		return errors.Wrap(err, "could not save state")
	}

	zap.S().Debugw("set default", "theme", current, "side", side)
	fmt.Printf("%s is now the default %s theme\n", current, side)
	return nil
}

// runState prints what shades has saved over the config, so anything worth
// keeping can be copied into shades.yaml by hand.
func runState() error {
	state, err := client.GetState()
	if err != nil {
		return errors.Wrap(err, "could not load state")
	}

	contents, err := yaml.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "could not encode state")
	}

	fmt.Printf("# %s\n%s", client.GetStatePath(), contents)
	return nil
}
