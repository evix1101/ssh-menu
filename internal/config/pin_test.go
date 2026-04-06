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
	if err != nil { t.Fatalf("unexpected error: %v", err) }

	result, _ := os.ReadFile(path)
	if !strings.Contains(string(result), "# Pinned") {
		t.Error("expected # Pinned to be added")
	}
	lines := strings.Split(string(result), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "# Pinned" {
			if i < 2 { t.Error("# Pinned should be after Group line") }
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
	if err != nil { t.Fatalf("unexpected error: %v", err) }

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
	if err != nil { t.Fatalf("unexpected error: %v", err) }

	result, _ := os.ReadFile(path)
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
	if err == nil { t.Error("expected error for host not found") }
}
