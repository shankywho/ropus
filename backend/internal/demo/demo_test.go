package demo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDemo_ScenariosExecution(t *testing.T) {
	orchestrator := NewDemoOrchestrator(nil)

	scenarios := []DemoScenarioType{
		ScenarioSyntheticIdentity,
		ScenarioAccountTakeover,
		ScenarioFraudRingAttack,
	}

	for _, sc := range scenarios {
		result := orchestrator.RunScenario(sc)
		assert.NotEmpty(t, result.ScenarioID)
		assert.Equal(t, 6, len(result.Steps))
		assert.NotEmpty(t, result.RiskResponse.Decision)
		assert.NotEmpty(t, result.StorySummary)
	}
}
