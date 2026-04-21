package worker_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/markusheinemann/scfleet/agent/internal/api"
	"github.com/markusheinemann/scfleet/agent/internal/scraper"
	"github.com/markusheinemann/scfleet/agent/internal/worker"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// --- mock scrapers ---

type mockScraper struct {
	html  string
	err   error
	calls atomic.Int32
}

func (m *mockScraper) Fetch(_ context.Context, _, _ string, _ time.Duration) (*scraper.Result, error) {
	m.calls.Add(1)
	if m.err != nil {
		return nil, m.err
	}
	return &scraper.Result{HTML: m.html}, nil
}

// partialScraper returns a result AND an error (simulates page_timeout with partial content).
type partialScraper struct {
	html string
	err  error
}

func (m *partialScraper) Fetch(_ context.Context, _, _ string, _ time.Duration) (*scraper.Result, error) {
	return &scraper.Result{HTML: m.html, Screenshot: []byte{0x89, 0x50}}, m.err
}

// capturingScraper records the timeout it was called with.
type capturingScraper struct {
	html            string
	capturedTimeout time.Duration
}

func (m *capturingScraper) Fetch(_ context.Context, _, _ string, timeout time.Duration) (*scraper.Result, error) {
	m.capturedTimeout = timeout
	return &scraper.Result{HTML: m.html}, nil
}

// --- mock orchestrator server ---

const (
	testJobID  = "01HWTEST0000000000000000"
	testJobURL = "https://example.com/product"
)

// claimResponse builds the JSON body the claim endpoint returns.
func claimResponse(jobID, jobURL string) []byte {
	tmpl := map[string]any{
		"version": "1",
		"fields": []map[string]any{
			{
				"name":     "title",
				"type":     "string",
				"required": true,
				"extractors": []map[string]any{
					{"strategy": "css", "selector": "h1"},
				},
			},
		},
	}
	b, _ := json.Marshal(map[string]any{
		"job_id":    jobID,
		"url":       jobURL,
		"template":  tmpl,
		"timeout_s": 30,
	})
	return b
}

// orchestratorMux builds an HTTP mux that simulates the orchestrator API.
// claimCount controls how many times the claim endpoint returns a job (then 204).
// It records complete/fail calls via channels.
func orchestratorMux(t *testing.T, claimsToServe int32) (mux *http.ServeMux, completed chan map[string]any, failed chan map[string]any) {
	t.Helper()
	completed = make(chan map[string]any, 10)
	failed = make(chan map[string]any, 10)

	var served atomic.Int32
	mux = http.NewServeMux()

	mux.HandleFunc("POST /api/v1/jobs/claim", func(w http.ResponseWriter, _ *http.Request) {
		if served.Add(1) > claimsToServe {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(claimResponse(testJobID, testJobURL)) //nolint:errcheck
	})

	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/artifacts", testJobID), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/complete", testJobID), func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck
		completed <- body
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/fail", testJobID), func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck
		failed <- body
		w.WriteHeader(http.StatusNoContent)
	})

	return mux, completed, failed
}

// waitChan blocks until a value arrives on ch or the test deadline is exceeded.
func waitChan[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for channel value")
		var zero T
		return zero
	}
}

// runWorkerOnce builds a Worker, runs it until it processes one job (or times out),
// then cancels.
func runWorkerOnce(t *testing.T, client *api.Client, s worker.Scraper) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	w := worker.New(client, s, time.Millisecond, 100*time.Millisecond, discardLogger)
	w.Run(ctx) //nolint:errcheck
}

// --- integration tests ---

func TestWorker_HappyPath_CompletesJob(t *testing.T) {
	// Serve one job that has an <h1> on the target page, then return 204.
	mux, completed, _ := orchestratorMux(t, 1)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &mockScraper{
		html: `<html><body><h1>Widget Pro</h1></body></html>`,
	}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)

	runWorkerOnce(t, client, sc)

	body := waitChan(t, completed)

	result, ok := body["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map, got %T", body["result"])
	}
	if result["title"] != "Widget Pro" {
		t.Errorf("expected title 'Widget Pro', got %q", result["title"])
	}
}

