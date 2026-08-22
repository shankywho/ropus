package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhook_ReliabilityAndSigning(t *testing.T) {
	platform := NewWebhookPlatform()
	secret := "whsec_live_99a818c7e41"

	sub := platform.RegisterWebhook(
		"org_webhook_test",
		"https://api.acmebank.com/v1/ropus-events",
		[]WebhookEventType{EventRiskDecisionCreated, EventCaseCreated},
		secret,
	)
	assert.NotEmpty(t, sub.SubscriptionID)

	// 1. Dispatch Event
	payload := map[string]interface{}{
		"decision_id":   "dec_99182",
		"decision":      "BLOCK",
		"risk_score":    0.96,
		"transaction_id": "tx_order_1001",
	}

	logs, err := platform.DispatchEvent(EventRiskDecisionCreated, "org_webhook_test", payload)
	require.NoError(t, err)
	require.Equal(t, 1, len(logs))

	log := logs[0]
	assert.True(t, log.Delivered)
	assert.NotEmpty(t, log.Signature)

	// 2. Verify Cryptographic HMAC Signature
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payloadBytes)
	expectedSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	assert.Equal(t, expectedSig, log.Signature)
	assert.True(t, strings.HasPrefix(log.Signature, "sha256="))
}
