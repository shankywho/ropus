package saas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaaS_OrganizationAndMetering(t *testing.T) {
	mgr := NewSaaSManager()

	// 1. Provision Organization
	org, err := mgr.CreateOrganization("org_test_bank", "Test Digital Bank", "BANKING", PlanEnterprise, "owner@testbank.com", "Owner User")
	require.NoError(t, err)
	assert.Equal(t, "org_test_bank", org.OrgID)
	assert.Equal(t, PlanEnterprise, org.Plan)

	// 2. Invite Member
	member, err := mgr.InviteMember("org_test_bank", "analyst@testbank.com", "Alice Analyst", RoleAnalyst)
	require.NoError(t, err)
	assert.Equal(t, RoleAnalyst, member.Role)

	// 3. Update Configuration
	cfg := TenantConfiguration{
		ActiveModelVersion:  "fraud-xgb-v6",
		BlockRiskThreshold:  85.0,
		ReviewRiskThreshold: 35.0,
		EnableAutonomousAI:  true,
	}
	require.NoError(t, mgr.UpdateConfiguration("org_test_bank", cfg))

	retrieved, err := mgr.GetOrganization("org_test_bank")
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-v6", retrieved.Configuration.ActiveModelVersion)

	// 4. Metering
	meter := NewUsageMeterEngine()
	meter.RecordRiskCheck("org_test_bank", 500)
	meter.RecordCaseCreation("org_test_bank")
	meter.RecordAgentCall("org_test_bank")

	usage := meter.GetTenantUsage("org_test_bank")
	assert.Equal(t, uint64(500), usage.RiskChecksTotal)
	assert.Equal(t, uint64(1), usage.CasesCreated)
	assert.Equal(t, uint64(1), usage.AgentCalls)
}
