package riskengine_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/riskengine"
)

func TestEvaluateRisk_OrchestratorPipeline(t *testing.T) {
	// 1. Mock ML Sidecar HTTP Server
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := riskengine.MLPredictResponse{
			RiskScore:           78,
			Probability:         0.78,
			ReasonCodes:         []string{"HIGH_IP_VELOCITY_1H"},
			FeatureAttributions: map[string]float64{"ip_velocity_1h": 0.45},
			LatencyMs:           4.2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mlServer.Close()

	// 2. Initialize Orchestrator with mock ML client and nil DB/velocity (degrades gracefully)
	mlClient := riskengine.NewMLClient(mlServer.URL)
	store := features.NewVelocityStore(nil)
	orchestrator := riskengine.NewOrchestrator(nil, store, nil, mlClient, nil)
	handler := riskengine.NewHandler(orchestrator)

	reqPayload := riskengine.RiskEvaluationRequest{
		TransactionID: "txn_test_orchestrator_1",
		Amount:        50000,
		Currency:      "INR",
		PaymentMethod: riskengine.PaymentMethod{
			Type:  "card",
			Token: "tkn_card_xyz456",
		},
		DeviceFingerprint: "fp_device_abc",
		IPAddress:         "192.168.1.100",
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant_demo_1")

	rr := httptest.NewRecorder()
	handler.EvaluateRisk(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: body=%s", rr.Code, rr.Body.String())
	}

	var resp riskengine.RiskEvaluationResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	if resp.TransactionID != "txn_test_orchestrator_1" {
		t.Errorf("expected transaction_id txn_test_orchestrator_1, got %s", resp.TransactionID)
	}

	if resp.RiskScore != 78 {
		t.Errorf("expected risk_score 78 from mock ML service, got %d", resp.RiskScore)
	}

	// Score 78 maps to MANUAL_REVIEW
	if resp.RecommendedAction != "MANUAL_REVIEW" {
		t.Errorf("expected recommended_action MANUAL_REVIEW for score 78, got %s", resp.RecommendedAction)
	}

	if len(resp.ReasonCodes) == 0 || resp.ReasonCodes[0] != "HIGH_IP_VELOCITY_1H" {
		t.Errorf("expected reason code HIGH_IP_VELOCITY_1H, got %v", resp.ReasonCodes)
	}
}

func TestEvaluateRisk_MLTimeoutFallback(t *testing.T) {
	// ML server that simulates non-responding / unreachable endpoint
	mlClient := riskengine.NewMLClient("http://127.0.0.1:59999") // invalid port
	store := features.NewVelocityStore(nil)
	orchestrator := riskengine.NewOrchestrator(nil, store, nil, mlClient, nil)
	handler := riskengine.NewHandler(orchestrator)

	reqPayload := riskengine.RiskEvaluationRequest{
		TransactionID: "txn_test_timeout",
		Amount:        2500,
		Currency:      "INR",
		PaymentMethod: riskengine.PaymentMethod{
			Type:  "card",
			Token: "tkn_card_timeout",
		},
		DeviceFingerprint: "fp_device_abc",
		IPAddress:         "192.168.1.100",
	}

	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handler.EvaluateRisk(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 on fallback, got %d", rr.Code)
	}

	var resp riskengine.RiskEvaluationResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if !resp.IsDegraded {
		t.Errorf("expected is_degraded=true when ML service fails")
	}

	if resp.RecommendedAction == "" {
		t.Errorf("expected valid recommended_action even on degraded fallback")
	}
}
