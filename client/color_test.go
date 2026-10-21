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
