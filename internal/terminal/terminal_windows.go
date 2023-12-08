//go:build windows

package terminal

import (
	"context"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sys/windows"
	"golang.org/x/term"

	"github.com/vimiix/ssx/internal/lg"
)

func ReadPassword() ([]byte, error) {
	return term.ReadPassword(int(windows.Stdin))
}

func GetAndWatchWindowSize(ctx context.Context, sess *ssh.Session) (int, int, error) {
	fd := windows.Stdout
	width, height, err := getConsoleSize(fd)
	if err != nil {
		return 0, 0, err
	}

	go func() {
		if err := watchWindowSize(ctx, sess, fd, width, height); err != nil {
			lg.Debug("watching window size err: %s", err)
		}
	}()

	return width, height, nil
}

func watchWindowSize(ctx context.Context, sess *ssh.Session, fd windows.Handle, width, height int) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(500 * time.Millisecond):
		}

		newWidth, newHeight, err := getConsoleSize(fd)
		if err != nil {
			return err
		}

		if newWidth == width && newHeight == height {
			continue
		}

		width = newWidth
		height = newHeight

		if err = sess.WindowChange(height, width); err != nil {
			return err
		}
	}
}

func getConsoleSize(fd windows.Handle) (int, int, error) {
	var csbi windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(fd, &csbi)
	if err != nil {
		return 0, 0, err
	}

	width := csbi.Window.Right - csbi.Window.Left + 1
	height := csbi.Window.Bottom - csbi.Window.Top + 1

	return int(width), int(height), nil
}
