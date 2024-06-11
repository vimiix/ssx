package ssx

import (
	"context"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/containerd/console"
	"golang.org/x/crypto/ssh"

	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/internal/terminal"
	"github.com/vimiix/ssx/ssx/entry"
)

const (
	NETWORK = "tcp"
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
func dialContext(ctx context.Context, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	d := net.Dialer{Timeout: config.Timeout}
	conn, err := d.DialContext(ctx, NETWORK, addr)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func dialThroughProxy(ctx context.Context, proxy *entry.Proxy, parentProxyCli *ssh.Client, targetEntry *entry.Entry) (*ssh.Client, error) {
	var err error
	if parentProxyCli == nil {
		proxyConfig, err := proxy.GenSSHConfig(ctx)
		if err != nil {
			return nil, err
		}
		lg.Debug("dialing proxy: %s", proxy.String())
		parentProxyCli, err = dialContext(ctx, proxy.Address(), proxyConfig)
		if err != nil {
			lg.Debug("dial proxy %s failed: %v", proxy.String(), err)
			return nil, err
		}
		lg.Debug("proxy client establised")
	}

	var (
		tmpTargetAddr   string
		tmpTargetConfig *ssh.ClientConfig
		tmpHostString   string
	)
	if proxy.Proxy != nil {
		tmpHostString = proxy.Proxy.String()
		tmpTargetAddr = proxy.Proxy.Address()
		tmpTargetConfig, err = proxy.Proxy.GenSSHConfig(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		tmpHostString = targetEntry.String()
		tmpTargetAddr = targetEntry.Address()
		tmpTargetConfig, err = targetEntry.GenSSHConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	lg.Debug("dialing to %s", tmpHostString)
	conn, err := parentProxyCli.DialContext(ctx, NETWORK, tmpTargetAddr)
	if err != nil {
		return nil, err
	}
	nc, chans, reqs, err := ssh.NewClientConn(conn, tmpTargetAddr, tmpTargetConfig)
	if err != nil {
		return nil, err
	}
	targetCli := ssh.NewClient(nc, chans, reqs)
	if proxy.Proxy == nil {
		return targetCli, nil
	}
	return dialThroughProxy(ctx, proxy.Proxy, parentProxyCli, targetEntry)
}

// Login connect remote server and touch enrty in storage
func (c *Client) Login(ctx context.Context) error {
	lg.Debug("connecting to %s", c.entry.String())
	cli, err := c.dial(ctx)
	if err != nil {
		// try fix authentication
		if c.entry.ID != 0 {
			lg.Error("login failed with stored authentication, try login with interactive")
			cli, err = c.tryLoginAgainWithEmptyPassword(ctx)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	c.cli = cli
	if err := c.touchEntry(c.entry); err != nil {
		lg.Error("failed to touch entry: %s", err)
	}
	return nil
}

func (c *Client) tryLoginAgainWithEmptyPassword(ctx context.Context) (*ssh.Client, error) {
	c.entry.ClearPassword()
	return c.dial(ctx)
}

func (c *Client) dial(ctx context.Context) (*ssh.Client, error) {
	if c.entry.Proxy != nil {
		return dialThroughProxy(ctx, c.entry.Proxy, nil, c.entry)
	}
	// connect directly
	sshConfig, err := c.entry.GenSSHConfig(ctx)
	if err != nil {
		return nil, err
	}
	return dialContext(ctx, c.entry.Address(), sshConfig)
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
