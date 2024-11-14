package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/vimiix/ssx/ssx/version"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/internal/utils"
)

const (
	GITHUB_LATEST_API = "https://api.github.com/repos/vimiix/ssx/releases/latest"
	GITHUB_PKG_FMT    = "https://github.com/vimiix/ssx/releases/download/v{VERSION}/ssx_v{VERSION}_{OS}_{ARCH}.tar.gz"
)

type upgradeOpt struct {
	PkgPath string
	Version string
}

type LatestPkgInfo struct {
	Version     string
	DownloadURL string
}

func newUpgradeCmd() *cobra.Command {
	opt := &upgradeOpt{}
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade ssx version",
		Example: `# Upgrade online
ssx upgrade [<version>]

# Upgrade with local filepath or specify new package URL path
ssx upgrade -p <PATH>

# If both version and package path are specified,
# ssx prefer to use package path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opt.Version = args[0]
			}
			return upgrade(cmd.Context(), opt)
		}}
	cmd.Flags().StringVarP(&opt.PkgPath, "package", "p", "", "new package file or URL path")
	return cmd
}

func unifyArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64", "x86_64":
		return "x86_64", nil
	case "arm64", "aarch64":
		return "arm64", nil
	default:
		return "", errors.Errorf("not supported architecture: %s", runtime.GOARCH)
	}
}

func upgrade(ctx context.Context, opt *upgradeOpt) error {
	tempDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return err
	}
	lg.Debug("make temp dir: %s", tempDir)
	defer os.RemoveAll(tempDir)
	var localPkg string
	if opt.PkgPath != "" {
		if strings.Contains(opt.PkgPath, "://") {
			localPkg = filepath.Join(tempDir, "ssx.tar.gz")
			lg.Info("downloading package from %s", opt.PkgPath)
			if err := utils.DownloadFile(ctx, opt.PkgPath, localPkg); err != nil {
				return err
			}
		} else {
			if !utils.FileExists(opt.PkgPath) {
				return errors.Errorf("file not found: %s", opt.PkgPath)
			}
			localPkg = opt.PkgPath
		}
	} else if opt.Version != "" {
		semVer := strings.TrimPrefix(opt.Version, "v")
		if len(strings.Split(semVer, ".")) != 3 {
			return errors.Errorf("bad version: %s", opt.Version)
		}
		arch, err := unifyArch()
		if err != nil {
			return err
		}
		replacer := strings.NewReplacer("{VERSION}", semVer, "{OS}", runtime.GOOS, "{ARCH}", arch)
		urlStr := replacer.Replace(GITHUB_PKG_FMT)
		localPkg = filepath.Join(tempDir, "ssx.tar.gz")
		lg.Info("downloading package from %s", urlStr)
		if err := utils.DownloadFile(ctx, urlStr, localPkg); err != nil {
			return err
		}
	} else {
		lg.Info("detecting latest package info")
		pkgInfo, err := getLatestPkgInfo()
		if err != nil {
			return err
		}
		// check version
		lg.Info("latest version: %s, current version: %s", pkgInfo.Version, version.Version)
		if pkgInfo.Version == version.Version {
			lg.Info("You are currently using the latest version.")
			return nil
		}
		if pkgInfo.DownloadURL == "" {
			return errors.New("failed to get latest package url")
		}
		localPkg = filepath.Join(tempDir, "ssx.tar.gz")
		lg.Info("downloading latest package from %s", pkgInfo.DownloadURL)
		if err := utils.DownloadFile(ctx, pkgInfo.DownloadURL, localPkg); err != nil {
			return err
		}
	}
	lg.Info("extracting package")
	if err := utils.Untar(localPkg, tempDir); err != nil {
		return err
	}
	newBin := filepath.Join(tempDir, "ssx")
	if !utils.FileExists(newBin) {
		return errors.New("not found ssx binary after extracting package")
	}
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execAbsPath, err := filepath.Abs(execPath)
	if err != nil {
		return err
	}
	lg.Info("replacing old binary with new binary")
	if err := replaceBinary(newBin, execAbsPath); err != nil {
		return err
	}
	lg.Info("upgrade success")
	return nil
}

func getLatestPkgInfo() (*LatestPkgInfo, error) {
	arch, err := unifyArch()
	if err != nil {
		return nil, err
	}
	r, err := http.Get(GITHUB_LATEST_API)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	jsonBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	stringBody := string(jsonBody)
	// get latest version by tag name
	latestVersion := gjson.Get(stringBody, "tag_name")
	// get download url
	downloadUrl := gjson.Get(stringBody,
		fmt.Sprintf(`assets.#(name%%"*%s_%s.tar.gz").browser_download_url`, runtime.GOOS, arch))
	return &LatestPkgInfo{
		Version:     latestVersion.String(),
		DownloadURL: downloadUrl.String(),
	}, nil
}

func replaceBinary(newBin string, oldBin string) error {
	bakBin := oldBin + ".bak"
	lg.Debug("backup old binary from %s to %s", oldBin, bakBin)
	if err := os.Link(oldBin, bakBin); err != nil {
		return err
	}

	lg.Debug("remove old binary")
	if err := os.RemoveAll(oldBin); err != nil {
		return err
	}

	lg.Debug("make the new binary effective")
	if err := utils.CopyFile(newBin, oldBin, 0700); err != nil {
		_ = os.RemoveAll(oldBin)
		renameErr := os.Rename(bakBin, oldBin)
		if renameErr != nil {
			lg.Warn("restore old binary failed, please rename it manually\n"+
				"    mv %s %s", bakBin, oldBin)
		}
		return err
	}
	_ = os.RemoveAll(bakBin)
	return nil
}
