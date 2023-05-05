package internal

import (
	"errors"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/vimiix/ssx/pkg/util"
)

type Auth struct {
	Password   string `yaml:"password"`
	KeyPath    string `yaml:"keypath"`
	Passphrase string `yaml:"passphrase"`
	Proxy      *Node  `yaml:"proxy"`
}

type User struct {
	Name string `yaml:"name"`
	Auth `yaml:"auth,omitempty"`
}

type Node struct {
	Name     string   `yaml:"name"`
	Alias    []string `yaml:"alias"`
	Hostname string   `yaml:"hostname"`
	IPv4     string   `yaml:"ipv4"`
	Port     string   `yaml:"port"`
	Users    []*User  `yaml:"users"`
}

/*
LoadConfig load config file in following order:
- environment "SSXCONFIG"
- .ssx
- .ssx.yml
- .ssx.yaml
- ~/.ssx
- ~/.ssx.yml
- ~/.ssx.yaml
*/
func LoadConfig() error {
	configFile, err := lookupConfigFile()
	if err != nil {
		return err
	}
	log.Printf("use config file: %q", configFile)
	// TODO
	_ = configFile
	return nil
}

func lookupConfigFile() (path string, err error) {
	path = os.Getenv("SSXCONFIG")
	if path != "" && util.FileExists(path) {
		path = filepath.Clean(path)
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
			if util.FileExists(path) {
				return
			}
		}
	}
	err = errors.New("not found config file")
	return
}
