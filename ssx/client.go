package ssx

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/ssx/entry"
)

type Client struct {
	repo      Repo
	entry     *entry.Entry
	cli       *ssh.Client
	closeOnce *sync.Once
}

func NewClient(e *entry.Entry, repo Repo) *Client {
	return &Client{
		entry:     e,
		repo:      repo,
		closeOnce: &sync.Once{},
	}
}

func (c *Client) touchEntry(e *entry.Entry) error {
	if e.Source != entry.SourceSSXStore {
		return nil
	}
	return c.repo.TouchEntry(e)
}

func (c *Client) Run(ctx context.Context) error {
	if err := c.login(ctx); err != nil {
		return err
	}
	defer c.close()

	if err := c.touchEntry(c.entry); err != nil {
		return err
	}
	lg.Info("connected server %s, version: %s",
		c.entry.String(), string(c.cli.ServerVersion()))

	session, err := c.cli.NewSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	return c.attach(ctx, session)
}

func (c *Client) attach(ctx context.Context, sess *ssh.Session) error {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() {
		_ = term.Restore(fd, state)
	}()

	w, h, err := getAndWatchWindowSize(ctx, sess)
	if err != nil {
		return err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err = sess.RequestPty("xterm", h, w, modes); err != nil {
		return err
	}

	var closeStdin sync.Once
	stdinPipe, err := sess.StdinPipe()
	if err != nil {
		return err
	}
	defer closeStdin.Do(func() {
		_ = stdinPipe.Close()
	})
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
	go func() {
		defer closeStdin.Do(func() {
			_ = stdinPipe.Close()
		})
		ioCopy(stdinPipe, os.Stdin)
	}()

	if err = sess.Shell(); err != nil {
		return err
	}

	go c.keepalive(ctx)

	_ = sess.Wait() // ignore *ExitError, always exit with code 130
	return nil
}

func ioCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		lg.Error(err.Error())
	}
}

func (c *Client) keepalive(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			ticker.Stop()
			return
		}
		_, _, err := c.cli.SendRequest("keepalive@openssh.com", false, nil)
		if err != nil {
			break
		}
	}
}

// code source: https://github.com/golang/go/issues/20288#issuecomment-832033017
func dialContext(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	d := net.Dialer{Timeout: config.Timeout}
	conn, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func (c *Client) login(ctx context.Context) error {
	network := "tcp"
	addr := net.JoinHostPort(c.entry.Host, c.entry.Port)
	clientConfig := c.entry.GenSSHConfig()
	lg.Info("connecting to %s", c.entry.String())
	cli, err := dialContext(ctx, network, addr, clientConfig)
	if err == nil {
		c.cli = cli
		return nil
	}

	if strings.Contains(err.Error(), "no supported methods remain") {
		fmt.Printf("%s@%s's password:", c.entry.User, c.entry.Host)
		bs, readErr := term.ReadPassword(syscall.Stdin)
		if readErr == nil {
			p := string(bs)
			if p != "" {
				clientConfig.Auth = []ssh.AuthMethod{ssh.Password(p)}
			}
			fmt.Println()
			cli, err = ssh.Dial(network, addr, clientConfig)
			if err == nil {
				c.entry.Password = p
				c.cli = cli
				return nil
			}
		}
	}
	return err
}

func (c *Client) close() {
	if c.cli == nil {
		return
	}
	c.closeOnce.Do(func() {
		_ = c.cli.Close()
		c.cli = nil
	})
}
