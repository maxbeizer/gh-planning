package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/maxbeizer/gh-planning/cmd"
)

func main() {
	ctx, cancel := signalContext()
	defer cancel()

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		<-sigch
		cancel()
	}()

	return ctx, cancel
}
