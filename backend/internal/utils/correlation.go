package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const (
	RequestIDHeader             = "X-Request-ID"
	requestIDContextKey contextKey = "request_id"
)

// GenerateRequestID creates a cryptographically unique request identifier.
func GenerateRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("req_%d_%s", time.Now().Unix(), hex.EncodeToString(b))
}

// SanitizeRequestID validates and sanitizes a client-provided request ID.
func SanitizeRequestID(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return GenerateRequestID()
	}
	if len(trimmed) > 128 {
		trimmed = trimmed[:128]
	}
	clean, err := SanitizeIdentifier(trimmed)
	if err != nil {
		return GenerateRequestID()
	}
	return clean
}

// WithRequestID injects a request ID into the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

// GetRequestID extracts the request ID from the context, or returns a new one if not present.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return GenerateRequestID()
	}
	if val, ok := ctx.Value(requestIDContextKey).(string); ok && val != "" {
		return val
	}
	return GenerateRequestID()
}

// CorrelationIDMiddleware attaches a valid X-Request-ID to context and response headers.
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawID := r.Header.Get(RequestIDHeader)
		reqID := SanitizeRequestID(rawID)

		// Set header in response
		w.Header().Set(RequestIDHeader, reqID)

		// Attach to context
		ctx := WithRequestID(r.Context(), reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
