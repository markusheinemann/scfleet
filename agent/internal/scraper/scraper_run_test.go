package scraper

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func newTestScraper(run func(context.Context, ...chromedp.Action) error) *Scraper {
	return &Scraper{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		run:    run,
	}
}

func TestFetch_HappyPath_ReturnsResult(t *testing.T) {
	s := newTestScraper(func(ctx context.Context, actions ...chromedp.Action) error {
		return nil
	})

	result, err := s.Fetch(context.Background(), "https://example.com", "", 30*time.Second)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestFetch_PageTimeout_ReturnsPartialResult(t *testing.T) {
	s := newTestScraper(func(ctx context.Context, actions ...chromedp.Action) error {
		return context.DeadlineExceeded
	})

	result, err := s.Fetch(context.Background(), "https://example.com", "", 30*time.Second)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	se, ok := err.(*ScrapeError)
	if !ok {
		t.Fatalf("expected *ScrapeError, got %T: %v", err, err)
	}
	if se.Type != "page_timeout" {
		t.Errorf("expected page_timeout, got %q", se.Type)
	}
	if result == nil {
		t.Error("expected partial result alongside timeout error")
	}
}

func TestFetch_NavigationError_ReturnsError(t *testing.T) {
	s := newTestScraper(func(ctx context.Context, actions ...chromedp.Action) error {
		return context.Canceled
	})

	result, err := s.Fetch(context.Background(), "https://example.com", "", 30*time.Second)

	if result != nil {
		t.Errorf("expected nil result on navigation error, got %+v", result)
	}
	se, ok := err.(*ScrapeError)
	if !ok {
		t.Fatalf("expected *ScrapeError, got %T: %v", err, err)
	}
	if se.Type != "navigation_error" {
		t.Errorf("expected navigation_error, got %q", se.Type)
	}
}
