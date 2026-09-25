package picker

import (
	"fmt"
	"strings"

	"github.com/brianmargolis/shades/client"

	"github.com/pkg/errors"
)

// FilterCommand is run by fzf's transform action. It works out the fzf actions
// for a key from the prompt, which is where the picker keeps its filter state
// between reloads.
const FilterCommand = "_picker-filter"

// The actions FilterCommand takes. Refresh keeps the filter as it is, for
// reloading after a favorite toggle.
const (
	FilterDark      = "dark"
	FilterLight     = "light"
	FilterAll       = "all"
	FilterFavorites = "favorites"
	FilterRefresh   = "refresh"
)

var FilterActions = []string{FilterDark, FilterLight, FilterAll, FilterFavorites, FilterRefresh}

var filterKeys = []struct {
	key    string
	action string
	label  string
}{
	{"alt-d", FilterDark, "dark"},
	{"alt-l", FilterLight, "light"},
	{"alt-a", FilterAll, "all"},
	{"alt-s", FilterFavorites, "starred only"},
}

const (
	promptDark      = "dark"
	promptLight     = "light"
	promptFavorites = "★"
	promptEnd       = "> "
)

func prompt(filter client.Filter) string {
	words := []string{}
	if filter.OnlyDark {
		words = append(words, promptDark)
	}
	if filter.OnlyLight {
		words = append(words, promptLight)
	}
	if filter.OnlyFavorites {
		words = append(words, promptFavorites)
	}
	return strings.Join(append(words, promptEnd), " ")
}

func parsePrompt(prompt string) client.Filter {
	filter := client.Filter{}
	for _, word := range strings.Fields(prompt) {
		switch word {
		case promptDark:
			filter.OnlyDark = true
		case promptLight:
			filter.OnlyLight = true
		case promptFavorites:
			filter.OnlyFavorites = true
		}
	}
	return filter
}

// FilterTransform returns the fzf actions that apply a filter key to the
// filter the prompt shows: a new prompt, and the list reloaded to match.
//
// It unbinds one:accept first, because a filter that leaves a single theme
// would otherwise pick it on the spot. Typing rebinds it.
func FilterTransform(currentPrompt string, action string) (string, error) {
	filter := parsePrompt(currentPrompt)
	switch action {
	case FilterDark:
		filter.OnlyDark, filter.OnlyLight = true, false
	case FilterLight:
		filter.OnlyDark, filter.OnlyLight = false, true
	case FilterAll:
		filter = client.Filter{}
	case FilterFavorites:
		filter.OnlyFavorites = !filter.OnlyFavorites
	case FilterRefresh:
	default:
		return "", errors.Errorf("unknown filter action %q", action)
	}

	return fmt.Sprintf(
		"unbind(one)+change-prompt(%s)+reload(%s)",
		prompt(filter),
		strings.Join(append([]string{"shades", ListCommand}, filter.Flags()...), " "),
	), nil
}

func filterBindings() []string {
	bindings := []string{
		fmt.Sprintf(
			"--bind=%s:execute-silent(shades %s {1})+transform(shades %s %s)",
			favoriteKey, ToggleFavoriteCommand, FilterCommand, FilterRefresh,
		),
		"--bind=change:rebind(one)",
	}
	for _, filterKey := range filterKeys {
		bindings = append(bindings, fmt.Sprintf(
			"--bind=%s:transform(shades %s %s)",
			filterKey.key, FilterCommand, filterKey.action,
		))
	}
	return bindings
}

func header() string {
	parts := []string{favoriteKey + " fav"}
	for _, filterKey := range filterKeys {
		parts = append(parts, filterKey.key+" "+filterKey.label)
	}
	return strings.Join(parts, " · ")
}
