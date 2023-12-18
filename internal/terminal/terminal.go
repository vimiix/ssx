package terminal

import (
	"context"

	"github.com/containerd/console"
)

func ReadPassword(ctx context.Context) ([]byte, error) {
	c := console.Current()
	defer func() {
		_ = c.Reset()
	}()

	var (
		errch    = make(chan error, 1)
		password []byte
	)

	go func() {
		bs, readErr := readPassword()
		if readErr != nil {
			errch <- readErr
		}
		password = bs
		errch <- nil
	}()

	select {
	case err := <-errch:
		return password, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
