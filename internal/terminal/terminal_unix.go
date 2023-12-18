//go:build !windows

package terminal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"github.com/vimiix/ssx/internal/lg"
)

func readPassword() ([]byte, error) {
	return term.ReadPassword(syscall.Stdin)
}

func GetAndWatchWindowSize(ctx context.Context, sess *ssh.Session) (int, int, error) {
	fd := int(os.Stdin.Fd())
	width, height, err := term.GetSize(fd)
	if err != nil {
		return 0, 0, err
	}

	go func() {
		if err := watchWindowSize(ctx, sess, fd); err != nil {
			lg.Debug("watching window size err: %s", err)
		}
	}()

	return width, height, nil
}

func watchWindowSize(ctx context.Context, sess *ssh.Session, fd int) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)

	for {
		select {
		case <-sigChan:
		case <-ctx.Done():
			return nil
		}

		w, h, err := term.GetSize(fd)
		if err != nil {
			return err
		}
		if err = sess.WindowChange(h, w); err != nil {
			return err
		}
	}
}
