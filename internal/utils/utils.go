package utils

import (
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/denisbrodbeck/machineid"
	"github.com/pkg/errors"
	"github.com/vimiix/ssx/ssx/env"
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

type Address struct {
	User string
	Host string
	Port string
}

var addrRegex = regexp.MustCompile(`^(?:(?P<user>[\w.-_]+)@)?(?P<host>[\w.-]+)(?::(?P<port>\d+))?(?:/(?P<path>[\w/.-]+))?$`)

func MatchAddress(addr string) (*Address, error) {
	matches := addrRegex.FindStringSubmatch(addr)
	if len(matches) == 0 {
		return nil, errors.Errorf("invalid address: %q", addr)
	}
	username, host, port := matches[1], matches[2], matches[3]
	addrObj := &Address{
		User: username,
		Host: host,
		Port: port,
	}
	return addrObj, nil
}

func to16chars(s string) string {
	if len(s) >= 16 {
		return s[:16]
	}
	return s + strings.Repeat("=", 16-len(s))
}

// GetSecretKey get secret key from env, if not set returns machine id
// always returns 16 characters key
func GetSecretKey() (string, error) {
	if os.Getenv(env.SSXSecretKey) != "" {
		return to16chars(os.Getenv(env.SSXSecretKey)), nil
	}
	// ref: https://man7.org/linux/man-pages/man5/machine-id.5.html
	machineID, err := machineid.ProtectedID("ssx")
	if err != nil {
		return "", err
	}
	return to16chars(machineID), nil
}
