package riskengine

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

// Role represents a distinct authorization privilege level.
type Role string

const (
	RoleAnonymous    Role = "ANONYMOUS"
	RoleReadOnly     Role = "READ_ONLY"
	RoleRiskOperator Role = "RISK_OPERATOR"
	RoleMLOperator   Role = "ML_OPERATOR"
	RoleAdmin        Role = "ADMIN"
)

// HasPermission checks if the role satisfies any of the required roles,
// with RoleAdmin having superuser privileges across all operations.
func (r Role) HasPermission(requiredRoles ...Role) bool {
	if r == RoleAdmin {
		return true // Admin satisfies all permission checks
	}
	for _, req := range requiredRoles {
		if r == req {
			return true
		}
	}
	return false
}

// Identity represents an authenticated caller identity.
type Identity struct {
	Subject  string            `json:"subject"`
	Role     Role              `json:"role"`
	TenantID string            `json:"tenant_id,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type authContextKey struct{}

var identityCtxKey = authContextKey{}

// WithIdentity injects an authenticated identity into the request context.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityCtxKey, id)
}

// GetIdentity extracts the authenticated identity from the context, if present.
func GetIdentity(ctx context.Context) (Identity, bool) {
	val := ctx.Value(identityCtxKey)
	if val == nil {
		return Identity{Role: RoleAnonymous}, false
	}
	id, ok := val.(Identity)
	return id, ok
}

// AuthManager manages API keys, role mappings, and authentication metrics.
type AuthManager struct {
	mu           sync.RWMutex
	credentials  map[string]Identity // APIKey -> Identity
	authFailures int64
	authSuccess  int64
}

// NewAuthManager initializes a thread-safe AuthManager.
func NewAuthManager() *AuthManager {
	return &AuthManager{
		credentials: make(map[string]Identity),
	}
}

// RegisterKey associates an API key with an authenticated identity.
func (am *AuthManager) RegisterKey(apiKey string, identity Identity) error {
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("api key cannot be empty")
	}
	am.mu.Lock()
	defer am.mu.Unlock()
	am.credentials[apiKey] = identity
	return nil
}

// AuthenticateCredential performs constant-time validation of the provided API key.
func (am *AuthManager) AuthenticateCredential(rawKey string) (Identity, bool) {
	if rawKey == "" {
		atomic.AddInt64(&am.authFailures, 1)
		return Identity{Role: RoleAnonymous}, false
	}

	am.mu.RLock()
	defer am.mu.RUnlock()

	rawKeyBytes := []byte(rawKey)
	var matchedIdentity Identity
	var found bool

	// Constant-time search across registered keys to prevent timing leaks
	for key, id := range am.credentials {
		keyBytes := []byte(key)
		if subtle.ConstantTimeCompare(keyBytes, rawKeyBytes) == 1 {
			matchedIdentity = id
			found = true
		}
	}

	if found {
		atomic.AddInt64(&am.authSuccess, 1)
		return matchedIdentity, true
	}

	atomic.AddInt64(&am.authFailures, 1)
	return Identity{Role: RoleAnonymous}, false
}

// AuthStats returns current authentication counters.
func (am *AuthManager) AuthStats() (successes int64, failures int64) {
	return atomic.LoadInt64(&am.authSuccess), atomic.LoadInt64(&am.authFailures)
}

// ExtractAPIKey extracts the API key from standard HTTP headers.
func ExtractAPIKey(r *http.Request) string {
	// 1. Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 2. X-API-Key header
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}

	// 3. X-Admin-API-Key header (backward compatibility)
	if key := r.Header.Get("X-Admin-API-Key"); key != "" {
		return key
	}

	return ""
}

// AuthenticateMiddleware extracts and validates credentials, populating the context with Identity.
func (am *AuthManager) AuthenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := ExtractAPIKey(r)
		if key != "" {
			if id, ok := am.AuthenticateCredential(key); ok {
				ctx := WithIdentity(r.Context(), id)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		// If no valid credentials found, proceed with RoleAnonymous
		ctx := WithIdentity(r.Context(), Identity{Role: RoleAnonymous})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole enforces that the authenticated identity possesses at least one of the specified roles.
func (am *AuthManager) RequireRole(requiredRoles ...Role) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id, ok := GetIdentity(r.Context())
			// If not already authenticated in context, attempt to authenticate from request headers
			if !ok || id.Role == RoleAnonymous {
				key := ExtractAPIKey(r)
				if key == "" {
					writeJSONAuthError(w, http.StatusUnauthorized, "unauthorized", "Missing authentication credentials")
					return
				}
				var authenticated bool
				id, authenticated = am.AuthenticateCredential(key)
				if !authenticated {
					writeJSONAuthError(w, http.StatusUnauthorized, "unauthorized", "Invalid authentication credentials")
					return
				}
			}

			if !id.Role.HasPermission(requiredRoles...) {
				writeJSONAuthError(w, http.StatusForbidden, "forbidden", fmt.Sprintf("Identity '%s' with role '%s' lacks required permissions", id.Subject, id.Role))
				return
			}

			next(w, r)
		}
	}
}

// Helper to write RFC 7807 compatible auth errors
func writeJSONAuthError(w http.ResponseWriter, statusCode int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"type":    fmt.Sprintf("https://errors.ropus.io/%s", errCode),
		"title":   strings.Title(strings.ReplaceAll(errCode, "_", " ")),
		"status":  statusCode,
		"detail":  message,
		"code":    errCode,
	})
}
