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
	"github.com/markusheinemann/scfleet/agent/internal/scraper"
	"github.com/markusheinemann/scfleet/agent/internal/worker"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	os.Exit(run(ctx, os.Args[1:], os.Stderr))
}

func run(ctx context.Context, args []string, stderr io.Writer) int {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.SetOutput(stderr)

	orchestratorURL := fs.String("url", "", "orchestrator base URL (required)")
	token := fs.String("token", "", "agent bearer token (required)")
	interval := fs.Duration("interval", 30*time.Second, "heartbeat interval")
	pollBase := fs.Duration("poll-base", 5*time.Second, "initial job poll interval")
	pollMax := fs.Duration("poll-max", 60*time.Second, "maximum job poll interval (with backoff)")
	debug := fs.Bool("debug", false, "enable debug logging")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *orchestratorURL == "" || *token == "" {
		fmt.Fprintln(stderr, "error: --url and --token are required")
		fs.Usage()
		return 1
	}

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: level}))
	logger.Info("agent starting", "url", *orchestratorURL, "interval", *interval)

	client := api.New(*orchestratorURL, *token, nil, logger)
	s := scraper.New(logger)
	w := worker.New(client, s, *pollBase, *pollMax, logger)
	agent := lifecycle.New(client, w, *interval, logger)

	if err := agent.Run(ctx); err != nil {
		logger.Error("agent stopped", "error", err)
		return 1
	}

	logger.Info("agent stopped")

	return 0
}
