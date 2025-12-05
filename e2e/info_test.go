package e2e

import (
	"regexp"
	"strings"
	"testing"
)

// TestInfoByID tests showing entry info by ID
func TestInfoByID(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// Create an entry
	args := []string{serverAddr(), "-c", "echo setup"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	_, _, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to setup entry: %v", err)
	}

	// Get entry ID
	stdout, _, _ := runSSX(t, "list")
	re := regexp.MustCompile(`\s+(\d+)\s+\|`)
	matches := re.FindStringSubmatch(stdout)
	if len(matches) < 2 {
		t.Fatalf("Could not find entry ID")
	}
	entryID := matches[1]

	// Get info by ID
	stdout, stderr, err := runSSX(t, "info", "--id", entryID)
	if err != nil {
		t.Fatalf("Failed to get info: %v, stderr: %s", err, stderr)
	}

	// Should be JSON format with entry details
	if !strings.Contains(stdout, "host") {
		t.Errorf("Expected JSON output with 'host' field, got: %s", stdout)
	}
	if !strings.Contains(stdout, cfg.Host) {
		t.Errorf("Expected output to contain host %q, got: %s", cfg.Host, stdout)
	}
	if !strings.Contains(stdout, cfg.User) {
		t.Errorf("Expected output to contain user %q, got: %s", cfg.User, stdout)
	}
}

// TestInfoByKeyword tests showing entry info by keyword
func TestInfoByKeyword(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// Create an entry
	args := []string{serverAddr(), "-c", "echo setup"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	_, _, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to setup entry: %v", err)
	}

	// Get info by keyword (partial host match)
	keyword := cfg.Host
	if len(keyword) > 4 {
		keyword = keyword[len(keyword)-4:]
	}

	stdout, stderr, err := runSSX(t, "info", keyword)
	if err != nil {
		t.Fatalf("Failed to get info by keyword: %v, stderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, cfg.Host) {
		t.Errorf("Expected output to contain host %q, got: %s", cfg.Host, stdout)
	}
}

// TestInfoByTag tests showing entry info by tag
func TestInfoByTag(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// Create an entry
	args := []string{serverAddr(), "-c", "echo setup"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	_, _, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to setup entry: %v", err)
	}

	// Get entry ID and add tag
	stdout, _, _ := runSSX(t, "list")
	re := regexp.MustCompile(`\s+(\d+)\s+\|`)
	matches := re.FindStringSubmatch(stdout)
	if len(matches) < 2 {
		t.Fatalf("Could not find entry ID")
	}
	entryID := matches[1]

	tagName := "info-test-tag"
	_, _, err = runSSX(t, "tag", "--id", entryID, "-t", tagName)
	if err != nil {
		t.Fatalf("Failed to add tag: %v", err)
	}

	// Get info by tag
	stdout, stderr, err := runSSX(t, "info", "--tag", tagName)
	if err != nil {
		t.Fatalf("Failed to get info by tag: %v, stderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, cfg.Host) {
		t.Errorf("Expected output to contain host %q, got: %s", cfg.Host, stdout)
	}
	if !strings.Contains(stdout, tagName) {
		t.Errorf("Expected output to contain tag %q, got: %s", tagName, stdout)
	}
}

// TestInfoNotFound tests info command with non-existent entry
func TestInfoNotFound(t *testing.T) {
	setupDB(t)

	_, stderr, err := runSSXWithDB(t, "info", "nonexistent-host-12345")
	if err == nil {
		t.Error("Expected error for non-existent entry")
	}

	if !strings.Contains(stderr, "not matched") && !strings.Contains(stderr, "no entry") && !strings.Contains(stderr, "not found") {
		t.Logf("Expected 'not matched' or 'no entry' error, got: %s", stderr)
	}
}

// TestInfoPasswordMasked tests that password is masked in info output
func TestInfoPasswordMasked(t *testing.T) {
	skipIfNoServer(t)
	if cfg.Password == "" {
		t.Skip("Skipping password mask test: no password configured")
	}
	cleanupDB(t)

	// Create an entry with password
	args := []string{serverAddr(), "-c", "echo setup"}
	_, _, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to setup entry: %v", err)
	}

	// Get entry ID
	stdout, _, _ := runSSX(t, "list")
	re := regexp.MustCompile(`\s+(\d+)\s+\|`)
	matches := re.FindStringSubmatch(stdout)
	if len(matches) < 2 {
		t.Fatalf("Could not find entry ID")
	}
	entryID := matches[1]

	// Get info
	stdout, _, err = runSSX(t, "info", "--id", entryID)
	if err != nil {
		t.Fatalf("Failed to get info: %v", err)
	}

	// Password should be masked (contains ***)
	if strings.Contains(stdout, cfg.Password) {
		t.Error("Password should be masked in info output")
	}
	if !strings.Contains(stdout, "***") {
		t.Log("Expected masked password with '***' in output")
	}
}
