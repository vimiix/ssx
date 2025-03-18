package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/ssx/env"
)

// FileExists check given filename if exists
func FileExists(filename string) bool {
	if filename == "" {
		return false
	}
	_, err := os.Stat(ExpandHomeDir(filename))
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

var addrRegex = regexp.MustCompile(`^(?:(?P<user>[\w.\-_]+)@)?(?P<host>[\w.-]+)(?::(?P<port>\d+))?(?:/(?P<path>[\w/.-]+))?$`)

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

// GetDeviceID get secret key from env, if not set returns machine id
// always returns 16 characters key
func GetDeviceID() (string, error) {
	if os.Getenv(env.SSXDeviceID) != "" {
		return os.Getenv(env.SSXDeviceID), nil
	}
	if os.Getenv(env.SSXSecretKey) != "" {
		lg.Warn("env SSX_SECRET_KEY is deprecated, please use SSX_DEVICE_ID instead")
		return os.Getenv(env.SSXSecretKey), nil
	}
	// ref: https://man7.org/linux/man-pages/man5/machine-id.5.html
	machineID, err := machineid.ProtectedID("ssx")
	if err != nil {
		return "", errors.Wrap(err, "failed to get machine id")
	}
	return machineID, nil
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
					// The specified files to be extracted have all been found,
					// and should be returned immediately.
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
			if !FileExists(dirpath) {
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

func Unzip(zipPath string, targetDir string, filenames ...string) error {
	// Open the zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Check if specific files are specified
	specifiedUnzip := len(filenames) > 0

	// Iterate through all files in zip
	for _, file := range reader.File {
		// Prevent zip slip vulnerability
		if strings.Contains(file.Name, "..") {
			lg.Warn("ignore file %s due to zip slip vulnerability", file.Name)
			continue
		}

		// If files are specified, check if current file is in the list
		if specifiedUnzip {
			if len(filenames) == 0 {
				// All specified files have been extracted, return early
				return nil
			}
			targetIdx := -1
			for idx, fn := range filenames {
				if strings.TrimPrefix(fn, "./") == strings.TrimPrefix(file.Name, "./") {
					targetIdx = idx
				}
			}
			if targetIdx == -1 {
				continue
			}
			// Remove processed file from the pending list
			filenames = append(filenames[:targetIdx], filenames[targetIdx+1:]...)
		}

		// Build target path
		target := filepath.Join(targetDir, filepath.FromSlash(file.Name))

		// Handle directories
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0700); err != nil {
				return err
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return err
		}

		// Create target file
		targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		// Open source file
		srcFile, err := file.Open()
		if err != nil {
			targetFile.Close()
			return err
		}

		// Copy contents
		_, err = io.Copy(targetFile, srcFile)

		// Close files
		srcFile.Close()
		targetFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func Extract(pkg string, targetDir string) error {
	if strings.HasSuffix(pkg, ".tar.gz") {
		return Untar(pkg, targetDir)
	}

	if strings.HasSuffix(pkg, ".zip") {
		return Unzip(pkg, targetDir)
	}

	return errors.New("unsupported archive format")
}

// CopyFile copies the contents of src to dst
func CopyFile(src, dst string, perm os.FileMode) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()
	tf, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer tf.Close()
	_, err = io.Copy(tf, sf)
	return err
}

// HashWithSHA256 hashes the input string using SHA-256 and returns the hexadecimal representation of the hash
func HashWithSHA256(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	hashedBytes := hash.Sum(nil)
	return hex.EncodeToString(hashedBytes)
}
