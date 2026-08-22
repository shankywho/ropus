package billing

import (
	"testing"

	"github.com/shankywho/ropus/backend/internal/saas"
	"github.com/stretchr/testify/assert"
)

func TestBilling_InvoiceGenerationAndOverages(t *testing.T) {
	engine := NewBillingEngine()

	// 1. Starter Tier under quota
	u1 := &saas.TenantUsageSnapshot{
		OrgID:           "org_starter_01",
		RiskChecksTotal: 45000,
	}
	inv1 := engine.GenerateInvoice("org_starter_01", saas.PlanStarter, u1)
	assert.Equal(t, 499.0, inv1.TotalAmount)
	assert.Equal(t, 0.0, inv1.OverageFee)

	// 2. Starter Tier with overage (150,000 checks > 100,000 included -> 50,000 * 0.005 = $250)
	u2 := &saas.TenantUsageSnapshot{
		OrgID:           "org_starter_02",
		RiskChecksTotal: 150000,
	}
	inv2 := engine.GenerateInvoice("org_starter_02", saas.PlanStarter, u2)
	assert.Equal(t, 250.0, inv2.OverageFee)
	assert.Equal(t, 749.0, inv2.TotalAmount)

	// 3. Invoice retrieval
	invs := engine.GetInvoices("org_starter_02")
	assert.Equal(t, 1, len(invs))
}
