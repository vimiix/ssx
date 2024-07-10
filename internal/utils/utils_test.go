package utils

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vimiix/ssx/ssx/env"
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

func TestGetSecretKeyShort(t *testing.T) {
	os.Setenv(env.SSXSecretKey, "abc")
	res, err := GetDeviceID()
	assert.NoError(t, err)
	assert.Equal(t, "abc=============", res)
}

func TestGetSecretKeyLong(t *testing.T) {
	os.Setenv(env.SSXSecretKey, "abcdefghijklmnopqrstuvwxyz")
	res, err := GetDeviceID()
	assert.NoError(t, err)
	assert.Equal(t, "abcdefghijklmnop", res)
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
