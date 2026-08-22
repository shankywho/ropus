package product_api

import (
	"testing"

	"github.com/shankywho/ropus/backend/internal/auth"
	"github.com/shankywho/ropus/backend/internal/tenant"
	"github.com/shankywho/ropus/backend/internal/webhooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProduct_RiskEvaluationPipeline(t *testing.T) {
	evaluator := NewProductRiskEvaluator()

	// 1. Clean Low-Risk Transaction
	cleanReq := &EvaluateRiskRequest{
		TransactionID: "tx_clean_1001",
		UserID:        "usr_john_doe",
		Amount:        125.50,
		Currency:      "USD",
		Merchant:      "AmazonRetail",
		Device: DeviceDetails{
			DeviceFingerprint: "fp_known_macbook_pro",
			IPAddress:         "192.0.2.1",
			UserAgent:         "Mozilla/5.0",
		},
		Location: LocationDetails{
			Country: "US",
			City:    "Seattle",
		},
	}
	cleanResp := evaluator.EvaluateTransaction(cleanReq)
	assert.Equal(t, DecisionApprove, cleanResp.Decision)
	assert.Less(t, cleanResp.RiskScore, 30.0)
	assert.Equal(t, "v3.34-ensemble-prod", cleanResp.ModelVersion)

	// 2. High-Risk Fraud Ring Transaction
	fraudReq := &EvaluateRiskRequest{
		TransactionID: "tx_fraud_9001",
		UserID:        "usr_mule_layer",
		Amount:        15000.0,
		Currency:      "USD",
		Merchant:      "CryptoLiquidityExpress",
		Device: DeviceDetails{
			DeviceFingerprint: "fp_compromised_mule_cluster",
			IPAddress:         "198.51.100.44",
			IsEmulator:        true,
			IsVPN:             true,
		},
		Location: LocationDetails{
			Country: "CY",
			City:    "Limassol",
		},
	}
	fraudResp := evaluator.EvaluateTransaction(fraudReq)
	assert.Equal(t, DecisionBlock, fraudResp.Decision)
	assert.GreaterOrEqual(t, fraudResp.RiskScore, 80.0)
	assert.NotEmpty(t, fraudResp.Reasons)
	assert.NotEmpty(t, fraudResp.GraphSignals)
	assert.NotEmpty(t, fraudResp.HumanExplanation)
}

func TestProduct_TenantAndAPIKeyManager(t *testing.T) {
	tm := tenant.NewTenantManager()
	org := tm.CreateOrganization("org_fintech_alpha", "Alpha Payments Inc", "GROWTH", 50000)
	assert.Equal(t, "Alpha Payments Inc", org.Name)

	rawKey, keyRec, err := tm.GenerateAPIKey("org_fintech_alpha", "Production Gateway Key")
	require.NoError(t, err)
	assert.NotEmpty(t, rawKey)
	assert.True(t, keyRec.IsActive)

	// Authenticate Key
	authenticated, err := tm.AuthenticateAPIKey(rawKey)
	require.NoError(t, err)
	assert.Equal(t, "org_fintech_alpha", authenticated.OrgID)

	// Rotate Key
	newKey, newRec, err := tm.RotateAPIKey(keyRec.KeyID)
	require.NoError(t, err)
	assert.NotEmpty(t, newKey)
	assert.True(t, newRec.IsActive)

	// Old Key should now fail
	_, err = tm.AuthenticateAPIKey(rawKey)
	assert.Error(t, err)
}

func TestProduct_AuthenticationAndRBAC(t *testing.T) {
	authPlat := auth.NewAuthPlatform("test_jwt_secret_key_12345")
	token, err := authPlat.GenerateToken("usr_alice", "org_default", "alice@acme.com", tenant.RoleAdmin)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := authPlat.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "usr_alice", claims.UserID)
	assert.Equal(t, tenant.RoleAdmin, claims.Role)

	// RBAC
	assert.True(t, authPlat.AuthorizeRole(claims.Role, tenant.RoleAnalyst))
	assert.False(t, authPlat.AuthorizeRole(tenant.RoleViewer, tenant.RoleAdmin))
}

func TestProduct_WebhookPlatform(t *testing.T) {
	wp := webhooks.NewWebhookPlatform()
	sub := wp.RegisterWebhook("org_default", "https://api.acme.com/ropus-webhook", []webhooks.WebhookEventType{webhooks.EventFraudDetected}, "whsec_secret_123")
	assert.NotEmpty(t, sub.SubscriptionID)

	payload := map[string]interface{}{
		"transaction_id": "tx_test_123",
		"decision":       "BLOCK",
		"risk_score":     94.5,
	}

	logs, err := wp.DispatchEvent(webhooks.EventFraudDetected, "org_default", payload)
	require.NoError(t, err)
	assert.Equal(t, 1, len(logs))
	assert.True(t, logs[0].Delivered)
	assert.Contains(t, logs[0].Signature, "sha256=")
}
