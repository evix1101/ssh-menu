package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/evix1101/ssh-menu/internal/host"
	"github.com/evix1101/ssh-menu/internal/theme"
)

func renderHostList(hosts []host.Host, cursor, scrollOffset, width, height int) string {
	if len(hosts) == 0 {
		return theme.DimStyle().Render("No hosts match your filter")
	}

	colors := theme.Current()
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.Selected))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Foreground))

	var b strings.Builder

	visibleCount := height
	if visibleCount <= 0 {
		visibleCount = len(hosts)
	}

	start := scrollOffset
	end := start + visibleCount
	if end > len(hosts) {
		end = len(hosts)
	}

	for i := start; i < end; i++ {
		h := hosts[i]
		pointer := " "
		if i == cursor {
			pointer = "▸"
		}

		pin := " "
		if h.Pinned {
			pin = "★"
		}

		line := fmt.Sprintf("%s%s%2d) %s", pointer, pin, h.MenuNumber, h.ShortName)

		maxWidth := width - 2
		if maxWidth > 0 && len(line) > maxWidth {
			line = line[:maxWidth-1] + "…"
		}

		if i == cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func calculateScrollOffset(cursor, currentOffset, visibleHeight, totalItems int) int {
	if visibleHeight <= 0 || totalItems <= visibleHeight {
		return 0
	}

	offset := currentOffset

	if cursor < offset {
		offset = cursor
	}

	if cursor >= offset+visibleHeight {
		offset = cursor - visibleHeight + 1
	}

	maxOffset := totalItems - visibleHeight
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}

	return offset
}
