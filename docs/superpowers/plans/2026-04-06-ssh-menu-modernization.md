# SSH Menu Modernization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restructure ssh-menu into focused packages with a two-pane UI, fuzzy search, pinned hosts, config validation, and comprehensive tests.

**Architecture:** Bottom-up build — data types first, then business logic (filtering, grouping, validation), then theme, then UI components, then wire it all together in main.go. Each task produces testable, committable work. Old `internal/` files are deleted in the final task.

**Tech Stack:** Go 1.24, Bubble Tea v1.3.10, Lipgloss v1.1.0

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `internal/host/host.go` | Host type, Warning type |
| Create | `internal/host/groups.go` | AssignMenuNumbers, GetAllGroups, SortWithPins |
| Create | `internal/host/groups_test.go` | Tests for above |
| Create | `internal/host/filter.go` | Fuzzy scoring, Match, FilterHosts |
| Create | `internal/host/filter_test.go` | Tests for above |
| Create | `internal/host/validate.go` | ValidateHosts |
| Create | `internal/host/validate_test.go` | Tests for above |
| Create | `internal/config/parser.go` | SSH config parsing |
| Create | `internal/config/parser_test.go` | Tests for above |
| Create | `internal/config/colors.go` | Color config from SSH config comments |
| Create | `internal/config/colors_test.go` | Tests for above |
| Create | `internal/config/pin.go` | TogglePin (write back to config file) |
| Create | `internal/config/pin_test.go` | Tests for above |
| Create | `internal/theme/theme.go` | Consolidated color system, styles |
| Create | `internal/theme/theme_test.go` | Tests for above |
| Create | `internal/ui/keys.go` | Key bindings |
| Create | `internal/ui/model.go` | Top-level Bubble Tea model |
| Create | `internal/ui/hostlist.go` | Left pane component |
| Create | `internal/ui/detail.go` | Right pane component |
| Create | `internal/ui/viewbar.go` | Group tab bar |
| Rewrite | `main.go` | Updated to use new packages |
| Delete | `internal/config.go` | Replaced by config/ and theme/ |
| Delete | `internal/host.go` | Replaced by host/ |
| Delete | `internal/theme.go` | Replaced by theme/ |
| Delete | `internal/ui.go` | Replaced by ui/ |

---

### Task 1: Host Type and Warning Type

**Files:**
- Create: `internal/host/host.go`

- [ ] **Step 1: Create the host package with types**

```go
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/nick/Sync/ssh-menu && go build ./internal/host/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/host/host.go
git commit -m "feat: add host and warning types in host package"
```

---

### Task 2: Group Logic and Menu Numbering

**Files:**
- Create: `internal/host/groups.go`
- Create: `internal/host/groups_test.go`

- [ ] **Step 1: Write tests for AssignMenuNumbers**

```go
package host

import (
	"testing"
)

func TestAssignMenuNumbers_ExplicitNumbers(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", DescText: "desc", MenuNumber: 3},
		{ShortName: "b", DescText: "desc", MenuNumber: 1},
	}
	result, err := AssignMenuNumbers(hosts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be sorted by menu number
	if result[0].ShortName != "b" || result[0].MenuNumber != 1 {
		t.Errorf("expected first host to be 'b' with number 1, got %s with %d", result[0].ShortName, result[0].MenuNumber)
	}
	if result[1].ShortName != "a" || result[1].MenuNumber != 3 {
		t.Errorf("expected second host to be 'a' with number 3, got %s with %d", result[1].ShortName, result[1].MenuNumber)
	}
}

func TestAssignMenuNumbers_AutoAssign(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", DescText: "desc", MenuNumber: 2},
		{ShortName: "b", DescText: "desc", MenuNumber: 0},
		{ShortName: "c", DescText: "desc", MenuNumber: 0},
	}
	result, err := AssignMenuNumbers(hosts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// b and c should get numbers 1 and 3 (2 is taken)
	numbers := map[string]int{}
	for _, h := range result {
		numbers[h.ShortName] = h.MenuNumber
	}
	if numbers["a"] != 2 {
		t.Errorf("expected a=2, got %d", numbers["a"])
	}
	if numbers["b"] != 1 {
		t.Errorf("expected b=1, got %d", numbers["b"])
	}
	if numbers["c"] != 3 {
		t.Errorf("expected c=3, got %d", numbers["c"])
	}
}

func TestAssignMenuNumbers_DuplicateError(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", DescText: "desc", MenuNumber: 1},
		{ShortName: "b", DescText: "desc", MenuNumber: 1},
	}
	_, err := AssignMenuNumbers(hosts)
	if err == nil {
		t.Fatal("expected error for duplicate menu numbers")
	}
}

func TestGetAllGroups_SortedUngroupedLast(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", Groups: []string{"Zebra"}},
		{ShortName: "b", Groups: []string{"Alpha"}},
		{ShortName: "c", Groups: []string{}},
	}
	groups := GetAllGroups(hosts)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d: %v", len(groups), groups)
	}
	if groups[0] != "Alpha" {
		t.Errorf("expected first group Alpha, got %s", groups[0])
	}
	if groups[1] != "Zebra" {
		t.Errorf("expected second group Zebra, got %s", groups[1])
	}
	if groups[2] != "Ungrouped" {
		t.Errorf("expected last group Ungrouped, got %s", groups[2])
	}
}

func TestSortWithPins_PinnedFirst(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", MenuNumber: 1, Pinned: false},
		{ShortName: "b", MenuNumber: 2, Pinned: true},
		{ShortName: "c", MenuNumber: 3, Pinned: false},
	}
	result := SortWithPins(hosts)
	if result[0].ShortName != "b" {
		t.Errorf("expected pinned host first, got %s", result[0].ShortName)
	}
	if result[1].ShortName != "a" {
		t.Errorf("expected a second, got %s", result[1].ShortName)
	}
	if result[2].ShortName != "c" {
		t.Errorf("expected c third, got %s", result[2].ShortName)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/host/ -v`
Expected: FAIL — functions not defined

- [ ] **Step 3: Implement groups.go**

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/host/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/host/groups.go internal/host/groups_test.go
git commit -m "feat: add group logic and menu numbering with tests"
```

---

### Task 3: Fuzzy Search

**Files:**
- Create: `internal/host/filter.go`
- Create: `internal/host/filter_test.go`

- [ ] **Step 1: Write tests for fuzzy scoring and filtering**

```go
package host

import (
	"testing"
)

func TestScore_ExactSubstring(t *testing.T) {
	s := Score("prod", "web-prod-01")
	if s <= 0 {
		t.Errorf("expected positive score for substring match, got %d", s)
	}
}

func TestScore_NoMatch(t *testing.T) {
	s := Score("xyz", "web-prod-01")
	if s != 0 {
		t.Errorf("expected 0 for no match, got %d", s)
	}
}

