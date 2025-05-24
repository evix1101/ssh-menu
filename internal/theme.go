package internal

import (
	"os"
	"regexp"
	"strings"
)

// DefaultColors returns the default color scheme
func DefaultColors() ColorConfig {
	return ColorConfig{
		Background: "#1e1e2e", // Dark blue/purple
		Foreground: "#cdd6f4", // Light gray
		Border:     "#9399b2", // Medium gray
		Selected:   "#a6e3a1", // Green
		Accent:     "#89dceb", // Cyan
		Dimmed:     "#585b70", // Dark gray
	}
}

// colorEnvVars holds the environment variable names for colors
var colorEnvVars = map[string]func(*ColorConfig, string){
	"SSH_MENU_COLOR_BACKGROUND": func(c *ColorConfig, v string) { c.Background = v },
	"SSH_MENU_COLOR_FOREGROUND": func(c *ColorConfig, v string) { c.Foreground = v },
	"SSH_MENU_COLOR_BORDER":     func(c *ColorConfig, v string) { c.Border = v },
	"SSH_MENU_COLOR_SELECTED":   func(c *ColorConfig, v string) { c.Selected = v },
	"SSH_MENU_COLOR_ACCENT":     func(c *ColorConfig, v string) { c.Accent = v },
	"SSH_MENU_COLOR_DIMMED":     func(c *ColorConfig, v string) { c.Dimmed = v },
}

// applyEnvVarColors applies environment variable colors to the config
func applyEnvVarColors(config *ColorConfig) {
	for envVar, setter := range colorEnvVars {
		if val := os.Getenv(envVar); val != "" {
			setter(config, val)
		}
	}
}

// mergeConfigColors merges colors from SSH config if defaults are still in use
func mergeConfigColors(config *ColorConfig, configColors ColorConfig) {
	defaults := DefaultColors()
	if config.Background == defaults.Background && configColors.Background != "" {
		config.Background = configColors.Background
	}
	if config.Foreground == defaults.Foreground && configColors.Foreground != "" {
		config.Foreground = configColors.Foreground
	}
	if config.Border == defaults.Border && configColors.Border != "" {
		config.Border = configColors.Border
	}
	if config.Selected == defaults.Selected && configColors.Selected != "" {
		config.Selected = configColors.Selected
	}
	if config.Accent == defaults.Accent && configColors.Accent != "" {
		config.Accent = configColors.Accent
	}
	if config.Dimmed == defaults.Dimmed && configColors.Dimmed != "" {
		config.Dimmed = configColors.Dimmed
	}
}

// ApplyColorConfig reads and applies color configuration
func ApplyColorConfig(configPath string) {
	config := DefaultColors()

	// Read from environment variables first (highest priority)
	applyEnvVarColors(&config)

	// Read from SSH config file if no env vars are set
	if configPath != "" {
		configColors := readColorsFromConfig(configPath)
		mergeConfigColors(&config, configColors)
	}

	// Store the config for use by the UI
	currentColorConfig = config
}

// Global variable to store current color config
var currentColorConfig ColorConfig

// GetCurrentColors returns the current color configuration
func GetCurrentColors() ColorConfig {
	if currentColorConfig.Foreground == "" {
		return DefaultColors()
	}
	return currentColorConfig
}

// readColorsFromConfig reads color configuration from SSH config file
func readColorsFromConfig(configPath string) ColorConfig {
	config := ColorConfig{}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return config
	}

	// Regular expressions for color settings
	reColorBg := regexp.MustCompile(`^#\s*ColorBackground:\s*(.+)$`)
	reColorFg := regexp.MustCompile(`^#\s*ColorForeground:\s*(.+)$`)
	reColorBorder := regexp.MustCompile(`^#\s*ColorBorder:\s*(.+)$`)
	reColorSelected := regexp.MustCompile(`^#\s*ColorSelected:\s*(.+)$`)
	reColorAccent := regexp.MustCompile(`^#\s*ColorAccent:\s*(.+)$`)
	reColorDimmed := regexp.MustCompile(`^#\s*ColorDimmed:\s*(.+)$`)

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if matches := reColorBg.FindStringSubmatch(line); len(matches) > 1 {
			config.Background = strings.TrimSpace(matches[1])
		} else if matches := reColorFg.FindStringSubmatch(line); len(matches) > 1 {
			config.Foreground = strings.TrimSpace(matches[1])
		} else if matches := reColorBorder.FindStringSubmatch(line); len(matches) > 1 {
			config.Border = strings.TrimSpace(matches[1])
		} else if matches := reColorSelected.FindStringSubmatch(line); len(matches) > 1 {
			config.Selected = strings.TrimSpace(matches[1])
		} else if matches := reColorAccent.FindStringSubmatch(line); len(matches) > 1 {
			config.Accent = strings.TrimSpace(matches[1])
		} else if matches := reColorDimmed.FindStringSubmatch(line); len(matches) > 1 {
			config.Dimmed = strings.TrimSpace(matches[1])
		}
	}

	return config
}
