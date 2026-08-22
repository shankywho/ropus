package api_keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeys_GenerationAndVerification(t *testing.T) {
	service := NewAPIKeyService()

	// 1. Generate Key
	perms := []string{"risk:evaluate", "cases:write"}
	res, err := service.GenerateKey("org_test_bank", "Production Gateway Key", "live", perms, 30)
	require.NoError(t, err)

	assert.Contains(t, res.PlaintextKey, "rop_live_")
	assert.Equal(t, "org_test_bank", res.Metadata.OrgID)
	assert.False(t, res.Metadata.IsRevoked)

	// 2. Verify Valid Key
	meta, err := service.VerifyKey(res.PlaintextKey)
	require.NoError(t, err)
	assert.Equal(t, res.Metadata.KeyID, meta.KeyID)
	assert.False(t, meta.LastUsedAt.IsZero())

	// 3. Verify Invalid / Fake Key
	_, err = service.VerifyKey("rop_live_fakefakefakefakefakefake")
	assert.Error(t, err)

	// 4. Revoke Key
	require.NoError(t, service.RevokeKey("org_test_bank", res.Metadata.KeyID))
	_, err = service.VerifyKey(res.PlaintextKey)
	assert.Error(t, err, "Revoked key must fail authentication")

	// 5. List Keys
	keys := service.ListKeys("org_test_bank")
	assert.Equal(t, 1, len(keys))
}
