package riskengine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/shankywho/ropus/backend/internal/utils"
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
	AccountID         string        `json:"account_id,omitempty"`
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
	rateLimiter  *TenantRateLimiter
}

// NewHandler constructs a new risk evaluation Handler using the Orchestrator.
func NewHandler(orchestrator *Orchestrator) *Handler {
	return &Handler{
		orchestrator: orchestrator,
		rateLimiter:  NewTenantRateLimiter(DefaultRateLimiterConfig()),
	}
}

// SetRateLimiter updates the rate limiter instance.
func (h *Handler) SetRateLimiter(rl *TenantRateLimiter) {
	h.rateLimiter = rl
}

// EvaluateRisk handles POST /v1/risk-evaluations with tenant isolation, rate limiting, and request validation.
func (h *Handler) EvaluateRisk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Propagate Correlation ID
	correlationID := r.Header.Get("X-Correlation-ID")
	if correlationID != "" {
		w.Header().Set("X-Correlation-ID", correlationID)
	}

	// 2. Resolve & Sanitize Tenant ID
	rawTenantID := r.Header.Get("X-Tenant-ID")
	if rawTenantID == "" {
		rawTenantID = "00000000-0000-0000-0000-000000000001"
	}
	tenantID, err := utils.SanitizeIdentifier(rawTenantID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_tenant_id",
			"message": fmt.Sprintf("Invalid tenant identifier '%s': %v", rawTenantID, err),
		})
		return
	}

	// 3. Per-Tenant Rate Limiting
	if h.rateLimiter != nil && !h.rateLimiter.Allow(tenantID) {
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "rate_limit_exceeded",
			"message": fmt.Sprintf("Request quota exceeded for tenant '%s'; retry after backoff", tenantID),
		})
		return
	}

	// 4. Decode Request Body (bounded to 1 MB)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req RiskEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": fmt.Sprintf("Failed to parse request JSON or payload exceeds 1MB limit: %v", err),
		})
		return
	}

	// 5. Validate Required Fields
	req.TransactionID = strings.TrimSpace(req.TransactionID)
	if req.TransactionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": "Field 'transaction_id' is required and cannot be empty",
		})
		return
	}
	if len(req.TransactionID) > 128 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": "Field 'transaction_id' exceeds maximum allowed length of 128 characters",
		})
		return
	}

	if req.Amount < 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": "Field 'amount' cannot be negative",
		})
		return
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	// If IP address is missing from payload, extract from remote address
	if req.IPAddress == "" {
		req.IPAddress = r.RemoteAddr
	}

	// 6. Execute Synchronous Risk Orchestration Pipeline
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
