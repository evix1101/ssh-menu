package config

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// ColorConfig holds UI color settings.
type ColorConfig struct {
	Background string
	Foreground string
	Border     string
	Selected   string
	Accent     string
	Dimmed     string
}

var colorRegexes = map[string]*regexp.Regexp{
	"Background": regexp.MustCompile(`^#\s*ColorBackground:\s*(.+)$`),
	"Foreground": regexp.MustCompile(`^#\s*ColorForeground:\s*(.+)$`),
	"Border":     regexp.MustCompile(`^#\s*ColorBorder:\s*(.+)$`),
	"Selected":   regexp.MustCompile(`^#\s*ColorSelected:\s*(.+)$`),
	"Accent":     regexp.MustCompile(`^#\s*ColorAccent:\s*(.+)$`),
	"Dimmed":     regexp.MustCompile(`^#\s*ColorDimmed:\s*(.+)$`),
}

// ParseColors reads color configuration from a reader.
func ParseColors(r io.Reader) ColorConfig {
	var config ColorConfig

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		for field, re := range colorRegexes {
			if m := re.FindStringSubmatch(line); m != nil {
				value := strings.TrimSpace(m[1])
				switch field {
				case "Background":
					config.Background = value
				case "Foreground":
					config.Foreground = value
				case "Border":
					config.Border = value
				case "Selected":
					config.Selected = value
				case "Accent":
					config.Accent = value
				case "Dimmed":
					config.Dimmed = value
				}
			}
		}
	}

	return config
}
