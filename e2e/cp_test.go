package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestCpUpload tests uploading a local file to remote
func TestCpUpload(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// Create a temporary local file
	tmpDir, err := os.MkdirTemp("", "ssx-cp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	localFile := filepath.Join(tmpDir, "upload_test.txt")
	testContent := "Hello from ssx e2e test - upload"
	if err := os.WriteFile(localFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Upload file
	remotePath := "/tmp/ssx_e2e_upload_test.txt"
	args := []string{"cp", localFile, serverAddr() + ":" + remotePath}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	stdout, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to upload file: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Verify file was uploaded by reading it back
	args = []string{serverAddr(), "-c", "cat " + remotePath}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	stdout, stderr, err = runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to verify upload: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, testContent) {
		t.Errorf("Expected uploaded content %q, got: %s", testContent, stdout)
	}

	// Cleanup remote file
	args = []string{serverAddr(), "-c", "rm -f " + remotePath}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	runSSXWithInput(t, cfg.Password+"\n", args...)
}

// TestCpDownload tests downloading a remote file to local
func TestCpDownload(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// Create a file on remote server
	remotePath := "/tmp/ssx_e2e_download_test.txt"
	testContent := "Hello from ssx e2e test - download"

	args := []string{serverAddr(), "-c", "echo '" + testContent + "' > " + remotePath}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	_, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to create remote file: %v\nstderr: %s", err, stderr)
	}

	// Create temp dir for download
	tmpDir, err := os.MkdirTemp("", "ssx-cp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	localFile := filepath.Join(tmpDir, "download_test.txt")

	// Download file
	args = []string{"cp", serverAddr() + ":" + remotePath, localFile}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	stdout, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to download file: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Verify downloaded content
	content, err := os.ReadFile(localFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if !strings.Contains(string(content), testContent) {
		t.Errorf("Expected downloaded content to contain %q, got: %s", testContent, string(content))
	}

	// Cleanup remote file
	args = []string{serverAddr(), "-c", "rm -f " + remotePath}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	runSSXWithInput(t, cfg.Password+"\n", args...)
}

// TestCpWithTag tests file copy using tag to reference remote host
func TestCpWithTag(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	// First create an entry and tag it
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

	tagName := "cp-test-server"
	_, _, err = runSSX(t, "tag", "--id", entryID, "-t", tagName)
	if err != nil {
		t.Fatalf("Failed to add tag: %v", err)
	}

	// Create local file
	tmpDir, err := os.MkdirTemp("", "ssx-cp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	localFile := filepath.Join(tmpDir, "tag_test.txt")
	testContent := "Tag-based copy test"
	if err := os.WriteFile(localFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Upload using tag
	remotePath := "/tmp/ssx_e2e_tag_test.txt"
	args = []string{"cp", localFile, tagName + ":" + remotePath}

	stdout, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to upload using tag: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Verify
	args = []string{serverAddr(), "-c", "cat " + remotePath}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	stdout, _, err = runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to verify: %v", err)
	}

	if !strings.Contains(stdout, testContent) {
		t.Errorf("Expected content %q, got: %s", testContent, stdout)
	}

	// Cleanup
	args = []string{serverAddr(), "-c", "rm -f " + remotePath}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	runSSXWithInput(t, cfg.Password+"\n", args...)
}

// TestCpRemoteToRemote tests copying file between two remote hosts
func TestCpRemoteToRemote(t *testing.T) {
	skipIfNoSecondServer(t)
	cleanupDB(t)

	// Create a file on first remote server
	remotePath1 := "/tmp/ssx_e2e_r2r_source.txt"
	testContent := "Remote to remote test content"

	args := []string{serverAddr(), "-c", "echo '" + testContent + "' > " + remotePath1}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	_, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to create source file: %v\nstderr: %s", err, stderr)
	}

	// Copy from first server to second server
	remotePath2 := "/tmp/ssx_e2e_r2r_dest.txt"
	args = []string{"cp", serverAddr() + ":" + remotePath1, serverAddr2() + ":" + remotePath2}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	stdout, stderr, err := runSSXWithInput(t, cfg.Password+"\n"+cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed remote-to-remote copy: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Verify file on second server
	args = []string{serverAddr2(), "-c", "cat " + remotePath2}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	stdout, stderr, err = runSSXWithInput(t, cfg.Password+"\n", args...)
	if err != nil {
		t.Fatalf("Failed to verify on second server: %v\nstderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, testContent) {
		t.Errorf("Expected content %q on second server, got: %s", testContent, stdout)
	}

	// Cleanup both servers
	args = []string{serverAddr(), "-c", "rm -f " + remotePath1}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	runSSXWithInput(t, cfg.Password+"\n", args...)

	args = []string{serverAddr2(), "-c", "rm -f " + remotePath2}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	runSSXWithInput(t, cfg.Password+"\n", args...)
}

// TestCpLocalToLocal tests that local-to-local copy is rejected
func TestCpLocalToLocal(t *testing.T) {
	setupDB(t)

	stdout, stderr, err := runSSXWithDB(t, "cp", "/tmp/file1", "/tmp/file2")
	if err == nil {
		t.Error("Expected error for local-to-local copy")
	}

	combined := stdout + stderr
	if !strings.Contains(combined, "local to local") && !strings.Contains(combined, "local") {
		t.Errorf("Expected 'local to local' error message, got stdout: %s, stderr: %s", stdout, stderr)
	}
}

// TestCpMissingArgs tests cp command with missing arguments
func TestCpMissingArgs(t *testing.T) {
	_, stderr, err := runSSX(t, "cp", "/tmp/file1")
	if err == nil {
		t.Error("Expected error for missing target argument")
	}

	if !strings.Contains(stderr, "requires") || !strings.Contains(stderr, "arg") {
		t.Logf("Expected argument error, got: %s", stderr)
	}
}

// TestCpNonExistentLocalFile tests uploading non-existent file
func TestCpNonExistentLocalFile(t *testing.T) {
	skipIfNoServer(t)
	cleanupDB(t)

	args := []string{"cp", "/nonexistent/file/path.txt", serverAddr() + ":/tmp/dest.txt"}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}

	_, stderr, err := runSSXWithInput(t, cfg.Password+"\n", args...)
	if err == nil {
		t.Error("Expected error for non-existent local file")
	}

	if !strings.Contains(stderr, "no such file") && !strings.Contains(stderr, "does not exist") && !strings.Contains(stderr, "failed") {
		t.Logf("Expected file not found error, got: %s", stderr)
	}
}
