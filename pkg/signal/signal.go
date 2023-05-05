package signal

import (
	"context"
	"os"
	"os/signal"
)

func NotifyContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, os.Interrupt, os.Kill)
}
