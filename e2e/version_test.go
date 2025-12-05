package e2e

import (
	"strings"
	"testing"
)

// TestVersion tests the version flag
func TestVersion(t *testing.T) {
	stdout, _, err := runSSX(t, "--version")
	if err != nil {
		t.Fatalf("ssx --version failed: %v", err)
	}

	// Version output contains version info
	if !strings.Contains(stdout, "Version") {
		t.Errorf("Expected version output to contain 'Version', got: %s", stdout)
	}
}

// TestHelp tests the help flag
func TestHelp(t *testing.T) {
	stdout, _, err := runSSX(t, "--help")
	if err != nil {
		t.Fatalf("ssx --help failed: %v", err)
	}

	expectedStrings := []string{
		"ssx is a retentive ssh client",
		"Available Commands:",
		"list",
		"delete",
		"tag",
		"info",
		"cp",
		"upgrade",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help output to contain %q, got: %s", expected, stdout)
		}
	}
}

// TestCpHelp tests the cp command help
func TestCpHelp(t *testing.T) {
	stdout, _, err := runSSX(t, "cp", "--help")
	if err != nil {
		t.Fatalf("ssx cp --help failed: %v", err)
	}

	expectedStrings := []string{
		"Copy files between local and remote hosts",
		"remote-to-remote",
		"SOURCE",
		"TARGET",
		"--identity-file",
		"--jump-server",
		"--port",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected cp help to contain %q, got: %s", expected, stdout)
		}
	}
}
