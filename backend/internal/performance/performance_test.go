package performance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPerformance_ScaleSimulation(t *testing.T) {
	engine := NewPerformanceEngine()

	res := engine.RunScaleSimulation(1000000, 10*time.Second)
	assert.GreaterOrEqual(t, res.ThroughputPerSec, 100000.0)
	assert.Less(t, res.LatencyP99Ms, 50.0, "P99 SLA must be under 50ms")
	assert.GreaterOrEqual(t, res.AvailabilityPercent, 99.99)
}