func TestScore_ExactBeatsFuzzy(t *testing.T) {
	exact := Score("prod", "prod-server")
	fuzzy := Score("prod", "p-r-o-d-server")
	if fuzzy >= exact {
		t.Errorf("exact substring (%d) should beat scattered match (%d)", exact, fuzzy)
	}
}

func TestScore_BoundaryBeatsMiddle(t *testing.T) {
	boundary := Score("prod", "web-prod")
	middle := Score("prod", "reproduce")
	if middle >= boundary {
		t.Errorf("boundary match (%d) should beat mid-word match (%d)", boundary, middle)
	}
}

func TestScore_EmptyQuery(t *testing.T) {
	s := Score("", "anything")
	if s != 0 {
		t.Errorf("expected 0 for empty query, got %d", s)
	}
}

func TestMatch_SearchesAllFields(t *testing.T) {
	h := Host{
		ShortName: "web-01",
		DescText:  "production server",
		LongName:  "web01.example.com",
		IP:        "10.0.1.5",
		Groups:    []string{"Servers"},
	}
	// Should match on description
	s := Match("production", h)
	if s <= 0 {
		t.Errorf("expected match on description, got %d", s)
	}
	// Should match on IP
	s = Match("10.0", h)
	if s <= 0 {
		t.Errorf("expected match on IP, got %d", s)
	}
	// Should match on group
	s = Match("server", h)
	if s <= 0 {
		t.Errorf("expected match on group, got %d", s)
	}
}

func TestFilterHosts_NumericPrefixMatchesMenuNumber(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", MenuNumber: 12, DescText: "desc"},
		{ShortName: "b", MenuNumber: 3, DescText: "desc"},
		{ShortName: "c", MenuNumber: 123, DescText: "desc"},
	}
	result := FilterHosts("12", hosts)
	// Should match menu numbers 12 and 123
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
}

func TestFilterHosts_EmptyQueryReturnsAll(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", MenuNumber: 1},
		{ShortName: "b", MenuNumber: 2},
	}
	result := FilterHosts("", hosts)
	if len(result) != 2 {
		t.Errorf("expected all hosts returned, got %d", len(result))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/host/ -v -run "TestScore|TestMatch|TestFilter"`
Expected: FAIL — functions not defined

- [ ] **Step 3: Implement filter.go**

```go
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
		score := 100
		// Bonus for match at start
		if idx == 0 {
			score += 20
		}
		// Bonus for boundary match
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
			// Bonus for consecutive matches
			if lastMatchIdx == ti-1 {
				consecutive++
				score += consecutive * 5
			} else {
				consecutive = 0
			}
			// Bonus for boundary match
			if ti == 0 || isBoundary(rune(text[ti-1])) {
				score += 8
			}
			// Bonus for earlier matches
			score += (len(text) - ti)
			lastMatchIdx = ti
			qi++
		}
	}

	// All query characters must be found
	if qi < len(queryLower) {
		return 0
	}

	return score
}

// isBoundary returns true if the character is a word boundary.
func isBoundary(r rune) bool {
	return r == '-' || r == '_' || r == '.' || r == '/' || unicode.IsSpace(r)
}

// Match computes a weighted fuzzy match score across all host fields.
func Match(query string, h Host) int {
	best := 0

	// Fields in priority order with weights
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
// Pure numeric queries match menu number prefixes.
// Everything else uses fuzzy matching.
func FilterHosts(query string, hosts []Host) []Host {
	if query == "" {
		result := make([]Host, len(hosts))
		copy(result, hosts)
		return result
	}

	// Check if query is purely numeric
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

	// Fuzzy match
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/host/ -v -run "TestScore|TestMatch|TestFilter"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/host/filter.go internal/host/filter_test.go
git commit -m "feat: add fuzzy search with weighted multi-field matching"
```

---

### Task 4: Host Validation

**Files:**
- Create: `internal/host/validate.go`
- Create: `internal/host/validate_test.go`

- [ ] **Step 1: Write tests for validation**

```go
package host

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateHosts_MissingIdentityFile(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", IdentityFile: "/nonexistent/path/key"},
	}
	result := ValidateHosts(hosts)
	if len(result[0].Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result[0].Warnings))
	}
	if result[0].Warnings[0].Level != "warn" {
		t.Errorf("expected warn level, got %s", result[0].Warnings[0].Level)
	}
}

func TestValidateHosts_ExistingIdentityFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key")
	os.WriteFile(keyPath, []byte("fake key"), 0600)

	hosts := []Host{
		{ShortName: "a", IdentityFile: keyPath},
	}
	result := ValidateHosts(hosts)
	if len(result[0].Warnings) != 0 {
		t.Errorf("expected no warnings, got %v", result[0].Warnings)
	}
}

func TestValidateHosts_TildeExpansion(t *testing.T) {
	// Host with ~ in identity file that doesn't exist after expansion
	hosts := []Host{
		{ShortName: "a", IdentityFile: "~/.ssh/nonexistent_key_12345"},
	}
	result := ValidateHosts(hosts)
	if len(result[0].Warnings) != 1 {
		t.Fatalf("expected 1 warning for missing key, got %d", len(result[0].Warnings))
	}
}

func TestValidateHosts_DuplicateAliases(t *testing.T) {
	hosts := []Host{
		{ShortName: "server-a", SourceFile: "/etc/ssh/config"},
		{ShortName: "server-a", SourceFile: "/etc/ssh/config.d/extra"},
	}
	result := ValidateHosts(hosts)
	// Both should get a warning
	if len(result[0].Warnings) != 1 || len(result[1].Warnings) != 1 {
		t.Errorf("expected 1 warning each for duplicates, got %d and %d",
			len(result[0].Warnings), len(result[1].Warnings))
	}
}

func TestValidateHosts_EmptyHostname(t *testing.T) {
	hosts := []Host{
		{ShortName: "not-an-ip", LongName: ""},
	}
	result := ValidateHosts(hosts)
	if len(result[0].Warnings) != 1 {
		t.Fatalf("expected 1 warning for empty hostname, got %d", len(result[0].Warnings))
	}
}

func TestValidateHosts_EmptyHostnameSkippedForFQDN(t *testing.T) {
	hosts := []Host{
		{ShortName: "server.example.com", LongName: ""},
	}
	result := ValidateHosts(hosts)
	if len(result[0].Warnings) != 0 {
		t.Errorf("expected no warnings for FQDN alias, got %v", result[0].Warnings)
	}
}

