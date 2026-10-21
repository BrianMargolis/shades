package client

import (
	"fmt"
	"math"
)

// HexToRGB converts a hex color code (e.g., "#1d2021") to RGB values, or black
// if it can't parse it.
func HexToRGB(hex string) (int, int, int) {
	var r, g, b int
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, 0, 0
	}
	return r, g, b
}

// Luminance is the WCAG relative luminance of the variant's background, from 0
// for black to 1 for white.
func (v ThemeVariant) Luminance() float64 {
	r, g, b := HexToRGB(v.Colors[ColorBG0])
	return 0.2126*linearize(r) + 0.7152*linearize(g) + 0.0722*linearize(b)
}

// linearize undoes sRGB gamma, since the channels have to be weighted in
// linear light for the sum to track how bright the color looks.
func linearize(channel int) float64 {
	value := float64(channel) / 255
	if value <= 0.04045 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}
