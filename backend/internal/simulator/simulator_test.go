package simulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimulator_SyntheticWorldGeneration(t *testing.T) {
	cfg := WorldConfig{
		Seed:              12345,
		CustomerCount:     200,
		MerchantCount:     20,
		TransactionsCount: 1000,
		FraudRatio:        0.05,
	}

	engine := NewSyntheticWorldEngine(cfg)
	engine.GenerateWorld()

	customers, merchants, txs, fraudCount := engine.GetSummary()
	assert.Equal(t, 200, customers)
	assert.Equal(t, 20, merchants)
	assert.Equal(t, 1000, txs)
	assert.Greater(t, fraudCount, 10, "Fraud transactions should be generated based on ratio")
}
