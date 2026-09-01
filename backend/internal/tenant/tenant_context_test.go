package tenant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenant_MissingIdentity(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/rules", nil)
	id, err := ResolveTenant(req)
	require.NoError(t, err)
	assert.Equal(t, DefaultTenantID, id.TenantID)
	assert.Equal(t, "DEFAULT_FALLBACK", id.Source)
}

func TestTenant_DemoHeaderValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/rules", nil)
	req.Header.Set("X-Tenant-ID", "tenant_fintech_corp_100")
	id, err := ResolveTenant(req)
	require.NoError(t, err)
	assert.Equal(t, "tenant_fintech_corp_100", id.TenantID)
	assert.Equal(t, "DEMO_HEADER", id.Source)
}

func TestTenant_MalformedHeaderReturnsError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/rules", nil)
	// Malformed control characters and SQL injection attempts
	req.Header.Set("X-Tenant-ID", "../../escape;DROP TABLE--")
	_, err := ResolveTenant(req)
	assert.Error(t, err, "Malformed tenant header must return error")
	assert.Contains(t, err.Error(), "invalid tenant identifier")
}

func TestTenant_HeaderDoesNotOverrideAuthenticatedIdentity(t *testing.T) {
	// Authenticated principal binds tenant-alpha
	authIdentity := Identity{
		TenantID: "tenant-alpha-auth-principal",
		Source:   "AUTHENTICATED_PRINCIPAL",
	}
	ctx := WithTenantIdentity(context.Background(), authIdentity)

	req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", nil).WithContext(ctx)
	// Attacker tries to spoof another tenant via untrusted client header
	req.Header.Set("X-Tenant-ID", "tenant-victim-spoofed")

	resolved, err := ResolveTenant(req)
	require.NoError(t, err)
	// Must resolve to authenticated principal, ignoring the spoofed header
	assert.Equal(t, "tenant-alpha-auth-principal", resolved.TenantID)
	assert.Equal(t, "AUTHENTICATED_PRINCIPAL", resolved.Source)
}

func TestTenant_MiddlewareContextPropagation(t *testing.T) {
	var capturedID Identity
	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = ResolveTenantOrFallback(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-Tenant-ID", "tenant_bank_777")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "tenant_bank_777", capturedID.TenantID)
	assert.Equal(t, "DEMO_HEADER", capturedID.Source)
}
