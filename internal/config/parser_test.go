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
	if h.ShortName != "web-01" { t.Errorf("ShortName: expected web-01, got %s", h.ShortName) }
	if h.LongName != "10.0.1.5" { t.Errorf("LongName: expected 10.0.1.5, got %s", h.LongName) }
	if h.User != "admin" { t.Errorf("User: expected admin, got %s", h.User) }
	if h.Port != "2222" { t.Errorf("Port: expected 2222, got %s", h.Port) }
	if h.IdentityFile != "~/.ssh/web_key" { t.Errorf("IdentityFile: expected ~/.ssh/web_key, got %s", h.IdentityFile) }
	if h.MenuNumber != 1 { t.Errorf("MenuNumber: expected 1, got %d", h.MenuNumber) }
	if h.DescText != "Web server" { t.Errorf("DescText: expected 'Web server', got '%s'", h.DescText) }
	if h.SourceFile != "test.config" { t.Errorf("SourceFile: expected test.config, got %s", h.SourceFile) }
}

func TestParseReader_MenuWithoutNumber(t *testing.T) {
	input := `# Menu: Auto-numbered host
Host auto-01
    HostName 10.0.1.1
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(hosts) != 1 { t.Fatalf("expected 1 host, got %d", len(hosts)) }
	if hosts[0].MenuNumber != 0 { t.Errorf("expected MenuNumber 0 for auto, got %d", hosts[0].MenuNumber) }
	if hosts[0].DescText != "Auto-numbered host" { t.Errorf("expected desc 'Auto-numbered host', got '%s'", hosts[0].DescText) }
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
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	h := hosts[0]
	if len(h.Groups) != 2 || h.Groups[0] != "Production" || h.Groups[1] != "Web" {
		t.Errorf("Groups: expected [Production, Web], got %v", h.Groups)
	}
	if h.IP != "203.0.113.50" { t.Errorf("IP: expected 203.0.113.50, got %s", h.IP) }
}

func TestParseReader_Pinned(t *testing.T) {
	input := `# Menu: Pinned host
# Pinned
Host pinned-01
    HostName 10.0.1.1
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if !hosts[0].Pinned { t.Error("expected host to be pinned") }
}

func TestParseReader_SkipsHostsWithoutMenu(t *testing.T) {
	input := `Host no-menu
    HostName 10.0.1.1

# Menu: Has menu
Host has-menu
    HostName 10.0.1.2
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(hosts) != 1 { t.Fatalf("expected 1 host (with menu), got %d", len(hosts)) }
	if hosts[0].ShortName != "has-menu" { t.Errorf("expected has-menu, got %s", hosts[0].ShortName) }
}

func TestParseReader_NoDefaultUserOrPort(t *testing.T) {
	input := `# Menu: Minimal
Host minimal
    HostName 10.0.1.1
`
	hosts, err := ParseReader(strings.NewReader(input), "test.config")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if hosts[0].User != "" { t.Errorf("expected empty User, got '%s'", hosts[0].User) }
	if hosts[0].Port != "" { t.Errorf("expected empty Port, got '%s'", hosts[0].Port) }
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
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(hosts) != 2 { t.Fatalf("expected 2 hosts, got %d", len(hosts)) }
}

func TestParseReader_EmptyInput(t *testing.T) {
	hosts, err := ParseReader(strings.NewReader(""), "test.config")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(hosts) != 0 { t.Errorf("expected 0 hosts from empty input, got %d", len(hosts)) }
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
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(hosts) != 1 { t.Fatalf("expected 1 host (skipping wildcard), got %d", len(hosts)) }
	if hosts[0].ShortName != "real" { t.Errorf("expected 'real', got '%s'", hosts[0].ShortName) }
}
