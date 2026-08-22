package observability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObservability_TelemetryMetrics(t *testing.T) {
	tracker := &PlatformMetricsTracker{}

	tracker.RecordRiskEvaluation(4.5, false)
	tracker.RecordRiskEvaluation(12.0, true)
	tracker.RecordAgentExecution()
	tracker.RecordCaseResolved()

	snap := tracker.GetSnapshot()
	assert.Equal(t, uint64(2), snap["risk_requests_total"])
	assert.Equal(t, uint64(1), snap["fraud_detected_total"])
	assert.Equal(t, uint64(1), snap["agent_evaluations_total"])
	assert.Equal(t, uint64(1), snap["cases_resolved_total"])
	assert.Equal(t, 8.25, snap["avg_latency_ms"])
}
