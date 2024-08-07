package entry

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/jinzhu/copier"
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

var (
	defaultIdentityFile = "~/.ssh/id_rsa"
	defaultUser         = "root"
	defaultPort         = "22"
)

const (
	ModeUninit = ""
	ModeSafe   = "safe"
	ModeUnsafe = "unsafe"
)

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
	Proxy      *Proxy    `json:"proxy"`
}

func (e *Entry) String() string {
	return fmt.Sprintf("%s@%s:%s", e.User, e.Host, e.Port)
}

func (e *Entry) Address() string {
	return net.JoinHostPort(e.Host, e.Port)
}

func (e *Entry) JSON() ([]byte, error) {
	entryCopy, err := e.Copy()
	if err != nil {
		return nil, err
	}

	entryCopy.Mask()
	return json.MarshalIndent(entryCopy, "", "    ")
}

func (e *Entry) Copy() (*Entry, error) {
	entryCopy := &Entry{}
	if err := copier.Copy(entryCopy, e); err != nil {
		return nil, err
	}
	return entryCopy, nil
}

func (e *Entry) Mask() {
	e.Password = utils.MaskString(e.Password)
	e.Passphrase = utils.MaskString(e.Passphrase)
	if e.Proxy != nil {
		e.Proxy.Mask()
	}
}

func (e *Entry) ClearPassword() {
	e.Password = ""
	if e.Proxy != nil {
		e.Proxy.ClearPassword()
	}
}

func (e *Entry) KeyFileAbsPath() string {
	return utils.ExpandHomeDir(e.KeyPath)
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

func (e *Entry) GenSSHConfig(ctx context.Context) (*ssh.ClientConfig, error) {
	cb, err := sshHostKeyCallback()
	if err != nil {
		return nil, err
	}
	auths, err := e.AuthMethods(ctx)
	if err != nil {
		return nil, err
	}
	cfg := &ssh.ClientConfig{
		User:            e.User,
		Auth:            auths,
		HostKeyCallback: cb,
		Timeout:         getConnectTimeout(),
	}
	cfg.SetDefaults()
	return cfg, nil
}

func sshHostKeyCallback() (ssh.HostKeyCallback, error) {
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
		lg.Error("failed to read known_hosts: %s", err)
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
				lg.Info("added host %s to known_hosts", hostname)
			} else {
				lg.Warn("failed to add host %s to known_hosts: %v", hostname, ferr)
			}
			return nil
		}
		return err
	})
	return cb, nil
}

// Tidy performs cleanup and validation on the Entry struct.
func (e *Entry) Tidy() error {
	if len(e.User) <= 0 {
		e.User = defaultUser
	}
	if len(e.Port) <= 0 {
		e.Port = defaultPort
	}
	if e.KeyPath == "" && utils.FileExists(defaultIdentityFile) {
		e.KeyPath = defaultIdentityFile
	}
	if e.Proxy != nil {
		e.Proxy.tidy()
	}
	return nil
}

// AuthMethods all possible auth methods
func (e *Entry) AuthMethods(ctx context.Context) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod
	// password auth
	if e.Password != "" {
		authMethods = append(authMethods, ssh.Password(e.Password))
	}

	// key file auth methods
	keyfileAuths, err := e.privateKeyAuthMethods(ctx)
	if err != nil {
		return nil, err
	}
	if len(keyfileAuths) > 0 {
		authMethods = append(authMethods, keyfileAuths...)
	}
	authMethods = append(authMethods, passwordCallback(ctx, e.User, e.Host, func(password string) { e.Password = password }))
	return authMethods, nil
}

func passwordCallback(ctx context.Context, user, host string, storePassFunc func(password string)) ssh.AuthMethod {
	prompt := func() (string, error) {
		lg.Debug("login through password callback")
		fmt.Printf("%s@%s's password:", user, host)
		bs, readErr := terminal.ReadPassword(ctx)
		fmt.Println()
		if readErr != nil {
			return "", readErr
		}
		p := string(bs)
		if storePassFunc != nil {
			storePassFunc(p)
		}
		return p, nil
	}
	return ssh.PasswordCallback(prompt)
}

