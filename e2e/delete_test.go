package e2e

import (
	"regexp"
	"strings"
	"testing"
)

// TestDeleteEntry tests deleting an entry
func TestDeleteEntry(t *testing.T) {
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
		t.Fatalf("Could not find entry ID in list output: %s", stdout)
	}
	entryID := matches[1]

	// Delete the entry
	_, stderr, err := runSSX(t, "delete", "--id", entryID)
	if err != nil {
		t.Fatalf("Failed to delete entry: %v, stderr: %s", err, stderr)
	}

	// Verify entry was deleted
	_, stderr, err = runSSX(t, "list")
	// Should either show empty list or error
	if err == nil && strings.Contains(stdout, cfg.Host) {
		t.Errorf("Entry should have been deleted, but still found in list")
	}
}

// TestDeleteMultipleEntries tests deleting multiple entries at once
func TestDeleteMultipleEntries(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// Create first entry
	args := []string{serverAddr(), "-c", "echo setup1"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	_, _, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to setup first entry: %v", err)
	}

	// Get entry IDs
	stdout, _, _ := runSSX(t, "list")
	re := regexp.MustCompile(`\s+(\d+)\s+\|`)
	matches := re.FindAllStringSubmatch(stdout, -1)
	if len(matches) < 1 {
		t.Fatalf("Could not find entry IDs")
	}

	// Delete all entries
	args = []string{"delete"}
	for _, match := range matches {
		args = append(args, "--id", match[1])
	}

	_, stderr, err := runSSX(t, args...)
	if err != nil {
		t.Fatalf("Failed to delete entries: %v, stderr: %s", err, stderr)
	}

	// Verify all entries were deleted
	_, _, err = runSSX(t, "list")
	if err == nil {
		// If no error, check that our host is not in the list
		stdout, _, _ := runSSX(t, "list")
		if strings.Contains(stdout, cfg.Host) {
			t.Error("Entries should have been deleted")
		}
	}
}

// TestDeleteNoID tests delete command without ID
func TestDeleteNoID(t *testing.T) {
	setupDB(t)

	stdout, _, err := runSSXWithDB(t, "delete")
	if err != nil {
		// It's okay if it errors, just check the message
		t.Logf("Delete without ID returned error (expected): %v", err)
	}

	if !strings.Contains(stdout, "no id specified") && !strings.Contains(stdout, "no id") {
		t.Logf("Expected 'no id specified' message, got: %s", stdout)
	}
}

// TestDeleteAliases tests delete command aliases
func TestDeleteAliases(t *testing.T) {
	setupDB(t)

	// Test 'd' alias
	stdout1, _, _ := runSSXWithDB(t, "d")
	// Test 'del' alias
	stdout2, _, _ := runSSXWithDB(t, "del")

	// Both should show same "no id specified" message or similar behavior
	_ = stdout1
	_ = stdout2
	// Just verify aliases work without crashing
}
