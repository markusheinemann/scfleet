package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/markusheinemann/scfleet/agent/internal/api"
	"github.com/markusheinemann/scfleet/agent/internal/lifecycle"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	os.Exit(run(ctx, os.Args[1:], os.Stderr))
}

func run(ctx context.Context, args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.SetOutput(stderr)

	url := fs.String("url", "", "orchestrator base URL (required)")
	token := fs.String("token", "", "agent bearer token (required)")
	interval := fs.Duration("interval", 30*time.Second, "heartbeat interval")
	debug := fs.Bool("debug", false, "enable debug logging")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *url == "" || *token == "" {
		fmt.Fprintln(stderr, "error: --url and --token are required")
		fs.Usage()
		return 1
	}

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: level}))
	logger.Info("agent starting", "url", *url, "interval", *interval)

	client := api.New(*url, *token, nil, logger)
	agent := lifecycle.New(client, *interval, logger)

	if err := agent.Run(ctx); err != nil {
		logger.Error("agent stopped", "error", err)
		return 1
	}

	logger.Info("agent stopped")
	return 0
}
