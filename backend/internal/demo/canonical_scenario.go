package demo

import (
	"context"
	"fmt"
	"time"

	"github.com/shankywho/ropus/backend/internal/product_api"
)

// CanonicalScenarioRunner executes the 17-stage "Cross-Border Account Takeover -> Synthetic Identity -> Mule Cashout" narrative.
type CanonicalScenarioRunner struct {
	pipeline *product_api.UnifiedRiskPipeline
}

// NewCanonicalScenarioRunner initializes the canonical demo engine.
func NewCanonicalScenarioRunner(pipeline *product_api.UnifiedRiskPipeline) *CanonicalScenarioRunner {
	if pipeline == nil {
		pipeline = product_api.NewUnifiedRiskPipeline(nil, nil, nil)
	}
	return &CanonicalScenarioRunner{pipeline: pipeline}
}

// CanonicalScenarioResult holds the 17-stage demo verification output.
type CanonicalScenarioResult struct {
	ScenarioName   string                          `json:"scenario_name"`
	TotalStages    int                             `json:"total_stages"`
	Stages         []ScenarioExecutionStep         `json:"stages"`
	DecisionResult *product_api.CanonicalRiskResponse `json:"decision_result"`
	ExecutedAt     time.Time                       `json:"executed_at"`
}

// ExecuteCanonicalAttack executes the 17-stage sequence deterministically.
func (r *CanonicalScenarioRunner) ExecuteCanonicalAttack(ctx context.Context, apiKeyToken string) (*CanonicalScenarioResult, error) {
	now := time.Now().UTC()

	stages := []ScenarioExecutionStep{
		{1, "CUSTOMER_HISTORY", "Customer profile usr_sarah_connor has 3-year spotless history with domestic US transactions", "Baseline: 2.1 tx/day, Average amount: $64.50", now},
		{2, "DEVICE_CHANGE", "New login session initiated from unrecognized device identifier", "DeviceID: dev_mule_cluster_99 (No prior association with customer)", now.Add(10 * time.Millisecond)},
		{3, "IMPOSSIBLE_TRAVEL", "Physical distance between consecutive sessions is 7,850 km within 12 minutes", "Origin: Limassol, Cyprus vs Previous: New York, USA", now.Add(20 * time.Millisecond)},
		{4, "IP_REPUTATION", "Network IP matched against known bulletproof proxy / anonymous VPN subnet", "IP: 198.51.100.44 (ISP: Anonymous Hosting Consortium)", now.Add(30 * time.Millisecond)},
		{5, "BEHAVIORAL_ANOMALY", "Immediate high-value wire transfer request ($14,500.00) initiated seconds after credential update", "Amount exceeds 99th percentile customer ceiling", now.Add(40 * time.Millisecond)},
		{6, "SHARED_INFRASTRUCTURE", "Hardware canvas hash matches device fingerprint seen across 14 synthetic identities", "Cluster ID: clus_transnational_mule_88", now.Add(50 * time.Millisecond)},
		{7, "SUSPICIOUS_TRANSACTION", "Wire instruction dispatched to high-risk OTC crypto exchange", "Merchant: CryptoLiquidityExpress (MCC 6051)", now.Add(60 * time.Millisecond)},
		{8, "GRAPH_DISCOVERY", "Graph engine identifies 3-hop cyclic relationship linking beneficiary to known fraud syndicate", "Path: Beneficiary -> MuleAccount_44 -> SyndicateNode_01", now.Add(70 * time.Millisecond)},
		{9, "ML_RISK_INCREASE", "XGBoost inference model evaluates composite features: FraudProbability rises to 0.982", "Feature importance: Velocity (0.24), DeviceEntropy (0.22), GeoDistance (0.21)", now.Add(80 * time.Millisecond)},
		{10, "AI_INVESTIGATOR_ACTIVATION", "Autonomous Investigator Agent triggers multi-agent inquiry into historical graph edges", "Agent Role: Senior Fraud Investigator (Claude 3.7 / GPT-4o)", now.Add(90 * time.Millisecond)},
		{11, "RELATED_ENTITIES_DISCOVERED", "Investigator identifies 6 co-conspirator accounts sharing the same device canvas hash", "Identities: usr_mule_01, usr_mule_02, usr_synth_09", now.Add(100 * time.Millisecond)},
		{12, "RISK_DECISION", "Risk engine issues hard BLOCK verdict (Risk Score: 0.96, Confidence: 0.94)", "Decision: BLOCK | Settlement halted", now.Add(110 * time.Millisecond)},
		{13, "CASE_CREATION", "Automated Case Management creates Priority P0 Investigation Case #CASE-88419", "Case assigned to Senior Financial Crime Analyst Queue", now.Add(120 * time.Millisecond)},
		{14, "ANALYST_EXPLANATION", "System synthesizes plain-English explanation distinguishing observed facts from inferences", "Fact: Impossible travel & bulletproof IP | Inference: Credential stuffing ATO", now.Add(130 * time.Millisecond)},
		{15, "RECOMMENDED_ACTION", "Recommended action: Restrict online banking credentials, hold settlement, notify cardholder", "Action Code: ACT_FREEZE_AND_NOTIFY", now.Add(140 * time.Millisecond)},
		{16, "AUDIT_TRAIL", "All decision parameters and factor weights written to SHA-256 hash-chained audit ledger", "Audit Block: aud_88419_blk_01 (Hash: 4f8a...e91c)", now.Add(150 * time.Millisecond)},
		{17, "CUSTOMER_WEBHOOK", "HMAC-SHA256 signed event 'risk.decision.created' dispatched to bank webhook endpoint", "Webhook Payload: EventID evt_88419 (Status: Delivered 200 OK)", now.Add(160 * time.Millisecond)},
	}

	req := product_api.CanonicalRiskRequest{
		TransactionID: "tx_canonical_ato_88419",
		CustomerID:    "usr_synthetic_bot_01",
		Amount:        14500.0,
		Currency:      "USD",
		MerchantID:    "CryptoLiquidityExpress",
		DeviceID:      "dev_mule_cluster_99",
		IPAddress:     "198.51.100.44",
		Country:       "CY",
		Timestamp:     now,
	}

	res, err := r.pipeline.EvaluateRisk(ctx, apiKeyToken, req)
	if err != nil {
		return nil, fmt.Errorf("canonical scenario risk evaluation failed: %w", err)
	}

	return &CanonicalScenarioResult{
		ScenarioName:   "Cross-Border Account Takeover -> Synthetic Identity -> Mule Cashout",
		TotalStages:    len(stages),
		Stages:         stages,
		DecisionResult: res,
		ExecutedAt:     now,
	}, nil
}
