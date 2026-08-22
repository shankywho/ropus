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

	"github.com/shankywho/ropus/backend/internal/audit"
	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/rules"
	"github.com/shankywho/ropus/backend/internal/utils"
)

// Orchestrator orchestrates the real-time synchronous risk evaluation pipeline.
type Orchestrator struct {
	db                    *pgxpool.Pool
	velocityStore         *features.VelocityStore
	deviceFeatureStore    *features.DeviceFeatureStore
	graphStore            *features.AccountDeviceGraphStore
	paymentTokenStore     *features.PaymentTokenStore
	deviceVelocityStore   *features.DeviceVelocityStore
	deviceReputationStore *features.DeviceReputationStore
	rulesService          *rules.Service
	mlClient              *MLClient
	shadowScorer          *ShadowScorer
	canaryRouter          *CanaryRouter
	driftDetector         *DriftDetector
	retrainingCoordinator *RetrainingCoordinator
	metricsEngine         *MetricsEngine
	sloEngine             *SLOEngine
	kms                   utils.KMS
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
		db:                    db,
		velocityStore:         velocityStore,
		deviceFeatureStore:    nil,
		graphStore:            nil,
		paymentTokenStore:     nil,
		deviceVelocityStore:   nil,
		deviceReputationStore: nil,
		rulesService:          rulesService,
		mlClient:              mlClient,
		shadowScorer:          nil,
		canaryRouter:          nil,
		driftDetector:         nil,
		retrainingCoordinator: nil,
		kms:                   kms,
	}
}

// SetRetrainingCoordinator attaches the RetrainingCoordinator to the Orchestrator.
func (o *Orchestrator) SetRetrainingCoordinator(rc *RetrainingCoordinator) {
	o.retrainingCoordinator = rc
}

// SetDriftDetector attaches the DriftDetector to the Orchestrator.
func (o *Orchestrator) SetDriftDetector(dd *DriftDetector) {
	o.driftDetector = dd
}

// GetDriftDetector returns the attached DriftDetector.
func (o *Orchestrator) GetDriftDetector() *DriftDetector {
	return o.driftDetector
}

// SetCanaryRouter attaches a CanaryRouter for staged production rollout.
func (o *Orchestrator) SetCanaryRouter(cr *CanaryRouter) {
	o.canaryRouter = cr
}

// GetCanaryRouter returns the attached CanaryRouter.
func (o *Orchestrator) GetCanaryRouter() *CanaryRouter {
	return o.canaryRouter
}

// SetShadowScorer attaches an asynchronous ShadowScorer.
func (o *Orchestrator) SetShadowScorer(ss *ShadowScorer) {
	o.shadowScorer = ss
}

// SetDeviceFeatureStore attaches a Redis-backed DeviceFeatureStore.
func (o *Orchestrator) SetDeviceFeatureStore(dfs *features.DeviceFeatureStore) {
	o.deviceFeatureStore = dfs
}

// SetAccountDeviceGraphStore attaches a Redis-backed AccountDeviceGraphStore.
func (o *Orchestrator) SetAccountDeviceGraphStore(gs *features.AccountDeviceGraphStore) {
	o.graphStore = gs
}

// SetPaymentTokenStore attaches a Redis-backed PaymentTokenStore.
func (o *Orchestrator) SetPaymentTokenStore(pts *features.PaymentTokenStore) {
	o.paymentTokenStore = pts
}

// SetDeviceVelocityStore attaches a Redis-backed DeviceVelocityStore.
func (o *Orchestrator) SetDeviceVelocityStore(dvs *features.DeviceVelocityStore) {
	o.deviceVelocityStore = dvs
}

// SetDeviceReputationStore attaches a Redis-backed DeviceReputationStore.
func (o *Orchestrator) SetDeviceReputationStore(drs *features.DeviceReputationStore) {
	o.deviceReputationStore = drs
}

// SetMetricsEngine attaches a MetricsEngine.
func (o *Orchestrator) SetMetricsEngine(me *MetricsEngine) {
	o.metricsEngine = me
}

// SetSLOEngine attaches an SLOEngine.
func (o *Orchestrator) SetSLOEngine(se *SLOEngine) {
	o.sloEngine = se
}

// GetMetricsEngine returns the attached MetricsEngine.
func (o *Orchestrator) GetMetricsEngine() *MetricsEngine {
	return o.metricsEngine
}

