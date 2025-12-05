package ssx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCpPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *CpPath
	}{
		{
			name:  "local absolute path",
			input: "/tmp/file.txt",
			expected: &CpPath{
				IsRemote: false,
				Path:     "/tmp/file.txt",
			},
		},
		{
			name:  "local relative path",
			input: "./file.txt",
			expected: &CpPath{
				IsRemote: false,
				Path:     "./file.txt",
			},
		},
		{
			name:  "remote with user and host",
			input: "root@192.168.1.100:/tmp/file.txt",
			expected: &CpPath{
				IsRemote: true,
				User:     "root",
				Host:     "192.168.1.100",
				Port:     "",
				Path:     "/tmp/file.txt",
			},
		},
		{
			name:  "remote with user, host and port",
			input: "root@192.168.1.100:22:/tmp/file.txt",
			expected: &CpPath{
				IsRemote: true,
				User:     "root",
				Host:     "192.168.1.100",
				Port:     "22",
				Path:     "/tmp/file.txt",
			},
		},
		{
			name:  "remote with host only",
			input: "192.168.1.100:/tmp/file.txt",
			expected: &CpPath{
				IsRemote: true,
				User:     "",
				Host:     "192.168.1.100",
				Port:     "",
				Path:     "/tmp/file.txt",
			},
		},
		{
			name:  "remote with hostname",
			input: "myserver.example.com:/tmp/file.txt",
			expected: &CpPath{
				IsRemote: true,
				User:     "",
				Host:     "myserver.example.com",
				Port:     "",
				Path:     "/tmp/file.txt",
			},
		},
		{
			name:  "tag with path",
			input: "myserver:/tmp/file.txt",
			expected: &CpPath{
				IsRemote:   true,
				RawKeyword: "myserver",
				Path:       "/tmp/file.txt",
			},
		},
		{
			name:  "tag with home path",
			input: "myserver:~/file.txt",
			expected: &CpPath{
				IsRemote:   true,
				RawKeyword: "myserver",
				Path:       "~/file.txt",
			},
		},
		{
			name:  "remote with underscore in user",
			input: "my_user@host:/path",
			expected: &CpPath{
				IsRemote: true,
				User:     "my_user",
				Host:     "host",
				Port:     "",
				Path:     "/path",
			},
		},
		{
			name:  "remote with dot in user",
			input: "user.name@host:/path",
			expected: &CpPath{
				IsRemote: true,
				User:     "user.name",
				Host:     "host",
				Port:     "",
				Path:     "/path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCpPath(tt.input)
			assert.Equal(t, tt.expected.IsRemote, result.IsRemote, "IsRemote mismatch")
			assert.Equal(t, tt.expected.Path, result.Path, "Path mismatch")
			if tt.expected.IsRemote {
				assert.Equal(t, tt.expected.User, result.User, "User mismatch")
				assert.Equal(t, tt.expected.Host, result.Host, "Host mismatch")
				assert.Equal(t, tt.expected.Port, result.Port, "Port mismatch")
				assert.Equal(t, tt.expected.RawKeyword, result.RawKeyword, "RawKeyword mismatch")
			}
		})
	}
}