func TestValidateHosts_NoIdentityFileNoWarning(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", IdentityFile: ""},
	}
	result := ValidateHosts(hosts)
	if len(result[0].Warnings) != 0 {
		t.Errorf("expected no warnings when no identity file set, got %v", result[0].Warnings)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/host/ -v -run TestValidate`
Expected: FAIL — ValidateHosts not defined

- [ ] **Step 3: Implement validate.go**

```go
package host

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// ValidateHosts runs validation checks on all hosts and populates their Warnings field.
func ValidateHosts(hosts []Host) []Host {
	checkIdentityFiles(hosts)
	checkDuplicateAliases(hosts)
	checkEmptyHostnames(hosts)
	return hosts
}

// checkIdentityFiles warns if an identity file doesn't exist on disk.
func checkIdentityFiles(hosts []Host) {
	for i := range hosts {
		if hosts[i].IdentityFile == "" {
			continue
		}
		path := expandTilde(hosts[i].IdentityFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			hosts[i].Warnings = append(hosts[i].Warnings, Warning{
				Level:   "warn",
				Message: fmt.Sprintf("Identity file not found: %s", hosts[i].IdentityFile),
			})
		}
	}
}

// checkDuplicateAliases warns if the same ShortName appears in multiple files.
func checkDuplicateAliases(hosts []Host) {
	seen := make(map[string][]int) // ShortName -> indices
	for i, h := range hosts {
		seen[h.ShortName] = append(seen[h.ShortName], i)
	}
	for name, indices := range seen {
		if len(indices) <= 1 {
			continue
		}
		// Only warn if they come from different files
		files := make(map[string]bool)
		for _, idx := range indices {
			files[hosts[idx].SourceFile] = true
		}
		if len(files) > 1 {
			for _, idx := range indices {
				hosts[idx].Warnings = append(hosts[idx].Warnings, Warning{
					Level:   "warn",
					Message: fmt.Sprintf("Duplicate host alias '%s' found in multiple files", name),
				})
			}
		}
	}
}

// checkEmptyHostnames warns if a host has no HostName and the alias isn't an IP or FQDN.
func checkEmptyHostnames(hosts []Host) {
	for i := range hosts {
		if hosts[i].LongName != "" {
			continue
		}
		// If the alias looks like an IP or FQDN, SSH will use it directly
		if net.ParseIP(hosts[i].ShortName) != nil || strings.Contains(hosts[i].ShortName, ".") {
			continue
		}
		hosts[i].Warnings = append(hosts[i].Warnings, Warning{
			Level:   "warn",
			Message: fmt.Sprintf("No HostName set for '%s'", hosts[i].ShortName),
		})
	}
}

// expandTilde replaces a leading ~ with the user's home directory.
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return strings.Replace(path, "~", home, 1)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/host/ -v -run TestValidate`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/host/validate.go internal/host/validate_test.go
git commit -m "feat: add SSH config validation with identity file and alias checks"
```

---

### Task 5: Config Parser

**Files:**
- Create: `internal/config/parser.go`
- Create: `internal/config/parser_test.go`

- [ ] **Step 1: Write tests for config parsing**

```go
package config

import (
	"strings"
	"testing"
)

func TestParseReader_BasicHost(t *testing.T) {
	input := `# Menu 1: Web server
Host web-01
    HostName 10.0.1.5
    User admin
    Port 2222
    IdentityFile ~/.ssh/web_key
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	h := hosts[0]
	if h.ShortName != "web-01" {
		t.Errorf("ShortName: expected web-01, got %s", h.ShortName)
	}
	if h.LongName != "10.0.1.5" {
		t.Errorf("LongName: expected 10.0.1.5, got %s", h.LongName)
	}
	if h.User != "admin" {
		t.Errorf("User: expected admin, got %s", h.User)
	}
	if h.Port != "2222" {
		t.Errorf("Port: expected 2222, got %s", h.Port)
	}
	if h.IdentityFile != "~/.ssh/web_key" {
		t.Errorf("IdentityFile: expected ~/.ssh/web_key, got %s", h.IdentityFile)
	}
	if h.MenuNumber != 1 {
		t.Errorf("MenuNumber: expected 1, got %d", h.MenuNumber)
	}
	if h.DescText != "Web server" {
		t.Errorf("DescText: expected 'Web server', got '%s'", h.DescText)
	}
	if h.SourceFile != "test.config" {
		t.Errorf("SourceFile: expected test.config, got %s", h.SourceFile)
	}
}

func TestParseReader_MenuWithoutNumber(t *testing.T) {
	input := `# Menu: Auto-numbered host
Host auto-01
    HostName 10.0.1.1
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].MenuNumber != 0 {
		t.Errorf("expected MenuNumber 0 for auto, got %d", hosts[0].MenuNumber)
	}
	if hosts[0].DescText != "Auto-numbered host" {
		t.Errorf("expected desc 'Auto-numbered host', got '%s'", hosts[0].DescText)
	}
}

func TestParseReader_GroupsAndIP(t *testing.T) {
	input := `# Menu: Server
# Group: Production
# Group: Web
# IP: 203.0.113.50
Host web-prod
    HostName 10.0.1.5
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h := hosts[0]
	if len(h.Groups) != 2 || h.Groups[0] != "Production" || h.Groups[1] != "Web" {
		t.Errorf("Groups: expected [Production, Web], got %v", h.Groups)
	}
	if h.IP != "203.0.113.50" {
		t.Errorf("IP: expected 203.0.113.50, got %s", h.IP)
	}
}

func TestParseReader_Pinned(t *testing.T) {
	input := `# Menu: Pinned host
# Pinned
Host pinned-01
    HostName 10.0.1.1
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hosts[0].Pinned {
		t.Error("expected host to be pinned")
	}
}

func TestParseReader_SkipsHostsWithoutMenu(t *testing.T) {
	input := `Host no-menu
    HostName 10.0.1.1

# Menu: Has menu
Host has-menu
    HostName 10.0.1.2
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host (with menu), got %d", len(hosts))
	}
	if hosts[0].ShortName != "has-menu" {
		t.Errorf("expected has-menu, got %s", hosts[0].ShortName)
	}
}

func TestParseReader_NoDefaultUserOrPort(t *testing.T) {
	input := `# Menu: Minimal
Host minimal
    HostName 10.0.1.1
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hosts[0].User != "" {
		t.Errorf("expected empty User, got '%s'", hosts[0].User)
	}
	if hosts[0].Port != "" {
		t.Errorf("expected empty Port, got '%s'", hosts[0].Port)
	}
}

func TestParseReader_MultipleHosts(t *testing.T) {
	input := `# Menu 1: First
Host first
    HostName 10.0.1.1

# Menu 2: Second
Host second
    HostName 10.0.1.2
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}
}

func TestParseReader_EmptyInput(t *testing.T) {
	hosts, err := ParseReader(strings.NewReader(""), "test.config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 0 {
		t.Errorf("expected 0 hosts from empty input, got %d", len(hosts))
	}
}

