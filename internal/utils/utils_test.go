package utils

import (
	"archive/zip"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	tests := []struct {
		caseName string
		file     string
		expect   bool
	}{
		{"file-should-not-exists", "foo/bar", false},
		{"file-should-exists", t.TempDir(), true},
	}
	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			actual := FileExists(tt.file)
			if actual != tt.expect {
				t.Errorf("expect %t, got %t", tt.expect, actual)
			}
		})
	}
}

func TestExpandHomeDir(t *testing.T) {
	tmpHome := "/home"
	getCurrentUserFunc = func() (*user.User, error) {
		return &user.User{HomeDir: tmpHome}, nil
	}
	tests := []struct {
		path   string
		expect string
	}{
		{"~", tmpHome},
		{"~/a/b", path.Join(tmpHome, "a/b")},
		{"/a/b", "/a/b"},
		{"", ""},
		{"./a", "./a"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("expand %q", tt.path), func(t *testing.T) {
			actual := ExpandHomeDir(tt.path)
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		s      string
		expect string
	}{
		{"", ""},
		{"a", "a***"},
		{"ab", "a***"},
		{"abc", "a***"},
		{"abcd", "ab***d"},
		{"abcdefgh", "ab***h"},
	}
	for _, tt := range tests {
		actual := MaskString(tt.s)
		assert.Equal(t, tt.expect, actual)
	}
}

func TestMatchAddress(t *testing.T) {
	tests := []struct {
		addr     string
		username string
		host     string
		port     string
	}{
		{"user@host:22", "user", "host", "22"},
		{"host:2222", "", "host", "2222"},
		{"host", "", "host", ""},
		{"a.b@1.1.1.1", "a.b", "1.1.1.1", ""},
		{"a_b@1.1.1.1", "a_b", "1.1.1.1", ""},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			addr, err := MatchAddress(tt.addr)
			assert.NoError(t, err)
			assert.Equal(t, tt.username, addr.User)
			assert.Equal(t, tt.host, addr.Host)
			assert.Equal(t, tt.port, addr.Port)
		})
	}
}

func TestHashWithSHA256(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Hash empty string",
			args: args{
				input: "",
			},
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name: "Hash non-empty string",
			args: args{
				input: "hello world",
			},
			want: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashWithSHA256(tt.args.input); got != tt.want {
				t.Errorf("HashWithSHA256() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnzip(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	// Create a test zip file
	testZipPath := filepath.Join(tmpDir, "test.zip")
	if err := createTestZip(testZipPath); err != nil {
		t.Fatal(err)
	}

	// Create target directory for extraction
	targetDir := filepath.Join(tmpDir, "extracted")

	tests := []struct {
		name      string
		zipPath   string
		targetDir string
		files     []string
		wantErr   bool
		checkFn   func(t *testing.T, targetDir string) error
	}{
		{
			name:      "extract all files",
			zipPath:   testZipPath,
			targetDir: targetDir,
			files:     nil,
			wantErr:   false,
			checkFn: func(t *testing.T, targetDir string) error {
				// Check if all files exist
				files := []string{
					"file1.txt",
					"dir1/file2.txt",
					"dir1/dir2/file3.txt",
				}
				for _, f := range files {
					path := filepath.Join(targetDir, f)
					if !FileExists(path) {
						return fmt.Errorf("file not found: %s", path)
					}
				}
				return nil
			},
		},
		{
			name:      "extract specific file",
			zipPath:   testZipPath,
			targetDir: filepath.Join(targetDir, "specific"),
			files:     []string{"file1.txt"},
			wantErr:   false,
			checkFn: func(t *testing.T, targetDir string) error {
				// Check if only specified file exists
				if !FileExists(filepath.Join(targetDir, "file1.txt")) {
					return fmt.Errorf("file1.txt not found")
				}
				if FileExists(filepath.Join(targetDir, "dir1/file2.txt")) {
					return fmt.Errorf("file2.txt should not exist")
				}
				return nil
			},
		},
		{
			name:      "invalid zip path",
			zipPath:   "nonexistent.zip",
			targetDir: targetDir,
			wantErr:   true,
			checkFn:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean target directory before each test
			os.RemoveAll(tt.targetDir)

			err := Unzip(tt.zipPath, tt.targetDir, tt.files...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unzip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFn != nil {
				if err := tt.checkFn(t, tt.targetDir); err != nil {
					t.Errorf("check failed: %v", err)
				}
			}
		})
	}
}

// createTestZip creates a test zip file with some test content
func createTestZip(zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Test files to create
	files := map[string]string{
		"file1.txt":           "content of file 1",
		"dir1/file2.txt":      "content of file 2",
		"dir1/dir2/file3.txt": "content of file 3",
	}

	// Create each file in the zip
	for name, content := range files {
		f, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			return err
		}
	}

	return nil
}
