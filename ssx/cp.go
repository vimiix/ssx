package ssx

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"

	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/internal/utils"
	"github.com/vimiix/ssx/ssx/entry"
)

// CpPath represents a parsed path (local or remote)
type CpPath struct {
	IsRemote   bool
	User       string
	Host       string
	Port       string
	Path       string
	Entry      *entry.Entry // resolved entry for remote path
	RawKeyword string       // original keyword/tag used
}

// remotePathRegex matches [user@]host[:port]:/path format
// Examples: root@192.168.1.1:/tmp/file, user@host:22:/path
// Host must contain a dot (domain) or be an IP address to be recognized as a real host
var remotePathRegex = regexp.MustCompile(`^(?:(?P<user>[\w.\-_]+)@)?(?P<host>[\w.\-]+)(?::(?P<port>\d+))?:(?P<path>.+)$`)

// ipOrDomainRegex checks if a string looks like an IP address or domain name
var ipOrDomainRegex = regexp.MustCompile(`^(\d{1,3}\.){1,3}\d{1,3}$|\.`)

// ParseCpPath parses a path string into CpPath struct
// It determines if the path is local or remote based on the format
func ParseCpPath(pathStr string) *CpPath {
	// Try to match remote path format
	matches := remotePathRegex.FindStringSubmatch(pathStr)
	if len(matches) > 0 {
		user, host, port, path := matches[1], matches[2], matches[3], matches[4]

		// If host looks like an IP or domain (contains dot), treat as real host
		// Otherwise, if no user specified and no port, it might be a tag
		if ipOrDomainRegex.MatchString(host) || user != "" || port != "" {
			return &CpPath{
				IsRemote: true,
				User:     user,
				Host:     host,
				Port:     port,
				Path:     path,
			}
		}

		// Looks like a tag/keyword (e.g., myserver:/path)
		return &CpPath{
			IsRemote:   true,
			RawKeyword: host,
			Path:       path,
		}
	}

	// Check if it's a tag/keyword:path format (e.g., myserver:/path)
	if idx := strings.Index(pathStr, ":"); idx > 0 {
		// Make sure it's not a Windows absolute path like C:\
		prefix := pathStr[:idx]
		suffix := pathStr[idx+1:]
		// If prefix doesn't look like a drive letter and suffix starts with /
		if len(prefix) > 1 && (strings.HasPrefix(suffix, "/") || strings.HasPrefix(suffix, "~")) {
			return &CpPath{
				IsRemote:   true,
				RawKeyword: prefix,
				Path:       suffix,
			}
		}
	}

	// It's a local path
	return &CpPath{
		IsRemote: false,
		Path:     pathStr,
	}
}

// CpOption holds options for cp command
type CpOption struct {
	Source       string
	Target       string
	IdentityFile string
	JumpServers  string
	Port         int
	Recursive    bool
}

// Copy performs file copy between local and remote, or remote to remote
func (s *SSX) Copy(ctx context.Context, opt *CpOption) error {
	srcPath := ParseCpPath(opt.Source)
	dstPath := ParseCpPath(opt.Target)

	// Local to local: not supported
	if !srcPath.IsRemote && !dstPath.IsRemote {
		return errors.New("local to local copy should use system cp command")
	}

	// Remote to remote: stream transfer through local
	if srcPath.IsRemote && dstPath.IsRemote {
		return s.copyRemoteToRemote(ctx, srcPath, dstPath, opt)
	}

	// Local to remote or remote to local
	var (
		remotePath *CpPath
		localPath  string
		isUpload   bool
	)

	if srcPath.IsRemote {
		remotePath = srcPath
		localPath = dstPath.Path
		isUpload = false
	} else {
		remotePath = dstPath
		localPath = srcPath.Path
		isUpload = true
	}

	// Resolve remote entry
	e, err := s.resolveRemotePath(remotePath, opt)
	if err != nil {
		return errors.Wrap(err, "failed to resolve remote path")
	}
	remotePath.Entry = e

	// Create SSH client and connect
	client := NewClient(e, s.repo)
	if err := client.Login(ctx); err != nil {
		return errors.Wrap(err, "failed to connect to remote host")
	}
	defer client.close()

	// Create SCP client from existing SSH connection
	scpClient, err := scp.NewClientBySSH(client.cli)
	if err != nil {
		return errors.Wrap(err, "failed to create SCP client")
	}

	if isUpload {
		return s.upload(ctx, scpClient, client.cli, localPath, remotePath)
	}
	return s.download(ctx, scpClient, remotePath, localPath)
}

