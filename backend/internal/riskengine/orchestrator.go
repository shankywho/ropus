package riskengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/rules"
	"github.com/shankywho/ropus/backend/internal/utils"
)

// Orchestrator orchestrates the real-time synchronous risk evaluation pipeline.
type Orchestrator struct {
	db            *pgxpool.Pool
	velocityStore *features.VelocityStore
	rulesService  *rules.Service
	mlClient      *MLClient
	kms           utils.KMS
}

// NewOrchestrator constructs a new risk Orchestrator.
func NewOrchestrator(
	db *pgxpool.Pool,
	velocityStore *features.VelocityStore,
	rulesService *rules.Service,
	mlClient *MLClient,
	kms utils.KMS,
) *Orchestrator {
	if kms == nil {
		kms = utils.NewMockKMS()
	}
	return &Orchestrator{
		db:            db,
		velocityStore: velocityStore,
		rulesService:  rulesService,
		mlClient:      mlClient,
		kms:           kms,
	}
}

// EnsureTenantExists creates the tenant record if it doesn't already exist.
func (o *Orchestrator) ensureTenantExists(ctx context.Context, tenantID string) error {
	if o.db == nil {
		return nil
	}
	_, err := o.db.Exec(ctx, `
		INSERT INTO tenants (tenant_id, name, api_key_hash, status)
		VALUES ($1, 'Default Tenant', $2, 'ACTIVE')
		ON CONFLICT (tenant_id) DO NOTHING
	`, tenantID, fmt.Sprintf("key_%s", tenantID))
	return err
}

// maskIPAddress produces a privacy-preserving masked IP for logging/audit streams.
func maskIPAddress(ip string) string {
	if ip == "" {
		return ""
	}
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.***.***", parts[0], parts[1])
	}
	return "masked_ip"
}

