package e2e

import (
	"regexp"
	"strings"
	"testing"
)

// TestTagAddAndDelete tests adding and deleting tags
func TestTagAddAndDelete(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// First, create an entry
	args := []string{serverAddr(), "-c", "echo setup"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	_, _, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to setup entry: %v", err)
	}

	// Get entry ID from list
	stdout, _, err := runSSX(t, "list")
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	// Extract ID (assuming format like "  1  | root@host...")
	re := regexp.MustCompile(`\s+(\d+)\s+\|`)
	matches := re.FindStringSubmatch(stdout)
	if len(matches) < 2 {
		t.Fatalf("Could not find entry ID in list output: %s", stdout)
	}
	entryID := matches[1]

	// Add tags
	_, stderr, err := runSSX(t, "tag", "--id", entryID, "-t", "test-tag", "-t", "production")
	if err != nil {
		t.Fatalf("Failed to add tags: %v, stderr: %s", err, stderr)
	}

	// Verify tags were added
	stdout, _, err = runSSX(t, "list")
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	if !strings.Contains(stdout, "test-tag") {
		t.Errorf("Expected list to contain 'test-tag', got: %s", stdout)
	}
	if !strings.Contains(stdout, "production") {
		t.Errorf("Expected list to contain 'production', got: %s", stdout)
	}

	// Delete one tag
	_, stderr, err = runSSX(t, "tag", "--id", entryID, "-d", "test-tag")
	if err != nil {
		t.Fatalf("Failed to delete tag: %v, stderr: %s", err, stderr)
	}

	// Verify tag was deleted
	stdout, _, err = runSSX(t, "list")
	if err != nil {
		t.Fatalf("Failed to list entries: %v", err)
	}

	if strings.Contains(stdout, "test-tag") {
		t.Errorf("Expected 'test-tag' to be deleted, but still found in: %s", stdout)
	}
	if !strings.Contains(stdout, "production") {
		t.Errorf("Expected 'production' tag to remain, got: %s", stdout)
	}
}

// TestTagRequiresID tests that tag command requires --id flag
func TestTagRequiresID(t *testing.T) {
	setupDB(t)

	_, stderr, err := runSSXWithDB(t, "tag", "-t", "sometag")
	if err == nil {
		t.Error("Expected error when --id is not provided")
	}

	// The error could be about required flag or invalid id
	_ = stderr // Error is expected
}

// TestTagNoTagSpecified tests error when no tag is specified
func TestTagNoTagSpecified(t *testing.T) {
	setupDB(t)

	_, stderr, err := runSSXWithDB(t, "tag", "--id", "1")
	if err == nil {
		t.Error("Expected error when no tag is specified")
	}

	if !strings.Contains(stderr, "no tag") {
		t.Logf("Expected 'no tag' error, got: %s", stderr)
	}
}

// TestConnectByTag tests connecting using tag
func TestConnectByTag(t *testing.T) {
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

	// Add a unique tag
	tagName := "e2e-test-server"
	_, _, err = runSSX(t, "tag", "--id", entryID, "-t", tagName)
	if err != nil {
		t.Fatalf("Failed to add tag: %v", err)
	}

	// Connect using tag
	args = []string{tagName, "-c", "echo tag_connect_test"}
	stdout, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to connect by tag: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if !strings.Contains(stdout, "tag_connect_test") {
		t.Errorf("Expected output to contain 'tag_connect_test', got: %s", stdout)
	}
}
