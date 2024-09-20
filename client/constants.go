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

const (
	ThemeEverforestVariantDarkHard    ThemeEverforestVariant = "dark-hard"
	ThemeEverforestVariantDarkMedium  ThemeEverforestVariant = "dark-medium"
	ThemeEverforestVariantDarkSoft    ThemeEverforestVariant = "dark-low"
	ThemeEverforestVariantLightHard   ThemeEverforestVariant = "light-hard"
	ThemeEverforestVariantLightMedium ThemeEverforestVariant = "light-medium"
	ThemeEverforestVariantLightSoft   ThemeEverforestVariant = "light-low"
)

type ThemeEverforest struct {
	Name     string                                  `yaml:"name"`
	Variants map[ThemeEverforestVariant]ThemeVariant `yaml:"variants"`
}

// CATPPUCCIN
type ThemeCatppuccinVariant string

const (
	ThemeCatppuccinVariantLatte ThemeCatppuccinVariant = "latte"
	ThemeCatppuccinVariantMocha ThemeCatppuccinVariant = "mocha"
)

type ThemeCatppuccin struct {
	Name     string                                  `yaml:"name"`
	Variants map[ThemeCatppuccinVariant]ThemeVariant `yaml:"variants"`
}

// GRUVBOX
type ThemeGruvboxVariant string

const (
	ThemeGruvboxVariantLight ThemeGruvboxVariant = "light"
	ThemeGruvboxVariantDark  ThemeGruvboxVariant = "dark"
)

type ThemeGruvbox struct {
	Name     string                               `yaml:"name"`
	Variants map[ThemeGruvboxVariant]ThemeVariant `yaml:"variants"`
}

type Themes struct {
	Everforest ThemeEverforest `yaml:"everforest"`
	Catppuccin ThemeCatppuccin `yaml:"catppuccin"`
	Gruvbox    ThemeGruvbox    `yaml:"gruvbox"`
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
	case "gruvbox":
		return t.Gruvbox.Variants[ThemeGruvboxVariant(variant)], nil
	}

	return ThemeVariant{}, errors.New("unknown theme or variant")
}

type Color string

const (
	ColorBGDim  Color = "BGDIM"
	ColorBG0    Color = "BG0"
	ColorBG1    Color = "BG1"
	ColorBG2    Color = "BG2"
	ColorBG3    Color = "BG3"
	ColorBG4    Color = "BG4"
	ColorBG5    Color = "BG5"
	ColorRed    Color = "RED"
	ColorOrange Color = "ORANGE"
	ColorYellow Color = "YELLOW"
	ColorGreen  Color = "GREEN"
	ColorBlue   Color = "BLUE"
	ColorAqua   Color = "AQUA"
	ColorPurple Color = "PURPLE"
	ColorFG     Color = "FG"
	ColorGray1  Color = "GRAY1"
	ColorGray2  Color = "GRAY2"
	ColorGray3  Color = "GRAY3"
)
