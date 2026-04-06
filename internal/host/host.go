package host

import (
	"fmt"
	"strings"
)

// Warning represents a validation warning for a host.
type Warning struct {
	Level   string // "warn"
	Message string
}

// Host represents an SSH config host entry.
type Host struct {
	ShortName    string
	LongName     string
	User         string
	Port         string
	IP           string
	IdentityFile string
	DescText     string
	MenuNumber   int
	Groups       []string
	Pinned       bool
	SourceFile   string
	Warnings     []Warning
}

// Title returns a formatted string for displaying the host in the list.
func (h Host) Title() string {
	prefix := "  "
	if h.Pinned {
		prefix = "★ "
	}
	return fmt.Sprintf("%s%2d) %s", prefix, h.MenuNumber, h.ShortName)
}

// FilterValue returns a string used for filtering.
func (h Host) FilterValue() string {
	return fmt.Sprintf("%d %s %s %s %s %s",
		h.MenuNumber, h.ShortName, h.DescText, h.LongName, h.IP, strings.Join(h.Groups, " "))
}
