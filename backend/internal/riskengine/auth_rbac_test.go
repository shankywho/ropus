package riskengine

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupTestAuthManager() *AuthManager {
	am := NewAuthManager()
	_ = am.RegisterKey("admin_secret_key_123", Identity{Subject: "admin_user", Role: RoleAdmin, TenantID: "master"})
	_ = am.RegisterKey("ml_operator_key_456", Identity{Subject: "ml_engineer", Role: RoleMLOperator, TenantID: "ml_team"})
	_ = am.RegisterKey("risk_operator_key_789", Identity{Subject: "risk_analyst", Role: RoleRiskOperator, TenantID: "risk_team"})
	_ = am.RegisterKey("readonly_key_000", Identity{Subject: "auditor_bot", Role: RoleReadOnly, TenantID: "audit_team"})
	return am
}

func TestRBAC_AuthenticationAndRolePermutations(t *testing.T) {
	am := setupTestAuthManager()

	// Handler requiring RoleMLOperator
	mlHandler := am.RequireRole(RoleMLOperator)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ml_success"}`))
	})

	// Handler requiring RoleAdmin exclusively
	adminHandler := am.RequireRole(RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"admin_success"}`))
	})

	// Handler requiring RoleReadOnly or higher
	readHandler := am.RequireRole(RoleReadOnly, RoleRiskOperator, RoleMLOperator)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"read_success"}`))
	})

	// 1. Missing credentials -> 401
	t.Run("MissingCredentials_401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", nil)
		rr := httptest.NewRecorder()
		mlHandler(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Missing authentication")
	})

	// 2. Invalid credentials -> 401
	t.Run("InvalidCredentials_401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", nil)
		req.Header.Set("Authorization", "Bearer bad_token_999")
		rr := httptest.NewRecorder()
		mlHandler(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid authentication")
	})

	// 3. Exact RoleMLOperator accessing ML endpoint -> 200
	t.Run("MLOperator_Accessing_MLEndpoint_200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", nil)
		req.Header.Set("Authorization", "Bearer ml_operator_key_456")
		rr := httptest.NewRecorder()
		mlHandler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "ml_success")
	})

	// 4. RoleAdmin accessing ML endpoint (Superuser Privilege) -> 200
	t.Run("Admin_Accessing_MLEndpoint_200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", nil)
		req.Header.Set("X-API-Key", "admin_secret_key_123")
		rr := httptest.NewRecorder()
		mlHandler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "ml_success")
	})

	// 5. RoleReadOnly attempting ML mutation (Insufficient Role) -> 403
	t.Run("ReadOnly_Attempting_MLMutation_403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", nil)
		req.Header.Set("Authorization", "Bearer readonly_key_000")
		rr := httptest.NewRecorder()
		mlHandler(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "lacks required permissions")
	})

	// 6. RoleRiskOperator attempting ML mutation -> 403
	t.Run("RiskOperator_Attempting_MLMutation_403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", nil)
		req.Header.Set("Authorization", "Bearer risk_operator_key_789")
		rr := httptest.NewRecorder()
		mlHandler(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	// 7. RoleMLOperator attempting Admin-only mutation (Recovery) -> 403
	t.Run("MLOperator_Attempting_AdminEndpoint_403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/operations/recovery/trigger", nil)
		req.Header.Set("Authorization", "Bearer ml_operator_key_456")
		rr := httptest.NewRecorder()
		adminHandler(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	// 8. RoleReadOnly accessing Read endpoint -> 200
	t.Run("ReadOnly_Accessing_ReadEndpoint_200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/operations/health", nil)
		req.Header.Set("X-Admin-API-Key", "readonly_key_000")
		rr := httptest.NewRecorder()
		readHandler(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestRBAC_ConcurrentAuthVerification(t *testing.T) {
	am := setupTestAuthManager()
	var wg sync.WaitGroup

	numWorkers := 32
	reqsPerWorker := 500

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < reqsPerWorker; j++ {
				var key string
				var shouldSucceed bool

				switch (workerID + j) % 4 {
				case 0:
					key = "admin_secret_key_123"
					shouldSucceed = true
				case 1:
					key = "ml_operator_key_456"
					shouldSucceed = true
				case 2:
					key = "invalid_secret_attack"
					shouldSucceed = false
				case 3:
					key = ""
					shouldSucceed = false
				}

				id, ok := am.AuthenticateCredential(key)
				if shouldSucceed {
					assert.True(t, ok)
					assert.NotEmpty(t, id.Subject)
				} else {
					assert.False(t, ok)
				}
			}
		}(i)
	}

	wg.Wait()

	successes, failures := am.AuthStats()
	assert.Equal(t, int64(numWorkers*reqsPerWorker/2), successes)
	assert.Equal(t, int64(numWorkers*reqsPerWorker/2), failures)
}

func TestRBAC_EmptyKeyRegistration(t *testing.T) {
	am := NewAuthManager()
	err := am.RegisterKey("", Identity{Role: RoleAdmin})
	assert.Error(t, err)
}
