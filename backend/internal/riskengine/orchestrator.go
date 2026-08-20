package riskengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/rules"
)

// Orchestrator orchestrates the real-time synchronous risk evaluation pipeline.
type Orchestrator struct {
	db            *pgxpool.Pool
	velocityStore *features.VelocityStore
	rulesService  *rules.Service
	mlClient      *MLClient
}

// NewOrchestrator constructs a new risk Orchestrator.
func NewOrchestrator(
	db *pgxpool.Pool,
	velocityStore *features.VelocityStore,
	rulesService *rules.Service,
	mlClient *MLClient,
) *Orchestrator {
	return &Orchestrator{
		db:            db,
		velocityStore: velocityStore,
		rulesService:  rulesService,
		mlClient:      mlClient,
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

	// Build the unified evaluation context
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
	// STEP 5: Persistence into PostgreSQL risk_decisions
	// -------------------------------------------------------------
	latencyMs := int(time.Since(startTime).Milliseconds())

	featureSnapshotBytes, _ := json.Marshal(evalContext)
	rawPayloadBytes, _ := json.Marshal(req)
	reasonCodesBytes, _ := json.Marshal(reasonCodes)

	if o.db != nil {
		_ = o.ensureTenantExists(ctx, tenantID)

		insertQuery := `
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
			    latency_ms = EXCLUDED.latency_ms
		`

		_, err = o.db.Exec(ctx, insertQuery,
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
			time.Now().UTC(),
		)
		if err != nil {
			log.Printf("Warning: Failed to persist risk decision into postgres: %v", err)
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
		EvaluatedAt:        time.Now().UTC().Format(time.RFC3339),
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
