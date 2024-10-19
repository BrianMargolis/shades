package client

import (
	"strings"

	"github.com/pkg/errors"
)

type ThemeVariant struct {
	Light       bool             `yaml:"light"`
	Colors      map[Color]string `yaml:"colors"`
	ThemeName   string           // not present in the yaml, gets faked in from the key name
	VariantName string           // same
}

type Theme struct {
	Name     string                  `yaml:"name"`
	Variants map[string]ThemeVariant `yaml:"variants"`
}

type Themes map[string]Theme

func (t Themes) GetVariant(themeAndVariant string) (ThemeVariant, error) {
	themeName, variantName, err := t.parse(themeAndVariant)
	if err != nil {
		return ThemeVariant{}, err
	}

	theme, ok := t[themeName]
	if !ok {
		return ThemeVariant{}, errors.Errorf("unknown theme: %s", themeName)
	}

	themeVariant, ok := theme.Variants[variantName]
	if !ok {
		return ThemeVariant{}, errors.Errorf("unknown variant: %s", variantName)
	}

	themeVariant.ThemeName = themeName
	themeVariant.VariantName = variantName

	return themeVariant, nil
}

func (t Themes) parse(themeAndVariant string) (string, string, error) {
	parts := strings.Split(themeAndVariant, ";")
	if len(parts) != 2 {
		return "", "", errors.New("invalid theme and variant")
	}

	return parts[0], parts[1], nil
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

var AllColors = []Color{
	ColorBGDim,
	ColorBG0,
	ColorBG1,
	ColorBG2,
	ColorBG3,
	ColorBG4,
	ColorBG5,
	ColorRed,
	ColorOrange,
	ColorYellow,
	ColorGreen,
	ColorBlue,
	ColorAqua,
	ColorPurple,
	ColorFG,
	ColorGray1,
	ColorGray2,
	ColorGray3,
}
