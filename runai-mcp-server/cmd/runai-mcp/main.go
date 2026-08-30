package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/run-ai/runai/runai-mcp-server/internal/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rootCmd, err := cli.NewRootCmd()
	if err != nil {
		os.Exit(1)
	}
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
