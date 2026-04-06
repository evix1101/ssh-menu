package ui

import (
	"strings"

	"github.com/evix1101/ssh-menu/internal/theme"
)

func renderViewBar(groups []string, activeIndex int) string {
	totalViews := 1 + len(groups)
	tabs := make([]string, totalViews)

	activeStyle := theme.ActiveTabStyle()
	inactiveStyle := theme.InactiveTabStyle()

	if activeIndex == 0 {
		tabs[0] = activeStyle.Render("All")
	} else {
		tabs[0] = inactiveStyle.Render("All")
	}

	for i, group := range groups {
		displayName := group
		if len(group) > 12 {
			displayName = group[:12] + "…"
		}
		if activeIndex == i+1 {
			tabs[i+1] = activeStyle.Render(displayName)
		} else {
			tabs[i+1] = inactiveStyle.Render(displayName)
		}
	}

	separator := theme.DimStyle().Render(" • ")
	return strings.Join(tabs, separator)
}
