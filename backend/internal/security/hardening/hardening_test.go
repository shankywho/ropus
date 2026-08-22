package hardening

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHardening_EncryptionAndSanitization(t *testing.T) {
	enc, err := NewEncryptionManager("")
	require.NoError(t, err)

	// 1. AES-256 GCM Roundtrip
	secret := "sensitive_card_cvv_and_ssn_payload"
	ciphertext, err := enc.EncryptField(secret)
	require.NoError(t, err)
	assert.NotEqual(t, secret, ciphertext)

	decrypted, err := enc.DecryptField(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, secret, decrypted)

	// 2. SQL Injection / XSS Sanitizer
	valid, _ := SanitizeInput("normal customer name")
	assert.True(t, valid)

	validSQL, reasonSQL := SanitizeInput("SELECT * FROM users; DROP TABLE accounts; --")
	assert.False(t, validSQL)
	assert.NotEmpty(t, reasonSQL)

	validXSS, _ := SanitizeInput("<script>alert('pwned')</script>")
	assert.False(t, validXSS)
}