// At present, I do not know how to correctly capture password information,
// so I need to write promt by myself through passwordCallback to achieve it
// func interactAuth(ctx context.Context, who string) ssh.AuthMethod {
// 	return ssh.KeyboardInteractive(func(name, instruction string, questions []string, echos []bool) (answers []string, err error) {
// 		answers = make([]string, 0, len(questions))
// 		for i, q := range questions {
// 			fmt.Printf("[%s] %s", who, q)
// 			if echos[i] {
// 				scan := bufio.NewScanner(os.Stdin)
// 				if scan.Scan() {
// 					answers = append(answers, scan.Text())
// 				}
// 				if err := scan.Err(); err != nil {
// 					return nil, err
// 				}
// 			} else {
// 				b, err := terminal.ReadPassword(ctx)
// 				if err != nil {
// 					return nil, err
// 				}
// 				fmt.Println()
// 				answers = append(answers, string(b))
// 			}
// 		}
// 		return answers, nil
// 	})
// }

func (e *Entry) privateKeyAuthMethods(ctx context.Context) ([]ssh.AuthMethod, error) {
	keyfiles := e.collectKeyfiles()
	if len(keyfiles) == 0 {
		return nil, nil
	}
	var methods []ssh.AuthMethod
	for _, f := range keyfiles {
		if !utils.FileExists(f) {
			lg.Debug("keyfile %s not found, skip", f)
			continue
		}
		auth, err := e.keyfileAuth(ctx, f)
		if err != nil {
			lg.Debug("skip use keyfile: %s", f)
			continue
		}
		if auth != nil {
			methods = append(methods, auth)
		}
	}
	return methods, nil
}

func (e *Entry) keyfileAuth(ctx context.Context, keypath string) (ssh.AuthMethod, error) {
	lg.Debug("parsing key file: %s", keypath)
	pemBytes, err := os.ReadFile(keypath)
	if err != nil {
		lg.Error("failed to read file %q: %s", keypath, err)
		return nil, err
	}
	var signer ssh.Signer
	signer, err = ssh.ParsePrivateKey(pemBytes)
	passphraseMissingError := &ssh.PassphraseMissingError{}
	if err != nil {
		if keypath != e.KeyFileAbsPath() {
			lg.Debug("parse failed, ignore keyfile %q", keypath)
			return nil, err
		}
		if errors.As(err, &passphraseMissingError) {
			if e.Passphrase != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(e.Passphrase))
			} else {
				fmt.Print("please enter passphrase of key file:")
				bs, readErr := terminal.ReadPassword(ctx)
				fmt.Println()
				if readErr != nil {
					return nil, readErr
				}
				// write back to entry instance
				e.Passphrase = string(bs)
				signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, bs)
			}
		}
	}
	if err != nil {
		lg.Error("failed to parse private key file: %s", err)
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

// defaultRSAKeyFiles List of possible key files
// The order of the list represents the priority
var defaultRSAKeyFiles = []string{
	"id_rsa", "id_ecdsa", "id_ecdsa_sk",
	"id_ed25519", "id_ed25519_sk",
}

func (e *Entry) collectKeyfiles() []string {
	var keypaths []string
	if e.KeyPath != "" && utils.FileExists(e.KeyPath) {
		keypaths = append(keypaths, e.KeyFileAbsPath())
	}
	u, err := user.Current()
	if err != nil {
		lg.Debug("failed to get current user, ignore default rsa keys")
		return keypaths
	}
	for _, fn := range defaultRSAKeyFiles {
		fp := filepath.Join(u.HomeDir, ".ssh", fn)
		if fp == utils.ExpandHomeDir(e.KeyPath) || !utils.FileExists(fp) {
			continue
		}
		keypaths = append(keypaths, fp)
	}
	return keypaths
}
