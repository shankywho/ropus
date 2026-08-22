package riskengine

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIError represents an RFC 7807 compliant standardized problem details payload.
type APIError struct {
	Type          string                 `json:"type"`
	Title         string                 `json:"title"`
	Status        int                    `json:"status"`
	Detail        string                 `json:"detail"`
	Instance      string                 `json:"instance,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	InvalidParams []map[string]string    `json:"invalid_params,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// WriteAPIError serializes an RFC 7807 error response.
func WriteAPIError(w http.ResponseWriter, r *http.Request, status int, errorType, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	correlationID := w.Header().Get("X-Correlation-ID")
	if correlationID == "" && r != nil {
		correlationID = r.Header.Get("X-Correlation-ID")
	}

	instance := ""
	if r != nil {
		instance = r.URL.Path
	}

	errPayload := APIError{
		Type:          errorType,
		Title:         title,
		Status:        status,
		Detail:        detail,
		Instance:      instance,
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC(),
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errPayload)
}
