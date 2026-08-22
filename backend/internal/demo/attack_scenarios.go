package demo

import (
	"fmt"
	"time"

	"github.com/shankywho/ropus/backend/internal/product_api"
)

// DemoScenarioType defines the live attack simulation stories.
type DemoScenarioType string

const (
	ScenarioSyntheticIdentity DemoScenarioType = "SYNTHETIC_IDENTITY_FRAUD"
	ScenarioAccountTakeover    DemoScenarioType = "ACCOUNT_TAKEOVER"
	ScenarioFraudRingAttack    DemoScenarioType = "FRAUD_RING_ATTACK"
)

// ScenarioExecutionStep represents a discrete phase in the demo narrative.
type ScenarioExecutionStep struct {
	StepIndex   int       `json:"step_index"`
	Phase       string    `json:"phase"` // "ATTACK", "DETECTION", "REASONING", "INVESTIGATION", "RESPONSE", "LEARNING"
	Description string    `json:"description"`
	Details     string    `json:"details"`
	Timestamp   time.Time `json:"timestamp"`
}

// ScenarioRunResult encapsulates the full narrative output of a simulated attack.
type ScenarioRunResult struct {
	ScenarioID       string                     `json:"scenario_id"`
	Type             DemoScenarioType           `json:"type"`
	Title            string                     `json:"title"`
	StorySummary     string                     `json:"story_summary"`
	Steps            []ScenarioExecutionStep    `json:"steps"`
	SampleRequest    product_api.EvaluateRiskRequest  `json:"sample_request"`
	RiskResponse     product_api.EvaluateRiskResponse `json:"risk_response"`
	ExecutedAt       time.Time                  `json:"executed_at"`
}

// DemoOrchestrator coordinates end-to-end interactive demo attacks.
type DemoOrchestrator struct {
	evaluator *product_api.ProductRiskEvaluator
}

// NewDemoOrchestrator initializes the demo simulator.
func NewDemoOrchestrator(evaluator *product_api.ProductRiskEvaluator) *DemoOrchestrator {
	if evaluator == nil {
		evaluator = product_api.NewProductRiskEvaluator()
	}
	return &DemoOrchestrator{evaluator: evaluator}
}

