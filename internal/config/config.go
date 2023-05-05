package config

import (
	"context"
	"os"
	"os/user"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"

	"github.com/vimiix/ssx/internal/types"
	"github.com/vimiix/ssx/pkg/lg"
	"github.com/vimiix/ssx/pkg/util"
)

var (
	configFile string
)

// Filepath returns effective config file path
func Filepath() string {
	return configFile
}

/*
Load config file in following order:
- environment "SSXCONFIG"
- .ssx
- .ssx.yml
- .ssx.yaml
- ~/.ssx
- ~/.ssx.yml
- ~/.ssx.yaml (default)
*/
func Load(ctx context.Context) (nodes []*types.Node, err error) {
	configFile, err = lookupConfigFile()
	if err != nil {
		return
	}
	if !util.FileExists(configFile) {
		// empty nodes
		return
	}
	lg.Debugf("load config file: %s", configFile)
	var fp *os.File
	fp, err = os.Open(configFile)
	if err != nil {
		err = errors.Wrapf(err, "open file %q", configFile)
		return
	}
	defer fp.Close()
	if err = yaml.NewDecoder(fp).DecodeContext(ctx, &nodes); err != nil {
		err = errors.Wrapf(err, "decode nodes from %s", configFile)
	}
	return
}

// Store save all nodes to config file
func Store(ctx context.Context, nodes []*types.Node) (err error) {
	var fp *os.File
	fp, err = os.OpenFile(configFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	if err = yaml.NewEncoder(fp).EncodeContext(ctx, nodes); err != nil {
		err = errors.Wrapf(err, "encode nodes to %s", configFile)
	}
	return
}

func lookupConfigFile() (path string, err error) {
	path = os.Getenv("SSXCONFIG")
	if path != "" {
		path = filepath.Clean(path)
		lg.Debugf("hit SSXCONFIG env: %s", path)
		return
	}

	var u *user.User
	u, err = user.Current()
	if err != nil {
		return
	}
	for _, fn := range []string{".ssx", ".ssx.yml", ".ssx.yaml"} {
		for _, dir := range []string{".", u.HomeDir} {
			path = filepath.Join(dir, fn)
			lg.Debugf("searching: %s", path)
			if util.FileExists(path) {
				lg.Debugf("hit: %s", path)
				return
			}
		}
	}
	path = filepath.Join(u.HomeDir, ".ssx.yaml")
	lg.Debugf("not found existing config file, use default: %s", path)
	return
}