// GetSLOEngine returns the attached SLOEngine.
func (o *Orchestrator) GetSLOEngine() *SLOEngine {
	return o.sloEngine
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

	decisionUUID := uuid.New().String()
	decisionID := fmt.Sprintf("dec_%s", decisionUUID)
	snapshotRef := fmt.Sprintf("snap_%s", uuid.New().String()[:10])

	// -------------------------------------------------------------
	// STEP 1: Context Aggregation & Velocity Queries
	// -------------------------------------------------------------
	ip := req.IPAddress
	token := req.PaymentMethod.Token

	// Parse, canonicalize, and validate untrusted device telemetry
	devIdentity := features.ParseDeviceIdentity(tenantID, req.DeviceFingerprint)

	var velocityMetrics *features.VelocityMetrics
	var err error
	if o.velocityStore != nil {
		velocityMetrics, err = o.velocityStore.GetVelocityMetrics(ctx, tenantID, ip, token)
		if err != nil {
			log.Printf("Velocity fetch failed, degrading gracefully: %v", err)
			velocityMetrics = &features.VelocityMetrics{TxnCountIP1h: 0, TxnCountToken24h: 0}
		}
	} else {
		velocityMetrics = &features.VelocityMetrics{TxnCountIP1h: 0, TxnCountToken24h: 0}
	}

	// Point-in-Time Safe: Fetch real-time device features BEFORE recording the current transaction
	var devFeatures *features.DeviceFeatures
	if o.deviceFeatureStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		df, err := o.deviceFeatureStore.GetDeviceFeatures(ctx, tenantID, devIdentity.DeviceID)
		if err != nil {
			log.Printf("Device feature store fetch failed, degrading gracefully: %v", err)
			devFeatures = &features.DeviceFeatures{
				IsDegraded:    true,
				DegradeReason: "DEVICE_FEATURE_STORE_UNAVAILABLE",
			}
		} else {
			devFeatures = df
		}
	} else {
		devFeatures = &features.DeviceFeatures{
			DeviceTxCount1m:         0,
			DeviceTxCount1h:         0,
			DeviceTxCount24h:        0,
			DeviceAmountSum24h:      0,
			DeviceUniqueAccounts24h: 0,
			DeviceUniqueTokens24h:   0,
			DeviceSeenBefore:        0,
			IsDegraded:              false,
		}
	}

	// Point-in-Time Safe: Fetch real-time account/device graph features BEFORE recording the current transaction
	var graphFeatures *features.AccountDeviceGraphFeatures
	if o.graphStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		gf, err := o.graphStore.GetGraphFeatures(ctx, tenantID, devIdentity.DeviceID, req.AccountID)
		if err != nil {
			log.Printf("Graph feature store fetch failed, degrading gracefully: %v", err)
			graphFeatures = &features.AccountDeviceGraphFeatures{
				IsDegraded:    true,
				DegradeReason: "ACCOUNT_DEVICE_GRAPH_UNAVAILABLE",
			}
		} else {
			graphFeatures = gf
		}
	} else {
		graphFeatures = &features.AccountDeviceGraphFeatures{
			MultiAccountSignal: features.MultiAccountNormal,
			IsDegraded:         false,
		}
	}

	// Point-in-Time Safe: Fetch real-time payment token features BEFORE recording the current transaction
	var tokenFeatures *features.PaymentTokenFeatures
	if o.paymentTokenStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		ptf, err := o.paymentTokenStore.GetPaymentTokenFeatures(ctx, tenantID, devIdentity.DeviceID, token)
		if err != nil {
			log.Printf("Payment token feature store fetch failed, degrading gracefully: %v", err)
			tokenFeatures = &features.PaymentTokenFeatures{
				IsDegraded:    true,
				DegradeReason: "PAYMENT_TOKEN_FEATURE_STORE_UNAVAILABLE",
			}
		} else {
			tokenFeatures = ptf
		}
	} else {
		tokenFeatures = &features.PaymentTokenFeatures{
			CardTestingSignal: features.CardTestingNormal,
			TokenFanOutSignal: features.TokenFanOutNormal,
			IsDegraded:        false,
		}
	}

	// Point-in-Time Safe: Fetch real-time multi-window velocity features BEFORE recording the current transaction
	var velFeatures *features.DeviceVelocityFeatures
	if o.deviceVelocityStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		vf, err := o.deviceVelocityStore.GetVelocityFeatures(ctx, tenantID, devIdentity.DeviceID)
		if err != nil {
			log.Printf("Device velocity store fetch failed, degrading gracefully: %v", err)
			velFeatures = &features.DeviceVelocityFeatures{
				IsDegraded:    true,
				DegradeReason: "VELOCITY_FEATURE_STORE_UNAVAILABLE",
			}
		} else {
			velFeatures = vf
		}
	} else {
		velFeatures = &features.DeviceVelocityFeatures{
			VelocitySignal: features.VelocityNormal,
			IsDegraded:     false,
		}
	}

	// Point-in-Time Safe: Fetch real-time device reputation features BEFORE recording the current transaction
	var repFeatures *features.DeviceReputationFeatures
	if o.deviceReputationStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		rf, err := o.deviceReputationStore.GetReputationFeatures(ctx, tenantID, devIdentity.DeviceID)
		if err != nil {
			log.Printf("Device reputation store fetch failed, degrading gracefully: %v", err)
			repFeatures = &features.DeviceReputationFeatures{
				DeviceReputationScore: 0.50,
				IsDegraded:            true,
				DegradeReason:         "DEVICE_REPUTATION_FEATURE_STORE_UNAVAILABLE",
			}
		} else {
			repFeatures = rf
		}
	} else {
		repFeatures = &features.DeviceReputationFeatures{
			DeviceReputationScore: 0.50,
			IsDegraded:            false,
		}
	}

	// Point-in-Time Safe: Construct Canonical 25-Feature ML Vector & Legacy 15-Feature Adapter
	canonical25Vector := BuildCanonical25FeatureVector(
		req.Amount,
		&devIdentity,
		devFeatures,
		tokenFeatures,
		velFeatures,
		repFeatures,
		velocityMetrics,
		startTime,
	)
	legacy15Vector := ExtractLegacy15FeatureVector(canonical25Vector)
	if o.driftDetector != nil { o.driftDetector.IngestVector(canonical25Vector.FeatureMap) }

	// Build the in-memory evaluation context for real-time inference & rules
	evalContext := map[string]interface{}{
		"transaction_id":      req.TransactionID,
		"account_id":          req.AccountID,
		"amount":              req.Amount,
		"currency":            req.Currency,
		"device_fingerprint":  devIdentity.CanonicalFingerprint,
		"device_id":           devIdentity.DeviceID,
		"device_status":       string(devIdentity.Status),
		"ip_address":          ip,
		"ml_feature_contract": map[string]interface{}{
			"canonical_version": MLFeatureContractV25,
			"legacy_version":    MLFeatureContractV15,
			"canonical_25":      canonical25Vector.FeatureMap,
			"legacy_15":         legacy15Vector.FeatureMap,
		},
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
		"device_features": map[string]interface{}{
			"device_tx_count_1m":         devFeatures.DeviceTxCount1m,
			"device_tx_count_1h":         devFeatures.DeviceTxCount1h,
			"device_tx_count_24h":        devFeatures.DeviceTxCount24h,
			"device_amount_sum_24h":      devFeatures.DeviceAmountSum24h,
			"device_unique_accounts_24h": devFeatures.DeviceUniqueAccounts24h,
			"device_unique_tokens_24h":   devFeatures.DeviceUniqueTokens24h,
			"device_seen_before":         devFeatures.DeviceSeenBefore,
		},
		"account_device_features": map[string]interface{}{
			"device_unique_accounts_1h":          graphFeatures.DeviceUniqueAccounts1h,
			"device_unique_accounts_24h":         graphFeatures.DeviceUniqueAccounts24h,
			"device_account_switches_1h":         graphFeatures.DeviceAccountSwitches1h,
			"device_account_switches_24h":        graphFeatures.DeviceAccountSwitches24h,
			"device_new_account_on_known_device": graphFeatures.DeviceNewAccountOnKnownDevice,
			"account_unique_devices_1h":          graphFeatures.AccountUniqueDevices1h,
			"account_unique_devices_24h":         graphFeatures.AccountUniqueDevices24h,
			"account_new_device_1h":              graphFeatures.AccountNewDevice1h,
			"account_device_switches_24h":        graphFeatures.AccountDeviceSwitches24h,
			"device_account_seen_before":         graphFeatures.DeviceAccountSeenBefore,
			"multi_account_signal":               string(graphFeatures.MultiAccountSignal),
		},
		"payment_token_features": map[string]interface{}{
			"device_unique_tokens_5m":     tokenFeatures.DeviceUniqueTokens5m,
			"device_unique_tokens_1h":     tokenFeatures.DeviceUniqueTokens1h,
			"device_unique_tokens_24h":    tokenFeatures.DeviceUniqueTokens24h,
			"device_token_tx_count_5m":    tokenFeatures.DeviceTokenTxCount5m,
			"device_token_tx_count_1h":    tokenFeatures.DeviceTokenTxCount1h,
			"device_token_tx_count_24h":   tokenFeatures.DeviceTokenTxCount24h,
			"device_token_amount_sum_24h": tokenFeatures.DeviceTokenAmountSum24h,
			"card_testing_signal":         string(tokenFeatures.CardTestingSignal),
			"token_unique_devices_1h":     tokenFeatures.TokenUniqueDevices1h,
			"token_unique_devices_24h":    tokenFeatures.TokenUniqueDevices24h,
			"token_tx_count_1h":           tokenFeatures.TokenTxCount1h,
			"token_tx_count_24h":          tokenFeatures.TokenTxCount24h,
			"token_fan_out_signal":        string(tokenFeatures.TokenFanOutSignal),
			"device_token_seen_before":    tokenFeatures.DeviceTokenSeenBefore,
		},
		"device_velocity_features": map[string]interface{}{
			"device_tx_count_10s":                 velFeatures.DeviceTxCount10s,
			"device_tx_count_1m":                  velFeatures.DeviceTxCount1m,
			"device_tx_count_5m":                  velFeatures.DeviceTxCount5m,
			"device_tx_count_15m":                 velFeatures.DeviceTxCount15m,
			"device_tx_count_1h":                  velFeatures.DeviceTxCount1h,
			"device_tx_count_6h":                  velFeatures.DeviceTxCount6h,
			"device_tx_count_24h":                 velFeatures.DeviceTxCount24h,
			"device_amount_sum_10s":               velFeatures.DeviceAmountSum10s,
			"device_amount_sum_1m":                velFeatures.DeviceAmountSum1m,
			"device_amount_sum_5m":                velFeatures.DeviceAmountSum5m,
			"device_amount_sum_15m":               velFeatures.DeviceAmountSum15m,
			"device_amount_sum_1h":                velFeatures.DeviceAmountSum1h,
			"device_amount_sum_6h":                velFeatures.DeviceAmountSum6h,
			"device_amount_sum_24h":               velFeatures.DeviceAmountSum24h,
			"device_avg_amount_1m":                velFeatures.DeviceAvgAmount1m,
			"device_avg_amount_5m":                velFeatures.DeviceAvgAmount5m,
			"device_avg_amount_1h":                velFeatures.DeviceAvgAmount1h,
			"device_avg_amount_24h":               velFeatures.DeviceAvgAmount24h,
			"device_max_amount_1h":                velFeatures.DeviceMaxAmount1h,
			"device_max_amount_24h":               velFeatures.DeviceMaxAmount24h,
			"device_tx_rate_10s":                  velFeatures.DeviceTxRate10s,
			"device_tx_rate_1m":                   velFeatures.DeviceTxRate1m,
			"device_tx_rate_5m":                   velFeatures.DeviceTxRate5m,
			"device_tx_rate_15m":                  velFeatures.DeviceTxRate15m,
			"device_tx_rate_1h":                   velFeatures.DeviceTxRate1h,
			"tx_acceleration_1m_15m":              velFeatures.TxAcceleration1m15m,
			"tx_acceleration_5m_1h":               velFeatures.TxAcceleration5m1h,
			"tx_acceleration_15m_1h":              velFeatures.TxAcceleration15m1h,
			"amount_acceleration_5m_1h":           velFeatures.AmountAcceleration5m1h,
			"amount_acceleration_15m_1h":          velFeatures.AmountAcceleration15m1h,
			"device_amount_concentration_5m_1h":   velFeatures.DeviceAmountConcentration5m1h,
			"device_amount_concentration_15m_24h": velFeatures.DeviceAmountConcentration15m24h,
			"velocity_signal":                     string(velFeatures.VelocitySignal),
		},
		"device_reputation_features": map[string]interface{}{
			"device_total_transactions":      repFeatures.DeviceTotalTransactions,
			"device_successful_transactions": repFeatures.DeviceSuccessfulTransactions,
			"device_failed_transactions":     repFeatures.DeviceFailedTransactions,
			"device_disputed_transactions":   repFeatures.DeviceDisputedTransactions,
			"device_fraud_transactions":      repFeatures.DeviceFraudTransactions,
			"device_refunded_transactions":   repFeatures.DeviceRefundedTransactions,
			"device_chargeback_count":        repFeatures.DeviceChargebackCount,
			"device_dispute_rate":            repFeatures.DeviceDisputeRate,
			"device_fraud_rate":              repFeatures.DeviceFraudRate,
			"device_refund_rate":             repFeatures.DeviceRefundRate,
			"device_success_rate":            repFeatures.DeviceSuccessRate,
			"device_recent_dispute_count":    repFeatures.DeviceRecentDisputeCount,
			"device_recent_fraud_count":      repFeatures.DeviceRecentFraudCount,
			"device_recent_chargeback_count": repFeatures.DeviceRecentChargebackCount,
			"device_days_since_first_seen":   repFeatures.DeviceDaysSinceFirstSeen,
			"device_days_since_last_dispute": repFeatures.DeviceDaysSinceLastDispute,
			"device_days_since_last_fraud":   repFeatures.DeviceDaysSinceLastFraud,
			"device_reputation_score":        repFeatures.DeviceReputationScore,
		},
		"device_feature_store": map[string]interface{}{
			"source":   "redis",
			"degraded": devFeatures.IsDegraded || graphFeatures.IsDegraded || tokenFeatures.IsDegraded || velFeatures.IsDegraded || repFeatures.IsDegraded,
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

	// Record telemetry warnings if client provided malformed/oversized device data
	if devIdentity.Status == features.DeviceStatusOversized || devIdentity.Status == features.DeviceStatusInvalid {
		reasonCodes = append(reasonCodes, "INVALID_DEVICE_TELEMETRY")
	}

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
	if devIdentity.Status == features.DeviceStatusOversized || devIdentity.Status == features.DeviceStatusInvalid {
		isDegraded = true
	}
	if devFeatures.IsDegraded {
		isDegraded = true
		reasonCodes = append(reasonCodes, "DEVICE_FEATURE_STORE_UNAVAILABLE")
	}
	if graphFeatures.IsDegraded {
		isDegraded = true
		reasonCodes = append(reasonCodes, "ACCOUNT_DEVICE_GRAPH_UNAVAILABLE")
	}
	if tokenFeatures.IsDegraded {
		isDegraded = true
		reasonCodes = append(reasonCodes, "PAYMENT_TOKEN_FEATURE_STORE_UNAVAILABLE")
	}
	if velFeatures.IsDegraded {
		isDegraded = true
		reasonCodes = append(reasonCodes, "VELOCITY_FEATURE_STORE_UNAVAILABLE")
	}

	// Graph Risk Signals & Reason Codes (Point-in-Time Safe)
	if req.AccountID != "" && graphFeatures.DeviceAccountSeenBefore == 0 {
		reasonCodes = append(reasonCodes, "ACCOUNT_NEW_TO_DEVICE")
	}
	if graphFeatures.DeviceNewAccountOnKnownDevice == 1 {
		reasonCodes = append(reasonCodes, "ACCOUNT_NEW_TO_ESTABLISHED_DEVICE")
	}
	if graphFeatures.DeviceUniqueAccounts24h >= 3 {
		reasonCodes = append(reasonCodes, "DEVICE_MULTI_ACCOUNT_ACTIVITY")
	}
	if graphFeatures.DeviceAccountSwitches1h >= 3 || graphFeatures.DeviceAccountSwitches24h >= 5 {
		reasonCodes = append(reasonCodes, "DEVICE_ACCOUNT_SWITCH_BURST")
	}
	if graphFeatures.AccountUniqueDevices24h >= 3 {
		reasonCodes = append(reasonCodes, "ACCOUNT_MULTI_DEVICE_ACTIVITY")
	}
	if graphFeatures.AccountUniqueDevices24h >= 5 {
		reasonCodes = append(reasonCodes, "ACCOUNT_DEVICE_FANOUT")
	}

	// Payment Token Risk Signals & Reason Codes (Point-in-Time Safe)
	if token != "" && tokenFeatures.DeviceTokenSeenBefore == 0 {
		reasonCodes = append(reasonCodes, "DEVICE_TOKEN_NEW_RELATIONSHIP")
	}
	if tokenFeatures.CardTestingSignal == features.CardTestingHighSignal || tokenFeatures.CardTestingSignal == features.CardTestingSuspicious {
		reasonCodes = append(reasonCodes, "CARD_TESTING_DEVICE_TOKEN_BURST")
	}
	if tokenFeatures.DeviceUniqueTokens24h >= 3 {
		reasonCodes = append(reasonCodes, "DEVICE_HIGH_TOKEN_DIVERSITY")
	}
	if tokenFeatures.DeviceUniqueTokens5m >= 3 {
		reasonCodes = append(reasonCodes, "DEVICE_RAPID_TOKEN_ROTATION")
	}
	if tokenFeatures.TokenUniqueDevices24h >= 3 {
		reasonCodes = append(reasonCodes, "TOKEN_MULTI_DEVICE_ACTIVITY")
	}
	if tokenFeatures.TokenFanOutSignal == features.TokenFanOutHighSignal || tokenFeatures.TokenFanOutSignal == features.TokenFanOutSuspicious {
		reasonCodes = append(reasonCodes, "TOKEN_DEVICE_FANOUT")
	}

	// Multi-Window Velocity & Anomaly Reason Codes (Point-in-Time Safe)
	if velFeatures.DeviceTxCount10s >= features.GlobalVelocityThresholds.Tx10sSuspicious {
		reasonCodes = append(reasonCodes, "DEVICE_TX_BURST_10S")
	}
	if velFeatures.DeviceTxCount1m >= features.GlobalVelocityThresholds.Tx1mSuspicious {
		reasonCodes = append(reasonCodes, "DEVICE_TX_BURST_1M")
	}
	if velFeatures.DeviceTxCount5m >= features.GlobalVelocityThresholds.Tx5mSuspicious {
		reasonCodes = append(reasonCodes, "DEVICE_TX_BURST_5M")
	}
	if velFeatures.DeviceAmountSum5m >= features.GlobalVelocityThresholds.Amount5mSuspicious {
		reasonCodes = append(reasonCodes, "DEVICE_AMOUNT_BURST_5M")
	}
	if velFeatures.DeviceAmountSum15m >= features.GlobalVelocityThresholds.Amount5mHigh {
		reasonCodes = append(reasonCodes, "DEVICE_AMOUNT_BURST_15M")
	}
	if velFeatures.TxAcceleration5m1h >= features.GlobalVelocityThresholds.AccelerationSuspicious || velFeatures.TxAcceleration1m15m >= 10.0 {
		reasonCodes = append(reasonCodes, "DEVICE_VELOCITY_ACCELERATION")
	}
	if velFeatures.VelocitySignal == features.VelocityHighSignal {
		reasonCodes = append(reasonCodes, "DEVICE_EXTREME_VELOCITY")
	}
	if velFeatures.DeviceAmountConcentration5m1h >= 0.8 && velFeatures.DeviceAmountSum1h >= features.GlobalVelocityThresholds.Amount5mLow {
		reasonCodes = append(reasonCodes, "DEVICE_AMOUNT_CONCENTRATION")
	}
	if velFeatures.DeviceTxRate10s >= 0.4 {
		reasonCodes = append(reasonCodes, "DEVICE_HIGH_FREQUENCY_ACTIVITY")
	}

	// Cross-Entity Burst Correlations (Point-in-Time Safe)
	if devFeatures.DeviceUniqueAccounts24h >= 2 && devFeatures.DeviceUniqueTokens24h >= 2 && velFeatures.DeviceTxCount5m >= 3 {
		reasonCodes = append(reasonCodes, "COORDINATED_ACCOUNT_TOKEN_BURST")
	}
	if devFeatures.DeviceUniqueAccounts24h >= 3 && velFeatures.DeviceTxCount1h >= 5 {
		reasonCodes = append(reasonCodes, "MULTI_ACCOUNT_HIGH_VELOCITY")
	}
	if devFeatures.DeviceUniqueTokens24h >= 3 && velFeatures.DeviceTxCount1h >= 5 {
		reasonCodes = append(reasonCodes, "MULTI_TOKEN_HIGH_VELOCITY")
	}
	if tokenFeatures.TokenUniqueDevices24h >= 3 && velFeatures.DeviceTxCount1h >= 5 {
		reasonCodes = append(reasonCodes, "TOKEN_FANOUT_HIGH_VELOCITY")
	}

	if repFeatures.IsDegraded {
		isDegraded = true
		reasonCodes = append(reasonCodes, "DEVICE_REPUTATION_FEATURE_STORE_UNAVAILABLE")
	}

	// Device Reputation & Dispute History Reason Codes (Point-in-Time Safe)
	if repFeatures.DeviceFraudTransactions >= 1 {
		reasonCodes = append(reasonCodes, "DEVICE_FRAUD_HISTORY")
	}
	if repFeatures.DeviceRecentFraudCount >= 1 {
		reasonCodes = append(reasonCodes, "DEVICE_RECENT_FRAUD_ACTIVITY")
	}
	if repFeatures.DeviceReputationScore >= 0.75 {
		reasonCodes = append(reasonCodes, "DEVICE_BAD_REPUTATION")
	}
	if repFeatures.DeviceFraudRate >= 0.05 && repFeatures.DeviceTotalTransactions >= 3 {
		reasonCodes = append(reasonCodes, "DEVICE_HIGH_FRAUD_RATE")
	}
	if repFeatures.DeviceDisputeRate >= 0.10 && repFeatures.DeviceTotalTransactions >= 3 {
		reasonCodes = append(reasonCodes, "DEVICE_HIGH_DISPUTE_RATE")
	}
	if repFeatures.DeviceRecentDisputeCount >= 2 {
		reasonCodes = append(reasonCodes, "DEVICE_RECENT_DISPUTE_BURST")
	}
	if repFeatures.DeviceDisputedTransactions >= 1 && repFeatures.DeviceRecentDisputeCount < 2 {
		reasonCodes = append(reasonCodes, "DEVICE_DISPUTE_HISTORY")
	}
	if repFeatures.DeviceDaysSinceFirstSeen >= 14 && repFeatures.DeviceSuccessfulTransactions >= 5 && repFeatures.DeviceDisputedTransactions == 0 && repFeatures.DeviceFraudTransactions == 0 && repFeatures.DeviceReputationScore <= 0.20 {
		reasonCodes = append(reasonCodes, "DEVICE_LONG_TRUSTED_HISTORY")
	}

	// -------------------------------------------------------------
	// STEP 3: ML Inference & Staged Canary Routing (if pre-rules did not halt pipeline)
	// -------------------------------------------------------------
	modelRoute := RouteLegacy
	if o.canaryRouter != nil {
		modelRoute = o.canaryRouter.Route(tenantID, req.TransactionID)
	}
	evalContext["model_route"] = modelRoute.String()

	if !preRuleTriggered {
		if modelRoute == RouteCandidate && o.mlClient != nil {
			o.canaryRouter.RecordCandidateRequest()
			candReq := MLShadowPredictRequest{
				FeaturesDict:           canonical25Vector.FeatureMap,
				EvaluationID:           decisionID,
				TenantID:               tenantID,
				TransactionID:          req.TransactionID,
				FeatureContractVersion: MLFeatureContractV25,
			}

			candStart := time.Now()
			candResp, candErr := o.mlClient.PredictShadow(ctx, candReq)
			candLatencyMs := float64(time.Since(candStart).Microseconds()) / 1000.0

			if candErr != nil {
				// FAIL-SAFE: Automatic fallback to legacy 15F model on candidate error
				log.Printf("Canary candidate inference failed (%v). Falling back to legacy 15F model.", candErr)
				o.canaryRouter.RecordCandidateFallback(candLatencyMs, candErr.Error())
				reasonCodes = append(reasonCodes, "CANARY_FALLBACK_TO_LEGACY")

				if o.canaryRouter != nil {
					o.canaryRouter.LogEvaluation(ctx, audit.CanaryRolloutEvaluation{
						EvaluationID:           decisionID,
						TenantID:               tenantID,
						TransactionID:          req.TransactionID,
						Timestamp:              time.Now().UTC(),
						ModelRoute:             modelRoute.String(),
						ProductionModelVersion: "xgb-ieee-canonical-v2-calibrated",
						CandidateModelVersion:  "fraud-xgb-25f-candidate-v1",
						CandidateLatencyMs:     candLatencyMs,
						FallbackUsed:           1,
						Error:                  candErr.Error(),
					})
				}

				// Legacy model fallback
				legacyResp, legErr := o.callLegacyML(ctx, legacy15Vector)
				if legErr != nil {
					log.Printf("ML inference degraded (%v). Falling back to rules/heuristics.", legErr)
					isDegraded = true
					riskScore = o.calculateFallbackRiskScore(req.Amount, velocityMetrics)
					reasonCodes = append(reasonCodes, "ML_SERVICE_DEGRADED")
				} else {
					riskScore = legacyResp.RiskScore
					if len(legacyResp.ReasonCodes) > 0 {
						reasonCodes = append(reasonCodes, legacyResp.ReasonCodes...)
					}
				}
			} else {
				// Candidate inference success
				riskScore = candResp.RiskScore
				o.canaryRouter.RecordCandidateSuccess(candLatencyMs, candResp.ShadowDecision)
				reasonCodes = append(reasonCodes, "MODEL_SIGNAL:CANDIDATE_25F_ACTIVE")

				if o.canaryRouter != nil {
					o.canaryRouter.LogEvaluation(ctx, audit.CanaryRolloutEvaluation{
						EvaluationID:           decisionID,
						TenantID:               tenantID,
						TransactionID:          req.TransactionID,
						Timestamp:              time.Now().UTC(),
						ModelRoute:             modelRoute.String(),
						ProductionModelVersion: "xgb-ieee-canonical-v2-calibrated",
						CandidateModelVersion:  candResp.ModelVersion,
						CandidateScore:         candResp.CalibratedProbability,
						CandidateDecision:      candResp.ShadowDecision,
						CandidateLatencyMs:     candLatencyMs,
						FallbackUsed:           0,
						Error:                  "",
					})
				}
			}
		} else {
			// Legacy Route (default 100% path)
			if o.canaryRouter != nil {
				o.canaryRouter.RecordLegacyRequest()
			}
			mlResp, err := o.callLegacyML(ctx, legacy15Vector)
			if err != nil {
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
	encryptedDeviceFP := devIdentity.CanonicalFingerprint
	if len(tenantKey) > 0 {
		if encIP, err := utils.EncryptString(ip, tenantKey); err == nil {
			encryptedIP = encIP
		}
		if devIdentity.CanonicalFingerprint != "" {
			if encFP, err := utils.EncryptString(devIdentity.CanonicalFingerprint, tenantKey); err == nil {
				encryptedDeviceFP = encFP
			}
		}
	}

	// 3. Prepare Encrypted Feature Snapshot (Stored in Postgres)
	encryptedFeatureSnapshot := make(map[string]interface{}, len(evalContext))
	for k, v := range evalContext {
		encryptedFeatureSnapshot[k] = v
	}
	encryptedFeatureSnapshot["ip_address"] = encryptedIP
	encryptedFeatureSnapshot["device_fingerprint"] = encryptedDeviceFP
	encryptedFeatureSnapshot["device_id"] = devIdentity.DeviceID
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
	outboxSnapshot["device_id"] = devIdentity.DeviceID

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

	// -------------------------------------------------------------
	// STEP 5: Post-Scoring State Update (Point-in-Time Safe)
	// -------------------------------------------------------------
	// Update Redis feature stores with the current transaction AFTER evaluation
	if o.velocityStore != nil {
		_ = o.velocityStore.RecordEvent(ctx, tenantID, ip, token, req.Amount)
	}
	if o.deviceFeatureStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		_ = o.deviceFeatureStore.RecordDeviceTransaction(ctx, tenantID, devIdentity.DeviceID, req.TransactionID, req.Amount, req.AccountID, token)
	}
	if o.graphStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		_ = o.graphStore.RecordGraphTransaction(ctx, tenantID, devIdentity.DeviceID, req.AccountID, req.TransactionID)
	}
	if o.paymentTokenStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		_ = o.paymentTokenStore.RecordPaymentTokenTransaction(ctx, tenantID, devIdentity.DeviceID, token, req.TransactionID, req.Amount)
	}
	if o.deviceVelocityStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		_ = o.deviceVelocityStore.RecordVelocityTransaction(ctx, tenantID, devIdentity.DeviceID, req.TransactionID, req.Amount)
	}
	if o.deviceReputationStore != nil && devIdentity.IsValid && devIdentity.DeviceID != "" {
		_ = o.deviceReputationStore.RecordOutcome(ctx, tenantID, devIdentity.DeviceID, req.TransactionID, features.OutcomeSuccess)
	}

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
				decisionUUID,
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

			// Persist device record and record event in durable relational ledger
			if devIdentity.IsValid && devIdentity.DeviceID != "" {
				var deviceUUID string
				upsertDeviceQuery := `
					INSERT INTO devices (
						tenant_id, device_hash, encrypted_fingerprint, first_seen_at, last_seen_at,
						total_tx_count, unique_account_count, trust_score, created_at, updated_at
					)
					VALUES ($1, $2, $3, NOW(), NOW(), 1, 1, 50, NOW(), NOW())
					ON CONFLICT (tenant_id, device_hash) DO UPDATE
					SET last_seen_at = NOW(),
					    total_tx_count = devices.total_tx_count + 1,
					    encrypted_fingerprint = COALESCE(EXCLUDED.encrypted_fingerprint, devices.encrypted_fingerprint),
					    updated_at = NOW()
					RETURNING device_id
				`
				devErr := tx.QueryRow(ctx, upsertDeviceQuery, tenantID, devIdentity.DeviceID, encryptedDeviceFP).Scan(&deviceUUID)
				if devErr == nil && deviceUUID != "" {
					if req.AccountID != "" {
						linkAccountQuery := `
							INSERT INTO device_accounts (
								tenant_id, device_id, account_id, first_seen_at, last_seen_at, transaction_count, created_at, updated_at
							)
							VALUES ($1, $2, $3, NOW(), NOW(), 1, NOW(), NOW())
							ON CONFLICT (tenant_id, device_id, account_id) DO UPDATE
							SET last_seen_at = NOW(),
							    transaction_count = device_accounts.transaction_count + 1,
							    updated_at = NOW()
						`
						_, _ = tx.Exec(ctx, linkAccountQuery, tenantID, deviceUUID, req.AccountID)
					}

					if token != "" {
						linkTokenQuery := `
							INSERT INTO device_payment_instruments (
								tenant_id, device_id, payment_token, first_seen_at, last_seen_at, transaction_count, created_at, updated_at
							)
							VALUES ($1, $2, $3, NOW(), NOW(), 1, NOW(), NOW())
							ON CONFLICT (tenant_id, device_id, payment_token) DO UPDATE
							SET last_seen_at = NOW(),
							    transaction_count = device_payment_instruments.transaction_count + 1,
							    updated_at = NOW()
						`
						_, _ = tx.Exec(ctx, linkTokenQuery, tenantID, deviceUUID, token)
					}

					eventMetadataBytes, _ := json.Marshal(map[string]interface{}{
						"risk_score": riskScore,
						"action":     finalAction,
					})
					eventQuery := `
						INSERT INTO device_events (
							tenant_id, device_id, event_type, account_id, payment_token,
							event_time, amount, currency, risk_decision_id, metadata, created_at
						)
						VALUES ($1, $2, 'TRANSACTION', $3, $4, NOW(), $5, $6, $7, $8, NOW())
					`
					_, _ = tx.Exec(ctx, eventQuery, tenantID, deviceUUID, req.AccountID, token, req.Amount, req.Currency, decisionUUID, eventMetadataBytes)
				}
			}

			if commitErr := tx.Commit(ctx); commitErr != nil {
				log.Printf("Error committing risk decision transaction: %v", commitErr)
			}
		}
	}

	// -------------------------------------------------------------
	// STEP 5.5: Asynchronous Shadow Scoring (Candidate 25F Model)
	// -------------------------------------------------------------
	if o.shadowScorer != nil {
		prodProb := float64(riskScore) / 100.0
		o.shadowScorer.Enqueue(ShadowScoreTask{
			EvaluationID:              decisionID,
			TenantID:                  tenantID,
			TransactionID:             req.TransactionID,
			Timestamp:                 nowUTC,
			Amount:                    float64(req.Amount),
			Canonical25Vector:         canonical25Vector,
			ProductionModelVersion:    "xgb-ieee-canonical-v2-calibrated",
			ProductionFeatureContract: MLFeatureContractV15,
			ProductionRawScore:        prodProb,
			ProductionCalibratedScore: prodProb,
			ProductionDecision:        finalAction,
			ProductionLatencyMs:       float64(latencyMs),
		})
	}

	// -------------------------------------------------------------
	// STEP 6: Structured Production Logging & Response Return
	// -------------------------------------------------------------
	// Structured, sensitive-free production log
	cbStateStr := "UNKNOWN"
	gateStatusStr := "UNKNOWN"
	canaryPct := 0
	if o.canaryRouter != nil {
		st := o.canaryRouter.GetStatus()
		if cb, ok := st["circuit_breaker"].(map[string]interface{}); ok {
			if s, ok := cb["state"].(string); ok {
				cbStateStr = s
			}
		}
		if g, ok := st["safety_gate_status"].(string); ok {
			gateStatusStr = g
		}
		if p, ok := st["target_percentage"].(int); ok {
			canaryPct = p
		}
	}

	fallbackUsedNum := 0
	if isDegraded {
		fallbackUsedNum = 1
	}

	// Record telemetry to MetricsEngine and SLOEngine
	if o.metricsEngine != nil {
		o.metricsEngine.RecordRequest(float64(latencyMs), finalAction, riskScore, !isDegraded, isDegraded, false)
	}
	if o.sloEngine != nil {
		o.sloEngine.RecordEvaluation(float64(latencyMs), !isDegraded, isDegraded, false)
	}

	reqID := utils.GetRequestID(ctx)

	structuredLog := map[string]interface{}{
		"event":                 "risk_evaluation_completed",
		"request_id":            reqID,
		"transaction_id":        req.TransactionID,
		"tenant_id":             tenantID,
		"decision_id":           decisionID,
		"model_version":         "fraud-xgb-25f-v3.0",
		"model_route":           modelRoute.String(),
		"decision":              finalAction,
		"risk_score":            riskScore,
		"latency_ms":            latencyMs,
		"fallback_used":         fallbackUsedNum,
		"is_degraded":           isDegraded,
		"canary_percentage":     canaryPct,
		"circuit_breaker_state": cbStateStr,
		"safety_gate_status":    gateStatusStr,
		"timestamp":             nowUTC.Format(time.RFC3339),
	}
	if logJSON, err := json.Marshal(structuredLog); err == nil {
		log.Printf("[RISK_EVALUATION] %s", string(logJSON))
	}

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

