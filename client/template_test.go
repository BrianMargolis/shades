package client

import "testing"

func TestRenderTemplateBatTheme(t *testing.T) {
	variant := ThemeVariant{
		ThemeName:   "ayu",
		VariantName: "dark",
		Colors:      map[Color]string{"RED": "#F07178"},
	}

	got := RenderTemplate("theme = %BAT_THEME%, red = %RED%", "%", variant)

	want := "theme = ayu-dark, red = #F07178"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
