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

// pendingMeta holds annotations that appear before a Host line.
type pendingMeta struct {
	descText   string
	menuNumber int
	ip         string
	groups     []string
	pinned     bool
}

// ParseReader parses SSH config from a reader and returns host entries.
func ParseReader(r io.Reader, sourceFile string) ([]host.Host, error) {
	var hosts []host.Host
	var current *host.Host
	var pending pendingMeta

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if m := reHost.FindStringSubmatch(line); m != nil {
			// Save current host if it has required fields
			if current != nil && current.ShortName != "" && current.DescText != "" {
				hosts = append(hosts, *current)
			}

			hostName := strings.TrimSpace(m[1])
			if hostName == "*" {
				// Wildcard: discard pending meta and don't start a host
				current = nil
				pending = pendingMeta{}
				continue
			}

			// Start new host, applying any buffered pending metadata
			current = &host.Host{
				ShortName:  hostName,
				Groups:     pending.groups,
				DescText:   pending.descText,
				MenuNumber: pending.menuNumber,
				IP:         pending.ip,
				Pinned:     pending.pinned,
				SourceFile: sourceFile,
			}
			if current.Groups == nil {
				current.Groups = []string{}
			}
			pending = pendingMeta{}
			continue
		}

		if m := reMenu.FindStringSubmatch(line); m != nil {
			var num int
			if m[1] != "" {
				var err error
				num, err = strconv.Atoi(m[1])
				if err != nil {
					return nil, fmt.Errorf("invalid menu number: %s", m[1])
				}
			}
			pending.menuNumber = num
			pending.descText = strings.TrimSpace(m[2])
			continue
		}
		if m := reIP.FindStringSubmatch(line); m != nil {
			pending.ip = strings.TrimSpace(m[1])
			continue
		}
		if m := reGroup.FindStringSubmatch(line); m != nil {
			g := strings.TrimSpace(m[1])
			if !sliceContains(pending.groups, g) {
				pending.groups = append(pending.groups, g)
			}
			continue
		}
		if rePinned.MatchString(line) {
			pending.pinned = true
			continue
		}

		if current == nil {
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

	if current != nil && current.ShortName != "" && current.DescText != "" {
		hosts = append(hosts, *current)
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
