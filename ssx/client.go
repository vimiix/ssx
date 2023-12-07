package ssx

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/ssx/entry"
)

type Client struct {
	entry *entry.Entry
	cli   *ssh.Client
}

func NewClient(e *entry.Entry) *Client {
	return &Client{entry: e}
}

func (c *Client) Run() error {
	if err := c.login(); err != nil {
		return err
	}
	defer c.close()

	lg.Info("connected server %s, version: %s",
		c.entry.String(), string(c.cli.ServerVersion()))

	session, err := c.cli.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return c.interact(session)
}

func (c *Client) interact(sess *ssh.Session) error {
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() {
		_ = term.Restore(fd, state)
	}()

	w, h, err := term.GetSize(fd)
	if err != nil {
		return err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1, // disable echo command
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	err = sess.RequestPty("xterm", h, w, modes)
	if err != nil {
		return err
	}

	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
	stdinPipe, err := sess.StdinPipe()
	if err != nil {
		return err
	}

	if err = sess.Shell(); err != nil {
		return err
	}

	go func() {
		_, err = io.Copy(stdinPipe, os.Stdin)
		lg.Error(err.Error())
	}()

	go monitorWindowSize(sess)
	go c.keepalive()

	return sess.Wait()
}

func monitorWindowSize(sess *ssh.Session) {
	var (
		fd = int(os.Stdin.Fd())
	)
	ow, oh, err := term.GetSize(fd)
	if err != nil {
		return
	}
	for {
		cw, ch, err := term.GetSize(fd)
		if err != nil {
			break
		}

		if cw != ow || ch != oh {
			if err = sess.WindowChange(ch, cw); err != nil {
				break
			}
			ow = cw
			oh = ch
		}
		time.Sleep(time.Second)
	}
}

func (c *Client) keepalive() {
	for {
		time.Sleep(time.Second * 10)
		_, _, err := c.cli.SendRequest("keepalive@openssh.com", false, nil)
		if err != nil {
			break
		}
	}
}

func (c *Client) login() error {
	network := "tcp"
	addr := net.JoinHostPort(c.entry.Host, c.entry.Port)
	clientConfig := c.entry.GenSSHConfig()
	lg.Info("login to %s", c.entry.String())
	cli, err := ssh.Dial(network, addr, clientConfig)
	if err == nil {
		c.cli = cli
		return nil
	}

	if strings.Contains(err.Error(), "no supported methods remain") && !strings.Contains(err.Error(), "password") {
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
	_ = c.cli.Close()
	c.cli = nil
}
