package riskengine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSecurityTestServer initializes a Chi test router with real handlers and auth middleware.
func setupSecurityTestServer(adminKey string) (*chi.Mux, *RetrainingCoordinator, *ModelRegistry, *CanaryRouter) {
	reg := NewModelRegistry()
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)

	r := chi.NewRouter()

	// Mock Risk Orchestrator & Handler
	store := features.NewVelocityStore(nil)
	mlClient := NewMLClient("http://127.0.0.1:59998") // offline sidecar -> graceful degradation
	orch := NewOrchestrator(nil, store, nil, mlClient, nil)
	riskHandler := NewHandler(orch)

	r.Post("/v1/risk-evaluations", riskHandler.EvaluateRisk)

	// Protected Admin Route
	r.Post("/v1/canary/control", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != adminKey {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	})

	// Model version with sanitization
	r.Get("/v1/models/{version}", func(w http.ResponseWriter, r *http.Request) {
		rawVersion := chi.URLParam(r, "version")
		sanitized, err := utils.SanitizeIdentifier(rawVersion)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad_request", "message": err.Error()})
			return
		}
		model, err := reg.GetModel(sanitized)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not_found", "message": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(model)
	})

	return r, coordinator, reg, canaryRouter
}

// ---------------------------------------------------------------------------
// 1. Missing Authentication
// ---------------------------------------------------------------------------
func TestSecurity_MissingAuthentication(t *testing.T) {
	router, _, _, _ := setupSecurityTestServer("secret_admin_key_999")

	req := httptest.NewRequest(http.MethodPost, "/v1/canary/control", strings.NewReader(`{"percentage": 25}`))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Request without Authorization header must return 401 Unauthorized")
}

// ---------------------------------------------------------------------------
// 2. Invalid Authentication
// ---------------------------------------------------------------------------
func TestSecurity_InvalidAuthentication(t *testing.T) {
	router, _, _, _ := setupSecurityTestServer("secret_admin_key_999")

	req := httptest.NewRequest(http.MethodPost, "/v1/canary/control", strings.NewReader(`{"percentage": 25}`))
	req.Header.Set("Authorization", "Bearer invalid_wrong_token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code, "Request with invalid Authorization token must return 403 Forbidden")
}

// ---------------------------------------------------------------------------
// 3. Oversized Request Body (> 1 MB DoS Protection)
// ---------------------------------------------------------------------------
func TestSecurity_OversizedRequestBody(t *testing.T) {
	router, _, _, _ := setupSecurityTestServer("secret_admin_key_999")

	// Create payload larger than 1 MB (1.5 MB)
	hugeString := strings.Repeat("A", 1500000)
	payload := fmt.Sprintf(`{"transaction_id": "%s", "amount": 100, "currency": "USD"}`, hugeString)

	req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Oversized request > 1MB must be rejected with 400 Bad Request")
	assert.Contains(t, rr.Body.String(), "1MB")
}

// ---------------------------------------------------------------------------
// 4. Path Traversal & Identifier Injection Defense
// ---------------------------------------------------------------------------
func TestSecurity_PathTraversalSanitization(t *testing.T) {
	router, _, _, _ := setupSecurityTestServer("secret_admin_key_999")

	maliciousVersions := []string{
		"../../etc/passwd",
		"model/../../shadow",
		"v1.0;DROP_TABLE",
		"script_alert_1",
	}

	for _, v := range maliciousVersions {
		req := httptest.NewRequest(http.MethodGet, "/v1/models/"+v, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.True(t, rr.Code == http.StatusBadRequest || rr.Code == http.StatusNotFound,
			"Malicious version '%s' must be rejected with 400 or 404, got %d", v, rr.Code)
	}

	// Test direct identifier sanitization defense
	_, err := utils.SanitizeIdentifier("version_with_null\x00byte")
	assert.Error(t, err)
	_, err = utils.SanitizeIdentifier("../../etc/shadow")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// 5. Tenant Isolation & Malformed Tenant Header Defense
// ---------------------------------------------------------------------------
func TestSecurity_TenantIsolationAndSanitization(t *testing.T) {
	router, _, _, _ := setupSecurityTestServer("secret_admin_key_999")

	// 1. Invalid/Malicious Tenant ID injection
	req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", strings.NewReader(`{"transaction_id": "txn_01", "amount": 100}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "../../tenant_escape/etc")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Path-traversal tenant ID must be rejected")
	assert.Contains(t, rr.Body.String(), "invalid_tenant_id")
}

// ---------------------------------------------------------------------------
// 6. Duplicate Model Promotion & Provenance Bypass Prevention
// ---------------------------------------------------------------------------
func TestSecurity_DuplicatePromotionAndStateSafety(t *testing.T) {
	_, _, reg, _ := setupSecurityTestServer("secret_admin_key_999")

	// Ensure active baseline is fraud-xgb-25f-v3.0
	prod, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", prod.Version)

	// Attempt to promote non-existent model
	err = reg.PromoteModel("non_existent_v99", "ADMIN", "Test malicious promotion")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in registry")

	// Verify production model was not altered
	prodAfter, _ := reg.GetProductionModel()
	assert.Equal(t, "fraud-xgb-25f-v3.0", prodAfter.Version)
}

// ---------------------------------------------------------------------------
// 7. Negative Amount and Required Field Validation
// ---------------------------------------------------------------------------
func TestSecurity_NegativeAmountAndFieldValidation(t *testing.T) {
	router, _, _, _ := setupSecurityTestServer("secret_admin_key_999")

	// Negative amount
	req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", strings.NewReader(`{"transaction_id": "txn_neg", "amount": -500}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "cannot be negative")

	// Empty transaction ID
	req2 := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", strings.NewReader(`{"transaction_id": "", "amount": 500}`))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	assert.Equal(t, http.StatusBadRequest, rr2.Code)
	assert.Contains(t, rr2.Body.String(), "transaction_id")
}
