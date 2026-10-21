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
	r, g, b := linearRGB(v.Colors[ColorBG0])
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// warmHue is the OKLCH hue, in degrees, that Warmth measures toward. It sits
// between orange and yellow, so sepia and peach count as warm, and green
// counts for less than its yellow component alone would give it.
const warmHue = 70

// Warmth is how far the variant's background leans toward orange-yellow in
// OKLab, positive for cream and brown, negative for blue, and 0 for gray. HSL
// saturation won't do for this, since it swings wildly near black and white,
// which is where backgrounds live.
func (v ThemeVariant) Warmth() float64 {
	_, a, b := okLab(linearRGB(v.Colors[ColorBG0]))
	angle := warmHue * math.Pi / 180
	return a*math.Cos(angle) + b*math.Sin(angle)
}

func linearRGB(hex string) (float64, float64, float64) {
	r, g, b := HexToRGB(hex)
	return linearize(r), linearize(g), linearize(b)
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

// okLab converts linear sRGB to OKLab's L, a, and b, with Björn Ottosson's
// published matrices.
func okLab(r, g, b float64) (float64, float64, float64) {
	l := math.Cbrt(0.4122214708*r + 0.5363325363*g + 0.0514459929*b)
	m := math.Cbrt(0.2119034982*r + 0.6806995451*g + 0.1073969566*b)
	s := math.Cbrt(0.0883024619*r + 0.2817188376*g + 0.6299787005*b)
	return 0.2104542553*l + 0.7936177850*m - 0.0040720468*s,
		1.9779984951*l - 2.4285922050*m + 0.4505937099*s,
		0.0259040371*l + 0.7827717662*m - 0.8086757660*s
}
