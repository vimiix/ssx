package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"

	"github.com/vimiix/ssx/cmd/ssx/cmd"
	"github.com/vimiix/ssx/ssx/cleaner"
)

func main() {
	var (
		exitCode = 0
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := cmd.NewRoot().ExecuteContext(ctx); err != nil {
		fmt.Println(color.HiRedString(err.Error()))
		exitCode = 1
	}

	cleaner.Clean()
	os.Exit(exitCode)
}
