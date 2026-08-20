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

func TestEvaluateRisk_MockResponse(t *testing.T) {
	// Initialize handler with nil redis (falls back gracefully to default velocity)
	store := features.NewVelocityStore(nil)
	handler := riskengine.NewHandler(store)

	reqPayload := riskengine.RiskEvaluationRequest{
		TransactionID: "txn_test_123",
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

	if resp.TransactionID != "txn_test_123" {
		t.Errorf("expected transaction_id txn_test_123, got %s", resp.TransactionID)
	}

	if resp.RecommendedAction != "ALLOW_RECOMMENDATION" {
		t.Errorf("expected recommended_action ALLOW_RECOMMENDATION, got %s", resp.RecommendedAction)
	}

	if resp.DecisionID == "" {
		t.Errorf("expected non-empty decision_id")
	}

	if resp.FeatureSnapshotRef == "" {
		t.Errorf("expected non-empty feature_snapshot_ref")
	}
}
