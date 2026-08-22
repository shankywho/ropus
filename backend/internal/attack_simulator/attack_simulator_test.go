package attack_simulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttackSimulator_Scenarios(t *testing.T) {
	sim := NewRealFraudAttackSimulator()

	scenarios := []AttackScenarioType{
		ScenarioAccountTakeover,
		ScenarioSyntheticIdentity,
		ScenarioCardTesting,
		ScenarioMoneyLaundering,
	}

	for _, sc := range scenarios {
		res := sim.ExecuteScenario(sc)
		assert.NotEmpty(t, res.CampaignID)
		assert.Equal(t, sc, res.ScenarioType)
		assert.Greater(t, res.TotalGrossLossAtRisk, 0.0)
		assert.Equal(t, 6, len(res.Timeline))
		assert.NotEmpty(t, res.InvestigationReport)
		assert.GreaterOrEqual(t, res.Confidence, 0.95)
	}
}
