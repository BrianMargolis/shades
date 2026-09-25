package client

import "testing"

func TestTogglerOpposite(t *testing.T) {
	config := testConfig()
	config.Themes["gruvbox"] = Theme{
		Variants: map[string]ThemeVariant{
			"dark":  {Light: false},
			"light": {Light: true},
		},
	}
	toggler := TogglerClient{
		DarkTheme:  config.DefaultDarkTheme,
		LightTheme: config.DefaultLightTheme,
		Themes:     config.Themes,
	}

	tests := []struct {
		current string
		want    string
	}{
		{current: "everforest;dark-medium", want: "everforest;light-medium"},
		{current: "everforest;light-medium", want: "everforest;dark-medium"},
		{current: "gruvbox;dark", want: "everforest;light-medium"},
		{current: "gruvbox;light", want: "everforest;dark-medium"},
	}

	for _, test := range tests {
		got, err := toggler.opposite(test.current)
		if err != nil {
			t.Errorf("opposite(%q): %v", test.current, err)
			continue
		}
		if got != test.want {
			t.Errorf("opposite(%q) = %q, want %q", test.current, got, test.want)
		}
	}

	if _, err := toggler.opposite("gone;variant"); err == nil {
		t.Error("opposite() of a theme not in the config should fail")
	}
}