func TestWorker_NoJob_DoesNotCallScraper(t *testing.T) {
	// Claim always returns 204.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/jobs/claim", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &mockScraper{}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	w := worker.New(client, sc, 5*time.Millisecond, 50*time.Millisecond, discardLogger)
	w.Run(ctx) //nolint:errcheck

	if sc.calls.Load() > 0 {
		t.Errorf("expected no scraper calls when no job available, got %d", sc.calls.Load())
	}
}

func TestWorker_ScraperError_ReportsFailure(t *testing.T) {
	mux, _, failed := orchestratorMux(t, 1)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &mockScraper{
		err: &scraper.ScrapeError{Type: "page_timeout", Message: "timed out after 30s"},
	}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)

	runWorkerOnce(t, client, sc)

	body := waitChan(t, failed)

	if body["error_type"] != "page_timeout" {
		t.Errorf("expected error_type 'page_timeout', got %q", body["error_type"])
	}
}

func TestWorker_MissingRequiredField_ReportsFailure(t *testing.T) {
	mux, _, failed := orchestratorMux(t, 1)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// HTML has no <h1>, but the template requires it.
	sc := &mockScraper{
		html: `<html><body><p>no heading</p></body></html>`,
	}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)

	runWorkerOnce(t, client, sc)

	body := waitChan(t, failed)

	if body["error_type"] != "missing_required_field" {
		t.Errorf("expected error_type 'missing_required_field', got %q", body["error_type"])
	}
}

func TestWorker_ProcessesMultipleJobsSequentially(t *testing.T) {
	// Serve 3 jobs then 204.
	mux, completed, _ := orchestratorMux(t, 3)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &mockScraper{
		html: `<html><body><h1>Product</h1></body></html>`,
	}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	w := worker.New(client, sc, time.Millisecond, 50*time.Millisecond, discardLogger)

	done := make(chan struct{})
	go func() {
		w.Run(ctx) //nolint:errcheck
		close(done)
	}()

	// Collect 3 completions.
	for range 3 {
		waitChan(t, completed)
	}
	cancel()
	<-done

	if sc.calls.Load() < 3 {
		t.Errorf("expected at least 3 scraper calls, got %d", sc.calls.Load())
	}
}

func TestWorker_PollError_BacksOff(t *testing.T) {
	mux := http.NewServeMux()
	var callCount atomic.Int32
	mux.HandleFunc("POST /api/v1/jobs/claim", func(w http.ResponseWriter, _ *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &mockScraper{}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	w := worker.New(client, sc, 5*time.Millisecond, 50*time.Millisecond, discardLogger)
	w.Run(ctx) //nolint:errcheck

	if callCount.Load() == 0 {
		t.Error("expected at least one poll attempt")
	}
	if sc.calls.Load() > 0 {
		t.Error("expected no scraper calls when poll errors")
	}
}

func TestWorker_ArtifactUploadFails_JobStillCompletes(t *testing.T) {
	completedCh := make(chan map[string]any, 1)
	var served atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/jobs/claim", func(w http.ResponseWriter, _ *http.Request) {
		if served.Add(1) > 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(claimResponse(testJobID, testJobURL)) //nolint:errcheck
	})
	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/artifacts", testJobID), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // upload fails
	})
	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/complete", testJobID), func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck
		completedCh <- body
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &mockScraper{html: `<html><body><h1>Widget Pro</h1></body></html>`}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)
	runWorkerOnce(t, client, sc)

	body := waitChan(t, completedCh)
	result, ok := body["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map, got %T", body["result"])
	}
	if result["title"] != "Widget Pro" {
		t.Errorf("expected title 'Widget Pro', got %v", result["title"])
	}
}

func TestWorker_GenericScrapeError_ReportsNavigationError(t *testing.T) {
	mux, _, failed := orchestratorMux(t, 1)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &mockScraper{err: fmt.Errorf("generic network error")}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)

	runWorkerOnce(t, client, sc)

	body := waitChan(t, failed)
	if body["error_type"] != "navigation_error" {
		t.Errorf("expected error_type 'navigation_error', got %q", body["error_type"])
	}
}

