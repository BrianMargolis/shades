package preview

import (
	"fmt"
	"strings"

	"github.com/brianmargolis/shades/client"
	"go.uber.org/zap"
)

type Previewer interface {
	Preview(theme client.ThemeVariant) (string, error)
	Swatches(theme client.ThemeVariant) string
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
		r, g, b := client.HexToRGB(hexCode)
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

// Swatches renders a palette as a single compact line, for the gallery.
//
// Unlike Preview, it emits one swatch per entry in client.AllColors instead of
// skipping hex codes it has already seen. Deduping would give palettes with
// repeated colors shorter lines, and the gallery stacks these rows on top of
// each other, so the columns have to line up for the themes to be comparable.
func (p *previewer) Swatches(theme client.ThemeVariant) string {
	var line strings.Builder
	for _, color := range client.AllColors {
		r, g, b := client.HexToRGB(theme.Colors[color])
		fmt.Fprintf(&line, "\033[48;2;%d;%d;%dm  \033[0m", r, g, b)
	}

	return line.String()
}
