package internal

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Host represents an SSH config host entry.
// It contains all the details needed to connect to an SSH server.
type Host struct {
	ShortName           string   // The host alias from SSH config
	LongName            string   // The actual hostname or IP address
	User                string   // Username for SSH connection
	Port                string   // Port number for SSH connection
	IP                  string   // IP address (from comments)
	IdentityFile        string   // Path to the SSH key file
	DescText            string   // Description text from Menu comment
	MenuNumber          int      // Menu number (0 means unspecified)
	Groups              []string // Categories/groups the host belongs to
	ServerAliveInterval int      // The ServerAliveInterval parameter (in seconds)
	ServerAliveCountMax int      // The ServerAliveCountMax parameter
	ConnectTimeout      int      // The ConnectTimeout parameter (in seconds)
}

// Title returns a formatted string for displaying the host in the list
func (h Host) Title() string {
	return fmt.Sprintf("%2d) %-20s %-5s", h.MenuNumber, h.ShortName, h.Port)
}

// Description returns a formatted description with optional IP address and groups
func (h Host) Description() string {
	desc := h.DescText
	if h.IP != "" {
		desc += fmt.Sprintf(" (%s)", h.IP)
	}
	if len(h.Groups) > 0 {
		desc += fmt.Sprintf(" [%s]", strings.Join(h.Groups, ", "))
	}
	return desc
}

// FilterValue returns a string used for filtering in the list
func (h Host) FilterValue() string {
	// Include groups in the filter value to enable filtering by group
	return fmt.Sprintf("%d %s %s", h.MenuNumber, h.ShortName, strings.Join(h.Groups, " "))
}

// AssignMenuNumbers ensures all hosts have valid menu numbers.
// This function validates menu numbers for duplicates and assigns
// numbers to hosts that don't have them.
func AssignMenuNumbers(hosts []Host) []Host {
	// Validate explicit menu numbers for duplicates
	usedNumbers := make(map[int]bool)
	for _, h := range hosts {
		if h.MenuNumber != 0 {
			if usedNumbers[h.MenuNumber] {
				fmt.Printf("Error: Duplicate menu number %d found for host %s.\n", h.MenuNumber, h.ShortName)
				os.Exit(1)
			}
			usedNumbers[h.MenuNumber] = true
		}
	}

	// For hosts with no explicit menu number, assign the first available numbers starting at 1
	nextAvailable := 1
	for i, h := range hosts {
		if h.MenuNumber == 0 {
			// Find the next available number
			for usedNumbers[nextAvailable] {
				nextAvailable++
			}
			hosts[i].MenuNumber = nextAvailable
			usedNumbers[nextAvailable] = true
		}
	}

	// Sort hosts by MenuNumber
	sort.Slice(hosts, func(i, j int) bool {
		return hosts[i].MenuNumber < hosts[j].MenuNumber
	})

	return hosts
}

// GroupHosts organizes hosts into a map by group
func GroupHosts(hosts []Host) map[string][]Host {
	groups := make(map[string][]Host)

	// First, handle hosts with no group - they go into "Ungrouped"
	for _, h := range hosts {
		if len(h.Groups) == 0 {
			groups["Ungrouped"] = append(groups["Ungrouped"], h)
			continue
		}

		// Add the host to each of its groups
		for _, g := range h.Groups {
			groups[g] = append(groups[g], h)
		}
	}

	return groups
}

// GetAllGroups returns a sorted list of all unique groups across all hosts
func GetAllGroups(hosts []Host) []string {
	groupMap := make(map[string]bool)

	// Collect all unique groups
	for _, h := range hosts {
		for _, g := range h.Groups {
			groupMap[g] = true
		}
	}

	// If we have hosts without groups, ensure "Ungrouped" is included
	hasUngrouped := false
	for _, h := range hosts {
		if len(h.Groups) == 0 {
			hasUngrouped = true
			break
		}
	}
	if hasUngrouped {
		groupMap["Ungrouped"] = true
	}

	// Convert map to slice
	groups := make([]string, 0, len(groupMap))
	for g := range groupMap {
		groups = append(groups, g)
	}

	// Sort groups alphabetically, but ensure "Ungrouped" is last if present
	sort.Slice(groups, func(i, j int) bool {
		if groups[i] == "Ungrouped" {
			return false
		}
		if groups[j] == "Ungrouped" {
			return true
		}
		return groups[i] < groups[j]
	})

	return groups
}
