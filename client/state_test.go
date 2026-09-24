package client

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func testConfig() ConfigModel {
	return ConfigModel{
		DefaultDarkTheme:  "everforest;dark-medium",
		DefaultLightTheme: "everforest;light-medium",
		Themes: Themes{
			"everforest": {
				Variants: map[string]ThemeVariant{
					"dark-medium":  {Light: false, Favorite: true},
					"light-medium": {Light: true},
				},
			},
		},
	}
}

func TestStateApplyTo(t *testing.T) {
	config := testConfig()
	State{
		Favorites: map[string]bool{
			"everforest;dark-medium":  false,
			"everforest;light-medium": true,
			"gone;variant":            true,
			"unparseable":             true,
		},
		DefaultDarkTheme: "gruvbox;dark",
	}.applyTo(&config)

	variants := config.Themes["everforest"].Variants
	if variants["dark-medium"].Favorite {
		t.Error("a false in state should unfavorite a variant the config tags")
	}
	if !variants["light-medium"].Favorite {
		t.Error("a true in state should favorite a variant the config doesn't tag")
	}
	if _, ok := config.Themes["gone"]; ok {
		t.Error("a favorite for a theme not in the config shouldn't add that theme")
	}
	if config.DefaultDarkTheme != "gruvbox;dark" {
		t.Errorf("DefaultDarkTheme = %q, want the one from state", config.DefaultDarkTheme)
	}
	if config.DefaultLightTheme != "everforest;light-medium" {
		t.Errorf("DefaultLightTheme = %q, want the config's, since state doesn't set one", config.DefaultLightTheme)
	}
}

func TestStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.yaml")
	t.Setenv("SHADES_STATE", path)

	missing, err := GetState()
	if err != nil {
		t.Fatalf("GetState() with no file: %v", err)
	}
	if !reflect.DeepEqual(missing, State{}) {
		t.Errorf("GetState() with no file = %+v, want empty", missing)
	}

	want := State{
		Favorites:         map[string]bool{"everforest;dark-medium": true, "gruvbox;dark": false},
		DefaultLightTheme: "flexoki;light",
	}
	if err := SaveState(want); err != nil {
		t.Fatalf("SaveState(): %v", err)
	}

	got, err := GetState()
	if err != nil {
		t.Fatalf("GetState(): %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetState() = %+v, want %+v", got, want)
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("state dir holds %d files, want only state.yaml (temp file left behind?)", len(entries))
	}
}
