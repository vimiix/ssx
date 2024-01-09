package utils

import (
	"fmt"
	"os/user"
	"path"
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
