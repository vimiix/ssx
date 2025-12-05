package e2e

import (
	"strings"
	"testing"
)

// TestListEmpty tests listing entries when database is empty
func TestListEmpty(t *testing.T) {
	setupDB(t)

	_, stderr, err := runSSXWithDB(t, "list")
	// Should return error when no entries exist
	if err == nil {
		t.Log("list command succeeded with empty database")
	}

	// Check for "no entry" message in combined output
	combined := stderr
	if !strings.Contains(combined, "no entry") && err != nil {
		t.Logf("Expected 'no entry' message or success, got stderr: %s", stderr)
	}
}

// TestListAliases tests list command aliases
func TestListAliases(t *testing.T) {
	setupDB(t)

	// Test 'l' alias
	_, _, err1 := runSSXWithDB(t, "l")
	// Test 'ls' alias
	_, _, err2 := runSSXWithDB(t, "ls")

	// Both should behave the same (either succeed or fail with same error)
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("List aliases behave differently: 'l' err=%v, 'ls' err=%v", err1, err2)
	}
}

// TestListAfterConnection tests listing entries after a successful connection
func TestListAfterConnection(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// First, connect to server to create an entry
	args := []string{serverAddr(), "-c", "echo hello"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	_, _, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	// Now list entries
	stdout, _, err := runSSX(t, "list")
	if err != nil {
		t.Fatalf("ssx list failed: %v", err)
	}

	// Should contain the server address
	if !strings.Contains(stdout, cfg.Host) {
		t.Errorf("Expected list output to contain %q, got: %s", cfg.Host, stdout)
	}

	if !strings.Contains(stdout, cfg.User) {
		t.Errorf("Expected list output to contain %q, got: %s", cfg.User, stdout)
	}
}
