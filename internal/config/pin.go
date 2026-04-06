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

func addPinComment(lines []string, hostLineIdx int) []string {
	// Check if already pinned
	for i := hostLineIdx - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if rePinnedLine.MatchString(trimmed) {
			return lines
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

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, "# Pinned")
	newLines = append(newLines, lines[insertIdx:]...)

	return newLines
}

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
	return lines
}
