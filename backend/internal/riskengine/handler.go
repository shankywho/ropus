package riskengine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shankywho/ropus/backend/internal/features"
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
}

// Handler handles risk evaluation HTTP requests.
type Handler struct {
	velocityStore *features.VelocityStore
}

// NewHandler constructs a new risk evaluation Handler.
func NewHandler(velocityStore *features.VelocityStore) *Handler {
	return &Handler{
		velocityStore: velocityStore,
	}
}

// EvaluateRisk handles POST /v1/risk-evaluations.
func (h *Handler) EvaluateRisk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Resolve Tenant ID (defaulting to demo tenant if not provided)
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000001"
	}

	// 2. Decode Request Body
	var req RiskEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"type":    "invalid_request",
			"message": fmt.Sprintf("Failed to parse request JSON: %v", err),
		})
		return
	}

	if req.TransactionID == "" {
		req.TransactionID = fmt.Sprintf("txn_%s", uuid.New().String()[:8])
	}

	// 3. Query Velocity Feature Store
	ip := req.IPAddress
	if ip == "" {
		ip = r.RemoteAddr
	}
	token := req.PaymentMethod.Token

	metrics, err := h.velocityStore.GetVelocityMetrics(r.Context(), tenantID, ip, token)
	if err != nil {
		// Log error, but fail-open/degrade gracefully for velocity features
		metrics = &features.VelocityMetrics{
			TxnCountIP1h:     0,
			TxnCountToken24h: 0,
		}
	}

	// 4. Record current event in velocity store asynchronously / best effort
	_ = h.velocityStore.RecordEvent(r.Context(), tenantID, ip, token, req.Amount)

	// 5. Construct Baseline Response (MVP mock ALLOW_RECOMMENDATION)
	decisionID := fmt.Sprintf("dec_%s", uuid.New().String())
	snapshotRef := fmt.Sprintf("snap_%s", uuid.New().String()[:8])

	resp := RiskEvaluationResponse{
		DecisionID:         decisionID,
		TransactionID:      req.TransactionID,
		RecommendedAction:  "ALLOW_RECOMMENDATION",
		RiskScore:          15, // Low risk baseline
		ReasonCodes:        []string{},
		FeatureSnapshotRef: snapshotRef,
		Features: map[string]interface{}{
			"velocity.ip.1hr":     metrics.TxnCountIP1h,
			"velocity.token.24hr": metrics.TxnCountToken24h,
			"amount":              req.Amount,
			"currency":            req.Currency,
		},
		EvaluatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
