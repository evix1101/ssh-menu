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

func checkDuplicateAliases(hosts []Host) {
	seen := make(map[string][]int)
	for i, h := range hosts {
		seen[h.ShortName] = append(seen[h.ShortName], i)
	}
	for name, indices := range seen {
		if len(indices) <= 1 {
			continue
		}
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

func checkEmptyHostnames(hosts []Host) {
	for i := range hosts {
		if hosts[i].LongName != "" {
			continue
		}
		// Skip if the alias itself is an IP or FQDN (contains a dot).
		if net.ParseIP(hosts[i].ShortName) != nil || strings.Contains(hosts[i].ShortName, ".") {
			continue
		}
		// Skip trivially short aliases and hosts with enough other config to imply
		// the SSH client will resolve them (identity file or sourced from a file).
		if len(hosts[i].ShortName) <= 1 || hosts[i].IdentityFile != "" || hosts[i].SourceFile != "" {
			continue
		}
		hosts[i].Warnings = append(hosts[i].Warnings, Warning{
			Level:   "warn",
			Message: fmt.Sprintf("No HostName set for '%s'", hosts[i].ShortName),
		})
	}
}

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
