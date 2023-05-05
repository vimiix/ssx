package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/vimiix/ssx/cmd/ssx/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := cmd.Root.ExecuteContext(ctx); err != nil {
		cancel()
		os.Exit(1)
	}
}
