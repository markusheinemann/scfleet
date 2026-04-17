package lifecycle

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// APIClient is the interface the Agent uses to communicate with the orchestrator.
type APIClient interface {
	Register(ctx context.Context) error
	Heartbeat(ctx context.Context) error
}

// Agent manages registration and periodic heartbeats to the orchestrator.
type Agent struct {
	client   APIClient
	interval time.Duration
	logger   *slog.Logger
}

// New creates an Agent that will send heartbeats on the given interval.
func New(client APIClient, interval time.Duration, logger *slog.Logger) *Agent {
	return &Agent{
		client:   client,
		interval: interval,
		logger:   logger,
	}
}

// Run registers the agent, then sends heartbeats until ctx is cancelled.
func (a *Agent) Run(ctx context.Context) error {
	if a.interval <= 0 {
		return fmt.Errorf("invalid heartbeat interval: %s", a.interval)
	}

	a.logger.Info("registering with orchestrator")

	if err := a.client.Register(ctx); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	a.logger.Info("registration successful, starting heartbeat loop", "interval", a.interval)

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
