package tenant

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type contextKey string

const (
	// ContextKey is the context key for trusted tenant identity.
	ContextKey contextKey = "trusted_tenant_identity"

	// DefaultTenantID is the fallback tenant identifier.
	DefaultTenantID string = "00000000-0000-0000-0000-000000000001"
)

var validTenantRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// Identity represents the resolved tenant identity and its provenance.
type Identity struct {
	TenantID string `json:"tenant_id"`
	Source   string `json:"source"` // "AUTHENTICATED_PRINCIPAL", "DEMO_HEADER", "DEFAULT_FALLBACK"
}

// WithTenantIdentity binds a trusted tenant identity into the request context.
func WithTenantIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ContextKey, id)
}

// ResolveTenant extracts and validates the trusted tenant identity from request context or header fallback.
func ResolveTenant(r *http.Request) (Identity, error) {
	if r == nil {
		return Identity{TenantID: DefaultTenantID, Source: "DEFAULT_FALLBACK"}, nil
	}

	// 1. Authenticated principal context takes highest precedence (cannot be spoofed by client header)
	if val := r.Context().Value(ContextKey); val != nil {
		if id, ok := val.(Identity); ok && id.TenantID != "" {
			return id, nil
		}
	}

	// 2. Demo / Standalone fallback via X-Tenant-ID header with strict character and length validation
	raw := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if raw == "" {
		return Identity{
			TenantID: DefaultTenantID,
			Source:   "DEFAULT_FALLBACK",
		}, nil
	}

	if !validTenantRegex.MatchString(raw) {
		return Identity{}, fmt.Errorf("invalid tenant identifier '%s': must match ^[a-zA-Z0-9_-]{1,64}$", raw)
	}

	return Identity{
		TenantID: raw,
		Source:   "DEMO_HEADER",
	}, nil
}

// ResolveTenantOrFallback extracts tenant identity or defaults to DefaultTenantID without returning an error.
func ResolveTenantOrFallback(r *http.Request) Identity {
	id, err := ResolveTenant(r)
	if err != nil || id.TenantID == "" {
		return Identity{
			TenantID: DefaultTenantID,
			Source:   "DEFAULT_FALLBACK",
		}
	}
	return id
}

// Middleware injects resolved tenant identity into every request context.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := ResolveTenant(r)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"invalid_tenant_id","message":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}
		ctx := WithTenantIdentity(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
