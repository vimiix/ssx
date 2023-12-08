package entry

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/internal/utils"
	"github.com/vimiix/ssx/ssx/env"
)

const (
	SourceSSHConfig = "ssh_config"
	SourceSSXStore  = "ssx_store"
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
	Source     string    `json:"source"`
	CreateAt   time.Time `json:"create_at"`
	UpdateAt   time.Time `json:"update_at"`
	// TODO support jump server
}

func (e *Entry) UniqueKey() string {
	return fmt.Sprintf("%s/%s", e.Host, e.User)
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

func (e *Entry) GenSSHConfig() *ssh.ClientConfig {
	cfg := &ssh.ClientConfig{
		User:            e.User,
		Auth:            e.AuthMethods(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         getConnectTimeout(),
	}
	cfg.SetDefaults()
	return cfg
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
				b, err := term.ReadPassword(syscall.Stdin)
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
	if e.KeyPath != "" {
		keypaths = append(keypaths, e.KeyPath)
	}
	u, err := user.Current()
	if err != nil {
		lg.Debug("failed to get current user, ignore default rsa keys")
		return keypaths
	}
	for _, fn := range defaultRSAKeyFiles {
		fp := filepath.Join(u.HomeDir, ".ssh", fn)
		if !utils.FileExists(fp) {
			continue
		}
		keypaths = append(keypaths, fp)
	}
	return keypaths
}
