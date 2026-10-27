package client

import "testing"

func TestRenderTemplate(t *testing.T) {
	variant := ThemeVariant{
		Colors:      map[Color]string{"RED": "#E67E80"},
		ThemeName:   "everforest",
		VariantName: "dark-medium",
	}

	got := RenderTemplate("%RED% %RED_HEX% syntax-theme = %BAT_THEME%", "%", variant)
	want := "#E67E80 E67E80 syntax-theme = everforest-dark-medium"
	if got != want {
		t.Errorf("RenderTemplate() = %q, want %q", got, want)
	}
}
