package e2e

import (
	"strings"
	"testing"
)

// TestConnectAndExecute tests connecting to a server and executing a command
func TestConnectAndExecute(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	args := []string{serverAddr(), "-c", "echo hello_ssx_test"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	stdout, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to connect and execute command: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if !strings.Contains(stdout, "hello_ssx_test") {
		t.Errorf("Expected output to contain 'hello_ssx_test', got: %s", stdout)
	}
}

// TestConnectWithPort tests connecting with explicit port
func TestConnectWithPort(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	addr := cfg.User + "@" + cfg.Host + ":" + cfg.Port
	args := []string{addr, "-c", "whoami"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	stdout, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to connect with port: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if !strings.Contains(stdout, cfg.User) {
		t.Errorf("Expected whoami output to contain %q, got: %s", cfg.User, stdout)
	}
}

// TestConnectWithIdentityFile tests connecting with SSH key
func TestConnectWithIdentityFile(t *testing.T) {
	if cfg.KeyPath == "" {
		t.Skip("Skipping key-based auth test: SSX_E2E_KEY not set")
	}
	skipIfNoServer(t)
	cleanupDB(t)

	args := []string{serverAddr(), "-i", cfg.KeyPath, "-c", "hostname"}

	stdout, stderr, err := runSSX(t, args...)
	if err != nil {
		t.Fatalf("Failed to connect with identity file: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Should get some output (hostname)
	if strings.TrimSpace(stdout) == "" {
		t.Error("Expected hostname output, got empty string")
	}
}

// TestConnectByKeyword tests connecting using partial keyword match
func TestConnectByKeyword(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// First, create an entry by connecting
	args := []string{serverAddr(), "-c", "echo setup"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	_, _, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to setup entry: %v", err)
	}

	// Now connect using partial host match
	// Extract a portion of the host for keyword search
	keyword := cfg.Host
	if len(keyword) > 4 {
		keyword = keyword[len(keyword)-4:] // Use last 4 characters
	}

	args = []string{keyword, "-c", "echo keyword_test"}
	stdout, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to connect by keyword: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if !strings.Contains(stdout, "keyword_test") {
		t.Errorf("Expected output to contain 'keyword_test', got: %s", stdout)
	}
}

// TestConnectWithTimeout tests command execution with timeout
func TestConnectWithTimeout(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	args := []string{serverAddr(), "-c", "sleep 1 && echo done", "--timeout", "5s"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	stdout, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed with timeout: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	if !strings.Contains(stdout, "done") {
		t.Errorf("Expected output to contain 'done', got: %s", stdout)
	}
}

// TestConnectTimeoutExceeded tests that timeout actually works
func TestConnectTimeoutExceeded(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	args := []string{serverAddr(), "-c", "sleep 10", "--timeout", "1s"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	_, _, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err == nil {
		t.Error("Expected timeout error, but command succeeded")
	}
}
