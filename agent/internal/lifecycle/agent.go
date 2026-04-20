package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

// APIClient is the interface the Agent uses to communicate with the orchestrator.
type APIClient interface {
	Register(ctx context.Context) error
	Heartbeat(ctx context.Context) error
}

// Worker is the interface for the job polling loop.
type Worker interface {
	Run(ctx context.Context) error
}

// Agent manages registration, periodic heartbeats, and job processing.
type Agent struct {
	client   APIClient
	worker   Worker
	interval time.Duration
	logger   *slog.Logger
}

// New creates an Agent that sends heartbeats on the given interval and runs the worker.
func New(client APIClient, worker Worker, interval time.Duration, logger *slog.Logger) *Agent {
	return &Agent{
		client:   client,
		worker:   worker,
		interval: interval,
		logger:   logger,
	}
}

// Run registers the agent, then runs heartbeats and job polling concurrently until ctx is cancelled.
func (a *Agent) Run(ctx context.Context) error {
	if a.interval <= 0 {
		return fmt.Errorf("invalid heartbeat interval: %s", a.interval)
	}

	a.logger.Info("registering with orchestrator")

	if err := a.client.Register(ctx); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	a.logger.Info("registration successful, starting heartbeat and worker loops", "interval", a.interval)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.heartbeatLoop(gctx)
	})

	g.Go(func() error {
		return a.worker.Run(gctx)
	})

	return g.Wait()
}

func (a *Agent) heartbeatLoop(ctx context.Context) error {
	if err := a.client.Heartbeat(ctx); err != nil {
		a.logger.Warn("heartbeat failed", "error", err)
	} else {
		a.logger.Debug("heartbeat sent")
	}

	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("shutting down")
			return nil
		case <-ticker.C:
			if err := a.client.Heartbeat(ctx); err != nil {
				a.logger.Warn("heartbeat failed", "error", err)
			} else {
				a.logger.Debug("heartbeat sent")
			}
		}
	}
}
