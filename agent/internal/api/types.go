package api

import "encoding/json"

// ClaimedJob is returned by POST /api/v1/jobs/claim (HTTP 200).
type ClaimedJob struct {
	JobID    string          `json:"job_id"`
	URL      string          `json:"url"`
	Template json.RawMessage `json:"template"`
	TimeoutS int             `json:"timeout_s"`
}

// CompleteJobRequest is the body for POST /api/v1/jobs/{id}/complete.
type CompleteJobRequest struct {
	Result      map[string]any    `json:"result"`
	FieldErrors map[string]string `json:"field_errors,omitempty"`
}

// FailJobRequest is the body for POST /api/v1/jobs/{id}/fail.
type FailJobRequest struct {
	ErrorType    string `json:"error_type"`
	ErrorMessage string `json:"error_message"`
}

// UploadArtifactsRequest is the body for POST /api/v1/jobs/{id}/artifacts.
type UploadArtifactsRequest struct {
	Screenshot string `json:"screenshot,omitempty"` // base64-encoded PNG
	HTML       string `json:"html,omitempty"`
}
