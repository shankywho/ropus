package intelligence_fabric

import (
	"fmt"
)

// IntelligenceSecurityGuard enforces integrity validation and Zero-Trust access to AFC-IOS subsystems.
type IntelligenceSecurityGuard struct {
	AllowedTenants map[string]bool
}

// NewIntelligenceSecurityGuard initializes the intelligence security guard.
func NewIntelligenceSecurityGuard() *IntelligenceSecurityGuard {
	return &IntelligenceSecurityGuard{
		AllowedTenants: map[string]bool{
			"consortium_root": true,
			"partner_bank_01": true,
			"default_tenant":  true,
		},
	}
}

// AuthorizeIngestion checks whether a signal source is authenticated and authorized.
func (g *IntelligenceSecurityGuard) AuthorizeIngestion(tenantID string, reliabilityScore float64) error {
	if !g.AllowedTenants[tenantID] {
		return fmt.Errorf("zero trust violation: tenant '%s' not authorized for intelligence fabric ingestion", tenantID)
	}
	if reliabilityScore < 0.50 {
		return fmt.Errorf("poison protection: signal rejected due to low source reliability (%.2f < 0.50)", reliabilityScore)
	}
	return nil
}

// AuthorizePolicyPromotion verifies security and audit criteria before policy deployment.
func (g *IntelligenceSecurityGuard) AuthorizePolicyPromotion(stage PolicyStage, approverRole string) error {
	if stage == PolicyStageProduction && approverRole != "ROLE_RISK_EXECUTIVE" {
		return fmt.Errorf("governance violation: production policy promotion requires ROLE_RISK_EXECUTIVE (got '%s')", approverRole)
	}
	return nil
}
