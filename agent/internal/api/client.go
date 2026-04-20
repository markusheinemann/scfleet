package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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

// PollJob attempts to claim a pending job.
// Returns (nil, nil) when no jobs are available (204 response).
func (c *Client) PollJob(ctx context.Context) (*ClaimedJob, error) {
	target, err := url.JoinPath(c.baseURL, "/api/v1/jobs/claim")
	if err != nil {
		return nil, fmt.Errorf("build url: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("poll job: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("poll job: unexpected status %d", resp.StatusCode)
	}

	var job ClaimedJob
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("decode job: %w", err)
	}

	return &job, nil
}

// CompleteJob reports successful extraction results.
func (c *Client) CompleteJob(ctx context.Context, jobID string, req CompleteJobRequest) error {
	return c.postJSON(ctx, "/api/v1/jobs/"+jobID+"/complete", req)
}

// FailJob reports that the job failed.
func (c *Client) FailJob(ctx context.Context, jobID string, req FailJobRequest) error {
	return c.postJSON(ctx, "/api/v1/jobs/"+jobID+"/fail", req)
}

// UploadArtifacts sends the page screenshot and HTML to the orchestrator for debugging.
func (c *Client) UploadArtifacts(ctx context.Context, jobID string, screenshot []byte, html string) error {
	req := UploadArtifactsRequest{HTML: html}
	if len(screenshot) > 0 {
		req.Screenshot = base64.StdEncoding.EncodeToString(screenshot)
	}

	return c.postJSON(ctx, "/api/v1/jobs/"+jobID+"/artifacts", req)
}

func (c *Client) postJSON(ctx context.Context, path string, body any) error {
	target, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return fmt.Errorf("build url: %w", err)
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		c.logger.Debug("request failed", "url", target, "error", err, "duration", elapsed)
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) post(ctx context.Context, path string) error {
	target, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return fmt.Errorf("build url: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
	req.Header.Set("Authorization", "Bearer "+c.token)

	c.logger.Debug("sending request", "method", http.MethodPost, "url", req.URL.String())

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		c.logger.Debug("request failed", "url", target, "error", err, "duration", elapsed)
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Debug("response received", "status", resp.StatusCode, "duration", elapsed)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}
