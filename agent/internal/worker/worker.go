package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/markusheinemann/scfleet/agent/internal/api"
	"github.com/markusheinemann/scfleet/agent/internal/extractor"
	"github.com/markusheinemann/scfleet/agent/internal/scraper"
)

// APIClient is the interface the Worker uses to communicate with the orchestrator.
type APIClient interface {
	PollJob(ctx context.Context) (*api.ClaimedJob, error)
	CompleteJob(ctx context.Context, jobID string, req api.CompleteJobRequest) error
	FailJob(ctx context.Context, jobID string, req api.FailJobRequest) error
	UploadArtifacts(ctx context.Context, jobID string, screenshot []byte, html string) error
}

// Scraper is the interface the Worker uses to fetch pages.
type Scraper interface {
	Fetch(ctx context.Context, url, waitSelector string, timeout time.Duration) (*scraper.Result, error)
}

// Worker runs the poll → scrape → extract → report loop.
type Worker struct {
	client       APIClient
	scraper      Scraper
	logger       *slog.Logger
	pollBase     time.Duration
	pollMax      time.Duration
	pollCurrent  time.Duration
	newExtractor func(string) (*extractor.Engine, error)
}

// New creates a Worker with exponential backoff polling between pollBase and pollMax.
func New(client APIClient, s Scraper, pollBase, pollMax time.Duration, logger *slog.Logger) *Worker {
	return &Worker{
		client:       client,
		scraper:      s,
		logger:       logger,
		pollBase:     pollBase,
		pollMax:      pollMax,
		newExtractor: extractor.New,
	}
}

// Run polls for jobs until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	w.pollCurrent = w.pollBase

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(w.pollCurrent):
			found, err := w.processOnce(ctx)
			if err != nil {
				w.logger.Warn("job processing error", "error", err)
				w.backoff()
			} else if found {
				w.pollCurrent = w.pollBase
			} else {
				w.backoff()
			}
		}
	}
}

// processOnce polls for one job and executes it. Returns true if a job was found.
func (w *Worker) processOnce(ctx context.Context) (bool, error) {
	job, err := w.client.PollJob(ctx)
	if err != nil {
		return false, fmt.Errorf("poll: %w", err)
	}

	if job == nil {
		w.logger.Debug("no pending jobs")
		return false, nil
	}

	w.logger.Info("claimed job", "job_id", job.JobID, "url", job.URL)

	return true, w.executeJob(ctx, job)
}

func (w *Worker) executeJob(ctx context.Context, job *api.ClaimedJob) error {
	var tmpl extractor.Template
	if err := json.Unmarshal(job.Template, &tmpl); err != nil {
		return w.reportFail(ctx, job.JobID, "extraction_error", "invalid template JSON: "+err.Error())
	}

	pageTimeout := time.Duration(job.TimeoutS) * time.Second
	if tmpl.PageTimeoutS > 0 {
		pageTimeout = time.Duration(tmpl.PageTimeoutS) * time.Second
	}

	if pageTimeout <= 0 {
		pageTimeout = 30 * time.Second
	}

	fetched, fetchErr := w.scraper.Fetch(ctx, job.URL, tmpl.JSWaitSelector, pageTimeout)

	if fetched != nil {
		if err := w.client.UploadArtifacts(ctx, job.JobID, fetched.Screenshot, fetched.HTML); err != nil {
			w.logger.Warn("failed to upload artifacts", "job_id", job.JobID, "error", err)
		}
	}

	if fetchErr != nil {
		var se *scraper.ScrapeError
		if errors.As(fetchErr, &se) {
			return w.reportFail(ctx, job.JobID, se.Type, se.Message)
		}

		return w.reportFail(ctx, job.JobID, "navigation_error", fetchErr.Error())
	}

	engine, err := w.newExtractor(fetched.HTML)
	if err != nil {
		return w.reportFail(ctx, job.JobID, "extraction_error", "failed to parse HTML: "+err.Error())
	}

	result, err := engine.Extract(&tmpl)
	if err != nil {
		var ee *extractor.ExtractionError
		errors.As(err, &ee)
		return w.reportFail(ctx, job.JobID, "missing_required_field",
			fmt.Sprintf("required field %q: no extractor yielded a value", ee.FieldName))
	}

	w.logger.Info("job completed", "job_id", job.JobID, "fields", len(result.Data))

	return w.client.CompleteJob(ctx, job.JobID, api.CompleteJobRequest{
		Result:      result.Data,
		FieldErrors: result.FieldErrors,
	})
}

func (w *Worker) reportFail(ctx context.Context, jobID, errType, errMsg string) error {
	w.logger.Warn("job failed", "job_id", jobID, "error_type", errType, "error", errMsg)

	return w.client.FailJob(ctx, jobID, api.FailJobRequest{
		ErrorType:    errType,
		ErrorMessage: errMsg,
	})
}

// backoff doubles the poll interval up to pollMax with ±25% jitter to prevent
// thundering herd when many idle agents poll simultaneously.
func (w *Worker) backoff() {
	next := w.pollCurrent * 2
	if next > w.pollMax {
		next = w.pollMax
	}

	// ±25% jitter
	jitter := time.Duration(rand.Int63n(int64(next/2))) - next/4 //nolint:gosec
	w.pollCurrent = next + jitter

	if w.pollCurrent < w.pollBase {
		w.pollCurrent = w.pollBase
	}
}
