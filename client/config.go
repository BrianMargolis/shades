package client

import (
	"os"

	"gopkg.in/yaml.v2"
)

type ConfigModel struct {
	SocketPath        string                       `yaml:"socket-path"`
	Client            map[string]map[string]string `yaml:"client"`
	Templates         []TemplateConfig             `yaml:"templates"`
	Themes            Themes                       `yaml:"themes"`
	DefaultDarkTheme  string                       `yaml:"defaultDarkTheme"`
	DefaultLightTheme string                       `yaml:"defaultLightTheme"`
	Daemon            DaemonConfig                 `yaml:"daemon"`
}

// TemplateConfig describes one arbitrary file to render for the "template"
// client: a template file, an output path, and the character that wraps color
// placeholders in the template (e.g. escape-char "%" turns "%RED%" into the
// red hex value).
type TemplateConfig struct {
	TemplatePath string `yaml:"template-path"`
	OutputPath   string `yaml:"output-path"`
	EscapeChar   string `yaml:"escape-char"`
}

type DaemonConfig struct {
	EnabledComponents []string `yaml:"enabled-components"`
}

func GetConfig() (ConfigModel, error) {
	f, err := os.Open(getConfigPath())
	if err != nil {
		return ConfigModel{}, err
	}

	defer f.Close()

	config := ConfigModel{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		return ConfigModel{}, err
	}

	state, err := GetState()
	if err != nil {
		return ConfigModel{}, err
	}
	state.applyTo(&config)

	return config, nil
}

func getConfigPath() string {
	// if SHADES_CONFIG is defined, use that
	envValue := os.Getenv("SHADES_CONFIG")
	if envValue != "" {
		return envValue
	}

	return os.Getenv("HOME") + "/.config/shades/shades.yaml"
}
