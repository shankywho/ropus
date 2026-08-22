package attack_simulator

import (
	"fmt"
	"time"
)

// AttackScenarioType categorizes the four primary real-world fraud campaigns.
type AttackScenarioType string

const (
	ScenarioAccountTakeover   AttackScenarioType = "ACCOUNT_TAKEOVER"
	ScenarioSyntheticIdentity AttackScenarioType = "SYNTHETIC_IDENTITY_FRAUD"
	ScenarioCardTesting       AttackScenarioType = "CARD_TESTING_CAMPAIGN"
	ScenarioMoneyLaundering   AttackScenarioType = "MONEY_LAUNDERING_NETWORK"
)

// AttackTimelineEvent documents a chronological milestone in the attack execution.
type AttackTimelineEvent struct {
	SecondOffset int       `json:"second_offset"`
	Timestamp    time.Time `json:"timestamp"`
	Phase        string    `json:"phase"`
	Summary      string    `json:"summary"`
	TechnicalIOC string    `json:"technical_ioc"`
}

// AttackSimulationResult details the end-to-end outcome of an executed campaign.
type AttackSimulationResult struct {
	CampaignID           string                `json:"campaign_id"`
	ScenarioType         AttackScenarioType    `json:"scenario_type"`
	Title                string                `json:"title"`
	TargetVictim         string                `json:"target_victim"`
	AttackerInfrastructure string              `json:"attacker_infrastructure"`
	TotalGrossLossAtRisk float64               `json:"total_gross_loss_at_risk"`
	MitigatedLossUSD     float64               `json:"mitigated_loss_usd"`
	FinalVerdict         string                `json:"final_verdict"` // "AUTONOMOUS_BLOCK", "STEP_UP_CHALLENGE", "CONSORTIUM_FREEZE"
	Confidence           float64               `json:"confidence"`
	Timeline             []AttackTimelineEvent `json:"timeline"`
	InvestigationReport  string                `json:"investigation_report"`
	ExecutedAt           time.Time             `json:"executed_at"`
}

// RealFraudAttackSimulator orchestrates live reproducible multi-vector attack scenarios.
type RealFraudAttackSimulator struct{}

// NewRealFraudAttackSimulator initializes the attack simulator.
func NewRealFraudAttackSimulator() *RealFraudAttackSimulator {
	return &RealFraudAttackSimulator{}
}

