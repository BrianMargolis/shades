package client

import (
	"os"

	"gopkg.in/yaml.v2"
)

type ConfigModel struct {
	SocketPath        string                       `yaml:"socket-path"`
	Client            map[string]map[string]string `yaml:"client"`
	Themes            Themes                       `yaml:"themes"`
	DefaultDarkTheme  string                       `yaml:"defaultDarkTheme"`
	DefaultLightTheme string                       `yaml:"defaultLightTheme"`
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
