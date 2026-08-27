package riskengine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shankywho/ropus/backend/internal/riskengine"
	"github.com/shankywho/ropus/backend/internal/rules"
)

// TestResilience_MLServiceTimeout tests that when the ML sidecar takes too long,
// the orchestrator falls back gracefully to deterministic rules and threat signals without failing the request.
func TestResilience_MLServiceTimeout(t *testing.T) {
	// Slow ML sidecar mock that sleeps for 2 seconds
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"fraud_probability": 0.50,
			"model_version":     "mock-slow-v1",
		})
	}))
	defer slowServer.Close()

	rulesService := rules.NewService(nil)
	mlClient := riskengine.NewMLClient(slowServer.URL)
	orchestrator := riskengine.NewOrchestrator(nil, nil, rulesService, mlClient, nil)

	handler := riskengine.NewHandler(orchestrator)
	r := chi.NewRouter()
	r.Post("/v1/risk-evaluations", handler.EvaluateRisk)
	srv := httptest.NewServer(r)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"transaction_id":     "txn_timeout_test_01",
		"amount":             25000,
		"currency":           "INR",
		"ip_address":         "106.51.0.1",
		"device_fingerprint": "dev_normal_phone",
		"account_id":         "acc_timeout_01",
		"payment_method": map[string]string{
			"type":  "CARD",
			"token": "tok_card_01",
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP 200 during ML timeout fallback, got %d", resp.StatusCode)
	}

	var evalResp riskengine.RiskEvaluationResponse
	_ = json.NewDecoder(resp.Body).Decode(&evalResp)

	if !evalResp.IsDegraded {
		t.Errorf("Expected is_degraded = true when ML service times out")
	}

	hasDegradedCode := false
	for _, code := range evalResp.ReasonCodes {
		if code == "ML_SERVICE_DEGRADED" {
			hasDegradedCode = true
			break
		}
	}
	if !hasDegradedCode {
		t.Errorf("Expected reason_codes to contain ML_SERVICE_DEGRADED, got: %v", evalResp.ReasonCodes)
	}
}

// TestResilience_MLServiceUnavailable tests behavior when the ML service endpoint is completely unreachable (HTTP 503 / connection refused).
func TestResilience_MLServiceUnavailable(t *testing.T) {
	rulesService := rules.NewService(nil)
	deadMLClient := riskengine.NewMLClient("http://127.0.0.1:59998") // Non-existent port
	orchestrator := riskengine.NewOrchestrator(nil, nil, rulesService, deadMLClient, nil)

	handler := riskengine.NewHandler(orchestrator)
	r := chi.NewRouter()
	r.Post("/v1/risk-evaluations", handler.EvaluateRisk)
	srv := httptest.NewServer(r)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"transaction_id":     "txn_dead_ml_01",
		"amount":             15000,
		"currency":           "INR",
		"ip_address":         "106.51.12.34",
		"device_fingerprint": "dev_clean",
		"account_id":         "acc_dead_01",
		"payment_method": map[string]string{
			"type":  "CARD",
			"token": "tok_clean_01",
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP 200 fallback when ML is unavailable, got %d", resp.StatusCode)
	}

	var evalResp riskengine.RiskEvaluationResponse
	_ = json.NewDecoder(resp.Body).Decode(&evalResp)

	if !evalResp.IsDegraded {
		t.Errorf("Expected is_degraded = true when ML is down")
	}
	if evalResp.RiskScore < 0 || evalResp.RiskScore > 100 {
		t.Errorf("Expected valid fallback risk score [0, 100], got %d", evalResp.RiskScore)
	}
}

// TestResilience_RedisStoreUnavailable tests that missing/nil Redis stores fall back to zero defaults without crashing.
func TestResilience_RedisStoreUnavailable(t *testing.T) {
	// Orchestrator initialized with nil velocity and feature stores
	orchestrator := riskengine.NewOrchestrator(nil, nil, nil, nil, nil)
	handler := riskengine.NewHandler(orchestrator)

	r := chi.NewRouter()
	r.Post("/v1/risk-evaluations", handler.EvaluateRisk)
	srv := httptest.NewServer(r)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"transaction_id":     "txn_no_redis_01",
		"amount":             5000,
		"currency":           "INR",
		"ip_address":         "106.51.0.1",
		"device_fingerprint": "dev_phone",
		"account_id":         "acc_no_redis_01",
		"payment_method": map[string]string{
			"type":  "CARD",
			"token": "tok_test_01",
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP 200 when Redis is unconfigured, got %d", resp.StatusCode)
	}

	var evalResp riskengine.RiskEvaluationResponse
	_ = json.NewDecoder(resp.Body).Decode(&evalResp)

	if evalResp.DecisionID == "" {
		t.Errorf("Expected generated decision_id")
	}
}

// TestResilience_ContextCancellation tests client disconnection / context cancellation handling.
func TestResilience_ContextCancellation(t *testing.T) {
	orchestrator := riskengine.NewOrchestrator(nil, nil, nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := riskengine.RiskEvaluationRequest{
		TransactionID:     "txn_canceled_01",
		Amount:            1000,
		Currency:          "INR",
		DeviceFingerprint: "dev_canceled",
	}

	resp, err := orchestrator.Evaluate(ctx, "tenant_test", req)
	// Evaluation with cancelled context should either succeed with fast path or handle cancellation gracefully without panicking
	if err != nil && err != context.Canceled {
		t.Errorf("Unexpected error on cancelled context: %v", err)
	}
	_ = resp
}

// TestResilience_IdempotentReplay verifies that repeated evaluations of the same payload yield consistent decisions.
func TestResilience_IdempotentReplay(t *testing.T) {
	rulesService := rules.NewService(nil)
	orchestrator := riskengine.NewOrchestrator(nil, nil, rulesService, nil, nil)

	req := riskengine.RiskEvaluationRequest{
		TransactionID:     "txn_idempotent_888",
		Amount:            75000,
		Currency:          "INR",
		IPAddress:         "106.51.12.34",
		DeviceFingerprint: "dev_stable_phone",
		AccountID:         "acc_stable_01",
		PaymentMethod: riskengine.PaymentMethod{
			Type:  "CARD",
			Token: "tok_stable_card",
		},
	}

	resp1, err1 := orchestrator.Evaluate(context.Background(), "tenant_test", req)
	if err1 != nil {
		t.Fatalf("First eval failed: %v", err1)
	}

	resp2, err2 := orchestrator.Evaluate(context.Background(), "tenant_test", req)
	if err2 != nil {
		t.Fatalf("Second eval failed: %v", err2)
	}

	if resp1.RecommendedAction != resp2.RecommendedAction {
		t.Errorf("Non-deterministic action across identical replays: %s vs %s", resp1.RecommendedAction, resp2.RecommendedAction)
	}
	if resp1.RiskScore != resp2.RiskScore {
		t.Errorf("Non-deterministic score across identical replays: %d vs %d", resp1.RiskScore, resp2.RiskScore)
	}
}
