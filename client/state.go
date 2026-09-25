package client

import (
	"os"
	"path/filepath"
	"strings"

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

func SaveState(state State) error {
	path := GetStatePath()

	contents, err := yaml.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "encode state")
	}
	if err := writeFileAtomically(path, contents); err != nil {
		return err
	}

	zap.S().Debugw("saved state", "path", path, "state", state)
	return nil
}

// GetCurrentTheme reads the theme the server last set, or "" if it hasn't set
// one yet.
func GetCurrentTheme() (string, error) {
	path := GetCurrentThemePath()

	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", errors.Wrapf(err, "read current theme file %s", path)
	}

	return strings.TrimSpace(string(contents)), nil
}

// SaveCurrentTheme records the theme the server just set, so a restarted
// server can pick up where it left off. It lives apart from the state file
// because the server writes it on every change, and sharing a file with the
// CLI's own read-modify-write of favorites would let one clobber the other.
func SaveCurrentTheme(theme string) error {
	return writeFileAtomically(GetCurrentThemePath(), []byte(theme+"\n"))
}

func GetCurrentThemePath() string {
	return filepath.Join(filepath.Dir(GetStatePath()), "current-theme")
}

// writeFileAtomically writes through a temp file in the same directory and
// renames it into place, so a crash mid-write can't leave a truncated file
// behind (a truncated state file would break every GetConfig).
func writeFileAtomically(path string, contents []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return errors.Wrapf(err, "create dir %s", dir)
	}

	temp, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*")
	if err != nil {
		return errors.Wrapf(err, "create temp file for %s", path)
	}
	defer os.Remove(temp.Name())

	if _, err := temp.Write(contents); err != nil {
		temp.Close()
		return errors.Wrapf(err, "write temp file for %s", path)
	}
	if err := temp.Close(); err != nil {
		return errors.Wrapf(err, "close temp file for %s", path)
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return errors.Wrapf(err, "replace %s", path)
	}
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
