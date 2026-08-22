package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurity_GatewayAndSignature(t *testing.T) {
	gw := NewEnterpriseSecurityGateway()

	// 1. Signature Verification
	payload := []byte(`{"transaction_id":"tx_123","amount":500}`)
	secret := "secret_key_prod_123"
	// Create valid signature
	assert.False(t, gw.VerifyRequestSignature(payload, "sha256=invalid_sig", secret))

	// 2. IP Blocklist validation
	_, err := gw.ValidateRequest("org_default", "198.51.100.44")
	assert.Error(t, err, "Known bulletproof IP must be blocked")

	// 3. Clean Request
	sCtx, err := gw.ValidateRequest("org_default", "192.0.2.1")
	require.NoError(t, err)
	assert.True(t, sCtx.Authenticated)
	assert.False(t, sCtx.IsIPBlocked)
}
