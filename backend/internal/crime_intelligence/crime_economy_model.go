package crime_intelligence

import (
	"math"
	"time"
)

// CrimeEconomyAnalysis evaluates the supply chain maturity and economic power of a fraud syndicate.
type CrimeEconomyAnalysis struct {
	SyndicateID             string    `json:"syndicate_id"`
	ThreatMaturityScore     float64   `json:"threat_maturity_score"`     // 0.0 to 1.0 (1.0 = highly mature, vertically integrated)
	ProjectedAnnualExposure float64   `json:"projected_annual_exposure"` // Estimated fraud gross turnover
	ExpansionVelocity       float64   `json:"expansion_velocity"`        // Network growth multiplier
	SupplyChainTiersActive  int       `json:"supply_chain_tiers_active"` // Out of 5 (Attackers, Infra, Data, Laundering, Cashout)
	ClassificationTier      string    `json:"classification_tier"`       // "TIER_1_APG_TRANSNATIONAL", "TIER_2_ORGANIZED", "TIER_3_OPPORTUNISTIC"
	EvaluatedAt             time.Time `json:"evaluated_at"`
}

// FinancialCrimeEcosystemModel analyzes crime syndicate supply chain structures and financial velocity.
type FinancialCrimeEcosystemModel struct{}

// NewFinancialCrimeEcosystemModel initializes the crime economy analyzer.
func NewFinancialCrimeEcosystemModel() *FinancialCrimeEcosystemModel {
	return &FinancialCrimeEcosystemModel{}
}

// ModelSyndicateEcosystem evaluates an organized crime group's structural sophistication.
func (m *FinancialCrimeEcosystemModel) ModelSyndicateEcosystem(
	syndicateID string,
	botnetSize int,
	launderingAccountsCount int,
	knownCashoutChannels int,
	monthlyVolume float64,
) *CrimeEconomyAnalysis {
	now := time.Now().UTC()

	// Calculate active supply chain tiers
	tiers := 1 // Base attacker tier
	if botnetSize > 100 {
		tiers++ // Infrastructure tier
	}
	if launderingAccountsCount > 5 {
		tiers++ // Laundering network tier
	}
	if knownCashoutChannels > 2 {
		tiers++ // Multi-channel cashout tier
	}
	if monthlyVolume > 500000.0 {
		tiers++ // Enterprise financing tier
	}

	maturity := float64(tiers) / 5.0
	annualExposure := monthlyVolume * 12.0
	expansionVel := math.Min(3.5, 1.0+(float64(launderingAccountsCount)/20.0))

	tierClass := "TIER_3_OPPORTUNISTIC"
	if maturity >= 0.80 {
		tierClass = "TIER_1_APG_TRANSNATIONAL"
	} else if maturity >= 0.50 {
		tierClass = "TIER_2_ORGANIZED"
	}

	return &CrimeEconomyAnalysis{
		SyndicateID:             syndicateID,
		ThreatMaturityScore:     maturity,
		ProjectedAnnualExposure: annualExposure,
		ExpansionVelocity:       expansionVel,
		SupplyChainTiersActive:  tiers,
		ClassificationTier:      tierClass,
		EvaluatedAt:             now,
	}
}
