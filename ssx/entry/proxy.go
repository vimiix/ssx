package entry

import (
	"context"
	"fmt"
	"net"

	"github.com/vimiix/ssx/internal/utils"
	"golang.org/x/crypto/ssh"
)

// Proxy represents a jump server
// Usage example: ssx -J <jump server1>[,<jump server2>,<jump server3>] <remote server>
type Proxy struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Proxy    *Proxy `json:"proxy"`
}

func (p *Proxy) Mask() {
	if p == nil {
		return
	}
	p.Password = utils.MaskString(p.Password)
	if p.Proxy != nil {
		p.Proxy.Mask()
	}
}

func (p *Proxy) tidy() {
	if p.User == "" {
		p.User = defaultUser
	}
	if p.Port == "" {
		p.Port = defaultPort
	}
	if p.Proxy != nil {
		p.Proxy.tidy()
	}
}

func (p *Proxy) Address() string {
	return net.JoinHostPort(p.Host, p.Port)
}

func (p *Proxy) String() string {
	return fmt.Sprintf("%s@%s:%s", p.User, p.Host, p.Port)
}

func (p *Proxy) GenSSHConfig(ctx context.Context) (*ssh.ClientConfig, error) {
	cb, err := sshHostKeyCallback()
	if err != nil {
		return nil, err
	}
	var auth []ssh.AuthMethod
	if p.Password != "" {
		auth = append(auth, ssh.Password(p.Password))
	} else {
		auth = append(auth, passwordCallback(
			ctx, p.User, p.Host, func(password string) { p.Password = password },
		))
	}
	cfg := &ssh.ClientConfig{
		User:            p.User,
		Auth:            auth,
		HostKeyCallback: cb,
		Timeout:         getConnectTimeout(),
	}
	return cfg, nil
}

func (p *Proxy) ClearPassword() {
	p.Password = ""
	if p.Proxy != nil {
		p.Proxy.ClearPassword()
	}
}
