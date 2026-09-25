package main

import (
	"fmt"

	"github.com/brianmargolis/shades/client"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// currentVariant resolves the theme the server says is set. The lookup against
// the config catches a server that is still holding a theme since removed from
// shades.yaml, before that theme gets written into the state file.
func currentVariant(config client.ConfigModel) (string, client.ThemeVariant) {
	current, err := client.CurrentTheme(socketPath)
	if err != nil {
		fatal("could not get the current theme (is 'shades -s' running?)", err)
	}

	variant, err := config.Themes.GetVariant(current)
	if err != nil {
		fatal(fmt.Sprintf("the current theme %q isn't in your config", current), err)
	}

	return current, variant
}

func loadState() client.State {
	state, err := client.GetState()
	if err != nil {
		fatal("could not load state", err)
	}

	return state
}

func saveState(state client.State) {
	if err := client.SaveState(state); err != nil {
		fatal("could not save state", err)
	}
}

// runFavorite marks or unmarks the current theme as a favorite.
func runFavorite(config client.ConfigModel, favorite bool) {
	current, _ := currentVariant(config)
	setFavorite(current, favorite)
}

// runToggleFavorite flips a named theme's favorite, for the picker's keybind.
// It takes the theme from fzf rather than asking the server, because the
// server only hears about the focused theme through an async set that may not
// have landed yet.
func runToggleFavorite(config client.ConfigModel, theme string) {
	variant, err := config.Themes.GetVariant(theme)
	if err != nil {
		fatal(fmt.Sprintf("%q isn't in your config", theme), err)
	}

	setFavorite(theme, !variant.Favorite)
}

func setFavorite(theme string, favorite bool) {
	state := loadState()
	if state.Favorites == nil {
		state.Favorites = map[string]bool{}
	}
	state.Favorites[theme] = favorite
	saveState(state)

	zap.S().Debugw("set favorite", "theme", theme, "favorite", favorite)
	if favorite {
		fmt.Printf("%s is now a favorite\n", theme)
	} else {
		fmt.Printf("%s is no longer a favorite\n", theme)
	}
}

// runDefault makes the current theme the default for whichever side, dark or
// light, it belongs to.
func runDefault(config client.ConfigModel) {
	current, variant := currentVariant(config)

	state := loadState()
	side := "dark"
	if variant.Light {
		side = "light"
		state.DefaultLightTheme = current
	} else {
		state.DefaultDarkTheme = current
	}
	saveState(state)

	zap.S().Debugw("set default", "theme", current, "side", side)
	fmt.Printf("%s is now the default %s theme\n", current, side)
}

// runState prints what shades has saved over the config, so anything worth
// keeping can be copied into shades.yaml by hand.
func runState() {
	state := loadState()

	contents, err := yaml.Marshal(state)
	if err != nil {
		fatal("could not encode state", err)
	}

	fmt.Printf("# %s\n%s", client.GetStatePath(), contents)
}
