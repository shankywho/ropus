package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraph_TenantIsolationSameIP(t *testing.T) {
	engine := NewGraphEngine(nil)

	sharedIP := "185.220.101.55"

	// Tenant A: confirmed fraud ring using sharedIP
	_ = engine.IngestTenantTransactionLinks("tenant_A", "txn_A1", "usr_A", "acc_A", "card_A", "dev_A", sharedIP, "merch_1", 50000, true)

	// Tenant B: clean merchant transaction using the same sharedIP
	_ = engine.IngestTenantTransactionLinks("tenant_B", "txn_B1", "usr_B", "acc_B", "card_B", "dev_B", sharedIP, "merch_1", 250, false)

	// Evaluate Tenant A
	evA := engine.EvaluateTenantEntityGraph("tenant_A", "acc_A", "dev_A", "card_A", sharedIP)
	assert.GreaterOrEqual(t, evA.FraudNodesCount, 1, "Tenant A must detect its own confirmed fraud node")
	assert.True(t, evA.GraphRiskContribution >= 0.85, "Tenant A risk score must be elevated")

	// Evaluate Tenant B
	evB := engine.EvaluateTenantEntityGraph("tenant_B", "acc_B", "dev_B", "card_B", sharedIP)
	assert.Equal(t, 0, evB.FraudNodesCount, "Tenant B must NOT inherit Tenant A's fraud node despite identical IP")
	assert.True(t, evB.GraphRiskContribution <= 0.30, "Tenant B risk score must remain clean")
}

func TestGraph_TenantIsolationSameDevice(t *testing.T) {
	engine := NewGraphEngine(nil)

	sharedDev := "dev_fingerprint_universal_emulator"

	// Tenant A: mule accounts linked to device
	_ = engine.IngestTenantTransactionLinks("tenant_A", "txn_A1", "usr_A1", "acc_A1", "card_A1", sharedDev, "10.0.0.1", "merch_1", 1000, true)
	_ = engine.IngestTenantTransactionLinks("tenant_A", "txn_A2", "usr_A2", "acc_A2", "card_A2", sharedDev, "10.0.0.2", "merch_1", 1000, true)

	// Tenant B: single legitimate account using the same device ID
	_ = engine.IngestTenantTransactionLinks("tenant_B", "txn_B1", "usr_B1", "acc_B1", "card_B1", sharedDev, "10.0.0.3", "merch_1", 100, false)

	// Tenant B evaluation must see 0 connected accounts from Tenant A
	evB := engine.EvaluateTenantEntityGraph("tenant_B", "acc_B1", sharedDev, "card_B1", "10.0.0.3")
	assert.Equal(t, 0, evB.ConnectedAccountCount, "Tenant B must see 0 external connected accounts")
	assert.Equal(t, 0, evB.FraudNodesCount, "Tenant B must see 0 fraud nodes")
}

func TestGraph_TenantIsolationSameAccount(t *testing.T) {
	engine := NewGraphEngine(nil)

	// Both tenants happen to use account identifier "acc_1001"
	sharedAcc := "acc_1001"

	_ = engine.IngestTenantTransactionLinks("tenant_A", "txn_A1", "usr_A", sharedAcc, "card_A", "dev_A", "10.0.0.1", "merch_1", 9000, true)
	_ = engine.IngestTenantTransactionLinks("tenant_B", "txn_B1", "usr_B", sharedAcc, "card_B", "dev_B", "10.0.0.2", "merch_1", 50, false)

	evA := engine.EvaluateTenantEntityGraph("tenant_A", sharedAcc, "dev_A", "card_A", "10.0.0.1")
	evB := engine.EvaluateTenantEntityGraph("tenant_B", sharedAcc, "dev_B", "card_B", "10.0.0.2")

	assert.GreaterOrEqual(t, evA.FraudNodesCount, 1)
	assert.Equal(t, 0, evB.FraudNodesCount)
}

func TestGraph_CannotTraverseIntoOtherTenant(t *testing.T) {
	engine := NewGraphEngine(nil)

	// Tenant A syndicate ring
	_ = engine.IngestTenantTransactionLinks("tenant_A", "txn_A1", "usr_A1", "acc_A1", "tok_A", "dev_A", "185.220.1.1", "merch_1", 10000, true)
	_ = engine.IngestTenantTransactionLinks("tenant_A", "txn_A2", "usr_A2", "acc_A2", "tok_A", "dev_A", "185.220.1.1", "merch_1", 10000, true)

	// Export graph for Tenant B
	graphB := engine.ExportTenantFraudGraph("tenant_B", "dec_test", "acc_A1")
	// Must NOT contain any of Tenant A's entities
	for _, entity := range graphB.Entities {
		assert.NotEqual(t, "acc_A1", entity.ID, "Tenant B graph export must not contain Tenant A entities")
		assert.NotEqual(t, "acc_A2", entity.ID, "Tenant B graph export must not contain Tenant A entities")
	}
}
