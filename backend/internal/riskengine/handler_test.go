package riskengine_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/riskengine"
)

func TestEvaluateRisk_OrchestratorPipeline(t *testing.T) {
	var capturedMLReq riskengine.MLPredictRequest

	// 1. Mock ML Sidecar HTTP Server
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewDecoder(r.Body).Decode(&capturedMLReq)

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

	t.Run("Valid Fingerprint Ingestion", func(t *testing.T) {
		reqPayload := riskengine.RiskEvaluationRequest{
			TransactionID: "txn_test_orchestrator_1",
			Amount:        50000,
			Currency:      "INR",
			PaymentMethod: riskengine.PaymentMethod{
				Type:  "card",
				Token: "tkn_card_xyz456",
			},
			DeviceFingerprint: "fp_known_iphone15_pro_v32",
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

		// Verify device_id was generated deterministically in features
		expectedDeviceID := features.HashDeviceID("tenant_demo_1", "fp_known_iphone15_pro_v32")
		devIDVal, ok := resp.Features["device_id"].(string)
		if !ok || devIDVal != expectedDeviceID {
			t.Errorf("expected device_id %s in features, got %v", expectedDeviceID, devIDVal)
		}

		// Verify that first-seen fingerprint passes IsNewDevice = 1 (novel device) to ML sidecar
		if capturedMLReq.IsNewDevice != 1 {
			t.Errorf("expected IsNewDevice=1 for first-seen fingerprint, got %d", capturedMLReq.IsNewDevice)
		}
	})

	t.Run("Missing Fingerprint Ingestion", func(t *testing.T) {
		reqPayload := riskengine.RiskEvaluationRequest{
			TransactionID: "txn_test_missing_fp",
			Amount:        1000,
			Currency:      "INR",
			PaymentMethod: riskengine.PaymentMethod{
				Type:  "card",
				Token: "tkn_card_123",
			},
			DeviceFingerprint: "",
			IPAddress:         "192.168.1.100",
		}

		body, _ := json.Marshal(reqPayload)
		req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant_demo_1")

		rr := httptest.NewRecorder()
		handler.EvaluateRisk(rr, req)

		var resp riskengine.RiskEvaluationResponse
		_ = json.NewDecoder(rr.Body).Decode(&resp)

		if resp.Features["device_status"] != "MISSING" {
			t.Errorf("expected device_status MISSING, got %v", resp.Features["device_status"])
		}

		// Missing fingerprint sets IsNewDevice = 1 (novelty indicator)
		if capturedMLReq.IsNewDevice != 1 {
			t.Errorf("expected IsNewDevice=1 for missing fingerprint, got %d", capturedMLReq.IsNewDevice)
		}
	})

	t.Run("Oversized Fingerprint (> 256 chars)", func(t *testing.T) {
		reqPayload := riskengine.RiskEvaluationRequest{
			TransactionID: "txn_test_oversized_fp",
			Amount:        1000,
			Currency:      "INR",
			PaymentMethod: riskengine.PaymentMethod{
				Type:  "card",
				Token: "tkn_card_123",
			},
			DeviceFingerprint: strings.Repeat("a", 300),
			IPAddress:         "192.168.1.100",
		}

		body, _ := json.Marshal(reqPayload)
		req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant_demo_1")

		rr := httptest.NewRecorder()
		handler.EvaluateRisk(rr, req)

		var resp riskengine.RiskEvaluationResponse
		_ = json.NewDecoder(rr.Body).Decode(&resp)

		if !resp.IsDegraded {
			t.Errorf("expected is_degraded=true for oversized fingerprint")
		}

		hasReasonCode := false
		for _, rc := range resp.ReasonCodes {
			if rc == "INVALID_DEVICE_TELEMETRY" {
				hasReasonCode = true
				break
			}
		}
		if !hasReasonCode {
			t.Errorf("expected INVALID_DEVICE_TELEMETRY reason code, got %v", resp.ReasonCodes)
		}
	})

	t.Run("Malformed Control Character Fingerprint", func(t *testing.T) {
		reqPayload := riskengine.RiskEvaluationRequest{
			TransactionID: "txn_test_control_char",
			Amount:        1000,
			Currency:      "INR",
			PaymentMethod: riskengine.PaymentMethod{
				Type:  "card",
				Token: "tkn_card_123",
			},
			DeviceFingerprint: "fp_injected\x00_null_byte",
			IPAddress:         "192.168.1.100",
		}

		body, _ := json.Marshal(reqPayload)
		req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant_demo_1")

		rr := httptest.NewRecorder()
		handler.EvaluateRisk(rr, req)

		var resp riskengine.RiskEvaluationResponse
		_ = json.NewDecoder(rr.Body).Decode(&resp)

		if !resp.IsDegraded {
			t.Errorf("expected is_degraded=true for control character injection")
		}
	})
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

func TestEvaluateRisk_CanaryRoutingCandidate(t *testing.T) {
	candidateHit := false
	legacyHit := false

	// Mock ML server supporting both /predict (legacy) and /predict/shadow (candidate)
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/predict/shadow" {
			candidateHit = true
			_ = json.NewEncoder(w).Encode(riskengine.MLShadowPredictResponse{
				ModelVersion:           "fraud-xgb-25f-candidate-v1",
				FeatureContractVersion: "v2.5",
				FeatureCount:           25,
				RawProbability:         0.04,
				CalibratedProbability:  0.035,
				RiskScore:              4,
				ShadowDecision:         "ALLOW_RECOMMENDATION",
				LatencyMs:              1.5,
				Runtime:                "mock_candidate",
			})
			return
		}
		if r.URL.Path == "/predict" {
			legacyHit = true
			_ = json.NewEncoder(w).Encode(riskengine.MLPredictResponse{
				RiskScore:           70,
				Probability:         0.70,
				ReasonCodes:         []string{"LEGACY_MODEL_CALLED"},
				FeatureAttributions: map[string]float64{"amount": 0.1},
				LatencyMs:           2.0,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mlServer.Close()

	mlClient := riskengine.NewMLClient(mlServer.URL)
	orchestrator := riskengine.NewOrchestrator(nil, nil, nil, mlClient, nil)

	// 100% candidate canary router
	canaryCfg := riskengine.CanaryRouterConfig{
		Enabled:                  true,
		Percentage:               100,
		CandidateModelVersion:    "fraud-xgb-25f-candidate-v1",
		CandidateFeatureContract: riskengine.MLFeatureContractV25,
	}
	canaryRouter := riskengine.NewCanaryRouter(canaryCfg, nil)
	orchestrator.SetCanaryRouter(canaryRouter)

	handler := riskengine.NewHandler(orchestrator)

	reqPayload := riskengine.RiskEvaluationRequest{
		TransactionID: "txn_canary_test_01",
		Amount:        1500,
		Currency:      "USD",
		PaymentMethod: riskengine.PaymentMethod{
			Type:  "card",
			Token: "tok_canary_01",
		},
		DeviceFingerprint: "ua=TestBrowser|sr=1920x1080",
		IPAddress:         "10.0.0.1",
	}
	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.EvaluateRisk(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if !candidateHit {
		t.Errorf("Expected candidate model endpoint /predict/shadow to be hit for 100%% canary")
	}
	if legacyHit {
		t.Errorf("Did not expect legacy model /predict to be hit when candidate succeeded")
	}

	var resp riskengine.RiskEvaluationResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.RiskScore != 4 {
		t.Errorf("Expected candidate risk_score 4, got %d", resp.RiskScore)
	}
	if resp.RecommendedAction != "ALLOW_RECOMMENDATION" {
		t.Errorf("Expected ALLOW_RECOMMENDATION, got %s", resp.RecommendedAction)
	}

	if resp.Features["model_route"] != "CANDIDATE" {
		t.Errorf("Expected model_route CANDIDATE in features, got %v", resp.Features["model_route"])
	}
}

func TestEvaluateRisk_CanaryCandidateFallbackToLegacy(t *testing.T) {
	candidateHit := false
	legacyHit := false

	// Mock ML server where /predict/shadow fails with 500, but /predict succeeds
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/predict/shadow" {
			candidateHit = true
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"detail":"simulated candidate failure"}`))
			return
		}
		if r.URL.Path == "/predict" {
			legacyHit = true
			_ = json.NewEncoder(w).Encode(riskengine.MLPredictResponse{
				RiskScore:           72,
				Probability:         0.72,
				ReasonCodes:         []string{"LEGACY_FALLBACK_ACTIVE"},
				FeatureAttributions: map[string]float64{"amount": 0.15},
				LatencyMs:           1.8,
			})
			return
		}
	}))
	defer mlServer.Close()

	mlClient := riskengine.NewMLClient(mlServer.URL)
	orchestrator := riskengine.NewOrchestrator(nil, nil, nil, mlClient, nil)

	canaryCfg := riskengine.CanaryRouterConfig{
		Enabled:                  true,
		Percentage:               100,
		CandidateModelVersion:    "fraud-xgb-25f-candidate-v1",
		CandidateFeatureContract: riskengine.MLFeatureContractV25,
	}
	canaryRouter := riskengine.NewCanaryRouter(canaryCfg, nil)
	orchestrator.SetCanaryRouter(canaryRouter)

	handler := riskengine.NewHandler(orchestrator)

	reqPayload := riskengine.RiskEvaluationRequest{
		TransactionID: "txn_fallback_test_01",
		Amount:        2000,
		Currency:      "USD",
		PaymentMethod: riskengine.PaymentMethod{
			Type:  "card",
			Token: "tok_fallback_01",
		},
		DeviceFingerprint: "ua=FallbackBrowser",
		IPAddress:         "10.0.0.2",
	}
	body, _ := json.Marshal(reqPayload)
	req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.EvaluateRisk(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 on fallback, got %d: %s", rr.Code, rr.Body.String())
	}

	if !candidateHit {
		t.Errorf("Expected candidate endpoint to be attempted")
	}
	if !legacyHit {
		t.Errorf("Expected automatic fallback to legacy model on candidate error")
	}

	var resp riskengine.RiskEvaluationResponse
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.RiskScore != 72 {
		t.Errorf("Expected legacy fallback score 72, got %d", resp.RiskScore)
	}
	if resp.RecommendedAction != "MANUAL_REVIEW" {
		t.Errorf("Expected MANUAL_REVIEW for score 72, got %s", resp.RecommendedAction)
	}

	// Verify CANARY_FALLBACK_TO_LEGACY reason code was attached
	hasFallbackCode := false
	for _, code := range resp.ReasonCodes {
		if code == "CANARY_FALLBACK_TO_LEGACY" {
			hasFallbackCode = true
			break
		}
	}
	if !hasFallbackCode {
		t.Errorf("Expected CANARY_FALLBACK_TO_LEGACY reason code in %v", resp.ReasonCodes)
	}

	// Verify router recorded the fallback
	status := canaryRouter.GetStatus()
	metrics := status["metrics"].(map[string]interface{})
	if metrics["candidate_fallback_total"].(int64) != 1 {
		t.Errorf("Expected candidate_fallback_total=1, got %v", metrics["candidate_fallback_total"])
	}
}

