package riskengine

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// RequireAdminAuth enforces constant-time administrative API key authentication.
func RequireAdminAuth(adminKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Admin-API-Key")
		if key == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				key = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(adminKey)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "unauthorized",
				"message": "Invalid or missing administrative API key",
			})
			return
		}
		next(w, r)
	}
}

func TestOperationalEndpoints_CanaryAndSystemStatus(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:                  true,
		Percentage:               50,
		CandidateModelVersion:    "fraud-xgb-25f-candidate-v1",
		CandidateFeatureContract: MLFeatureContractV25,
	}
	router := NewCanaryRouter(cfg, nil)

	// Record some sample activity
	router.RecordCandidateRequest()
	router.RecordCandidateSuccess(2.45, "ALLOW_RECOMMENDATION")
	router.RecordLegacyRequest()

	status := router.GetStatus()

	if status["enabled"] != true {
		t.Errorf("Expected enabled=true, got %v", status["enabled"])
	}
	if status["target_percentage"] != 50 {
		t.Errorf("Expected target_percentage=50, got %v", status["target_percentage"])
	}
	if status["production_model"] != "fraud-xgb-25f-v3.0" {
		t.Errorf("Expected production_model fraud-xgb-25f-v3.0, got %v", status["production_model"])
	}

	metrics, ok := status["metrics"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected metrics map in status response")
	}
	if metrics["total_requests"] != int64(2) {
		t.Errorf("Expected total_requests=2, got %v", metrics["total_requests"])
	}
	if metrics["candidate_requests_total"] != int64(1) {
		t.Errorf("Expected candidate_requests_total=1, got %v", metrics["candidate_requests_total"])
	}
}

func TestOperatorSafety_CanaryControlValidation(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:    false,
		Percentage: 0,
	}
	router := NewCanaryRouter(cfg, nil)

	// 1. Valid update with reason
	err := router.UpdateConfig(true, 25, "admin_user_ops", "Rollout to Tier 4 (25%)")
	if err != nil {
		t.Fatalf("Expected valid update to succeed, got: %v", err)
	}

	status := router.GetStatus()
	if status["target_percentage"] != 25 {
		t.Errorf("Expected target_percentage=25, got %v", status["target_percentage"])
	}

	// 2. Reject negative percentage
	err = router.UpdateConfig(true, -10, "admin", "test")
	if err == nil {
		t.Errorf("Expected error for negative percentage, got nil")
	}

	// 3. Reject percentage > 100
	err = router.UpdateConfig(true, 150, "admin", "test")
	if err == nil {
		t.Errorf("Expected error for percentage > 100, got nil")
	}
}

func TestRequireAdminAuth_Middleware(t *testing.T) {
	adminKey := "super_secret_test_key_xyz"
	mockNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"authorized"}`))
	})

	// 1. Missing header -> 401
	reqMissing, _ := http.NewRequest("POST", "/v1/canary/control", nil)
	rrMissing := httptest.NewRecorder()
	RequireAdminAuth(adminKey, mockNext)(rrMissing, reqMissing)
	if rrMissing.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing key, got %d", rrMissing.Code)
	}

	// 2. Invalid header -> 401
	reqInvalid, _ := http.NewRequest("POST", "/v1/canary/control", nil)
	reqInvalid.Header.Set("X-Admin-API-Key", "wrong_key")
	rrInvalid := httptest.NewRecorder()
	RequireAdminAuth(adminKey, mockNext)(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid key, got %d", rrInvalid.Code)
	}

	// 3. Valid X-Admin-API-Key -> 200
	reqValidHeader, _ := http.NewRequest("POST", "/v1/canary/control", nil)
	reqValidHeader.Header.Set("X-Admin-API-Key", adminKey)
	rrValidHeader := httptest.NewRecorder()
	RequireAdminAuth(adminKey, mockNext)(rrValidHeader, reqValidHeader)
	if rrValidHeader.Code != http.StatusOK {
		t.Errorf("Expected 200 for valid X-Admin-API-Key, got %d", rrValidHeader.Code)
	}

	// 4. Valid Bearer Token Authorization -> 200
	reqValidBearer, _ := http.NewRequest("POST", "/v1/canary/control", nil)
	reqValidBearer.Header.Set("Authorization", "Bearer "+adminKey)
	rrValidBearer := httptest.NewRecorder()
	RequireAdminAuth(adminKey, mockNext)(rrValidBearer, reqValidBearer)
	if rrValidBearer.Code != http.StatusOK {
		t.Errorf("Expected 200 for valid Bearer token, got %d", rrValidBearer.Code)
	}
}

func TestOperationalMetrics_RateCalculations(t *testing.T) {
	metrics := &CanaryMetrics{
		LatencyReservoir: NewLatencyReservoir(100),
	}

	// 10 legacy requests, 10 candidate requests (8 success, 2 fallbacks)
	for i := 0; i < 10; i++ {
		metrics.LegacyRequestsTotal.Add(1)
	}
	for i := 0; i < 8; i++ {
		metrics.CandidateRequestsTotal.Add(1)
		metrics.CandidateSuccessTotal.Add(1)
		metrics.TotalCandidateLatencyUs.Add(2500) // 2.5ms
		metrics.LatencyReservoir.Add(2.5)
	}
	for i := 0; i < 2; i++ {
		metrics.CandidateRequestsTotal.Add(1)
		metrics.CandidateErrorTotal.Add(1)
		metrics.CandidateFallbackTotal.Add(1)
		metrics.LatencyReservoir.Add(10.0)
	}

	snap := metrics.Snapshot()

	if snap["total_requests"] != int64(20) {
		t.Errorf("Expected total_requests=20, got %v", snap["total_requests"])
	}
	if snap["actual_canary_percentage"] != 50.0 {
		t.Errorf("Expected actual_canary_percentage=50.0, got %v", snap["actual_canary_percentage"])
	}
	if snap["candidate_error_rate"] != 0.20 {
		t.Errorf("Expected candidate_error_rate=0.20, got %v", snap["candidate_error_rate"])
	}
	if snap["candidate_fallback_rate"] != 0.20 {
		t.Errorf("Expected candidate_fallback_rate=0.20, got %v", snap["candidate_fallback_rate"])
	}
}