// ExecuteScenario executes a cinematic end-to-end fraud attack flow.
func (s *RealFraudAttackSimulator) ExecuteScenario(scenario AttackScenarioType) *AttackSimulationResult {
	now := time.Now().UTC()
	campaignID := fmt.Sprintf("camp_%d_%s", now.UnixNano(), scenario)

	var title string
	var victim string
	var infra string
	var atRisk float64
	var mitigated float64
	var verdict string
	var events []AttackTimelineEvent
	var report string

	switch scenario {
	case ScenarioAccountTakeover:
		title = "Scenario 1: High-Value Account Takeover with Geo-Anomaly"
		victim = "usr_enterprise_vip_8842"
		infra = "Residential VPN (Lagos, NG) + Unrecognized Android Emulator"
		atRisk = 48500.0
		mitigated = 48500.0
		verdict = "STEP_UP_CHALLENGE"
		events = []AttackTimelineEvent{
			{0, now, "CREDENTIAL_THEFT", "Attacker obtains session token from darkweb credential breach", "Breach ID: COMBO_CORP_2026"},
			{5, now.Add(5 * time.Second), "IMPOSSIBLE_TRAVEL", "Login attempted from Lagos, NG 12 minutes after customer was in New York, US", "Geo Velocity: 4,800 miles/hr"},
			{10, now.Add(10 * time.Second), "NEW_DEVICE", "Unrecognized device fingerprint observed", "Device: fp_unseen_android_14"},
			{15, now.Add(15 * time.Second), "BEHAVIORAL_ANOMALY", "High-value wire initiation ($48,500) to external crypto merchant", "Amount exceeds 99.8th percentile"},
			{20, now.Add(20 * time.Second), "AI_INVESTIGATION", "Autonomous Agent Council deliberates; Threat Hunter flags VPN proxy subnet", "Council Consensus: CHALLENGE_MFA"},
			{25, now.Add(25 * time.Second), "ACCOUNT_PROTECTED", "Hardware WebAuthn challenge prompted; unauthorized session terminated", "Loss Prevented: $48,500.00"},
		}
		report = "Forensic Investigation #FIR-ATO-8842: High confidence Account Takeover detected. Velocity anomaly (impossible travel between US and NG) and unrecognized device fingerprint triggered adaptive WebAuthn step-up. Session isolated and customer notified."

	case ScenarioSyntheticIdentity:
		title = "Scenario 2: Transnational Synthetic Identity Ring"
		victim = "Consortium Merchant Rails"
		infra = "Cloned Fingerprint Emulator Farm (AS99812 Bulletproof Subnet)"
		atRisk = 125000.0
		mitigated = 125000.0
		verdict = "AUTONOMOUS_BLOCK"
		events = []AttackTimelineEvent{
			{0, now, "IDENTITY_CREATION", "Attacker provisions 14 synthetic customer accounts using AI face synthesis", "Shared SSN/Tax Prefix Cluster"},
			{5, now.Add(5 * time.Second), "INFRASTRUCTURE_LINK", "Streaming engine detects identical canvas hash and hardware entropy", "Cloned Emulator ID: fp_cloned_root_01"},
			{10, now.Add(10 * time.Second), "GRAPH_RESOLUTION", "Fraud Knowledge Graph maps 14 accounts linked to 1 central device node", "Graph Ring Degree: 14 Nodes"},
			{15, now.Add(15 * time.Second), "ML_SCORING", "XGBoost tree model outputs 0.984 fraud probability", "Risk Score: 98.4% (Critical)"},
			{20, now.Add(20 * time.Second), "AI_INVESTIGATION", "LLM Investigator compiles multi-account evidentiary graph dossier", "Evidence: 14 linked synthetic accounts"},
			{25, now.Add(25 * time.Second), "RING_CONTAINMENT", "Entire 14-account syndicate ring frozen across gateway rails", "Loss Prevented: $125,000.00"},
		}
		report = "Forensic Investigation #FIR-SYNTH-9901: Synthetic Identity Ring confirmed. Graph community clustering identified 14 accounts originating from cloned emulator farm. All entities hard blocked."

	case ScenarioCardTesting:
		title = "Scenario 3: Distributed Card Testing Botnet Storm"
		victim = "Merchant Gateway Ecosystem"
		infra = "Distributed Residential Proxy Pool (250 IPs across 50 Merchants)"
		atRisk = 350000.0
		mitigated = 350000.0
		verdict = "CONSORTIUM_FREEZE"
		events = []AttackTimelineEvent{
			{0, now, "MICRO_AUTHORIZATION", "Botnet initiates $1.00 micro-transactions across 50 partner merchants", "Velocity: 180 auths/sec"},
			{5, now.Add(5 * time.Second), "VELOCITY_SURGE", "Consortium sliding window detects anomalous global card velocity spike", "Spike: 450x baseline rate"},
			{10, now.Add(10 * time.Second), "DISTRIBUTED_DETECTION", "Streaming fraud engine identifies distributed card testing campaign signature", "Campaign: CAMP-CARDING-SURGE-44"},
			{15, now.Add(15 * time.Second), "AI_DECISIONING", "AI Strategic Council deploys dynamic subnet rate limiting rule DSL", "Rule: Rule_BlockRotatingASN_v2"},
			{20, now.Add(20 * time.Second), "NETWORK_DEFENSE", "Global Response Network synchronizes block across all 24 consortium banks", "24 Banks Coordinated"},
			{25, now.Add(25 * time.Second), "ATTACK_NEUTRALIZED", "Carding attack halted with 0 legitimate merchant false declines", "Loss Prevented: $350,000.00"},
		}
		report = "Forensic Investigation #FIR-CARDING-4412: Distributed Carding Campaign neutralized. Dynamic rule DSL deployed to edge gateways within 15 seconds, preventing $350,000 in chargebacks."

	case ScenarioMoneyLaundering:
		title = "Scenario 4: Multi-Hop Circular Money Laundering Ring"
		victim = "Fintech Settlement Accounts"
		infra = "Rapid Mule Chain Layering (3 Hops in 4 Minutes)"
		atRisk = 280000.0
		mitigated = 280000.0
		verdict = "CONSORTIUM_FREEZE"
		events = []AttackTimelineEvent{
			{0, now, "FUNDS_INJECTION", "Initial illicit deposit of $280,000 transferred into Tier-1 mule account", "Account: acc_mule_alpha_01"},
			{5, now.Add(5 * time.Second), "RAPID_LAYERING", "Funds fragmented and rotated across 8 secondary intermediary accounts", "Layering Hops: 8 Nodes in 2 mins"},
			{10, now.Add(10 * time.Second), "CIRCULAR_MOVEMENT", "Money flow analyzer detects circular flow returning to origin entity", "Topology: CIRCULAR_SMURFING"},
			{15, now.Add(15 * time.Second), "GNN_CLASSIFICATION", "Graph Neural Network scores laundering risk at 0.992", "Laundering Score: 99.2%"},
			{20, now.Add(20 * time.Second), "AI_INVESTIGATION", "Investigator Agent synthesizes SAR (Suspicious Activity Report) filing dossier", "Regulatory SAR Package Ready"},
			{25, now.Add(25 * time.Second), "FUNDS_INTERCEPTED", "All 8 mule routing accounts frozen prior to external crypto rail cashout", "Loss Prevented: $280,000.00"},
		}
		report = "Forensic Investigation #FIR-AML-7731: Circular Money Laundering Scheme intercepted. Multi-hop fund flow traced across 8 mule accounts; $280,000 frozen before crypto cashout."
	}

	return &AttackSimulationResult{
		CampaignID:             campaignID,
		ScenarioType:           scenario,
		Title:                  title,
		TargetVictim:           victim,
		AttackerInfrastructure: infra,
		TotalGrossLossAtRisk:   atRisk,
		MitigatedLossUSD:       mitigated,
		FinalVerdict:           verdict,
		Confidence:             0.98,
		Timeline:               events,
		InvestigationReport:    report,
		ExecutedAt:             now,
	}
}