// callLegacyML formats and dispatches a 15-feature inference request to the legacy ONNX model.
func (o *Orchestrator) callLegacyML(ctx context.Context, legacy15Vector *MLFeatureVector) (*MLPredictResponse, error) {
	if o.mlClient == nil {
		return nil, fmt.Errorf("mlClient is nil")
	}

	isNewDev := 0
	if legacy15Vector.FeatureMap["device_seen_before"] == 0.0 {
		isNewDev = 1
	}
	ip24h := legacy15Vector.FeatureMap["ip_velocity_24h"]
	hour := int(legacy15Vector.FeatureMap["transaction_hour"])
	day := int(legacy15Vector.FeatureMap["transaction_day"])
	dist1 := int(legacy15Vector.FeatureMap["dist1_missing"])
	devMobile := int(legacy15Vector.FeatureMap["device_type_mobile"])
	devMissing := int(legacy15Vector.FeatureMap["device_info_missing"])
	amtRatio := legacy15Vector.FeatureMap["amount_to_mean_ratio"]
	devSeen := int(legacy15Vector.FeatureMap["device_seen_before"])

	mlReq := MLPredictRequest{
		Amount:                 legacy15Vector.FeatureMap["amount"],
		IPVelocity1h:           legacy15Vector.FeatureMap["ip_velocity_1h"],
		IPVelocity24h:          &ip24h,
		TokenVelocity24h:       legacy15Vector.FeatureMap["token_velocity_24h"],
		IsNewDevice:            isNewDev,
		DeviceSeenBefore:       &devSeen,
		HourOfDay:              &hour,
		DayOfWeek:              &day,
		Dist1Missing:           &dist1,
		DeviceTypeMobile:       &devMobile,
		DeviceInfoMissing:      &devMissing,
		AmountToMeanRatio:      &amtRatio,
		FeatureContractVersion: MLFeatureContractV15,
	}

	return o.mlClient.Predict(ctx, mlReq)
}