func TestParseReader_WildcardHostSkipped(t *testing.T) {
	input := `# Menu: Wildcard
Host *
    ServerAliveInterval 60

# Menu: Real host
Host real
    HostName 10.0.1.1
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host (skipping wildcard), got %d", len(hosts))
	}
	if hosts[0].ShortName != "real" {
		t.Errorf("expected 'real', got '%s'", hosts[0].ShortName)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/config/ -v`
Expected: FAIL — ParseReader not defined

- [ ] **Step 3: Implement parser.go**

```go
package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/evix1101/ssh-menu/internal/host"
)

var (
	reHost     = regexp.MustCompile(`^Host\s+(.+)$`)
	reHostname = regexp.MustCompile(`(?i)^Hostname\s+(.+)$`)
	reUser     = regexp.MustCompile(`^User\s+(.+)$`)
	rePort     = regexp.MustCompile(`^Port\s+(\d+)$`)
	reIdentity = regexp.MustCompile(`^IdentityFile\s+(.+)$`)
	reMenu     = regexp.MustCompile(`^#\s*Menu(?:\s+(\d+))?:\s*(.+)$`)
	reIP       = regexp.MustCompile(`^#\s*IP:\s*(.+)$`)
	reGroup    = regexp.MustCompile(`^#\s*Group:\s*(.+)$`)
	rePinned   = regexp.MustCompile(`^#\s*Pinned\s*$`)
)

// ParseReader parses SSH config from a reader and returns host entries.
// sourceFile is stored on each host for pin toggle write-back.
func ParseReader(r io.Reader, sourceFile string) ([]host.Host, error) {
	var hosts []host.Host
	current := host.Host{SourceFile: sourceFile}
	inHost := false

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Host line starts a new block
		if m := reHost.FindStringSubmatch(line); m != nil {
			// Save previous host if valid
			if inHost && current.ShortName != "" && current.DescText != "" {
				hosts = append(hosts, current)
			}
			hostName := strings.TrimSpace(m[1])
			// Skip wildcard hosts
			if hostName == "*" {
				current = host.Host{SourceFile: sourceFile}
				inHost = false
				continue
			}
			current = host.Host{
				ShortName:  hostName,
				Groups:     []string{},
				SourceFile: sourceFile,
			}
			inHost = true
			continue
		}

		// Comment lines can appear before or within a host block
		if m := reMenu.FindStringSubmatch(line); m != nil {
			if m[1] != "" {
				num, err := strconv.Atoi(m[1])
				if err != nil {
					return nil, fmt.Errorf("invalid menu number: %s", m[1])
				}
				current.MenuNumber = num
			}
			current.DescText = strings.TrimSpace(m[2])
			continue
		}
		if m := reIP.FindStringSubmatch(line); m != nil {
			current.IP = strings.TrimSpace(m[1])
			continue
		}
		if m := reGroup.FindStringSubmatch(line); m != nil {
			g := strings.TrimSpace(m[1])
			if !sliceContains(current.Groups, g) {
				current.Groups = append(current.Groups, g)
			}
			continue
		}
		if rePinned.MatchString(line) {
			current.Pinned = true
			continue
		}

		// SSH config directives (only within a host block)
		if !inHost {
			continue
		}
		if m := reHostname.FindStringSubmatch(line); m != nil {
			current.LongName = strings.TrimSpace(m[1])
		} else if m := reUser.FindStringSubmatch(line); m != nil {
			current.User = strings.TrimSpace(m[1])
		} else if m := rePort.FindStringSubmatch(line); m != nil {
			current.Port = strings.TrimSpace(m[1])
		} else if m := reIdentity.FindStringSubmatch(line); m != nil {
			current.IdentityFile = strings.TrimSpace(m[1])
		}
	}

	// Append last host
	if inHost && current.ShortName != "" && current.DescText != "" {
		hosts = append(hosts, current)
	}

	return hosts, nil
}

// ReadConfigFiles reads all SSH config files (main + config.d).
func ReadConfigFiles(configPath string) ([]host.Host, error) {
	mainHosts, err := parseFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading main config: %w", err)
	}

	configDirPath := filepath.Join(filepath.Dir(configPath), "config.d")
	dirInfo, err := os.Stat(configDirPath)
	if os.IsNotExist(err) || (err == nil && !dirInfo.IsDir()) {
		return mainHosts, nil
	} else if err != nil {
		return nil, fmt.Errorf("error checking config.d: %w", err)
	}

	files, err := os.ReadDir(configDirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config.d: %w", err)
	}

	allHosts := mainHosts
	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}
		filePath := filepath.Join(configDirPath, file.Name())
		additional, err := parseFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Error reading config file %s: %v\n", filePath, err)
			continue
		}
		allHosts = append(allHosts, additional...)
	}

	return allHosts, nil
}

func parseFile(path string) ([]host.Host, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseReader(f, path)
}

func sliceContains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/config/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/parser.go internal/config/parser_test.go
git commit -m "feat: add SSH config parser with reader-based API and source tracking"
```

---

### Task 6: Color Config Parsing

**Files:**
- Create: `internal/config/colors.go`
- Create: `internal/config/colors_test.go`

- [ ] **Step 1: Write tests for color config**

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/config/ -v -run TestParseColors`
Expected: FAIL — ParseColors not defined

- [ ] **Step 3: Implement colors.go**

