package client

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// State is what shades saves on the user's behalf. It lives apart from
// shades.yaml so that shades never rewrites a file the user maintains by hand,
// and GetConfig lays it over the config so everything downstream sees one
// merged view.
type State struct {
	// A false entry is an explicit "not a favorite", which lets the user
	// unfavorite a variant that shades.yaml tags without editing shades.yaml.
	Favorites         map[string]bool `yaml:"favorites,omitempty"`
	DefaultDarkTheme  string          `yaml:"defaultDarkTheme,omitempty"`
	DefaultLightTheme string          `yaml:"defaultLightTheme,omitempty"`
}

func GetState() (State, error) {
	path := GetStatePath()

	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return State{}, nil
	}
	if err != nil {
		return State{}, errors.Wrapf(err, "read state file %s", path)
	}

	state := State{}
	if err := yaml.Unmarshal(contents, &state); err != nil {
		return State{}, errors.Wrapf(err, "parse state file %s", path)
	}

	return state, nil
}

// SaveState writes through a temp file in the same directory and renames it
// into place, so a crash mid-write can't leave a truncated state file that
// would then break every GetConfig.
func SaveState(state State) error {
	path := GetStatePath()
	logger := zap.S().With("path", path)

	contents, err := yaml.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "encode state")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return errors.Wrapf(err, "create state dir %s", dir)
	}

	temp, err := os.CreateTemp(dir, ".state-*.yaml")
	if err != nil {
		return errors.Wrap(err, "create temp state file")
	}
	defer os.Remove(temp.Name())

	if _, err := temp.Write(contents); err != nil {
		temp.Close()
		return errors.Wrap(err, "write temp state file")
	}
	if err := temp.Close(); err != nil {
		return errors.Wrap(err, "close temp state file")
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return errors.Wrapf(err, "replace state file %s", path)
	}

	logger.Debugw("saved state", "state", state)
	return nil
}

func GetStatePath() string {
	envValue := os.Getenv("SHADES_STATE")
	if envValue != "" {
		return envValue
	}

	return os.Getenv("HOME") + "/.shades/state.yaml"
}

func (s State) applyTo(config *ConfigModel) {
	if s.DefaultDarkTheme != "" {
		config.DefaultDarkTheme = s.DefaultDarkTheme
	}
	if s.DefaultLightTheme != "" {
		config.DefaultLightTheme = s.DefaultLightTheme
	}

	for themeAndVariant, favorite := range s.Favorites {
		themeName, variantName, err := config.Themes.parse(themeAndVariant)
		if err != nil {
			zap.S().Debugw("skipping unparseable favorite in state", "theme", themeAndVariant, "error", err)
			continue
		}

		// A theme renamed or removed from shades.yaml leaves its entry behind
		// here, and that shouldn't stop the rest of the state from applying.
		variant, ok := config.Themes[themeName].Variants[variantName]
		if !ok {
			zap.S().Debugw("skipping favorite for a theme not in the config", "theme", themeAndVariant)
			continue
		}

		variant.Favorite = favorite
		config.Themes[themeName].Variants[variantName] = variant
	}
}
