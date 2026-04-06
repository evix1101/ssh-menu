package config

import (
	"strings"
	"testing"
)

func TestParseColors_FromConfig(t *testing.T) {
	input := `# ColorBackground: #000000
# ColorForeground: #ffffff
# ColorBorder: #aaaaaa
# ColorSelected: #00ff00
# ColorAccent: #0000ff
# ColorDimmed: #555555
`
	colors := ParseColors(strings.NewReader(input))
	if colors.Background != "#000000" {
		t.Errorf("Background: expected #000000, got %s", colors.Background)
	}
	if colors.Foreground != "#ffffff" {
		t.Errorf("Foreground: expected #ffffff, got %s", colors.Foreground)
	}
	if colors.Selected != "#00ff00" {
		t.Errorf("Selected: expected #00ff00, got %s", colors.Selected)
	}
}

func TestParseColors_EmptyInput(t *testing.T) {
	colors := ParseColors(strings.NewReader(""))
	if colors.Background != "" {
		t.Errorf("expected empty background, got %s", colors.Background)
	}
}

func TestParseColors_PartialConfig(t *testing.T) {
	input := `# ColorAccent: #ff0000
`
	colors := ParseColors(strings.NewReader(input))
	if colors.Accent != "#ff0000" {
		t.Errorf("Accent: expected #ff0000, got %s", colors.Accent)
	}
	if colors.Background != "" {
		t.Errorf("Background should be empty, got %s", colors.Background)
	}
}