```go
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
// Returns a ColorConfig with only the fields found in the input populated.
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/config/ -v -run TestParseColors`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/colors.go internal/config/colors_test.go
git commit -m "feat: add color config parser for SSH config comments"
```

---

### Task 7: Pin Toggle (Write Back to Config)

**Files:**
- Create: `internal/config/pin.go`
- Create: `internal/config/pin_test.go`

- [ ] **Step 1: Write tests for pin toggling**

```go
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTogglePin_AddPin(t *testing.T) {
	content := `# Menu 1: Web server
# Group: Production
Host web-01
    HostName 10.0.1.5
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config")
	os.WriteFile(path, []byte(content), 0644)

	err := TogglePin(path, "web-01", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, _ := os.ReadFile(path)
	if !strings.Contains(string(result), "# Pinned") {
		t.Error("expected # Pinned to be added")
	}
	// Should be after Group line, before Host line
	lines := strings.Split(string(result), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "# Pinned" {
			if i < 2 {
				t.Error("# Pinned should be after Group line")
			}
			if i > 0 && strings.HasPrefix(strings.TrimSpace(lines[i-1]), "Host ") {
				t.Error("# Pinned should be before Host line")
			}
			break
		}
	}
}

func TestTogglePin_RemovePin(t *testing.T) {
	content := `# Menu 1: Web server
# Pinned
Host web-01
    HostName 10.0.1.5
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config")
	os.WriteFile(path, []byte(content), 0644)

	err := TogglePin(path, "web-01", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, _ := os.ReadFile(path)
	if strings.Contains(string(result), "# Pinned") {
		t.Error("expected # Pinned to be removed")
	}
}

func TestTogglePin_PreservesFormatting(t *testing.T) {
	content := `# Menu 1: Web server
Host web-01
    HostName 10.0.1.5
    User admin

# Menu 2: DB server
Host db-01
    HostName 10.0.2.5
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config")
	os.WriteFile(path, []byte(content), 0644)

	err := TogglePin(path, "web-01", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, _ := os.ReadFile(path)
	// db-01 block should be unchanged
	if !strings.Contains(string(result), "# Menu 2: DB server\nHost db-01") {
		t.Error("other host blocks should be preserved")
	}
}

func TestTogglePin_HostNotFound(t *testing.T) {
	content := `# Menu 1: Web server
Host web-01
    HostName 10.0.1.5
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config")
	os.WriteFile(path, []byte(content), 0644)

	err := TogglePin(path, "nonexistent", true)
	if err == nil {
		t.Error("expected error for host not found")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/config/ -v -run TestTogglePin`
Expected: FAIL — TogglePin not defined

- [ ] **Step 3: Implement pin.go**

```go
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var reHostLine = regexp.MustCompile(`^Host\s+(.+)$`)
var rePinnedLine = regexp.MustCompile(`^#\s*Pinned\s*$`)
var reMenuLine = regexp.MustCompile(`^#\s*Menu`)
var reGroupLine = regexp.MustCompile(`^#\s*Group:`)

// TogglePin adds or removes a # Pinned comment for a host in a config file.
func TogglePin(filePath string, hostAlias string, pin bool) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	hostLineIdx := -1

	// Find the Host line for this alias
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if m := reHostLine.FindStringSubmatch(trimmed); m != nil {
			if strings.TrimSpace(m[1]) == hostAlias {
				hostLineIdx = i
				break
			}
		}
	}

	if hostLineIdx == -1 {
		return fmt.Errorf("host '%s' not found in %s", hostAlias, filePath)
	}

	if pin {
		lines = addPinComment(lines, hostLineIdx)
	} else {
		lines = removePinComment(lines, hostLineIdx)
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

// addPinComment inserts a # Pinned line before the Host line,
// after any # Menu or # Group comments.
func addPinComment(lines []string, hostLineIdx int) []string {
	// Check if already pinned
	for i := hostLineIdx - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if rePinnedLine.MatchString(trimmed) {
			return lines // Already pinned
		}
		if trimmed == "" || reHostLine.MatchString(trimmed) {
			break
		}
	}

	// Find insertion point: after the last Menu/Group/IP comment before the Host line
	insertIdx := hostLineIdx
	for i := hostLineIdx - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if reMenuLine.MatchString(trimmed) || reGroupLine.MatchString(trimmed) ||
			strings.HasPrefix(trimmed, "# IP:") {
			insertIdx = i + 1
			break
		}
		if trimmed == "" || reHostLine.MatchString(trimmed) {
			break
		}
	}

	// Insert # Pinned
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, "# Pinned")
	newLines = append(newLines, lines[insertIdx:]...)

	return newLines
}

// removePinComment removes the # Pinned line associated with a host.
func removePinComment(lines []string, hostLineIdx int) []string {
	for i := hostLineIdx - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if rePinnedLine.MatchString(trimmed) {
			return append(lines[:i], lines[i+1:]...)
		}
		if trimmed == "" || reHostLine.MatchString(trimmed) {
			break
		}
	}
	return lines // No pinned comment found
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/config/ -v -run TestTogglePin`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/pin.go internal/config/pin_test.go
git commit -m "feat: add pin toggle that writes back to SSH config files"
```

---

### Task 8: Theme System

**Files:**
- Create: `internal/theme/theme.go`
- Create: `internal/theme/theme_test.go`

- [ ] **Step 1: Write tests for theme**

```go
package theme

import (
	"os"
	"testing"

	"github.com/evix1101/ssh-menu/internal/config"
)

func TestDefaultColors(t *testing.T) {
	c := DefaultColors()
	if c.Background != "#1e1e2e" {
		t.Errorf("expected default background #1e1e2e, got %s", c.Background)
	}
	if c.Selected != "#a6e3a1" {
		t.Errorf("expected default selected #a6e3a1, got %s", c.Selected)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	os.Setenv("SSH_MENU_COLOR_BACKGROUND", "#ff0000")
	defer os.Unsetenv("SSH_MENU_COLOR_BACKGROUND")

	c := DefaultColors()
	ApplyEnvOverrides(&c)
	if c.Background != "#ff0000" {
		t.Errorf("expected env override #ff0000, got %s", c.Background)
	}
	// Other colors should remain default
	if c.Foreground != "#cdd6f4" {
		t.Errorf("foreground should remain default, got %s", c.Foreground)
	}
}

func TestMergeConfigColors_OnlyOverridesDefaults(t *testing.T) {
	c := DefaultColors()
	// Simulate env override for background
	c.Background = "#custom"

	fileColors := config.ColorConfig{
		Background: "#fromfile",
		Accent:     "#accentfile",
	}

	MergeConfigColors(&c, fileColors)
	// Background was already customized — should NOT be overridden by file
	if c.Background != "#custom" {
		t.Errorf("expected custom bg preserved, got %s", c.Background)
	}
	// Accent was still default — should be overridden by file
	if c.Accent != "#accentfile" {
		t.Errorf("expected accent from file, got %s", c.Accent)
	}
}

func TestInit_FallsBackToDefaults(t *testing.T) {
	// Clear any env vars
	for _, env := range []string{
		"SSH_MENU_COLOR_BACKGROUND", "SSH_MENU_COLOR_FOREGROUND",
		"SSH_MENU_COLOR_BORDER", "SSH_MENU_COLOR_SELECTED",
		"SSH_MENU_COLOR_ACCENT", "SSH_MENU_COLOR_DIMMED",
	} {
		os.Unsetenv(env)
	}

	Init("/nonexistent/path")
	c := Current()
	defaults := DefaultColors()
	if c.Background != defaults.Background {
		t.Errorf("expected default bg, got %s", c.Background)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/theme/ -v`
Expected: FAIL — functions not defined

- [ ] **Step 3: Implement theme.go**

