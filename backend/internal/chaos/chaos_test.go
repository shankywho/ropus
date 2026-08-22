package chaos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChaos_FailureDrills(t *testing.T) {
	chaos := NewChaosEngine()

	scenarios := []ChaosScenario{
		ChaosKafkaOutage,
		ChaosDBUnavailable,
		ChaosModelTimeout,
		ChaosNetworkLatency,
		ChaosRedisFailure,
	}

	for _, sc := range scenarios {
		res := chaos.ExecuteDrill(sc)
		assert.NotEmpty(t, res.DrillID)
		assert.True(t, res.DetectedBySystem)
		assert.True(t, res.FallbackActivated)
		assert.Equal(t, 0, res.CustomerImpactDrops, "Zero customer impact required under failure drills")
		assert.Less(t, res.RecoveryTimeMs, 100.0)
	}
}
