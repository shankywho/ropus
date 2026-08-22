package developer

import (
	"context"
	"testing"

	"github.com/shankywho/ropus/backend/internal/auth/api_keys"
	"github.com/shankywho/ropus/backend/internal/ml"
	"github.com/shankywho/ropus/backend/internal/saas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeveloper_EvaluateTransactionEndpoint(t *testing.T) {
	ctx := context.Background()
	keyService := api_keys.NewAPIKeyService()
	meter := saas.NewUsageMeterEngine()
	mlEngine := ml.NewRealMLInferenceEngine()

	gw := NewDeveloperAPIGateway(keyService, meter, mlEngine)

	// 1. Provision Key
	keyRes, err := keyService.GenerateKey("org_test_fintech", "Dev API Key", "live", []string{"risk:evaluate"}, 30)
	require.NoError(t, err)

	// 2. Clean Transaction
	cleanReq := EvaluateRiskRequest{
		CustomerID:    "cust_12345",
		TransactionID: "tx_clean_01",
		Amount:        45.0,
		Currency:      "USD",
		Merchant:      "AmazonRetail",
		DeviceID:      "dev_iphone_15",
	}

	resp, err := gw.EvaluateTransaction(ctx, keyRes.PlaintextKey, cleanReq)
	require.NoError(t, err)
	assert.Equal(t, "APPROVE", resp.Decision)
	assert.Equal(t, "tx_clean_01", resp.TransactionID)

	// 3. High Risk Transaction
	fraudReq := EvaluateRiskRequest{
		CustomerID:    "cust_99881",
		TransactionID: "tx_fraud_01",
		Amount:        12500.0,
		Currency:      "USD",
		Merchant:      "CryptoLiquidity",
		DeviceID:      "dev_emulator_root",
	}

	fraudResp, err := gw.EvaluateTransaction(ctx, keyRes.PlaintextKey, fraudReq)
	require.NoError(t, err)
	assert.Equal(t, "BLOCK", fraudResp.Decision)
	assert.GreaterOrEqual(t, fraudResp.RiskScore, 0.80)

	// 4. Verify Metering was updated
	usage := meter.GetTenantUsage("org_test_fintech")
	assert.Equal(t, uint64(2), usage.RiskChecksTotal)
}
