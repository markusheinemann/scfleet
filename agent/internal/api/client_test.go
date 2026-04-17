package api_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/markusheinemann/scfleet/agent/internal/api"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func newTestClient(t *testing.T, handler http.HandlerFunc) (*api.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return api.New(srv.URL, "test-token", srv.Client(), discardLogger), srv
}

func TestRegister_Success(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.Register(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRegister_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	})

	_ = client.Register(context.Background())

	if gotAuth != "Bearer test-token" {
		t.Errorf("expected 'Bearer test-token', got %q", gotAuth)
	}
}

func TestRegister_PostsToCorrectPath(t *testing.T) {
	var gotPath string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	_ = client.Register(context.Background())

	if gotPath != "/api/v1/register" {
		t.Errorf("expected /api/v1/register, got %q", gotPath)
	}
}

func TestRegister_ErrorOnNonSuccess(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	if err := client.Register(context.Background()); err == nil {
		t.Fatal("expected error on 401, got nil")
	}
}

func TestHeartbeat_Success(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.Heartbeat(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestHeartbeat_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	})

	_ = client.Heartbeat(context.Background())

	if gotAuth != "Bearer test-token" {
		t.Errorf("expected 'Bearer test-token', got %q", gotAuth)
	}
}

func TestHeartbeat_PostsToCorrectPath(t *testing.T) {
	var gotPath string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	_ = client.Heartbeat(context.Background())

	if gotPath != "/api/v1/heartbeat" {
		t.Errorf("expected /api/v1/heartbeat, got %q", gotPath)
	}
}

func TestHeartbeat_ErrorOnServerError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if err := client.Heartbeat(context.Background()); err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}

func TestClient_ErrorOnUnreachableURL(t *testing.T) {
	client := api.New("http://127.0.0.1:1", "token", nil, discardLogger)

	if err := client.Register(context.Background()); err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}

func TestClient_HandlesTrailingSlashInBaseURL(t *testing.T) {
	var gotPath string
	client, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	// Build a client with a trailing slash — should not produce double slashes.
	clientWithSlash := api.New(srv.URL+"/", "test-token", srv.Client(), discardLogger)
	_ = clientWithSlash.Register(context.Background())
	_ = client // suppress unused warning

	if gotPath != "/api/v1/register" {
		t.Errorf("expected /api/v1/register, got %q (possible double slash)", gotPath)
	}
}

func TestClient_BuildURLError(t *testing.T) {
	client := api.New("%", "token", nil, discardLogger)
	if err := client.Register(context.Background()); err == nil {
		t.Fatal("expected error for invalid base URL, got nil")
	}
}

func TestClient_DefaultHttpClient(t *testing.T) {
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := api.New(srv.URL, "tok", nil, discardLogger)
	_ = client.Register(context.Background())

	if !called {
		t.Fatal("expected server to be called with default http client")
	}
}
