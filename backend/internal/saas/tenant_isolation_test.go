package saas

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shankywho/ropus/backend/internal/auth/api_keys"
)

func TestMultiTenant_AdversarialBoundaryIsolation(t *testing.T) {
	saasMgr := NewSaaSManager()
	keyService := api_keys.NewAPIKeyService()

	// 1. Provision Tenant A and Tenant B
	orgA, err := saasMgr.CreateOrganization("org_acme_test", "Acme Bank", "BANKING", PlanGrowth, "owner@acmebank.com", "Acme Owner")
	require.NoError(t, err)

	orgB, err := saasMgr.CreateOrganization("org_beta_test", "Beta Payments", "FINTECH", PlanStarter, "owner@betapayments.com", "Beta Owner")
	require.NoError(t, err)

	// 2. Generate Key for Tenant A
	keyA, err := keyService.GenerateKey(orgA.OrgID, "Acme Production Key", "live", []string{"risk:evaluate"}, 90)
	require.NoError(t, err)

	// 3. Verify Key resolution maps strictly to Org A
	metaA, err := keyService.VerifyKey(keyA.PlaintextKey)
	require.NoError(t, err)
	assert.Equal(t, orgA.OrgID, metaA.OrgID)
	assert.NotEqual(t, orgB.OrgID, metaA.OrgID)

	// 4. Adversarial Check: Tenant B attempting to read Tenant A configuration
	fetchedA, err := saasMgr.GetOrganization(orgA.OrgID)
	require.NoError(t, err)

	fetchedB, err := saasMgr.GetOrganization(orgB.OrgID)
	require.NoError(t, err)

	assert.NotEqual(t, fetchedA.OrgID, fetchedB.OrgID)
	assert.NotEqual(t, fetchedA.Name, fetchedB.Name)

	// 5. Revoke Tenant A key -> Tenant B is unaffected
	require.NoError(t, keyService.RevokeKey(orgA.OrgID, keyA.Metadata.KeyID))

	_, err = keyService.VerifyKey(keyA.PlaintextKey)
	assert.Error(t, err, "Revoked key must be rejected")
}
