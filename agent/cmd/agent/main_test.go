package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRun_MissingURL(t *testing.T) {
	var buf bytes.Buffer
	code := run(context.Background(), []string{"--token", "abc"}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRun_MissingToken(t *testing.T) {
	var buf bytes.Buffer
	code := run(context.Background(), []string{"--url", "http://localhost"}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRun_MissingBothFlags(t *testing.T) {
	var buf bytes.Buffer
	code := run(context.Background(), []string{}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRun_InvalidFlag(t *testing.T) {
	var buf bytes.Buffer
	code := run(context.Background(), []string{"--unknown-flag"}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1 for invalid flag, got %d", code)
	}
}

func TestRun_RegisterFailureReturnsOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	code := run(context.Background(), []string{"--url", srv.URL, "--token", "bad-token"}, io.Discard)
	if code != 1 {
		t.Errorf("expected exit code 1 on register failure, got %d", code)
	}
}

func TestRun_DebugFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	code := run(context.Background(), []string{"--url", srv.URL, "--token", "bad", "--debug"}, io.Discard)
	if code != 1 {
		t.Errorf("expected exit code 1 on register failure with debug, got %d", code)
	}
}

func TestRun_SuccessfulShutdown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	code := run(ctx, []string{"--url", srv.URL, "--token", "tok", "--interval", "10s"}, io.Discard)
	if code != 0 {
		t.Errorf("expected exit code 0 on graceful shutdown, got %d", code)
	}
}
