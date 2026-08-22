package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeIdentifier_Valid(t *testing.T) {
	validIDs := []string{
		"model_cand_20260821",
		"fraud-xgb-25f-v3.1-candidate-01",
		"job_retrain_12345",
		"version.1.0-alpha",
	}

	for _, id := range validIDs {
		clean, err := SanitizeIdentifier(id)
		require.NoError(t, err, "Should be valid: %s", id)
		assert.Equal(t, id, clean)
	}
}

func TestSanitizeIdentifier_PathTraversalBlocked(t *testing.T) {
	badIDs := []string{
		"../etc/passwd",
		"..\\windows\\win.ini",
		"models/../../secret",
		"candidate\x00hidden",
		"/root/admin",
		"model; rm -rf /",
		"model name with spaces",
		"model$()",
		"",
	}

	for _, id := range badIDs {
		_, err := SanitizeIdentifier(id)
		assert.Error(t, err, "Should reject malicious/invalid identifier: %s", id)
	}
}

func TestConstantTimeCompare(t *testing.T) {
	assert.True(t, ConstantTimeCompare("secret_key_123", "secret_key_123"))
	assert.False(t, ConstantTimeCompare("secret_key_123", "wrong_key"))
	assert.False(t, ConstantTimeCompare("secret_key_123", "secret_key_124"))
}

func TestSanitizeLogMessage_MasksPANAndCVV(t *testing.T) {
	raw := "Customer payment with card 4111 2222 3333 4444 and cvv: 123 failed"
	sanitized := SanitizeLogMessage(raw)
	assert.NotContains(t, sanitized, "4111 2222 3333 4444")
	assert.NotContains(t, sanitized, "123")
	assert.Contains(t, sanitized, "cvv:***")
}
