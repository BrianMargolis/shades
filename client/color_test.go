package client

import "testing"

func TestLuminance(t *testing.T) {
	variant := func(background string) ThemeVariant {
		return ThemeVariant{Colors: map[Color]string{ColorBG0: background}}
	}

	black := variant("#000000").Luminance()
	gray := variant("#808080").Luminance()
	white := variant("#ffffff").Luminance()
	if !(black < gray && gray < white) {
		t.Errorf("Luminance() black %v, gray %v, white %v, want them increasing", black, gray, white)
	}
	if black != 0 || white < 0.999 {
		t.Errorf("Luminance() black %v, white %v, want 0 and 1", black, white)
	}
	if missing := (ThemeVariant{}).Luminance(); missing != 0 {
		t.Errorf("Luminance() with no BG0 = %v, want 0", missing)
	}
}

func TestWarmth(t *testing.T) {
	variant := func(background string) ThemeVariant {
		return ThemeVariant{Colors: map[Color]string{ColorBG0: background}}
	}

	tests := []struct {
		name   string
		cooler string
		warmer string
	}{
		{name: "blue-white is cooler than white", cooler: "#e1e2e7", warmer: "#ffffff"},
		{name: "white is cooler than sepia", cooler: "#ffffff", warmer: "#fdf6e3"},
		{name: "blue-black is cooler than gray-black", cooler: "#24283b", warmer: "#1c1c1c"},
		{name: "gray-black is cooler than brown-black", cooler: "#1c1c1c", warmer: "#282828"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cooler, warmer := variant(test.cooler).Warmth(), variant(test.warmer).Warmth()
			if cooler >= warmer {
				t.Errorf("Warmth() %s = %v, %s = %v, want the first lower", test.cooler, cooler, test.warmer, warmer)
			}
		})
	}
}
