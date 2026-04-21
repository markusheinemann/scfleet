package api_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
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

// --- PollJob ---

func TestPollJob_NoContent_ReturnsNilJob(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	job, err := client.PollJob(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if job != nil {
		t.Fatalf("expected nil job on 204, got %v", job)
	}
}

func TestPollJob_Success_DecodesJob(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"job_id":"job-1","url":"https://example.com","template":{},"timeout_s":30}`)) //nolint:errcheck
	})

	job, err := client.PollJob(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if job == nil {
		t.Fatal("expected job, got nil")
	}
	if job.JobID != "job-1" {
		t.Errorf("expected job_id 'job-1', got %q", job.JobID)
	}
	if job.TimeoutS != 30 {
		t.Errorf("expected timeout_s 30, got %d", job.TimeoutS)
	}
}

func TestPollJob_NonOKStatus_ReturnsError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.PollJob(context.Background())
	if err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}

func TestPollJob_InvalidJSON_ReturnsError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not-json")) //nolint:errcheck
	})

	_, err := client.PollJob(context.Background())
	if err == nil {
		t.Fatal("expected error on malformed JSON, got nil")
	}
}

func TestPollJob_PostsToCorrectPath(t *testing.T) {
	var gotPath string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	client.PollJob(context.Background()) //nolint:errcheck

	if gotPath != "/api/v1/jobs/claim" {
		t.Errorf("expected /api/v1/jobs/claim, got %q", gotPath)
	}
}

func TestPollJob_BuildURLError(t *testing.T) {
	client := api.New("%", "token", nil, discardLogger)
	_, err := client.PollJob(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid base URL, got nil")
	}
}

func TestPollJob_NetworkError_ReturnsError(t *testing.T) {
	client := api.New("http://127.0.0.1:1", "token", nil, discardLogger)
	_, err := client.PollJob(context.Background())
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}

func TestPollJob_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	})

	client.PollJob(context.Background()) //nolint:errcheck

	if gotAuth != "Bearer test-token" {
		t.Errorf("expected 'Bearer test-token', got %q", gotAuth)
	}
}

// --- CompleteJob ---

func TestCompleteJob_Success(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.CompleteJob(context.Background(), "job-1", api.CompleteJobRequest{
		Result: map[string]any{"title": "Widget"},
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCompleteJob_PostsToCorrectPath(t *testing.T) {
	var gotPath string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	client.CompleteJob(context.Background(), "job-abc", api.CompleteJobRequest{}) //nolint:errcheck

	if gotPath != "/api/v1/jobs/job-abc/complete" {
		t.Errorf("expected /api/v1/jobs/job-abc/complete, got %q", gotPath)
	}
}

func TestCompleteJob_SendsJSONBody(t *testing.T) {
	var gotBody []byte
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	})

	client.CompleteJob(context.Background(), "job-1", api.CompleteJobRequest{ //nolint:errcheck
		Result: map[string]any{"price": 9.99},
	})

	var decoded map[string]any
	if err := json.Unmarshal(gotBody, &decoded); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	result, ok := decoded["result"].(map[string]any)
	if !ok {
		t.Fatalf("expected result map, got %T", decoded["result"])
	}
	if result["price"] != 9.99 {
		t.Errorf("expected price 9.99, got %v", result["price"])
	}
}

func TestCompleteJob_SetsAuthorizationHeader(t *testing.T) {
	var gotAuth string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	})

	client.CompleteJob(context.Background(), "job-1", api.CompleteJobRequest{}) //nolint:errcheck

	if gotAuth != "Bearer test-token" {
		t.Errorf("expected 'Bearer test-token', got %q", gotAuth)
	}
}

func TestCompleteJob_BuildURLError(t *testing.T) {
	client := api.New("%", "token", nil, discardLogger)
	if err := client.CompleteJob(context.Background(), "job-1", api.CompleteJobRequest{}); err == nil {
		t.Fatal("expected error for invalid base URL, got nil")
	}
}

func TestCompleteJob_UnmarshalableBody_ReturnsError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	// channels cannot be JSON-marshaled
	err := client.CompleteJob(context.Background(), "job-1", api.CompleteJobRequest{
		Result: map[string]any{"bad": make(chan int)},
	})
	if err == nil {
		t.Fatal("expected error when body contains unmarshalable value")
	}
}

func TestCompleteJob_NetworkError_ReturnsError(t *testing.T) {
	client := api.New("http://127.0.0.1:1", "token", nil, discardLogger)
	if err := client.CompleteJob(context.Background(), "job-1", api.CompleteJobRequest{}); err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}

func TestCompleteJob_ErrorOnNonSuccess(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if err := client.CompleteJob(context.Background(), "job-1", api.CompleteJobRequest{}); err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}

// --- FailJob ---

func TestFailJob_Success(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.FailJob(context.Background(), "job-1", api.FailJobRequest{
		ErrorType: "page_timeout", ErrorMessage: "timed out",
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFailJob_PostsToCorrectPath(t *testing.T) {
	var gotPath string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	client.FailJob(context.Background(), "job-xyz", api.FailJobRequest{}) //nolint:errcheck

	if gotPath != "/api/v1/jobs/job-xyz/fail" {
		t.Errorf("expected /api/v1/jobs/job-xyz/fail, got %q", gotPath)
	}
}

func TestFailJob_SendsErrorFields(t *testing.T) {
	var gotBody []byte
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	})

	client.FailJob(context.Background(), "job-1", api.FailJobRequest{ //nolint:errcheck
		ErrorType: "navigation_error", ErrorMessage: "DNS failure",
	})

	var decoded map[string]any
	json.Unmarshal(gotBody, &decoded) //nolint:errcheck
	if decoded["error_type"] != "navigation_error" {
		t.Errorf("expected error_type 'navigation_error', got %v", decoded["error_type"])
	}
	if decoded["error_message"] != "DNS failure" {
		t.Errorf("expected error_message 'DNS failure', got %v", decoded["error_message"])
	}
}

func TestFailJob_ErrorOnNonSuccess(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if err := client.FailJob(context.Background(), "job-1", api.FailJobRequest{}); err == nil {
		t.Fatal("expected error on 500, got nil")
	}
}

// --- UploadArtifacts ---

func TestUploadArtifacts_WithScreenshot_EncodesBase64(t *testing.T) {
	var gotBody []byte
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	})

	screenshot := []byte{0x89, 0x50, 0x4E, 0x47}
	if err := client.UploadArtifacts(context.Background(), "job-1", screenshot, "<html/>"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var decoded map[string]any
	json.Unmarshal(gotBody, &decoded) //nolint:errcheck

	want := base64.StdEncoding.EncodeToString(screenshot)
	if decoded["screenshot"] != want {
		t.Errorf("expected base64 screenshot %q, got %q", want, decoded["screenshot"])
	}
	if decoded["html"] != "<html/>" {
		t.Errorf("expected html '<html/>', got %v", decoded["html"])
	}
}

func TestUploadArtifacts_NoScreenshot_OmitsField(t *testing.T) {
	var gotBody []byte
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.UploadArtifacts(context.Background(), "job-1", nil, "<html/>"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var decoded map[string]any
	json.Unmarshal(gotBody, &decoded) //nolint:errcheck
	if _, ok := decoded["screenshot"]; ok {
		t.Error("expected screenshot field absent when no screenshot provided")
	}
}

func TestUploadArtifacts_PostsToCorrectPath(t *testing.T) {
	var gotPath string
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	client.UploadArtifacts(context.Background(), "job-abc", nil, "") //nolint:errcheck

	if gotPath != "/api/v1/jobs/job-abc/artifacts" {
		t.Errorf("expected /api/v1/jobs/job-abc/artifacts, got %q", gotPath)
	}
}

func TestUploadArtifacts_ErrorOnNonSuccess(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if err := client.UploadArtifacts(context.Background(), "job-1", nil, ""); err == nil {
		t.Fatal("expected error on 500, got nil")
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
