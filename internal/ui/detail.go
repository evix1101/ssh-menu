package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/evix1101/ssh-menu/internal/host"
	"github.com/evix1101/ssh-menu/internal/theme"
)

func renderDetail(h host.Host, width, height int) string {
	if width < 20 {
		return ""
	}

	colors := theme.Current()
	var b strings.Builder

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.Accent))
	b.WriteString(nameStyle.Render(h.ShortName))
	b.WriteString("\n")

	if h.DescText != "" {
		b.WriteString(theme.NormalStyle().Render(h.DescText))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	labelStyle := theme.DimStyle()
	valueStyle := theme.NormalStyle()

	details := []struct {
		label string
		value string
	}{
		{"Host", h.LongName},
		{"User", h.User},
		{"Port", h.Port},
		{"Key", h.IdentityFile},
		{"IP", h.IP},
	}

	for _, d := range details {
		if d.value != "" {
			b.WriteString(fmt.Sprintf("%s  %s\n",
				labelStyle.Render(fmt.Sprintf("%-5s", d.label+":")),
				valueStyle.Render(d.value)))
		}
	}

	if len(h.Groups) > 0 {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%s  %s\n",
			labelStyle.Render("Groups:"),
			valueStyle.Render(strings.Join(h.Groups, ", "))))
	}

	if h.Pinned {
		b.WriteString("\n")
		pinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Selected))
		b.WriteString(pinStyle.Render("★ Pinned"))
		b.WriteString("\n")
	}

	if len(h.Warnings) > 0 {
		b.WriteString("\n")
		warnStyle := theme.WarningStyle()
		for _, w := range h.Warnings {
			b.WriteString(warnStyle.Render(fmt.Sprintf("⚠ %s", w.Message)))
			b.WriteString("\n")
		}
	}

	panel := lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		Padding(1, 2)

	return panel.Render(b.String())
}
