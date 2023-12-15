package entry

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/skeema/knownhosts"
	"golang.org/x/crypto/ssh"

	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/internal/terminal"
	"github.com/vimiix/ssx/internal/utils"
	"github.com/vimiix/ssx/ssx/env"
)

const (
	SourceSSHConfig = "ssh_config"
	SourceSSXStore  = "ssx_store"
)

const defaultIdentityFile = "~/.ssh/id_rsa"

// Entry represent a target server
type Entry struct {
	ID         uint64    `json:"id"`
	Host       string    `json:"host"`
	User       string    `json:"user"`
	Port       string    `json:"port"`
	VisitCount int       `json:"visit_count"` // Perhaps I will support sorting by VisitCount in the future
	KeyPath    string    `json:"key_path"`
	Passphrase string    `json:"passphrase"`
	Password   string    `json:"password"`
	Tags       []string  `json:"tags"`
	Source     string    `json:"source"` // Data source, used to distinguish that it is from ssx stored or local ssh configuration
	CreateAt   time.Time `json:"create_at"`
	UpdateAt   time.Time `json:"update_at"`
	// TODO support jump server
}

func (e *Entry) String() string {
	return fmt.Sprintf("%s@%s:%s", e.User, e.Host, e.Port)
}

func getConnectTimeout() time.Duration {
	var defaultTimeout = time.Second * 10
	val := os.Getenv(env.SSXConnectTimeout)
	if len(val) <= 0 {
		return defaultTimeout
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		lg.Debug("invalid %q value: %q", env.SSXConnectTimeout, val)
		d = defaultTimeout
	}
	return d
}

func (e *Entry) GenSSHConfig() (*ssh.ClientConfig, error) {
	cb, err := e.sshHostKeyCallback()
	if err != nil {
		return nil, err
	}
	cfg := &ssh.ClientConfig{
		User:            e.User,
		Auth:            e.AuthMethods(),
		HostKeyCallback: cb,
		Timeout:         getConnectTimeout(),
	}
	cfg.SetDefaults()
	return cfg, nil
}

func (e *Entry) sshHostKeyCallback() (ssh.HostKeyCallback, error) {
	khPath := utils.ExpandHomeDir("~/.ssh/known_hosts")
	if !utils.FileExists(khPath) {
		f, err := os.OpenFile(khPath, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return nil, err
		}
		_ = f.Close()
	}
	kh, err := knownhosts.New(khPath)
	if err != nil {
		lg.Error("failed to read known_hosts: ", err)
		return nil, err
	}
	// Create a custom permissive hostkey callback which still errors on hosts
	// with changed keys, but allows unknown hosts and adds them to known_hosts
	cb := ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := kh(hostname, remote, key)
		if knownhosts.IsHostKeyChanged(err) {
			lg.Error("REMOTE HOST IDENTIFICATION HAS CHANGED for host %s! This may indicate a MitM attack.", hostname)
			return errors.Errorf("host key changed for host %s", hostname)
		} else if knownhosts.IsHostUnknown(err) {
			f, ferr := os.OpenFile(khPath, os.O_APPEND|os.O_WRONLY, 0600)
			if ferr == nil {
				defer f.Close()
				ferr = knownhosts.WriteKnownHost(f, hostname, remote, key)
			}
			if ferr == nil {
				log.Printf("Added host %s to known_hosts\n", hostname)
			} else {
				log.Printf("Failed to add host %s to known_hosts: %v\n", hostname, ferr)
			}
			return nil
		}
		return err
	})
	return cb, nil
}

func (e *Entry) Tidy() error {
	if len(e.User) <= 0 {
		curUser, err := user.Current()
		if err != nil {
			return err
		}
		e.User = curUser.Username
	}
	if len(e.Port) <= 0 {
		e.Port = "22"
	}
	if e.KeyPath == "" {
		e.KeyPath = defaultIdentityFile
	}
	return nil
}

// AuthMethods all possible auth methods
func (e *Entry) AuthMethods() []ssh.AuthMethod {
	var authMethods []ssh.AuthMethod
	// password auth
	if e.Password != "" {
		authMethods = append(authMethods, ssh.Password(e.Password))
	}

	// key file auth methods
	keyfileAuths := e.privateKeyAuthMethods()
	if len(keyfileAuths) > 0 {
		authMethods = append(authMethods, keyfileAuths...)
	}

	authMethods = append(authMethods, e.interactAuth())
	return authMethods
}

func (e *Entry) interactAuth() ssh.AuthMethod {
	return ssh.KeyboardInteractive(func(name, instruction string, questions []string, echos []bool) (answers []string, err error) {
		answers = make([]string, 0, len(questions))
		for i, q := range questions {
			fmt.Print(q)
			if echos[i] {
				scan := bufio.NewScanner(os.Stdin)
				if scan.Scan() {
					answers = append(answers, scan.Text())
				}
				if err := scan.Err(); err != nil {
					return nil, err
				}
			} else {
				b, err := terminal.ReadPassword()
				if err != nil {
					return nil, err
				}
				fmt.Println()
				answers = append(answers, string(b))
			}
		}
		return answers, nil
	})
}

func (e *Entry) privateKeyAuthMethods() []ssh.AuthMethod {
	keyfiles := e.collectKeyfiles()
	if len(keyfiles) == 0 {
		return nil
	}
	var methods []ssh.AuthMethod
	for _, f := range keyfiles {
		auth := e.keyfileAuth(f)
		if auth != nil {
			methods = append(methods, auth)
		}
	}
	return methods
}

func (e *Entry) keyfileAuth(keypath string) ssh.AuthMethod {
	pemBytes, err := os.ReadFile(keypath)
	if err != nil {
		lg.Debug("generate rsaAuth: failed to read file %q: %s", keypath, err)
		return nil
	}
	var signer ssh.Signer
	if e.Passphrase == "" {
		signer, err = ssh.ParsePrivateKey(pemBytes)
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(e.Passphrase))
	}
	if err != nil {
		lg.Debug("generate rsaAuth: %s", err)
		return nil
	}
	return ssh.PublicKeys(signer)
}

// defaultRSAKeyFiles List of possible key files
// The order of the list represents the priority
var defaultRSAKeyFiles = []string{
	"id_rsa", "id_ecdsa", "id_ecdsa_sk",
	"id_ed25519", "id_ed25519_sk", "id_rsa",
}

func (e *Entry) collectKeyfiles() []string {
	var keypaths []string
	if e.KeyPath != "" && utils.FileExists(e.KeyPath) {
		keypaths = append(keypaths, e.KeyPath)
	}
	u, err := user.Current()
	if err != nil {
		lg.Debug("failed to get current user, ignore default rsa keys")
		return keypaths
	}
	for _, fn := range defaultRSAKeyFiles {
		fp := filepath.Join(u.HomeDir, ".ssh", fn)
		if fp == e.KeyPath || !utils.FileExists(fp) {
			continue
		}
		keypaths = append(keypaths, fp)
	}
	return keypaths
}
