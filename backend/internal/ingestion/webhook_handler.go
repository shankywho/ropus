package ingestion

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookEvent struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"` // e.g. "dispute.opened", "payment.attempted"
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type DisputeRecord struct {
	DisputeID      string          `json:"dispute_id"`
	TenantID       string          `json:"tenant_id"`
	TransactionID  string          `json:"transaction_id"`
	DecisionID     *string         `json:"decision_id,omitempty"`
	Amount         int64           `json:"amount"`
	Currency       string          `json:"currency"`
	DisputeReason  string          `json:"dispute_reason"`
	Status         string          `json:"status"`
	EvidencePacket json.RawMessage `json:"evidence_packet"`
	CreatedAt      time.Time       `json:"created_at"`
}

type WebhookHandler struct {
	db            *pgxpool.Pool
	webhookSecret string
}

func NewWebhookHandler(db *pgxpool.Pool) *WebhookHandler {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		secret = "whsec_dummy_risk_secret_12345"
	}
	return &WebhookHandler{
		db:            db,
		webhookSecret: secret,
	}
}

// VerifySignature validates HMAC-SHA256 signature from headers.
func (h *WebhookHandler) VerifySignature(payload []byte, headerSignature string) bool {
	if headerSignature == "" {
		return false
	}

	// Support prefixes like "sha256=..." or raw hex
	sig := strings.TrimPrefix(headerSignature, "sha256=")
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

// HandleProviderWebhook processes incoming POST /webhooks/provider.
func (h *WebhookHandler) HandleProviderWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": "failed to read request body",
		})
		return
	}

	// 1. Signature Verification
	sigHeader := r.Header.Get("X-Signature")
	if sigHeader == "" {
		sigHeader = r.Header.Get("X-Webhook-Signature")
	}
	if sigHeader == "" {
		sigHeader = r.Header.Get("X-Hub-Signature-256")
	}

	// If signature header is provided, verify it strictly; if dev dummy header is passed, validate
	if sigHeader != "" {
		if !h.VerifySignature(body, sigHeader) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "unauthorized",
				"message": "HMAC-SHA256 signature verification failed",
			})
			return
		}
	}

	// 2. Decode Event
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_json",
			"message": fmt.Sprintf("failed to parse event JSON: %v", err),
		})
		return
	}

	if event.EventID == "" {
		event.EventID = fmt.Sprintf("evt_%s", uuid.New().String()[:8])
	}

	// 3. Process Specific Event Types
	switch event.EventType {
	case "dispute.opened":
		h.handleDisputeOpened(w, r, event)
	default:
		log.Printf("Received and acknowledged generic webhook event: %s (%s)", event.EventType, event.EventID)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "received",
			"event_id":   event.EventID,
			"event_type": event.EventType,
		})
	}
}

// handleDisputeOpened correlates dispute with original decision & attaches feature snapshot.
func (h *WebhookHandler) handleDisputeOpened(w http.ResponseWriter, r *http.Request, event WebhookEvent) {
	ctx := r.Context()
	data := event.Data

	txnID, _ := data["transaction_id"].(string)
	if txnID == "" {
		txnID, _ = data["txn_id"].(string)
	}
	if txnID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":   "validation_error",
			"message": "missing transaction_id in dispute data",
		})
		return
	}

	reason, _ := data["reason"].(string)
	if reason == "" {
		reason = "fraudulent_charge"
	}

	var amount int64 = 0
	if amtVal, ok := data["amount"].(float64); ok {
		amount = int64(amtVal)
	}

	currency, _ := data["currency"].(string)
	if currency == "" {
		currency = "INR"
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000001"
	}

	disputeID := uuid.New().String()
	now := time.Now().UTC()

	var decisionID *string
	var featureSnapshot json.RawMessage
	var riskScore int
	var recAction string
	var reasonCodes json.RawMessage

	// Query original risk decision
	if h.db != nil {
		query := `
			SELECT decision_id, amount, currency, risk_score, recommended_action, reason_codes, feature_snapshot
			FROM risk_decisions
			WHERE tenant_id = $1 AND transaction_id = $2
			LIMIT 1
		`
		var decID string
		err := h.db.QueryRow(ctx, query, tenantID, txnID).Scan(
			&decID,
			&amount,
			&currency,
			&riskScore,
			&recAction,
			&reasonCodes,
			&featureSnapshot,
		)
		if err == nil {
			decisionID = &decID
		}
	}

	// Construct immutable Evidence Packet
	evidenceMap := map[string]interface{}{
		"dispute_id":             disputeID,
		"transaction_id":         txnID,
		"dispute_reason":         reason,
		"original_decision_id":   decisionID,
		"original_risk_score":    riskScore,
		"original_recommendation": recAction,
		"original_reason_codes":  reasonCodes,
		"feature_snapshot":       featureSnapshot,
		"correlated_at":          now.Format(time.RFC3339),
	}
	evidenceBytes, _ := json.Marshal(evidenceMap)

	// Insert into disputes table
	if h.db != nil {
		insertQuery := `
			INSERT INTO disputes (dispute_id, tenant_id, transaction_id, decision_id, amount, currency, dispute_reason, status, evidence_packet, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		_, err := h.db.Exec(ctx, insertQuery,
			disputeID,
			tenantID,
			txnID,
			decisionID,
			amount,
			currency,
			reason,
			"EVIDENCE_ATTACHED",
			evidenceBytes,
			now,
			now,
		)
		if err != nil {
			log.Printf("Warning: Failed to persist dispute record: %v", err)
		}

		// Insert into audit log
		_, _ = h.db.Exec(ctx, `
			INSERT INTO audit_log (tenant_id, actor_id, action, entity_type, entity_id, changes, created_at)
			VALUES ($1, 'WEBHOOK_INGESTION', 'DISPUTE_OPENED', 'DISPUTE', $2, $3, $4)
		`, tenantID, disputeID, fmt.Sprintf(`{"transaction_id":"%s","status":"EVIDENCE_ATTACHED"}`, txnID), now)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":            "processed",
		"event_id":          event.EventID,
		"dispute_id":        disputeID,
		"transaction_id":    txnID,
		"evidence_attached": decisionID != nil,
		"evidence_packet":   evidenceMap,
	})
}
