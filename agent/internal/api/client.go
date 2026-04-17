package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Client sends authenticated requests to the orchestrator API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	logger     *slog.Logger
}

// New creates a Client targeting baseURL with the given bearer token.
func New(baseURL, token string, httpClient *http.Client, logger *slog.Logger) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: httpClient,
		logger:     logger,
	}
}

// Register signals the orchestrator that this agent has started.
func (c *Client) Register(ctx context.Context) error {
	return c.post(ctx, "/api/v1/register")
}

// Heartbeat signals the orchestrator that this agent is still alive.
func (c *Client) Heartbeat(ctx context.Context) error {
	return c.post(ctx, "/api/v1/heartbeat")
}

func (c *Client) post(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	c.logger.Debug("sending request", "method", http.MethodPost, "url", req.URL.String())

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		c.logger.Debug("request failed", "url", c.baseURL+path, "error", err, "duration", elapsed)
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Debug("response received", "status", resp.StatusCode, "duration", elapsed)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}
