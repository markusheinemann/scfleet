package scraper_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/markusheinemann/scfleet/agent/internal/scraper"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestNew_ReturnsNonNil(t *testing.T) {
	s := scraper.New(discardLogger)
	if s == nil {
		t.Fatal("expected non-nil scraper")
	}
}

func TestScrapeError_Error(t *testing.T) {
	tests := []struct {
		errType string
		msg     string
		want    string
	}{
		{"page_timeout", "did not load within 30s", "page_timeout: did not load within 30s"},
		{"navigation_error", "DNS failure", "navigation_error: DNS failure"},
		{"navigation_error", "", "navigation_error: "},
	}
	for _, tc := range tests {
		err := &scraper.ScrapeError{Type: tc.errType, Message: tc.msg}
		if err.Error() != tc.want {
			t.Errorf("expected %q, got %q", tc.want, err.Error())
		}
	}
}

// TestFetch_CancelledContext verifies that Fetch propagates a pre-cancelled context as a
// navigation_error without attempting to launch Chrome. chromedp checks ctx.Done() before
// allocating the browser, so this is reliable regardless of whether Chrome is installed.
func TestFetch_CancelledContext_ReturnsNavigationError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Fetch is called

	s := scraper.New(discardLogger)
	result, err := s.Fetch(ctx, "https://example.com", "", 30*time.Second)

	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	se, ok := err.(*scraper.ScrapeError)
	if !ok {
		t.Fatalf("expected *ScrapeError, got %T: %v", err, err)
	}
	if se.Type != "navigation_error" {
		t.Errorf("expected navigation_error, got %q", se.Type)
	}
}

// skipIfNoChrome skips the test when no Chrome/Chromium binary is found in PATH.
func skipIfNoChrome(t *testing.T) {
	t.Helper()
	for _, bin := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
		if _, err := exec.LookPath(bin); err == nil {
			return
		}
	}
	t.Skip("Chrome/Chromium not found in PATH; skipping integration test")
}

// TestFetch_Integration_HappyPath covers the ActionFunc closure body (webdriver script
// injection) and the success return path, which both require a real Chrome session.
func TestFetch_Integration_HappyPath(t *testing.T) {
	skipIfNoChrome(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "<html><head></head><body><h1>Hello</h1></body></html>")
	}))
	defer srv.Close()

	s := scraper.New(discardLogger)
	result, err := s.Fetch(context.Background(), srv.URL, "", 30*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil || result.HTML == "" {
		t.Error("expected non-empty HTML in result")
	}
}

// TestFetch_WaitSelector_CancelledContext additionally exercises the waitSelector branch
// (appending WaitVisible to the task list) before the context cancellation is detected.
func TestFetch_WaitSelector_CancelledContext_ReturnsNavigationError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := scraper.New(discardLogger)
	_, err := s.Fetch(ctx, "https://example.com", "h1.title", 30*time.Second)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	se, ok := err.(*scraper.ScrapeError)
	if !ok {
		t.Fatalf("expected *ScrapeError, got %T: %v", err, err)
	}
	if se.Type != "navigation_error" {
		t.Errorf("expected navigation_error, got %q", se.Type)
	}
}
