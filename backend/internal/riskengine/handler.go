package riskengine

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// PaymentMethod details provided in the transaction context.
type PaymentMethod struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// RiskEvaluationRequest models the incoming POST /v1/risk-evaluations payload.
type RiskEvaluationRequest struct {
	TransactionID     string        `json:"transaction_id"`
	Amount            int64         `json:"amount"`
	Currency          string        `json:"currency"`
	PaymentMethod     PaymentMethod `json:"payment_method"`
	DeviceFingerprint string        `json:"device_fingerprint"`
	IPAddress         string        `json:"ip_address,omitempty"`
}

// RiskEvaluationResponse represents the standardized risk evaluation outcome.
type RiskEvaluationResponse struct {
	DecisionID         string                 `json:"decision_id"`
	TransactionID      string                 `json:"transaction_id"`
	RecommendedAction  string                 `json:"recommended_action"`
	RiskScore          int                    `json:"risk_score"`
	ReasonCodes        []string               `json:"reason_codes"`
	FeatureSnapshotRef string                 `json:"feature_snapshot_ref"`
	Features           map[string]interface{} `json:"features,omitempty"`
	EvaluatedAt        string                 `json:"evaluated_at"`
	IsDegraded         bool                   `json:"is_degraded,omitempty"`
	LatencyMs          int                    `json:"latency_ms"`
}

// Handler handles risk evaluation HTTP requests.
type Handler struct {
	orchestrator *Orchestrator
}

// NewHandler constructs a new risk evaluation Handler using the Orchestrator.
func NewHandler(orchestrator *Orchestrator) *Handler {
	return &Handler{
		orchestrator: orchestrator,
	}
}

// EvaluateRisk handles POST /v1/risk-evaluations.
func (h *Handler) EvaluateRisk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Resolve Tenant ID
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000001"
	}

	// 2. Decode Request Body
	var req RiskEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": fmt.Sprintf("Failed to parse request JSON: %v", err),
		})
		return
	}

	// If IP address is missing from payload, extract from remote address
	if req.IPAddress == "" {
		req.IPAddress = r.RemoteAddr
	}

	// 3. Execute Complete Synchronous Risk Orchestration Pipeline
	resp, err := h.orchestrator.Evaluate(r.Context(), tenantID, req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "evaluation_error",
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
