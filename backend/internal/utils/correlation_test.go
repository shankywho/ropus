package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrelationID_Generation(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.True(t, len(id1) >= 10)
}

func TestCorrelationID_Sanitization(t *testing.T) {
	// Valid preserved
	assert.Equal(t, "custom-request-123_v1", SanitizeRequestID("custom-request-123_v1"))

	// Path traversal stripped/regenerated
	bad := SanitizeRequestID("../secret")
	assert.NotContains(t, bad, "..")
	assert.NotContains(t, bad, "/")

	// Empty generated
	empty := SanitizeRequestID("")
	assert.NotEmpty(t, empty)
}

func TestCorrelationID_Context(t *testing.T) {
	ctx := context.Background()
	reqID := "req_test_12345"

	ctxWithID := WithRequestID(ctx, reqID)
	assert.Equal(t, reqID, GetRequestID(ctxWithID))

	// Nil context fallback
	assert.NotEmpty(t, GetRequestID(nil))
}

func TestCorrelationID_Middleware(t *testing.T) {
	handler := CorrelationIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := GetRequestID(r.Context())
		assert.NotEmpty(t, reqID)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(reqID))
	}))

	// 1. Without incoming header -> generates new and sets header
	req := httptest.NewRequest("GET", "/v1/risk-evaluations", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	respID := rec.Header().Get(RequestIDHeader)
	assert.NotEmpty(t, respID)
	assert.Equal(t, respID, rec.Body.String())

	// 2. With valid incoming header -> preserves it
	req2 := httptest.NewRequest("GET", "/v1/risk-evaluations", nil)
	req2.Header.Set(RequestIDHeader, "client-trace-id-abc")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.Equal(t, "client-trace-id-abc", rec2.Header().Get(RequestIDHeader))
	assert.Equal(t, "client-trace-id-abc", rec2.Body.String())
}
