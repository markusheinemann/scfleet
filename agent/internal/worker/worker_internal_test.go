package worker

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
	"github.com/markusheinemann/scfleet/agent/internal/extractor"
	"github.com/markusheinemann/scfleet/agent/internal/scraper"
)

type internalStubScraper struct{ html string }

func (s *internalStubScraper) Fetch(_ context.Context, _, _ string, _ time.Duration) (*scraper.Result, error) {
	return &scraper.Result{HTML: s.html}, nil
}

func TestWorker_ExtractorNewError_ReportsExtractionError(t *testing.T) {
	const jobID = "job-extractor-err"
	failCh := make(chan map[string]any, 1)

	var served atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/jobs/claim", func(w http.ResponseWriter, _ *http.Request) {
		if served.Add(1) > 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		tmpl := map[string]any{"version": "1", "fields": []any{}}
		body, _ := json.Marshal(map[string]any{
			"job_id": jobID, "url": "https://example.com", "template": tmpl, "timeout_s": 30,
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(body) //nolint:errcheck
	})
	mux.HandleFunc(fmt.Sprintf("POST /api/v1/jobs/%s/fail", jobID), func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck
		failCh <- body
		w.WriteHeader(http.StatusNoContent)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := New(
		api.New(srv.URL, "tok", srv.Client(), logger),
		&internalStubScraper{html: "<html></html>"},
		time.Millisecond, 100*time.Millisecond,
		logger,
	)
	w.newExtractor = func(string) (*extractor.Engine, error) {
		return nil, fmt.Errorf("injected parse failure")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go w.Run(ctx) //nolint:errcheck

	select {
	case body := <-failCh:
		cancel()
		if body["error_type"] != "extraction_error" {
			t.Errorf("expected extraction_error, got %q", body["error_type"])
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for fail report")
	}
}
