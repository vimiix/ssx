package utils

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/denisbrodbeck/machineid"
	"github.com/pkg/errors"
	"github.com/vimiix/ssx/internal/file"
	"github.com/vimiix/ssx/internal/lg"
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

func DownloadFile(ctx context.Context, urlStr string, saveFile string) error {
	_, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	fp, err := os.Create(saveFile)
	if err != nil {
		return err
	}
	defer fp.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.Errorf("request failed:\n- url: %s\n- response: %s", urlStr, resp.Status)
	}
	_, err = io.Copy(fp, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

type closeFunc func()

func openTarball(tarball string) (*tar.Reader, closeFunc, error) {
	if tarball == "" {
		return nil, nil, errors.Errorf("no tarball specified")
	}

	f, err := os.Open(tarball)
	if err != nil {
		return nil, nil, err
	}
	closers := []io.Closer{f}
	var tr *tar.Reader
	gr, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	closers = append(closers, gr)
	tr = tar.NewReader(gr)
	closeFunc := func() {
		for i := len(closers) - 1; i > -1; i-- {
			closers[i].Close()
		}
	}
	return tr, closeFunc, nil
}

func Untar(tarPath string, targetDir string, filenames ...string) error {
	specifiedUntar := false
	if len(filenames) > 0 {
		specifiedUntar = true
	}

	tr, closefunc, err := openTarball(tarPath)
	if err != nil {
		return err
	}
	defer closefunc()

	for {
		header, err := tr.Next()
		switch {
		// if no more files are found return
		case err == io.EOF:
			return nil
		// return any other error
		case err != nil:
			return err
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}
		if strings.Contains(header.Name, "..") {
			// code scanning: https://github.com/vimiix/ssx/security/code-scanning/3
			lg.Warn("ignore file %s due to zip slip vulnerability", header.Name)
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(targetDir, filepath.FromSlash(header.Name))
		switch header.Typeflag {
		// if it's a dir, and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0700); err != nil {
					return err
				}
			}
		// if it's a file create it
		case tar.TypeReg:
			if specifiedUntar {
				if len(filenames) == 0 {
					// 指定要解压的文件都已经找到，应立即返回
					return nil
				}
				targetIdx := -1
				for idx, fn := range filenames {
					if strings.TrimPrefix(fn, "./") == strings.TrimPrefix(header.Name, "./") {
						targetIdx = idx
					}
				}
				if targetIdx == -1 {
					continue
				}
				filenames = append(filenames[:targetIdx], filenames[targetIdx+1:]...)
			}

			dirpath := path.Dir(target)
			if !file.IsExist(dirpath) {
				if err := os.MkdirAll(dirpath, 0700); err != nil {
					return err
				}
			}

			targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			// copy over contents
			if _, err := io.Copy(targetFile, tr); err != nil {
				return err
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			targetFile.Close()
		}
	}
}
