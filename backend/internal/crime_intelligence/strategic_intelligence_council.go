package crime_intelligence

import (
	"fmt"
	"time"
)

// StrategicEcosystemDirective represents the top-level macro directive issued by the AI Strategic Council.
type StrategicEcosystemDirective struct {
	DirectiveID              string    `json:"directive_id"`
	EcosystemThreatLevel     string    `json:"ecosystem_threat_level"` // "DEFCON_1_CRITICAL", "DEFCON_2_ELEVATED", "DEFCON_3_NORMAL"
	ConsortiumPriorityActions []string  `json:"consortium_priority_actions"`
	TargetSyndicateFocus     string    `json:"target_syndicate_focus"`
	RecommendedBudgetUSD     float64   `json:"recommended_budget_usd"`
	StrategicRationale       string    `json:"strategic_rationale"`
	IssuedAt                 time.Time `json:"issued_at"`
}

// StrategicIntelligenceCouncil synthesizes multi-disciplinary AI assessments for macro threat posture.
type StrategicIntelligenceCouncil struct {
	analystAgent *AICrimeAnalystAgent
}

// NewStrategicIntelligenceCouncil initializes the strategic council.
func NewStrategicIntelligenceCouncil(analyst *AICrimeAnalystAgent) *StrategicIntelligenceCouncil {
	if analyst == nil {
		analyst = NewAICrimeAnalystAgent(nil)
	}
	return &StrategicIntelligenceCouncil{analystAgent: analyst}
}

// IssueMacroDirective produces strategic instructions for the global defense network.
func (c *StrategicIntelligenceCouncil) IssueMacroDirective(activeSyndicatesCount int, totalGrossExposure float64) *StrategicEcosystemDirective {
	now := time.Now().UTC()

	level := "DEFCON_3_NORMAL"
	var actions []string
	budget := 50000.0

	if activeSyndicatesCount >= 10 || totalGrossExposure >= 1000000.0 {
		level = "DEFCON_1_CRITICAL"
		actions = []string{
			"Activate global cross-bank residential proxy subnet isolation",
			"Mandate hardware-attestation MFA for high-value transactions (> $1,000)",
			"Disseminate federated laundering graph signatures to partner payment rails",
		}
		budget = 250000.0
	} else if activeSyndicatesCount >= 3 {
		level = "DEFCON_2_ELEVATED"
		actions = []string{
			"Increase velocity sampling frequency across carding hot-spots",
			"Enable synthetic identity clustering on new account openings",
		}
		budget = 100000.0
	} else {
		actions = []string{"Maintain standard baseline threat surveillance"}
	}

	rationale := fmt.Sprintf("Council issued %s based on %d active monitored syndicates and $%.2f gross network exposure", level, activeSyndicatesCount, totalGrossExposure)

	return &StrategicEcosystemDirective{
		DirectiveID:              fmt.Sprintf("dir_%d", now.UnixNano()),
		EcosystemThreatLevel:     level,
		ConsortiumPriorityActions: actions,
		TargetSyndicateFocus:     "Transnational Carding & Mule Laundering Networks",
		RecommendedBudgetUSD:     budget,
		StrategicRationale:       rationale,
		IssuedAt:                 now,
	}
}
