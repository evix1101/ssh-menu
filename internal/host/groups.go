package host

import (
	"fmt"
	"sort"
)

// AssignMenuNumbers validates and assigns menu numbers to hosts.
// Returns an error if duplicate explicit menu numbers are found.
func AssignMenuNumbers(hosts []Host) ([]Host, error) {
	usedNumbers := make(map[int]bool)
	for _, h := range hosts {
		if h.MenuNumber != 0 {
			if usedNumbers[h.MenuNumber] {
				return nil, fmt.Errorf("duplicate menu number %d found for host %s", h.MenuNumber, h.ShortName)
			}
			usedNumbers[h.MenuNumber] = true
		}
	}

	nextAvailable := 1
	for i, h := range hosts {
		if h.MenuNumber == 0 {
			for usedNumbers[nextAvailable] {
				nextAvailable++
			}
			hosts[i].MenuNumber = nextAvailable
			usedNumbers[nextAvailable] = true
		}
	}

	sort.Slice(hosts, func(i, j int) bool {
		return hosts[i].MenuNumber < hosts[j].MenuNumber
	})

	return hosts, nil
}

// GetAllGroups returns a sorted list of all unique groups.
// "Ungrouped" is placed last if any hosts have no groups.
func GetAllGroups(hosts []Host) []string {
	groupMap := make(map[string]bool)
	hasUngrouped := false

	for _, h := range hosts {
		if len(h.Groups) == 0 {
			hasUngrouped = true
		}
		for _, g := range h.Groups {
			groupMap[g] = true
		}
	}

	if hasUngrouped {
		groupMap["Ungrouped"] = true
	}

	groups := make([]string, 0, len(groupMap))
	for g := range groupMap {
		groups = append(groups, g)
	}

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

// SortWithPins returns a copy of hosts sorted with pinned hosts first,
// then by menu number within each group.
func SortWithPins(hosts []Host) []Host {
	sorted := make([]Host, len(hosts))
	copy(sorted, hosts)

	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Pinned != sorted[j].Pinned {
			return sorted[i].Pinned
		}
		return sorted[i].MenuNumber < sorted[j].MenuNumber
	})

	return sorted
}

// HostsForGroup returns hosts belonging to a specific group.
func HostsForGroup(hosts []Host, groupName string) []Host {
	var result []Host
	if groupName == "Ungrouped" {
		for _, h := range hosts {
			if len(h.Groups) == 0 {
				result = append(result, h)
			}
		}
	} else {
		for _, h := range hosts {
			for _, g := range h.Groups {
				if g == groupName {
					result = append(result, h)
					break
				}
			}
		}
	}
	return result
}
