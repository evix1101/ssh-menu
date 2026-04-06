package host

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateHosts_MissingIdentityFile(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", LongName: "10.0.1.1", IdentityFile: "/nonexistent/path/key"},
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
		{ShortName: "a", LongName: "10.0.1.1", IdentityFile: keyPath},
	}
	result := ValidateHosts(hosts)
	if len(result[0].Warnings) != 0 {
		t.Errorf("expected no warnings, got %v", result[0].Warnings)
	}
}

func TestValidateHosts_TildeExpansion(t *testing.T) {
	hosts := []Host{
		{ShortName: "a", LongName: "10.0.1.1", IdentityFile: "~/.ssh/nonexistent_key_12345"},
	}
	result := ValidateHosts(hosts)
	if len(result[0].Warnings) != 1 {
		t.Fatalf("expected 1 warning for missing key, got %d", len(result[0].Warnings))
	}
}

func TestValidateHosts_DuplicateAliases(t *testing.T) {
	hosts := []Host{
		{ShortName: "server-a", LongName: "10.0.1.1", SourceFile: "/etc/ssh/config"},
		{ShortName: "server-a", LongName: "10.0.1.1", SourceFile: "/etc/ssh/config.d/extra"},
	}
	result := ValidateHosts(hosts)
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
		{ShortName: "a", LongName: "10.0.1.1", IdentityFile: ""},
	}
	result := ValidateHosts(hosts)
	if len(result[0].Warnings) != 0 {
		t.Errorf("expected no warnings when no identity file set, got %v", result[0].Warnings)
	}
}
