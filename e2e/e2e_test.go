// Package e2e provides end-to-end tests for the ssx command-line tool.
// These tests require a real SSH server to run against.
//
// To run these tests, set the following environment variables:
//   - SSX_E2E_HOST: SSH server hostname or IP
//   - SSX_E2E_PORT: SSH server port (default: 22)
//   - SSX_E2E_USER: SSH username
//   - SSX_E2E_PASSWORD: SSH password (optional if using key)
//   - SSX_E2E_KEY: Path to SSH private key (optional if using password)
//   - SSX_E2E_HOST2: Second SSH server for remote-to-remote tests (optional)
//   - SSX_E2E_PORT2: Second SSH server port (default: 22)
//   - SSX_E2E_USER2: Second SSH username (optional)
//
// Example:
//
//	SSX_E2E_HOST=192.168.1.100 SSX_E2E_USER=root SSX_E2E_PASSWORD=secret go test -v ./e2e/...
package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test configuration from environment variables
type testConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	KeyPath  string

	// Second host for remote-to-remote tests
	Host2 string
	Port2 string
	User2 string
}

var (
	cfg        testConfig
	ssxBinary  string
	testDBPath string
)

func TestMain(m *testing.M) {
	// Load configuration from environment
	cfg = testConfig{
		Host:     os.Getenv("SSX_E2E_HOST"),
		Port:     getEnvOrDefault("SSX_E2E_PORT", "22"),
		User:     os.Getenv("SSX_E2E_USER"),
		Password: os.Getenv("SSX_E2E_PASSWORD"),
		KeyPath:  os.Getenv("SSX_E2E_KEY"),
		Host2:    os.Getenv("SSX_E2E_HOST2"),
		Port2:    getEnvOrDefault("SSX_E2E_PORT2", "22"),
		User2:    os.Getenv("SSX_E2E_USER2"),
	}

	// Build ssx binary for testing
	tmpDir, err := os.MkdirTemp("", "ssx-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	ssxBinary = filepath.Join(tmpDir, "ssx")
	testDBPath = filepath.Join(tmpDir, "test.db")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", ssxBinary, "../cmd/ssx")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build ssx binary: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll(tmpDir)

	os.Exit(code)
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// skipIfNoServer skips the test if no SSH server is configured
func skipIfNoServer(t *testing.T) {
	if cfg.Host == "" || cfg.User == "" {
		t.Skip("Skipping e2e test: SSX_E2E_HOST and SSX_E2E_USER must be set")
	}
	if cfg.Password == "" && cfg.KeyPath == "" {
		t.Skip("Skipping e2e test: SSX_E2E_PASSWORD or SSX_E2E_KEY must be set")
	}
}

// skipIfNoSecondServer skips the test if second SSH server is not configured
func skipIfNoSecondServer(t *testing.T) {
	skipIfNoServer(t)
	if cfg.Host2 == "" {
		t.Skip("Skipping remote-to-remote test: SSX_E2E_HOST2 must be set")
	}
}

// runSSX runs the ssx command with given arguments
func runSSX(t *testing.T, args ...string) (string, string, error) {
	return runSSXWithEnv(t, nil, args...)
}

// runSSXWithEnv runs the ssx command with given arguments and environment
func runSSXWithEnv(t *testing.T, env []string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ssxBinary, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("SSX_DB_PATH=%s", testDBPath))
	cmd.Env = append(cmd.Env, env...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// runSSXWithInput runs the ssx command with stdin input
func runSSXWithInput(t *testing.T, input string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ssxBinary, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("SSX_DB_PATH=%s", testDBPath))

	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// initDB initializes the test database with a password
func initDB(t *testing.T) {
	// Run a simple command to initialize the database with password
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ssxBinary, "list")
	cmd.Env = append(os.Environ(), fmt.Sprintf("SSX_DB_PATH=%s", testDBPath))
	cmd.Stdin = strings.NewReader("testpassword\n")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run() // Ignore error, just initializing
}

// runSSXWithDB runs ssx with initialized database (provides password if needed)
func runSSXWithDB(t *testing.T, args ...string) (string, string, error) {
	return runSSXWithInput(t, "testpassword\n", args...)
}

// serverAddr returns the full server address
func serverAddr() string {
	addr := cfg.User + "@" + cfg.Host
	if cfg.Port != "22" {
		addr += ":" + cfg.Port
	}
	return addr
}

// serverAddr2 returns the second server address
func serverAddr2() string {
	user := cfg.User2
	if user == "" {
		user = cfg.User
	}
	addr := user + "@" + cfg.Host2
	if cfg.Port2 != "22" {
		addr += ":" + cfg.Port2
	}
	return addr
}

// cleanupDB removes the test database
func cleanupDB(t *testing.T) {
	os.Remove(testDBPath)
}

// setupDB cleans and initializes the test database
func setupDB(t *testing.T) {
	cleanupDB(t)
	initDB(t)
}
