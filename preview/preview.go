package preview

import (
	"brianmargolis/shades/client"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type Previewer interface {
	Preview(theme client.ThemeVariant) (string, error)
}

type previewer struct {
}

func NewPreviewer() Previewer {
	return &previewer{}
}

func (p *previewer) Preview(theme client.ThemeVariant) (string, error) {
	hexCodesSeen := make(map[string]struct{})

	swatches := []string{}
	for _, color := range client.AllColors {
		hexCode := theme.Colors[color]
		if _, ok := hexCodesSeen[hexCode]; ok {
			continue
		}
		r, g, b := hexToRGB(hexCode)
		zap.S().Debug(
			"color swatch",
			zap.String("theme", theme.ThemeName),
			zap.Int("r", r),
			zap.Int("g", g),
			zap.Int("b", b),
			zap.String("color", string(color)),
		)
		swatches = append(swatches, fmt.Sprintf("\033[48;2;%d;%d;%dm  %-180s \033[0m", r, g, b, ""))
		hexCodesSeen[hexCode] = struct{}{}
	}

	return strings.Join(swatches, "\n"), nil
}

// hexToRGB converts a hex color code (e.g., "#1d2021") to RGB values
func hexToRGB(hex string) (int, int, int) {
	var r, g, b int
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, 0, 0 // Return black in case of error
	}
	return r, g, b
}