// copyRemoteToRemote copies file from one remote host to another via streaming
// The file is streamed through local without being stored on disk
func (s *SSX) copyRemoteToRemote(ctx context.Context, srcPath, dstPath *CpPath, opt *CpOption) error {
	// Resolve source entry
	srcEntry, err := s.resolveRemotePath(srcPath, opt)
	if err != nil {
		return errors.Wrap(err, "failed to resolve source remote path")
	}
	srcPath.Entry = srcEntry

	// Resolve destination entry
	dstEntry, err := s.resolveRemotePath(dstPath, opt)
	if err != nil {
		return errors.Wrap(err, "failed to resolve destination remote path")
	}
	dstPath.Entry = dstEntry

	lg.Info("copying %s:%s -> %s:%s (streaming)", srcEntry.Address(), srcPath.Path, dstEntry.Address(), dstPath.Path)

	// Connect to source host
	srcClient := NewClient(srcEntry, s.repo)
	if err := srcClient.Login(ctx); err != nil {
		return errors.Wrap(err, "failed to connect to source host")
	}
	defer srcClient.close()

	// Connect to destination host
	dstClient := NewClient(dstEntry, s.repo)
	if err := dstClient.Login(ctx); err != nil {
		return errors.Wrap(err, "failed to connect to destination host")
	}
	defer dstClient.close()

	// Get file info from source (size and permissions)
	fileInfo, err := getRemoteFileInfo(srcClient.cli, srcPath.Path)
	if err != nil {
		return errors.Wrap(err, "failed to get source file info")
	}

	lg.Debug("source file size: %d, mode: %s", fileInfo.size, fileInfo.mode)

	// Resolve destination path (handle directory case)
	finalDstPath, err := resolveRemoteDestPath(dstClient.cli, dstPath.Path, filepath.Base(srcPath.Path))
	if err != nil {
		return errors.Wrap(err, "failed to resolve destination path")
	}

	// Create pipe for streaming
	pr, pw := io.Pipe()

	// Error channels for goroutines
	downloadErrCh := make(chan error, 1)
	uploadErrCh := make(chan error, 1)

	// Start download goroutine (source -> pipe)
	go func() {
		defer pw.Close()
		scpSrc, err := scp.NewClientBySSH(srcClient.cli)
		if err != nil {
			downloadErrCh <- errors.Wrap(err, "failed to create source SCP client")
			return
		}
		err = scpSrc.CopyFromRemotePassThru(ctx, pw, srcPath.Path, nil)
		downloadErrCh <- err
	}()

	// Start upload goroutine (pipe -> destination)
	go func() {
		scpDst, err := scp.NewClientBySSH(dstClient.cli)
		if err != nil {
			uploadErrCh <- errors.Wrap(err, "failed to create destination SCP client")
			return
		}
		err = scpDst.CopyFile(ctx, pr, finalDstPath, fileInfo.mode)
		uploadErrCh <- err
	}()

	// Wait for both operations to complete
	var downloadErr, uploadErr error
	for i := 0; i < 2; i++ {
		select {
		case downloadErr = <-downloadErrCh:
			if downloadErr != nil {
				pr.Close() // Signal upload to stop
			}
		case uploadErr = <-uploadErrCh:
			if uploadErr != nil {
				pw.Close() // Signal download to stop
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if downloadErr != nil {
		return errors.Wrap(downloadErr, "failed to download from source")
	}
	if uploadErr != nil {
		return errors.Wrap(uploadErr, "failed to upload to destination")
	}

	lg.Info("remote to remote copy completed successfully")
	return nil
}

// remoteFileInfo holds basic file information from remote host
type remoteFileInfo struct {
	size int64
	mode string
}

// getRemoteFileInfo gets file size and permissions from remote host via SSH
func getRemoteFileInfo(client *ssh.Client, remotePath string) (*remoteFileInfo, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// Use stat command to get file info
	// Output format: size mode (e.g., "1234 0644")
	cmd := fmt.Sprintf(`stat -c '%%s %%a' '%s' 2>/dev/null || stat -f '%%z %%Lp' '%s'`, remotePath, remotePath)
	output, err := session.Output(cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to stat remote file %s", remotePath)
	}

	var size int64
	var mode string
	_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%d %s", &size, &mode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse stat output")
	}

	// Ensure mode is 4 digits
	if len(mode) < 4 {
		mode = "0" + mode
	}

	return &remoteFileInfo{size: size, mode: mode}, nil
}

// isRemoteDir checks if a remote path is a directory
func isRemoteDir(client *ssh.Client, remotePath string) (bool, error) {
	session, err := client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	// Use test -d to check if path is a directory
	cmd := fmt.Sprintf(`test -d '%s' && echo "dir" || echo "notdir"`, remotePath)
	output, err := session.Output(cmd)
	if err != nil {
		return false, nil // Path doesn't exist, treat as file
	}

	return strings.TrimSpace(string(output)) == "dir", nil
}

// resolveRemoteDestPath resolves the final destination path on remote host
// If destPath is a directory, appends srcFileName to it
func resolveRemoteDestPath(client *ssh.Client, destPath, srcFileName string) (string, error) {
	isDir, err := isRemoteDir(client, destPath)
	if err != nil {
		return "", err
	}

	if isDir {
		// Destination is a directory, append source filename
		return filepath.Join(destPath, srcFileName), nil
	}

	// Destination is a file path (or doesn't exist yet)
	return destPath, nil
}

// resolveRemotePath resolves a remote CpPath to an Entry
func (s *SSX) resolveRemotePath(cp *CpPath, opt *CpOption) (*entry.Entry, error) {
	// If we have a raw keyword (tag/partial match), search for it
	if cp.RawKeyword != "" {
		lg.Debug("resolving remote path by keyword: %s", cp.RawKeyword)
		e, err := s.searchEntry(cp.RawKeyword)
		if err != nil {
			return nil, err
		}
		// Apply identity file if specified
		if opt.IdentityFile != "" {
			e.KeyPath = opt.IdentityFile
		}
		return e, nil
	}

	// Build address string for search
	addr := cp.Host
	if cp.User != "" {
		addr = cp.User + "@" + addr
	}
	if cp.Port != "" {
		addr = addr + ":" + cp.Port
	}

	lg.Debug("resolving remote path by address: %s", addr)

	// Try to find existing entry or create new one
	e, err := s.searchEntry(addr)
	if err != nil {
		return nil, err
	}

	// Apply identity file if specified
	if opt.IdentityFile != "" {
		e.KeyPath = opt.IdentityFile
	}

	return e, nil
}

// upload copies a local file to remote host
func (s *SSX) upload(ctx context.Context, scpClient scp.Client, sshClient *ssh.Client, localPath string, remotePath *CpPath) error {
	localPath = utils.ExpandHomeDir(localPath)

	// Check if local file exists
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return errors.Wrapf(err, "failed to stat local file %s", localPath)
	}

	if fileInfo.IsDir() {
		return errors.New("directory upload is not supported yet, please use tar to archive first")
	}

	// Open local file
	f, err := os.Open(localPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open local file %s", localPath)
	}
	defer f.Close()

	// Get file permission
	perm := fmt.Sprintf("%04o", fileInfo.Mode().Perm())

	// Resolve destination path (handle directory case)
	finalRemotePath, err := resolveRemoteDestPath(sshClient, remotePath.Path, filepath.Base(localPath))
	if err != nil {
		return errors.Wrap(err, "failed to resolve remote destination path")
	}

	lg.Info("uploading %s -> %s:%s", localPath, remotePath.Entry.Address(), finalRemotePath)

	// Copy file to remote
	err = scpClient.CopyFromFile(ctx, *f, finalRemotePath, perm)
	if err != nil {
		return errors.Wrap(err, "failed to upload file")
	}

	lg.Info("upload completed successfully")
	return nil
}

// download copies a remote file to local
func (s *SSX) download(ctx context.Context, scpClient scp.Client, remotePath *CpPath, localPath string) error {
	localPath = utils.ExpandHomeDir(localPath)

	// Resolve local destination path (handle directory case)
	localInfo, err := os.Stat(localPath)
	if err == nil && localInfo.IsDir() {
		// Local path is a directory, append source filename
		localPath = filepath.Join(localPath, filepath.Base(remotePath.Path))
	}

	// Create local file
	f, err := os.Create(localPath)
	if err != nil {
		return errors.Wrapf(err, "failed to create local file %s", localPath)
	}
	defer f.Close()

	lg.Info("downloading %s:%s -> %s", remotePath.Entry.Address(), remotePath.Path, localPath)

	// Copy file from remote
	err = scpClient.CopyFromRemote(ctx, f, remotePath.Path)
	if err != nil {
		// Clean up partial file on error
		f.Close()
		os.Remove(localPath)
		return errors.Wrap(err, "failed to download file")
	}

	lg.Info("download completed successfully")
	return nil
}
