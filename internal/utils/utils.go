package utils

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// FileExists check given filename if exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

var getCurrentUserFunc = user.Current

// ExpandHomeDir expands the path to include the home directory if the path is prefixed with `~`.
// If it isn't prefixed with `~`, the path is returned as-is.
func ExpandHomeDir(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	path = filepath.Clean(path)

	u, err := getCurrentUserFunc()
	if err != nil || u.HomeDir == "" {
		return path
	}

	return filepath.Join(u.HomeDir, path[1:])
}

func MaskString(s string) string {
	mask := "***"
	if len(s) == 0 {
		return s
	} else if len(s) <= 3 {
		return s[:1] + mask
	} else {
		return s[:2] + mask + s[len(s)-1:]
	}
}

// CurrentUserName returns the UserName of the current os user
func CurrentUserName() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

// ContainsI a case-insensitive strings contains
func ContainsI(s, sub string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(sub),
	)
}