// RunScenario executes a realistic full-lifecycle attack simulation.
func (o *DemoOrchestrator) RunScenario(scenarioType DemoScenarioType) *ScenarioRunResult {
	now := time.Now().UTC()
	scenarioID := fmt.Sprintf("demo_%d_%s", now.UnixNano(), scenarioType)

	var title string
	var summary string
	var steps []ScenarioExecutionStep
	var req product_api.EvaluateRiskRequest

	switch scenarioType {
	case ScenarioSyntheticIdentity:
		title = "Scenario 1: Synthetic Identity Creation & Cashout"
		summary = "Adversary creates multiple synthetic profiles sharing a single emulator fingerprint and attempts high-velocity cashout."
		req = product_api.EvaluateRiskRequest{
			TransactionID: "tx_synth_9981",
			UserID:        "usr_synthetic_bot_01",
			Amount:        14500.0,
			Currency:      "USD",
			Merchant:      "CryptoLiquidityExpress",
			Device: product_api.DeviceDetails{
				DeviceFingerprint: "fp_compromised_mule_cluster",
				IPAddress:         "198.51.100.44",
				UserAgent:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				IsEmulator:        true,
				IsVPN:             true,
			},
			Location: product_api.LocationDetails{
				Country: "CY",
				City:    "Limassol",
			},
		}

		steps = []ScenarioExecutionStep{
			{1, "ATTACK", "Adversary generates synthetic profile using AI face synthesis and connects via VPN", "IP: 198.51.100.44 (Known Bulletproof Proxy)", now},
			{2, "DETECTION", "Risk engine ingests transaction and detects emulator signature & graph cluster link", "Risk Score: 94.0% | Threat Intel Match", now.Add(20 * time.Millisecond)},
			{3, "REASONING", "Agent Council deliberates: Investigator confirms 14 shared entity edges", "Decision: BLOCK | Recommended: Freeze linked accounts", now.Add(50 * time.Millisecond)},
			{4, "INVESTIGATION", "Autonomous Case Manager creates Case #CASE-9981 with evidentiary graph dossier", "Evidence: 6 linked synthetic identities identified", now.Add(100 * time.Millisecond)},
			{5, "RESPONSE", "Payment gateway settlement blocked; account frozen; consortium alert broadcast", "Consortium defense dispatched to 24 partner institutions", now.Add(150 * time.Millisecond)},
			{6, "LEARNING", "Closed-loop feedback adjusts emulator detection weights and creates canary rule", "Rule Rule_SynthCluster_v1 promoted to Shadow evaluation", now.Add(200 * time.Millisecond)},
		}

	case ScenarioAccountTakeover:
		title = "Scenario 2: High-Value Account Takeover with Impossible Travel"
		summary = "Legitimate customer account accessed from remote location within 15 minutes of domestic login."
		req = product_api.EvaluateRiskRequest{
			TransactionID: "tx_ato_4421",
			UserID:        "usr_vip_enterprise_88",
			Amount:        4200.0,
			Currency:      "USD",
			Merchant:      "GlobalWireTransfer",
			Device: product_api.DeviceDetails{
				DeviceFingerprint: "fp_new_unrecognized_browser",
				IPAddress:         "203.0.113.88",
				UserAgent:         "Mozilla/5.0 (Linux; Android 14) Chrome/120.0",
				IsEmulator:        false,
				IsVPN:             true,
			},
			Location: product_api.LocationDetails{
				Country: "NG",
				City:    "Lagos",
			},
		}

		steps = []ScenarioExecutionStep{
			{1, "ATTACK", "Attacker initiates $4,200 wire from unrecognized device in Lagos, Nigeria", "15 minutes after customer was active in New York, US", now},
			{2, "DETECTION", "Impossible travel velocity anomaly triggered (> 5,000 miles/hr velocity)", "Risk Score: 68.0% | Challenge Threshold Exceeded", now.Add(20 * time.Millisecond)},
			{3, "REASONING", "Decision Engine determines Step-Up Multi-Factor Authentication required", "Decision: CHALLENGE | Action: WebAuthn Hardware Push", now.Add(40 * time.Millisecond)},
			{4, "INVESTIGATION", "Security agent initiates proactive session freeze and alerts legitimate owner", "SMS & Email security notice dispatched", now.Add(80 * time.Millisecond)},
			{5, "RESPONSE", "Transaction paused awaiting biometric confirmation", "MFA Prompt dispatched; unauthorized session revoked", now.Add(120 * time.Millisecond)},
			{6, "LEARNING", "Device fingerprint added to suspicious monitoring watchlist", "Velocity model recalibrated for geo-hop telemetry", now.Add(160 * time.Millisecond)},
		}

	case ScenarioFraudRingAttack:
		title = "Scenario 3: Distributed Card Testing & Mule Laundering Ring"
		summary = "Organized syndicate launches coordinated micro-authorization carding attacks across 50 partner merchants."
		req = product_api.EvaluateRiskRequest{
			TransactionID: "tx_ring_7701",
			UserID:        "usr_mule_layer_12",
			Amount:        28500.0,
			Currency:      "USD",
			Merchant:      "PreciousMetalsDirect",
			Device: product_api.DeviceDetails{
				DeviceFingerprint: "fp_compromised_mule_cluster",
				IPAddress:         "198.51.100.44",
				UserAgent:         "Mozilla/5.0 (X11; Linux x86_64)",
				IsEmulator:        true,
				IsVPN:             true,
			},
			Location: product_api.LocationDetails{
				Country: "RO",
				City:    "Bucharest",
			},
		}

		steps = []ScenarioExecutionStep{
			{1, "ATTACK", "Syndicate initiates distributed carding wave across 50 merchants simultaneously", "Aggregated velocity: 120 tx/min across 100 accounts", now},
			{2, "DETECTION", "Streaming fraud engine detects cross-tenant distributed campaign signature", "Risk Score: 98.0% | Campaign Signature: CAMP-PHANTOM-09", now.Add(15 * time.Millisecond)},
			{3, "REASONING", "AI Strategic Council triggers DEFCON 1 CRITICAL macro defense response", "Decision: BLOCK | Defense: Isolate proxy subnet & freeze routing nodes", now.Add(45 * time.Millisecond)},
			{4, "INVESTIGATION", "Forensic Threat Hunter maps full syndicate graph tree (35 nodes, 42 edges)", "Lineage connected to Transnational Carding Syndicate Beta", now.Add(90 * time.Millisecond)},
			{5, "RESPONSE", "Global Response Network synchronizes block across all 24 consortium banks", "Estimated $2,400,000 gross syndicate exposure mitigated", now.Add(130 * time.Millisecond)},
			{6, "LEARNING", "Episodic memory logged; AI Red Team updates mutation defenses", "DSL countermeasure deployed to production risk gateways", now.Add(170 * time.Millisecond)},
		}
	}

	resp := o.evaluator.EvaluateTransaction(&req)

	return &ScenarioRunResult{
		ScenarioID:    scenarioID,
		Type:          scenarioType,
		Title:         title,
		StorySummary:  summary,
		Steps:         steps,
		SampleRequest: req,
		RiskResponse:  *resp,
		ExecutedAt:    now,
	}
}
