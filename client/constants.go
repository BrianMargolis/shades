package client

import "github.com/pkg/errors"

type ThemeVariant struct {
	Light       bool             `yaml:"light"`
	Colors      map[Color]string `yaml:"colors"`
	ThemeName   string           // not present in the yaml, gets faked in from the key name
	VariantName string           // same
}

// EVERFOREST
type ThemeEverforestVariant string

const ThemeEverforestVariantDarkHard ThemeEverforestVariant = "dark-hard"
const ThemeEverforestVariantDarkMedium ThemeEverforestVariant = "dark-medium"
const ThemeEverforestVariantDarkSoft ThemeEverforestVariant = "dark-low"
const ThemeEverforestVariantLightHard ThemeEverforestVariant = "light-hard"
const ThemeEverforestVariantLightMedium ThemeEverforestVariant = "light-medium"
const ThemeEverforestVariantLightSoft ThemeEverforestVariant = "light-low"

type ThemeEverforest struct {
	Name     string                                  `yaml:"name"`
	Variants map[ThemeEverforestVariant]ThemeVariant `yaml:"variants"`
}

// CATPPUCCIN
type ThemeCatppuccinVariant string

const ThemeCatppuccinVariantLatte ThemeCatppuccinVariant = "latte"
const ThemeCatppuccinVariantMocha ThemeCatppuccinVariant = "mocha"

type ThemeCatppuccin struct {
	Name     string                                  `yaml:"name"`
	Variants map[ThemeCatppuccinVariant]ThemeVariant `yaml:"variants"`
}

type Themes struct {
	Everforest ThemeEverforest `yaml:"everforest"`
	Catppuccin ThemeCatppuccin `yaml:"catppuccin"`
}

func (t Themes) GetVariant(themeAndVariant string) (ThemeVariant, error) {
	theme, variant, err := GetThemeAndVariant(themeAndVariant)
	if err != nil {
		return ThemeVariant{}, err
	}

	themeVariant, err := t.getVariant(theme, variant)
	if err != nil {
		return ThemeVariant{}, err
	}

	themeVariant.ThemeName = theme
	themeVariant.VariantName = variant

	return themeVariant, nil
}

func (t Themes) getVariant(theme, variant string) (ThemeVariant, error) {
	switch theme {
	case "everforest":
		return t.Everforest.Variants[ThemeEverforestVariant(variant)], nil
	case "catppuccin":
		return t.Catppuccin.Variants[ThemeCatppuccinVariant(variant)], nil
	}

	return ThemeVariant{}, errors.New("unknown theme or variant")
}

type Color string

const ColorBGDim Color = "BGDIM"
const ColorBG0 Color = "BG0"
const ColorBG1 Color = "BG1"
const ColorBG2 Color = "BG2"
const ColorBG3 Color = "BG3"
const ColorBG4 Color = "BG4"
const ColorBG5 Color = "BG5"
const ColorRed Color = "RED"
const ColorOrange Color = "ORANGE"
const ColorYellow Color = "YELLOW"
const ColorGreen Color = "GREEN"
const ColorBlue Color = "BLUE"
const ColorAqua Color = "AQUA"
const ColorPurple Color = "PURPLE"
const ColorFG Color = "FG"
const ColorGray1 Color = "GRAY1"
const ColorGray2 Color = "GRAY2"
const ColorGray3 Color = "GRAY3"
