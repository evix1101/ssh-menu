package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/evix1101/ssh-menu/internal/config"
)

// Colors is an alias for config.ColorConfig.
type Colors = config.ColorConfig

var current Colors

// DefaultColors returns the default Catppuccin Mocha-inspired color scheme.
func DefaultColors() Colors {
	return Colors{
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Border:     "#9399b2",
		Selected:   "#a6e3a1",
		Accent:     "#89dceb",
		Dimmed:     "#585b70",
	}
}

var envVars = []struct {
	envKey string
	field  func(*Colors) *string
}{
	{"SSH_MENU_COLOR_BACKGROUND", func(c *Colors) *string { return &c.Background }},
	{"SSH_MENU_COLOR_FOREGROUND", func(c *Colors) *string { return &c.Foreground }},
	{"SSH_MENU_COLOR_BORDER", func(c *Colors) *string { return &c.Border }},
	{"SSH_MENU_COLOR_SELECTED", func(c *Colors) *string { return &c.Selected }},
	{"SSH_MENU_COLOR_ACCENT", func(c *Colors) *string { return &c.Accent }},
	{"SSH_MENU_COLOR_DIMMED", func(c *Colors) *string { return &c.Dimmed }},
}

// ApplyEnvOverrides applies environment variable colors to the config.
func ApplyEnvOverrides(c *Colors) {
	for _, ev := range envVars {
		if val := os.Getenv(ev.envKey); val != "" {
			*ev.field(c) = val
		}
	}
}

// MergeConfigColors merges file-based colors into the config,
// only overriding fields that are still at their default values.
func MergeConfigColors(c *Colors, fileColors config.ColorConfig) {
	defaults := DefaultColors()
	merge := func(current *string, defaultVal, fileVal string) {
		if *current == defaultVal && fileVal != "" {
			*current = fileVal
		}
	}
	merge(&c.Background, defaults.Background, fileColors.Background)
	merge(&c.Foreground, defaults.Foreground, fileColors.Foreground)
	merge(&c.Border, defaults.Border, fileColors.Border)
	merge(&c.Selected, defaults.Selected, fileColors.Selected)
	merge(&c.Accent, defaults.Accent, fileColors.Accent)
	merge(&c.Dimmed, defaults.Dimmed, fileColors.Dimmed)
}

// Init loads color configuration from env vars and config file.
func Init(configPath string) {
	current = DefaultColors()
	ApplyEnvOverrides(&current)

	if configPath != "" {
		f, err := os.Open(configPath)
		if err == nil {
			defer f.Close()
			fileColors := config.ParseColors(f)
			MergeConfigColors(&current, fileColors)
		}
	}
}

// Current returns the active color configuration.
func Current() Colors {
	if current.Foreground == "" {
		return DefaultColors()
	}
	return current
}

// Style constructors

func TitleStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c.Accent))
}

func SelectedStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c.Selected))
}

func NormalStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Foreground))
}

func DimStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Dimmed))
}

func WarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af"))
}

func ActiveTabStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(c.Background)).
		Background(lipgloss.Color(c.Selected)).
		Padding(0, 1)
}

func InactiveTabStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Foreground)).
		Padding(0, 1)
}

func BorderStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Border))
}