```go
package theme

import (
	"os"

	"github.com/evix1101/ssh-menu/internal/config"
	"github.com/charmbracelet/lipgloss"
)

// Colors holds the resolved color configuration for the UI.
type Colors = config.ColorConfig

var current Colors

// DefaultColors returns the default Catppuccin Mocha-inspired color scheme.
func DefaultColors() Colors {
	return Colors{
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Border:     "#9399b2",
		Selected:   "#a6e3a1",
		Accent:     "#89dceb",
		Dimmed:     "#585b70",
	}
}

var envVars = []struct {
	envKey string
	field  func(*Colors) *string
}{
	{"SSH_MENU_COLOR_BACKGROUND", func(c *Colors) *string { return &c.Background }},
	{"SSH_MENU_COLOR_FOREGROUND", func(c *Colors) *string { return &c.Foreground }},
	{"SSH_MENU_COLOR_BORDER", func(c *Colors) *string { return &c.Border }},
	{"SSH_MENU_COLOR_SELECTED", func(c *Colors) *string { return &c.Selected }},
	{"SSH_MENU_COLOR_ACCENT", func(c *Colors) *string { return &c.Accent }},
	{"SSH_MENU_COLOR_DIMMED", func(c *Colors) *string { return &c.Dimmed }},
}

// ApplyEnvOverrides applies environment variable colors to the config.
func ApplyEnvOverrides(c *Colors) {
	for _, ev := range envVars {
		if val := os.Getenv(ev.envKey); val != "" {
			*ev.field(c) = val
		}
	}
}

// MergeConfigColors merges file-based colors into the config,
// only overriding fields that are still at their default values.
func MergeConfigColors(c *Colors, fileColors config.ColorConfig) {
	defaults := DefaultColors()
	merge := func(current *string, defaultVal, fileVal string) {
		if *current == defaultVal && fileVal != "" {
			*current = fileVal
		}
	}
	merge(&c.Background, defaults.Background, fileColors.Background)
	merge(&c.Foreground, defaults.Foreground, fileColors.Foreground)
	merge(&c.Border, defaults.Border, fileColors.Border)
	merge(&c.Selected, defaults.Selected, fileColors.Selected)
	merge(&c.Accent, defaults.Accent, fileColors.Accent)
	merge(&c.Dimmed, defaults.Dimmed, fileColors.Dimmed)
}

// Init loads color configuration from env vars and config file.
func Init(configPath string) {
	current = DefaultColors()
	ApplyEnvOverrides(&current)

	if configPath != "" {
		f, err := os.Open(configPath)
		if err == nil {
			defer f.Close()
			fileColors := config.ParseColors(f)
			MergeConfigColors(&current, fileColors)
		}
	}
}

// Current returns the active color configuration.
func Current() Colors {
	if current.Foreground == "" {
		return DefaultColors()
	}
	return current
}

// Convenience style constructors

// TitleStyle returns a style for the title bar.
func TitleStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c.Accent))
}

// SelectedStyle returns a style for selected items.
func SelectedStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(c.Selected))
}

// NormalStyle returns a style for normal text.
func NormalStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Foreground))
}

// DimStyle returns a style for dimmed text.
func DimStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Dimmed))
}

// WarningStyle returns a style for warning text.
func WarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af"))
}

// ActiveTabStyle returns a style for active view tabs.
func ActiveTabStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(c.Background)).
		Background(lipgloss.Color(c.Selected)).
		Padding(0, 1)
}

// InactiveTabStyle returns a style for inactive view tabs.
func InactiveTabStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Foreground)).
		Padding(0, 1)
}

// BorderStyle returns a style for borders.
func BorderStyle() lipgloss.Style {
	c := Current()
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Border))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/nick/Sync/ssh-menu && go test ./internal/theme/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/theme/theme.go internal/theme/theme_test.go
git commit -m "feat: add consolidated theme system with env, config, and default precedence"
```

---

### Task 9: UI Key Bindings

**Files:**
- Create: `internal/ui/keys.go`

- [ ] **Step 1: Create key bindings**

```go
package ui

import "github.com/charmbracelet/bubbletea"

// Key represents a key binding with its key type and rune value.
type keyAction int

const (
	keyQuit keyAction = iota
	keySelect
	keyUp
	keyDown
	keyLeft
	keyRight
	keyTab
	keyBackspace
	keyTogglePin
	keyRune
)

// classifyKey maps a tea.KeyMsg to a keyAction.
func classifyKey(msg bubbletea.KeyMsg) keyAction {
	switch msg.Type {
	case bubbletea.KeyEscape:
		return keyQuit
	case bubbletea.KeyEnter:
		return keySelect
	case bubbletea.KeyUp:
		return keyUp
	case bubbletea.KeyDown:
		return keyDown
	case bubbletea.KeyLeft:
		return keyLeft
	case bubbletea.KeyRight:
		return keyRight
	case bubbletea.KeyTab:
		return keyTab
	case bubbletea.KeyBackspace:
		return keyBackspace
	case bubbletea.KeyRunes:
		if msg.String() == "p" {
			return keyTogglePin
		}
		return keyRune
	}
	return keyRune
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/nick/Sync/ssh-menu && go build ./internal/ui/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/ui/keys.go
git commit -m "feat: add centralized key binding classification"
```

---

### Task 10: UI View Bar Component

**Files:**
- Create: `internal/ui/viewbar.go`

- [ ] **Step 1: Create the view bar**

```go
package ui

import (
	"strings"

	"github.com/evix1101/ssh-menu/internal/theme"
)

// renderViewBar renders the group tab bar.
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/nick/Sync/ssh-menu && go build ./internal/ui/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/ui/viewbar.go
git commit -m "feat: add group tab bar component"
```

---

### Task 11: UI Detail Panel Component

**Files:**
- Create: `internal/ui/detail.go`

- [ ] **Step 1: Create the detail panel**

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/evix1101/ssh-menu/internal/host"
	"github.com/evix1101/ssh-menu/internal/theme"
)

// renderDetail renders the right-pane detail panel for a host.
func renderDetail(h host.Host, width, height int) string {
	if width < 20 {
		return ""
	}

	colors := theme.Current()
	var b strings.Builder

	// Host alias (bold, accent)
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.Accent))
	b.WriteString(nameStyle.Render(h.ShortName))
	b.WriteString("\n")

	// Description
	if h.DescText != "" {
		b.WriteString(theme.NormalStyle().Render(h.DescText))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Connection details — only show fields that are explicitly set
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

	// Groups
	if len(h.Groups) > 0 {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%s  %s\n",
			labelStyle.Render("Groups:"),
			valueStyle.Render(strings.Join(h.Groups, ", "))))
	}

	// Pinned status
	if h.Pinned {
		b.WriteString("\n")
		pinStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Selected))
		b.WriteString(pinStyle.Render("★ Pinned"))
		b.WriteString("\n")
	}

	// Warnings
	if len(h.Warnings) > 0 {
		b.WriteString("\n")
		warnStyle := theme.WarningStyle()
		for _, w := range h.Warnings {
			b.WriteString(warnStyle.Render(fmt.Sprintf("⚠ %s", w.Message)))
			b.WriteString("\n")
		}
	}

	// Wrap in a bordered box
	panel := lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		Padding(1, 2)

	return panel.Render(b.String())
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/nick/Sync/ssh-menu && go build ./internal/ui/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/ui/detail.go
git commit -m "feat: add host detail panel component"
```

---

### Task 12: UI Host List Component

**Files:**
- Create: `internal/ui/hostlist.go`

- [ ] **Step 1: Create the scrollable host list**

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/evix1101/ssh-menu/internal/host"
	"github.com/evix1101/ssh-menu/internal/theme"
)

// renderHostList renders the left-pane scrollable host list.
// Returns the rendered string. cursor is the selected index.
// scrollOffset is managed by the caller.
func renderHostList(hosts []host.Host, cursor, scrollOffset, width, height int) string {
	if len(hosts) == 0 {
		return theme.DimStyle().Render("No hosts match your filter")
	}

	colors := theme.Current()
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.Selected))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Foreground))

	var b strings.Builder

	// Determine visible range
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

		// Truncate if needed
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

// calculateScrollOffset returns the scroll offset to keep cursor visible.
func calculateScrollOffset(cursor, currentOffset, visibleHeight, totalItems int) int {
	if visibleHeight <= 0 || totalItems <= visibleHeight {
		return 0
	}

	offset := currentOffset

	// Cursor above visible area
	if cursor < offset {
		offset = cursor
	}

	// Cursor below visible area
	if cursor >= offset+visibleHeight {
		offset = cursor - visibleHeight + 1
	}

	// Clamp
	maxOffset := totalItems - visibleHeight
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset < 0 {
		offset = 0
	}

	return offset
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/nick/Sync/ssh-menu && go build ./internal/ui/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/ui/hostlist.go
git commit -m "feat: add scrollable host list component"
```