func TestWorker_TimeoutWithPartialResult_UploadsArtifactsThenFails(t *testing.T) {
	mux, _, failed := orchestratorMux(t, 1)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &partialScraper{
		html: `<html><body></body></html>`,
		err:  &scraper.ScrapeError{Type: "page_timeout", Message: "timed out after 30s"},
	}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)

	runWorkerOnce(t, client, sc)

	body := waitChan(t, failed)
	if body["error_type"] != "page_timeout" {
		t.Errorf("expected error_type 'page_timeout', got %q", body["error_type"])
	}
}

func TestWorker_InvalidTemplateJSON_ReportsExtractionError(t *testing.T) {
	failCh := make(chan map[string]any, 1)
	const badJobID = "job-bad-tmpl"

	var served atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/jobs/claim", func(w http.ResponseWriter, _ *http.Request) {
		if served.Add(1) > 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// template field is a JSON string, not an object — unmarshal into Template struct will fail.
		body := `{"job_id":"` + badJobID + `","url":"https://example.com","template":"not-an-object","timeout_s":30}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body)) //nolint:errcheck
	})
	mux.HandleFunc("POST /api/v1/jobs/"+badJobID+"/fail", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck
		failCh <- body
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &mockScraper{html: `<html></html>`}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)
	runWorkerOnce(t, client, sc)

	body := waitChan(t, failCh)
	if body["error_type"] != "extraction_error" {
		t.Errorf("expected error_type 'extraction_error', got %q", body["error_type"])
	}
}

func TestWorker_TemplatePageTimeoutOverride(t *testing.T) {
	completedCh := make(chan struct{}, 1)

	var served atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/jobs/claim", func(w http.ResponseWriter, _ *http.Request) {
		if served.Add(1) > 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		tmpl := map[string]any{
			"version":        "1",
			"page_timeout_s": 60,
			"fields": []map[string]any{
				{"name": "title", "type": "string", "required": true,
					"extractors": []map[string]any{{"strategy": "css", "selector": "h1"}}},
			},
		}
		b, _ := json.Marshal(map[string]any{
			"job_id": testJobID, "url": testJobURL, "template": tmpl, "timeout_s": 30,
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(b) //nolint:errcheck
	})
	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/artifacts", testJobID), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/complete", testJobID), func(w http.ResponseWriter, _ *http.Request) {
		completedCh <- struct{}{}
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &capturingScraper{html: `<html><body><h1>Title</h1></body></html>`}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)
	runWorkerOnce(t, client, sc)

	waitChan(t, completedCh)

	if sc.capturedTimeout != 60*time.Second {
		t.Errorf("expected 60s page timeout from template, got %v", sc.capturedTimeout)
	}
}

func TestWorker_BackoffClampsToBase(t *testing.T) {
	// pollMax(2ns) < pollBase(4ns): after one backoff next=2ns, jitter=rand.Int63n(1)-0=0,
	// pollCurrent=2ns < pollBase=4ns → clamp fires deterministically.
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/jobs/claim", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &mockScraper{}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	w := worker.New(client, sc, 4*time.Nanosecond, 2*time.Nanosecond, discardLogger)
	w.Run(ctx) //nolint:errcheck
}

func TestWorker_ZeroTimeoutS_DefaultsTo30s(t *testing.T) {
	completedCh := make(chan struct{}, 1)

	var served atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/jobs/claim", func(w http.ResponseWriter, _ *http.Request) {
		if served.Add(1) > 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// timeout_s: 0 and no page_timeout_s in template → should default to 30s
		tmpl := map[string]any{
			"version": "1",
			"fields": []map[string]any{
				{"name": "title", "type": "string", "required": true,
					"extractors": []map[string]any{{"strategy": "css", "selector": "h1"}}},
			},
		}
		b, _ := json.Marshal(map[string]any{
			"job_id": testJobID, "url": testJobURL, "template": tmpl, "timeout_s": 0,
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(b) //nolint:errcheck
	})
	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/artifacts", testJobID), func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/complete", testJobID), func(w http.ResponseWriter, _ *http.Request) {
		completedCh <- struct{}{}
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	sc := &capturingScraper{html: `<html><body><h1>Title</h1></body></html>`}
	client := api.New(srv.URL, "test-token", srv.Client(), discardLogger)
	runWorkerOnce(t, client, sc)

	waitChan(t, completedCh)

	if sc.capturedTimeout != 30*time.Second {
		t.Errorf("expected default 30s timeout, got %v", sc.capturedTimeout)
	}
}
