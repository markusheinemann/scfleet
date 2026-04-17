package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRun_MissingURL(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"--token", "abc"}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRun_MissingToken(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"--url", "http://localhost"}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRun_MissingBothFlags(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRun_InvalidFlag(t *testing.T) {
	var buf bytes.Buffer
	code := run([]string{"--unknown-flag"}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1 for invalid flag, got %d", code)
	}
}

func TestRun_RegisterFailureReturnsOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	code := run([]string{"--url", srv.URL, "--token", "bad-token"}, &buf)
	if code != 1 {
		t.Errorf("expected exit code 1 on register failure, got %d", code)
	}
}