---

### Task 13: UI Model (Top-Level Bubble Tea)

**Files:**
- Create: `internal/ui/model.go`

- [ ] **Step 1: Create the main UI model**

```go
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evix1101/ssh-menu/internal/config"
	"github.com/evix1101/ssh-menu/internal/host"
	"github.com/evix1101/ssh-menu/internal/theme"
)

const minWidthForTwoPane = 60

// Model is the top-level Bubble Tea model.
type Model struct {
	hosts         []host.Host
	Selected      *host.Host
	PinToggled    bool // true if a pin was toggled during this session
	verbose       bool
	sshOpts       string
	cursor        int
	scrollOffset  int
	viewIndex     int
	groups        []string
	filteredHosts []host.Host
	filterText    string
	width         int
	height        int
}

// New creates a new UI model.
func New(hosts []host.Host, verbose bool, sshOpts string) *Model {
	m := &Model{
		hosts:   hosts,
		verbose: verbose,
		sshOpts: sshOpts,
		groups:  host.GetAllGroups(hosts),
	}
	m.updateFilteredHosts()
	return m
}

// Run starts the Bubble Tea program.
func Run(m *Model) error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	if fm, ok := finalModel.(*Model); ok {
		m.Selected = fm.Selected
		m.PinToggled = fm.PinToggled
	}
	return nil
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return tea.WindowSize()
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When filter is active and user types 'p', it should be a filter character
	// Only treat 'p' as pin toggle when filter is empty
	action := classifyKey(msg)

	// Override: if filterText is non-empty, 'p' is just a character
	if action == keyTogglePin && m.filterText != "" {
		action = keyRune
	}

	switch action {
	case keyQuit:
		return m, tea.Quit
	case keySelect:
		return m.selectHost()
	case keyUp:
		m.moveCursor(-1)
	case keyDown:
		m.moveCursor(1)
	case keyLeft:
		m.navigateView(-1)
	case keyRight:
		m.navigateView(1)
	case keyTab:
		m.navigateView(1)
	case keyBackspace:
		m.handleBackspace()
	case keyTogglePin:
		m.togglePin()
	case keyRune:
		m.filterText += msg.String()
		m.updateFilteredHosts()
		m.cursor = 0
		m.scrollOffset = 0
	}
	return m, nil
}

func (m *Model) selectHost() (tea.Model, tea.Cmd) {
	if len(m.filteredHosts) == 1 {
		m.Selected = &m.filteredHosts[0]
		return m, tea.Quit
	}
	if len(m.filteredHosts) > 0 && m.cursor < len(m.filteredHosts) {
		m.Selected = &m.filteredHosts[m.cursor]
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.filteredHosts) {
		m.cursor = len(m.filteredHosts) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) navigateView(delta int) {
	totalViews := 1 + len(m.groups)
	m.viewIndex += delta
	if m.viewIndex < 0 {
		m.viewIndex = totalViews - 1
	} else if m.viewIndex >= totalViews {
		m.viewIndex = 0
	}
	m.filterText = ""
	m.cursor = 0
	m.scrollOffset = 0
	m.updateFilteredHosts()
}

func (m *Model) handleBackspace() {
	if len(m.filterText) > 0 {
		m.filterText = m.filterText[:len(m.filterText)-1]
		m.updateFilteredHosts()
		m.cursor = 0
		m.scrollOffset = 0
	}
}

func (m *Model) togglePin() {
	if len(m.filteredHosts) == 0 || m.cursor >= len(m.filteredHosts) {
		return
	}
	selected := &m.filteredHosts[m.cursor]
	newPinState := !selected.Pinned

	// Write to config file
	if selected.SourceFile != "" {
		if err := config.TogglePin(selected.SourceFile, selected.ShortName, newPinState); err != nil {
			return // Silently fail — the detail panel won't update
		}
	}

	// Update all references to this host in our data
	for i := range m.hosts {
		if m.hosts[i].ShortName == selected.ShortName && m.hosts[i].SourceFile == selected.SourceFile {
			m.hosts[i].Pinned = newPinState
		}
	}

	m.PinToggled = true
	m.updateFilteredHosts()
}

func (m *Model) updateFilteredHosts() {
	// Get hosts for current view
	var viewHosts []host.Host
	if m.viewIndex == 0 {
		viewHosts = m.hosts
	} else {
		groupIndex := m.viewIndex - 1
		if groupIndex < len(m.groups) {
			viewHosts = host.HostsForGroup(m.hosts, m.groups[groupIndex])
		}
	}

	// Apply filter
	filtered := host.FilterHosts(m.filterText, viewHosts)

	// Sort with pins
	m.filteredHosts = host.SortWithPins(filtered)
}

// View implements tea.Model.
func (m *Model) View() string {
	colors := theme.Current()
	var s strings.Builder

	// Title + help
	helpText := "↑/↓ Navigate • ←/→ View • p Pin • Enter Select • Esc Quit"
	helpWidth := lipgloss.Width(helpText)
	title := "SSH Menu"
	titleWidth := lipgloss.Width(title)

	spacing := ""
	if m.width > 0 && m.width > titleWidth+helpWidth+2 {
		spacing = strings.Repeat(" ", m.width-titleWidth-helpWidth)
	}

	s.WriteString(theme.TitleStyle().Render(title))
	s.WriteString(spacing)
	s.WriteString(theme.DimStyle().Render(helpText))
	s.WriteString("\n")

	// View bar
	if len(m.groups) > 0 {
		s.WriteString(renderViewBar(m.groups, m.viewIndex))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Filter indicator
	if m.filterText != "" {
		filterStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colors.Accent))
		s.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s", m.filterText)))
		s.WriteString("\n\n")
	}

	// Calculate content area height
	headerLines := 3 // title + viewbar + blank
	if m.filterText != "" {
		headerLines += 2
	}
	contentHeight := m.height - headerLines
	if contentHeight < 1 {
		contentHeight = 20
	}

	// Two-pane or single-pane based on width
	if m.width >= minWidthForTwoPane {
		leftWidth := m.width * 55 / 100
		rightWidth := m.width - leftWidth - 1 // -1 for separator

		m.scrollOffset = calculateScrollOffset(m.cursor, m.scrollOffset, contentHeight, len(m.filteredHosts))

		leftPane := renderHostList(m.filteredHosts, m.cursor, m.scrollOffset, leftWidth, contentHeight)

		rightPane := ""
		if len(m.filteredHosts) > 0 && m.cursor < len(m.filteredHosts) {
			rightPane = renderDetail(m.filteredHosts[m.cursor], rightWidth, contentHeight)
		}

		// Join panes side by side
		separator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Border)).
			Render("│")

		leftLines := strings.Split(leftPane, "\n")
		rightLines := strings.Split(rightPane, "\n")

		// Pad to equal length
		maxLines := contentHeight
		for len(leftLines) < maxLines {
			leftLines = append(leftLines, "")
		}
		for len(rightLines) < maxLines {
			rightLines = append(rightLines, "")
		}

		leftStyle := lipgloss.NewStyle().Width(leftWidth)
		for i := 0; i < maxLines; i++ {
			left := leftStyle.Render(safeGet(leftLines, i))
			right := safeGet(rightLines, i)
			s.WriteString(left)
			s.WriteString(separator)
			s.WriteString(right)
			s.WriteString("\n")
		}
	} else {
		// Single pane fallback
		m.scrollOffset = calculateScrollOffset(m.cursor, m.scrollOffset, contentHeight, len(m.filteredHosts))
		s.WriteString(renderHostList(m.filteredHosts, m.cursor, m.scrollOffset, m.width, contentHeight))
	}

	return s.String()
}

func safeGet(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/nick/Sync/ssh-menu && go build ./internal/ui/`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add internal/ui/model.go
git commit -m "feat: add two-pane Bubble Tea UI model with scrolling and pin toggle"
```

---

### Task 14: Rewrite main.go

**Files:**
- Rewrite: `main.go`

- [ ] **Step 1: Rewrite main.go to use new packages**

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/evix1101/ssh-menu/internal/config"
	"github.com/evix1101/ssh-menu/internal/host"
	"github.com/evix1101/ssh-menu/internal/theme"
	"github.com/evix1101/ssh-menu/internal/ui"
)

func main() {
	verbosePtr := flag.Bool("V", false, "Enable SSH verbose mode (-v flag)")
	groupPtr := flag.String("g", "", "Filter hosts by group")
	listGroupsPtr := flag.Bool("l", false, "List all available groups")
	sshOptsPtr := flag.String("s", "", "Additional SSH options to pass through")
	flag.Parse()

	configPath := sshConfigPath()

	// Initialize theme
	theme.Init(configPath)

	// Read all config files
	hosts, err := config.ReadConfigFiles(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SSH config: %s\n", err)
		os.Exit(1)
	}

	if len(hosts) == 0 {
		fmt.Fprintln(os.Stderr, "No menu hosts found in SSH config. Ensure hosts have a '# Menu ...' comment.")
		os.Exit(1)
	}

	// Assign menu numbers
	hosts, err = host.AssignMenuNumbers(hosts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate hosts
	hosts = host.ValidateHosts(hosts)

	// List groups mode
	if *listGroupsPtr {
		listGroups(hosts)
		return
	}

	// Filter by group
	if *groupPtr != "" {
		hosts = host.HostsForGroup(hosts, *groupPtr)
		if len(hosts) == 0 {
			fmt.Fprintf(os.Stderr, "No hosts found in group '%s'\n", *groupPtr)
			os.Exit(1)
		}
	}

	// Direct host selection from CLI args
	if args := flag.Args(); len(args) > 0 {
		h := findHost(args[0], hosts)
		if h == nil {
			fmt.Fprintln(os.Stderr, "Host not found.")
			os.Exit(1)
		}
		if err := connectSSH(*h, *verbosePtr, *sshOptsPtr); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing SSH: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Interactive UI
	m := ui.New(hosts, *verbosePtr, *sshOptsPtr)
	if err := ui.Run(m); err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}

	if m.Selected != nil {
		if err := connectSSH(*m.Selected, *verbosePtr, *sshOptsPtr); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing SSH: %v\n", err)
			os.Exit(1)
		}
	}
}

func sshConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to determine home directory.")
		os.Exit(1)
	}
	return filepath.Join(home, ".ssh", "config")
}

func connectSSH(h host.Host, verbose bool, sshOpts string) error {
	var args []string
	if verbose {
		args = append(args, "-v")
	}
	if sshOpts != "" {
		args = append(args, strings.Fields(sshOpts)...)
	}
	args = append(args, h.ShortName)

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findHost(input string, hosts []host.Host) *host.Host {
	if num, err := strconv.Atoi(input); err == nil {
		for i, h := range hosts {
			if h.MenuNumber == num {
				return &hosts[i]
			}
		}
		return nil
	}
	for i, h := range hosts {
		if h.ShortName == input || h.LongName == input || h.IP == input {
			return &hosts[i]
		}
	}
	return nil
}

func listGroups(hosts []host.Host) {
	groups := host.GetAllGroups(hosts)
	if len(groups) == 0 {
		fmt.Println("No groups found in SSH config.")
		return
	}
	fmt.Println("Available groups:")
	for _, g := range groups {
		count := len(host.HostsForGroup(hosts, g))
		fmt.Printf("  %s (%d hosts)\n", g, count)
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/nick/Sync/ssh-menu && go build -o ssh-menu .`
Expected: No errors, binary produced

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "feat: rewrite main.go to use new package structure"
```

---

### Task 15: Delete Old Files and Clean Up

**Files:**
- Delete: `internal/config.go`
- Delete: `internal/host.go`
- Delete: `internal/theme.go`
- Delete: `internal/ui.go`

- [ ] **Step 1: Delete old internal files**

```bash
rm internal/config.go internal/host.go internal/theme.go internal/ui.go
```

- [ ] **Step 2: Run go mod tidy**

Run: `cd /home/nick/Sync/ssh-menu && go mod tidy`
Expected: No errors

- [ ] **Step 3: Verify everything builds**

Run: `cd /home/nick/Sync/ssh-menu && go build -o ssh-menu .`
Expected: No errors

- [ ] **Step 4: Run all tests**

Run: `cd /home/nick/Sync/ssh-menu && go test ./... -v`
Expected: All tests pass

- [ ] **Step 5: Run go vet**

Run: `cd /home/nick/Sync/ssh-menu && go vet ./...`
Expected: No issues

- [ ] **Step 6: Commit**

```bash
git rm internal/config.go internal/host.go internal/theme.go internal/ui.go
git add -A
git commit -m "refactor: remove old internal files, complete package restructure"
```