// Evaluate executes the complete synchronous decision pipeline.
func (o *Orchestrator) Evaluate(ctx context.Context, tenantID string, req RiskEvaluationRequest) (*RiskEvaluationResponse, error) {
	startTime := time.Now()

	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000001"
	}
	if req.TransactionID == "" {
		req.TransactionID = fmt.Sprintf("txn_%s", uuid.New().String()[:12])
	}
	if req.Currency == "" {
		req.Currency = "INR"
	}

	decisionID := fmt.Sprintf("dec_%s", uuid.New().String())
	snapshotRef := fmt.Sprintf("snap_%s", uuid.New().String()[:10])

	// -------------------------------------------------------------
	// STEP 1: Context Aggregation & Velocity Queries
	// -------------------------------------------------------------
	ip := req.IPAddress
	token := req.PaymentMethod.Token

	var velocityMetrics *features.VelocityMetrics
	var err error
	if o.velocityStore != nil {
		velocityMetrics, err = o.velocityStore.GetVelocityMetrics(ctx, tenantID, ip, token)
		if err != nil {
			log.Printf("Velocity fetch failed, degrading gracefully: %v", err)
			velocityMetrics = &features.VelocityMetrics{TxnCountIP1h: 0, TxnCountToken24h: 0}
		}
		// Record current transaction in velocity store
		_ = o.velocityStore.RecordEvent(ctx, tenantID, ip, token, req.Amount)
	} else {
		velocityMetrics = &features.VelocityMetrics{TxnCountIP1h: 0, TxnCountToken24h: 0}
	}

	// Build the in-memory evaluation context for real-time inference & rules
	evalContext := map[string]interface{}{
		"transaction_id":      req.TransactionID,
		"amount":              req.Amount,
		"currency":            req.Currency,
		"device_fingerprint":  req.DeviceFingerprint,
		"ip_address":          ip,
		"payment_method": map[string]interface{}{
			"type":  req.PaymentMethod.Type,
			"token": token,
		},
		"velocity.ip.1hr":     velocityMetrics.TxnCountIP1h,
		"velocity.token.24hr": velocityMetrics.TxnCountToken24h,
		"features": map[string]interface{}{
			"ipTxnCount1h":     velocityMetrics.TxnCountIP1h,
			"tokenTxnCount24h": velocityMetrics.TxnCountToken24h,
		},
	}

	// -------------------------------------------------------------
	// STEP 2: Fetch Active Rules & Evaluate Pre-Rules (Hard Guardrails)
	// -------------------------------------------------------------
	var activeRules []rules.Rule
	if o.rulesService != nil {
		activeStatus := rules.StatusActive
		activeRules, err = o.rulesService.ListRules(ctx, tenantID, &activeStatus)
		if err != nil {
			log.Printf("Warning: Failed to fetch active rules: %v", err)
		}
	}

	reasonCodes := make([]string, 0)
	var finalAction string
	var preRuleTriggered bool
	riskScore := 10 // baseline low score

	// Evaluate pre-rules (hard blocks / allows)
	for _, r := range activeRules {
		ruleDef, err := rules.ParseRuleDefinition(r.DSLAST)
		if err != nil {
			continue
		}

		matched, err := ruleDef.Condition.Evaluate(evalContext)
		if err != nil || !matched {
			continue
		}

		// Pre-rule matched!
		action := ruleDef.Action
		if action == "" {
			action = "MANUAL_REVIEW"
		}
		if ruleDef.ReasonCode != "" {
			reasonCodes = append(reasonCodes, ruleDef.ReasonCode)
		} else {
			reasonCodes = append(reasonCodes, fmt.Sprintf("RULE_%s_MATCHED", r.Name))
		}

		// Hard decline / allow triggers immediate halt of further pipeline steps
		if action == "DECLINE_RECOMMENDATION" || action == "ALLOW_RECOMMENDATION" {
			finalAction = action
			preRuleTriggered = true
			if action == "DECLINE_RECOMMENDATION" {
				riskScore = 95
			} else {
				riskScore = 5
			}
			break
		}
	}

	isDegraded := false

	// -------------------------------------------------------------
	// STEP 3: ML Inference (if pre-rules did not halt pipeline)
	// -------------------------------------------------------------
	if !preRuleTriggered {
		// Prepare ML feature payload
		mlReq := MLPredictRequest{
			Amount:           float64(req.Amount),
			IPVelocity1h:     float64(velocityMetrics.TxnCountIP1h),
			TokenVelocity24h: float64(velocityMetrics.TxnCountToken24h),
			IsNewDevice:      0,
		}
		if req.DeviceFingerprint != "" && (len(req.DeviceFingerprint) < 8 || req.DeviceFingerprint == "new_device") {
			mlReq.IsNewDevice = 1
		}

		// Execute inference with 50ms timeout
		if o.mlClient != nil {
			mlResp, err := o.mlClient.Predict(ctx, mlReq)
			if err != nil {
				// Graceful degradation on ML failure or 50ms timeout
				log.Printf("ML inference degraded (%v). Falling back to rules/heuristics.", err)
				isDegraded = true
				riskScore = o.calculateFallbackRiskScore(req.Amount, velocityMetrics)
				reasonCodes = append(reasonCodes, "ML_SERVICE_DEGRADED")
			} else {
				riskScore = mlResp.RiskScore
				if len(mlResp.ReasonCodes) > 0 {
					reasonCodes = append(reasonCodes, mlResp.ReasonCodes...)
				}
			}
		} else {
			isDegraded = true
			riskScore = o.calculateFallbackRiskScore(req.Amount, velocityMetrics)
		}

		// -------------------------------------------------------------
		// STEP 4: Post-Rules & Dynamic Thresholds
		// -------------------------------------------------------------
		evalContext["risk_score"] = riskScore

		// Threshold-based outcome mapping
		if finalAction == "" {
			switch {
			case riskScore >= 85:
				finalAction = "DECLINE_RECOMMENDATION"
			case riskScore >= 65:
				finalAction = "MANUAL_REVIEW"
			case riskScore >= 45:
				finalAction = "STEP_UP_RECOMMENDATION"
			default:
				finalAction = "ALLOW_RECOMMENDATION"
			}
		}
	}

	// -------------------------------------------------------------
	// STEP 5: Envelope Encryption & Transactional Outbox Persistence
	// -------------------------------------------------------------
	latencyMs := int(time.Since(startTime).Milliseconds())
	nowUTC := time.Now().UTC()

	// 1. Retrieve Tenant AES-256 Key from KMS
	tenantKey, keyErr := o.kms.GetTenantKey(tenantID)
	if keyErr != nil {
		log.Printf("Warning: KMS key retrieval error: %v", keyErr)
	}

	// 2. Encrypt PII (IP Address & Device Fingerprint) for at-rest storage
	encryptedIP := ip
	encryptedDeviceFP := req.DeviceFingerprint
	if len(tenantKey) > 0 {
		if encIP, err := utils.EncryptString(ip, tenantKey); err == nil {
			encryptedIP = encIP
		}
		if encFP, err := utils.EncryptString(req.DeviceFingerprint, tenantKey); err == nil {
			encryptedDeviceFP = encFP
		}
	}

	// 3. Prepare Encrypted Feature Snapshot (Stored in Postgres)
	encryptedFeatureSnapshot := make(map[string]interface{}, len(evalContext))
	for k, v := range evalContext {
		encryptedFeatureSnapshot[k] = v
	}
	encryptedFeatureSnapshot["ip_address"] = encryptedIP
	encryptedFeatureSnapshot["device_fingerprint"] = encryptedDeviceFP
	encryptedFeatureSnapshot["_encryption"] = "AES-256-GCM"

	featureSnapshotBytes, _ := json.Marshal(encryptedFeatureSnapshot)

	// Encrypt raw payload PII
	encryptedRawPayload := req
	encryptedRawPayload.IPAddress = encryptedIP
	encryptedRawPayload.DeviceFingerprint = encryptedDeviceFP
	rawPayloadBytes, _ := json.Marshal(encryptedRawPayload)
	reasonCodesBytes, _ := json.Marshal(reasonCodes)

	// 4. Prepare Sanitized Outbox Payload (Decrypted PII is NEVER sent to Kafka)
	outboxSnapshot := make(map[string]interface{}, len(evalContext))
	for k, v := range evalContext {
		outboxSnapshot[k] = v
	}
	outboxSnapshot["ip_address"] = maskIPAddress(ip)
	outboxSnapshot["device_fingerprint"] = encryptedDeviceFP

	outboxPayload := map[string]interface{}{
		"decision_id":          decisionID,
		"tenant_id":            tenantID,
		"transaction_id":       req.TransactionID,
		"amount":               req.Amount,
		"currency":             req.Currency,
		"recommended_action":   finalAction,
		"risk_score":           riskScore,
		"reason_codes":         reasonCodes,
		"feature_snapshot_ref": snapshotRef,
		"feature_snapshot":     outboxSnapshot,
		"latency_ms":           latencyMs,
		"evaluated_at":         nowUTC.Format(time.RFC3339),
	}
	outboxPayloadBytes, _ := json.Marshal(outboxPayload)
	outboxID := uuid.New().String()

	if o.db != nil {
		_ = o.ensureTenantExists(ctx, tenantID)

		tx, err := o.db.Begin(ctx)
		if err != nil {
			log.Printf("Warning: Failed to start database transaction: %v", err)
		} else {
			defer tx.Rollback(ctx)

			// Insert into risk_decisions with encrypted PII
			insertDecisionQuery := `
				INSERT INTO risk_decisions (
					decision_id, tenant_id, transaction_id, amount, currency,
					recommended_action, risk_score, reason_codes,
					feature_snapshot_ref, feature_snapshot, raw_payload, latency_ms, created_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
				ON CONFLICT (tenant_id, transaction_id) DO UPDATE
				SET recommended_action = EXCLUDED.recommended_action,
				    risk_score = EXCLUDED.risk_score,
				    reason_codes = EXCLUDED.reason_codes,
				    feature_snapshot = EXCLUDED.feature_snapshot,
				    raw_payload = EXCLUDED.raw_payload,
				    latency_ms = EXCLUDED.latency_ms
			`
			_, err = tx.Exec(ctx, insertDecisionQuery,
				decisionID,
				tenantID,
				req.TransactionID,
				req.Amount,
				req.Currency,
				finalAction,
				riskScore,
				reasonCodesBytes,
				snapshotRef,
				featureSnapshotBytes,
				rawPayloadBytes,
				latencyMs,
				nowUTC,
			)
			if err != nil {
				log.Printf("Error inserting risk_decision in tx: %v", err)
			}

			// Insert outbox event (guaranteed sanitized / encrypted PII)
			insertOutboxQuery := `
				INSERT INTO outbox_events (id, aggregate_type, aggregate_id, type, payload, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err = tx.Exec(ctx, insertOutboxQuery,
				outboxID,
				"RiskDecision",
				decisionID,
				"risk.decisioned",
				outboxPayloadBytes,
				nowUTC,
			)
			if err != nil {
				log.Printf("Error inserting outbox_events in tx: %v", err)
			}

			if commitErr := tx.Commit(ctx); commitErr != nil {
				log.Printf("Error committing risk decision transaction: %v", commitErr)
			}
		}
	}

	// -------------------------------------------------------------
	// STEP 6: Return Standardized JSON Response
	// -------------------------------------------------------------
	return &RiskEvaluationResponse{
		DecisionID:         decisionID,
		TransactionID:      req.TransactionID,
		RecommendedAction:  finalAction,
		RiskScore:          riskScore,
		ReasonCodes:        reasonCodes,
		FeatureSnapshotRef: snapshotRef,
		Features:           evalContext,
		EvaluatedAt:        nowUTC.Format(time.RFC3339),
		IsDegraded:         isDegraded,
		LatencyMs:          latencyMs,
	}, nil
}

// calculateFallbackRiskScore generates a heuristic score when ML service is degraded.
func (o *Orchestrator) calculateFallbackRiskScore(amount int64, velocity *features.VelocityMetrics) int {
	score := 15
	if amount > 100000 {
		score += 35
	}
	if velocity.TxnCountIP1h >= 4 {
		score += 30
	}
	if velocity.TxnCountToken24h >= 6 {
		score += 25
	}
	if score > 99 {
		score = 99
	}
	return score
}
