package ssx

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/containerd/console"
	"golang.org/x/crypto/ssh"

	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/internal/terminal"
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

type ExecuteOption struct {
	Command string
	Stdout  io.Writer
	Stderr  io.Writer
	Timeout time.Duration
}

// Execute a command combined stdout and stderr output, then exit
func (c *Client) Execute(ctx context.Context, opt *ExecuteOption) error {
	if opt.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opt.Timeout)
		defer cancel()
	}

	if err := c.Login(ctx); err != nil {
		return err
	}
	defer c.close()

	sess, err := c.cli.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	sess.Stdout = opt.Stdout
	sess.Stderr = opt.Stderr
	return sess.Run(opt.Command)
}

// Interact Bind the current terminal to provide an interactive interface
func (c *Client) Interact(ctx context.Context) error {
	if err := c.Login(ctx); err != nil {
		return err
	}
	defer c.close()

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
	current := console.Current()
	defer func() {
		_ = current.Reset()
	}()

	if err := current.SetRaw(); err != nil {
		return err
	}

	w, h, err := terminal.GetAndWatchWindowSize(ctx, sess)
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

// Login connect remote server and touch enrty in storage
func (c *Client) Login(ctx context.Context) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	if err := c.touchEntry(c.entry); err != nil {
		lg.Error("failed to touch entry: %s", err)
	}
	return nil
}

func (c *Client) connect(ctx context.Context) error {
	network := "tcp"
	addr := net.JoinHostPort(c.entry.Host, c.entry.Port)
	clientConfig, err := c.entry.GenSSHConfig(ctx)
	if err != nil {
		return err
	}
	lg.Debug("connecting to %s", c.entry.String())
	cli, err := dialContext(ctx, network, addr, clientConfig)
	if err == nil {
		c.cli = cli
		return nil
	}

	if strings.Contains(err.Error(), "no supported methods remain") {
		lg.Debug("failed connect by default auth methods, try password again")
		fmt.Printf("%s@%s's password:", c.entry.User, c.entry.Host)
		bs, readErr := terminal.ReadPassword(ctx)
		fmt.Println()
		if readErr == nil {
			p := string(bs)
			if p != "" {
				clientConfig.Auth = []ssh.AuthMethod{ssh.Password(p)}
			}
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
