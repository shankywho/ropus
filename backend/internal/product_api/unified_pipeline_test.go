package product_api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shankywho/ropus/backend/internal/auth/api_keys"
	"github.com/shankywho/ropus/backend/internal/saas"
)

func TestUnifiedRiskPipeline_EndToEnd(t *testing.T) {
	keyService := api_keys.NewAPIKeyService()
	meter := saas.NewUsageMeterEngine()
	pipeline := NewUnifiedRiskPipeline(keyService, meter, nil)

	// Create valid API key for tenant
	keyResp, err := keyService.GenerateKey("org_fintech_alpha", "Test Checkout Key", "live", []string{"risk:evaluate"}, 90)
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Low Risk Transaction -> Should APPROVE
	reqLow := CanonicalRiskRequest{
		TransactionID: "tx_low_001",
		CustomerID:    "usr_clean_101",
		Amount:        45.00,
		Currency:      "USD",
		MerchantID:    "Starbucks",
		DeviceID:      "dev_clean_phone",
		IPAddress:     "192.0.2.1",
		Country:       "US",
		Timestamp:     time.Now().UTC(),
	}

	resLow, err := pipeline.EvaluateRisk(ctx, keyResp.PlaintextKey, reqLow)
	require.NoError(t, err)
	assert.Equal(t, "APPROVE", resLow.Decision)
	assert.Less(t, resLow.RiskScore, 0.30)
	assert.NotEmpty(t, resLow.DecisionID)

	// 2. High Risk Transaction -> Should BLOCK with Exact Factor Attribution
	reqHigh := CanonicalRiskRequest{
		TransactionID: "tx_high_999",
		CustomerID:    "usr_synthetic_bot_01",
		Amount:        14500.00,
		Currency:      "USD",
		MerchantID:    "CryptoLiquidityExpress",
		DeviceID:      "dev_mule_cluster_99",
		IPAddress:     "198.51.100.44",
		Country:       "CY",
		Timestamp:     time.Now().UTC(),
	}

	resHigh, err := pipeline.EvaluateRisk(ctx, keyResp.PlaintextKey, reqHigh)
	require.NoError(t, err)
	assert.Equal(t, "BLOCK", resHigh.Decision)
	assert.GreaterOrEqual(t, resHigh.RiskScore, 0.80)
	assert.NotEmpty(t, resHigh.CaseID)
	assert.NotEmpty(t, resHigh.RiskFactors)
	assert.NotEmpty(t, resHigh.Reasons)

	// Verify persistence in stored decision records
	stored, exists := pipeline.GetStoredDecision(resHigh.DecisionID)
	assert.True(t, exists)
	assert.Equal(t, "org_fintech_alpha", stored.TenantID)
	assert.Equal(t, "BLOCK", stored.Decision)

	// Verify usage meter recorded the checks
	snapshot := meter.GetTenantUsage("org_fintech_alpha")
	assert.Equal(t, uint64(2), snapshot.RiskChecksTotal)
	assert.Equal(t, uint64(1), snapshot.CasesCreated)
}

func TestUnifiedRiskPipeline_BotCadenceAttackDetection(t *testing.T) {
	keyService := api_keys.NewAPIKeyService()
	meter := saas.NewUsageMeterEngine()
	pipeline := NewUnifiedRiskPipeline(keyService, meter, nil)

	keyResp, err := keyService.GenerateKey("org_fintech_bot_target", "Bot Target Key", "live", []string{"risk:evaluate"}, 90)
	require.NoError(t, err)

	ctx := context.Background()
	baseTime := time.Now().UTC().Add(-5 * time.Second)

	// Simulate rapid scripted loop with identical 250ms intervals from single bot device
	var lastRes *CanonicalRiskResponse
	for i := 0; i < 5; i++ {
		req := CanonicalRiskRequest{
			TransactionID: "tx_script_" + string(rune('a'+i)),
			CustomerID:    "usr_victim_target",
			Amount:        50.00,
			Currency:      "USD",
			MerchantID:    "E-Shop",
			DeviceID:      "fp_scripted_botnet_node",
			IPAddress:     "203.0.113.77",
			Country:       "US",
			Timestamp:     baseTime.Add(time.Duration(i*250) * time.Millisecond),
		}
		var err error
		lastRes, err = pipeline.EvaluateRisk(ctx, keyResp.PlaintextKey, req)
		require.NoError(t, err)
	}

	// Verify that Bot Risk Layer factor was added and elevated the risk score
	hasBotFactor := false
	for _, factor := range lastRes.RiskFactors {
		if factor.FactorName == "Automated Attack & Bot Risk Layer" {
			hasBotFactor = true
			assert.Greater(t, factor.Contribution, 0.0)
		}
	}
	assert.True(t, hasBotFactor, "Expected 'Automated Attack & Bot Risk Layer' factor contribution in response")
}
