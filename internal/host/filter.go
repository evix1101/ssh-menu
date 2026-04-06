package host

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// Score computes a fuzzy match score for query against text.
// Returns 0 if no match. Higher scores are better matches.
func Score(query, text string) int {
	if query == "" || text == "" {
		return 0
	}

	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(text)

	// Check for exact substring — highest score
	if idx := strings.Index(textLower, queryLower); idx >= 0 {
		score := 1000
		if idx == 0 {
			score += 20
		}
		if idx > 0 && isBoundary(rune(text[idx-1])) {
			score += 15
		}
		return score
	}

	// Fuzzy character-by-character matching
	qi := 0
	score := 0
	consecutive := 0
	lastMatchIdx := -1

	for ti := 0; ti < len(textLower) && qi < len(queryLower); ti++ {
		if textLower[ti] == queryLower[qi] {
			score += 10
			if lastMatchIdx == ti-1 {
				consecutive++
				score += consecutive * 5
			} else {
				consecutive = 0
			}
			if ti == 0 || isBoundary(rune(text[ti-1])) {
				score += 8
			}
			score += (len(text) - ti)
			lastMatchIdx = ti
			qi++
		}
	}

	if qi < len(queryLower) {
		return 0
	}

	return score
}

func isBoundary(r rune) bool {
	return r == '-' || r == '_' || r == '.' || r == '/' || unicode.IsSpace(r)
}

// Match computes a weighted fuzzy match score across all host fields.
func Match(query string, h Host) int {
	best := 0
	fields := []struct {
		text   string
		weight int
	}{
		{h.ShortName, 5},
		{h.DescText, 3},
		{h.LongName, 3},
		{h.IP, 2},
		{strings.Join(h.Groups, " "), 2},
	}
	for _, f := range fields {
		if s := Score(query, f.text); s > 0 {
			weighted := s * f.weight
			if weighted > best {
				best = weighted
			}
		}
	}
	return best
}

// FilterHosts filters and sorts hosts by query.
func FilterHosts(query string, hosts []Host) []Host {
	if query == "" {
		result := make([]Host, len(hosts))
		copy(result, hosts)
		return result
	}

	isNumeric := true
	for _, r := range query {
		if !unicode.IsDigit(r) {
			isNumeric = false
			break
		}
	}

	if isNumeric {
		var result []Host
		for _, h := range hosts {
			menuStr := fmt.Sprintf("%d", h.MenuNumber)
			if strings.HasPrefix(menuStr, query) {
				result = append(result, h)
			}
		}
		return result
	}

	type scored struct {
		host  Host
		score int
	}
	var matches []scored
	for _, h := range hosts {
		if s := Match(query, h); s > 0 {
			matches = append(matches, scored{host: h, score: s})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})
	result := make([]Host, len(matches))
	for i, m := range matches {
		result[i] = m.host
	}
	return result
}
