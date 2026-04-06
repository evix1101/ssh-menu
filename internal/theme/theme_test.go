package theme

import (
	"os"
	"testing"

	"github.com/evix1101/ssh-menu/internal/config"
)

func TestDefaultColors(t *testing.T) {
	c := DefaultColors()
	if c.Background != "#1e1e2e" { t.Errorf("expected default background #1e1e2e, got %s", c.Background) }
	if c.Selected != "#a6e3a1" { t.Errorf("expected default selected #a6e3a1, got %s", c.Selected) }
}

func TestApplyEnvOverrides(t *testing.T) {
	os.Setenv("SSH_MENU_COLOR_BACKGROUND", "#ff0000")
	defer os.Unsetenv("SSH_MENU_COLOR_BACKGROUND")

	c := DefaultColors()
	ApplyEnvOverrides(&c)
	if c.Background != "#ff0000" { t.Errorf("expected env override #ff0000, got %s", c.Background) }
	if c.Foreground != "#cdd6f4" { t.Errorf("foreground should remain default, got %s", c.Foreground) }
}

func TestMergeConfigColors_OnlyOverridesDefaults(t *testing.T) {
	c := DefaultColors()
	c.Background = "#custom"

	fileColors := config.ColorConfig{
		Background: "#fromfile",
		Accent:     "#accentfile",
	}

	MergeConfigColors(&c, fileColors)
	if c.Background != "#custom" { t.Errorf("expected custom bg preserved, got %s", c.Background) }
	if c.Accent != "#accentfile" { t.Errorf("expected accent from file, got %s", c.Accent) }
}

func TestInit_FallsBackToDefaults(t *testing.T) {
	for _, env := range []string{
		"SSH_MENU_COLOR_BACKGROUND", "SSH_MENU_COLOR_FOREGROUND",
		"SSH_MENU_COLOR_BORDER", "SSH_MENU_COLOR_SELECTED",
		"SSH_MENU_COLOR_ACCENT", "SSH_MENU_COLOR_DIMMED",
	} {
		os.Unsetenv(env)
	}

	Init("/nonexistent/path")
	c := Current()
	defaults := DefaultColors()
	if c.Background != defaults.Background { t.Errorf("expected default bg, got %s", c.Background) }
}
