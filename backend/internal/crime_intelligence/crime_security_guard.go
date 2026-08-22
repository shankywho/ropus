package crime_intelligence

import (
	"fmt"
)

// CrimeSecurityGuard validates intelligence access controls and defends against threat graph poisoning.
type CrimeSecurityGuard struct {
	AuthorizedTenants map[string]bool
}

// NewCrimeSecurityGuard initializes the crime security guard.
func NewCrimeSecurityGuard() *CrimeSecurityGuard {
	return &CrimeSecurityGuard{
		AuthorizedTenants: map[string]bool{
			"consortium_master": true,
			"partner_bank_01":   true,
			"default_tenant":    true,
		},
	}
}

// ValidateIntelligenceAccess ensures only authorized participants access transnational threat intelligence.
func (g *CrimeSecurityGuard) ValidateIntelligenceAccess(tenantID, requestedScope string) error {
	if !g.AuthorizedTenants[tenantID] {
		return fmt.Errorf("security violation: tenant '%s' is not authorized for threat intelligence scope '%s'", tenantID, requestedScope)
	}
	return nil
}

// ValidatePoisonProtection verifies threat signals to prevent adversarial data poisoning.
func (g *CrimeSecurityGuard) ValidatePoisonProtection(confidence float64, reporterReputation float64) error {
	if reporterReputation < 0.70 {
		return fmt.Errorf("poison protection: rejected threat submission from low reputation reporter (%.2f < 0.70)", reporterReputation)
	}
	if confidence <= 0.0 || confidence > 1.0 {
		return fmt.Errorf("poison protection: invalid confidence bounds %.2f", confidence)
	}
	return nil
}
