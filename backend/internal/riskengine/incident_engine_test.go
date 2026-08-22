package riskengine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncidentEngine_DetectionAndDeduplication(t *testing.T) {
	alertSink := NewInMemoryAlertSink(50)
	alertMgr := NewAlertManager(alertSink)
	defer alertMgr.Close()

	engine := NewIncidentEngine(alertMgr, nil)
	ctx := context.Background()

	health := SystemHealthReport{
		OverallStatus: HealthStatusDegraded,
		Timestamp:     time.Now().UTC(),
		Components: map[string]ComponentHealth{
			"clickhouse": {
				Name:     "clickhouse",
				Status:   HealthStatusDegraded,
				Message:  "ClickHouse high query latency",
				Critical: false,
			},
		},
	}

	slo := SLOSummary{
		OverallStatus: SLOStatusBreached,
		Measurements: map[string]SLOMetricRecord{
			"slo_availability": {
				SLOID:        "slo_availability",
				Name:         "Risk Evaluation Availability",
				CurrentValue: 0.985,
				Target:       0.999,
				Status:       SLOStatusBreached,
				BurnRate:     15.0,
			},
		},
	}

	// 1. Initial Evaluation -> Should raise 2 incidents
	incidents := engine.Evaluate(ctx, health, slo, CircuitStateHealthy, DriftStatusHealthy, StateIdle, "fraud-xgb-25f-v3.0")
	assert.Equal(t, 2, len(incidents))

	// Verify alert received in sink
	time.Sleep(20 * time.Millisecond)
	alerts := alertSink.GetRecentAlerts()
	assert.True(t, len(alerts) >= 2)

	// 2. Second Evaluation with same failure -> Should deduplicate and increment occurrence count
	incidents2 := engine.Evaluate(ctx, health, slo, CircuitStateHealthy, DriftStatusHealthy, StateIdle, "fraud-xgb-25f-v3.0")
	assert.Equal(t, 2, len(incidents2))
	for _, inc := range incidents2 {
		assert.Equal(t, int64(2), inc.OccurrenceCount)
		assert.Equal(t, IncidentStateOpen, inc.Status)
	}

	// 3. Acknowledge Incident
	incID := incidents2[0].IncidentID
	err := engine.AcknowledgeIncident(incID, "security_analyst_1", "Investigating ClickHouse latency spike")
	require.NoError(t, err)

	list := engine.ListIncidents()
	var foundAck bool
	for _, inc := range list {
		if inc.IncidentID == incID {
			assert.Equal(t, IncidentStateAcknowledged, inc.Status)
			assert.Equal(t, "security_analyst_1", inc.AcknowledgedBy)
			foundAck = true
		}
	}
	assert.True(t, foundAck)

	// 4. Resolve Condition in System -> Auto-resolution on next evaluation
	healthHealthy := SystemHealthReport{
		OverallStatus: HealthStatusHealthy,
		Components: map[string]ComponentHealth{
			"clickhouse": {
				Name:     "clickhouse",
				Status:   HealthStatusHealthy,
				Message:  "ClickHouse latency normal",
				Critical: false,
			},
		},
	}
	sloHealthy := SLOSummary{
		OverallStatus: SLOStatusHealthy,
		Measurements: map[string]SLOMetricRecord{
			"slo_availability": {
				SLOID:        "slo_availability",
				Name:         "Risk Evaluation Availability",
				CurrentValue: 0.9999,
				Target:       0.999,
				Status:       SLOStatusHealthy,
			},
		},
	}

	incidents3 := engine.Evaluate(ctx, healthHealthy, sloHealthy, CircuitStateHealthy, DriftStatusHealthy, StateIdle, "fraud-xgb-25f-v3.0")
	for _, inc := range incidents3 {
		assert.Equal(t, IncidentStateResolved, inc.Status)
		assert.NotNil(t, inc.ResolvedAt)
	}
}

func TestIncidentEngine_ManualResolution(t *testing.T) {
	alertMgr := NewAlertManager(&LogAlertSink{})
	defer alertMgr.Close()

	engine := NewIncidentEngine(alertMgr, nil)
	ctx := context.Background()

	health := SystemHealthReport{
		OverallStatus: HealthStatusUnhealthy,
		Components: map[string]ComponentHealth{
			"postgres": {
				Name:     "postgres",
				Status:   HealthStatusUnhealthy,
				Message:  "Database down",
				Critical: true,
			},
		},
	}

	incidents := engine.Evaluate(ctx, health, SLOSummary{}, CircuitStateHealthy, DriftStatusHealthy, StateIdle, "v1")
	require.NotEmpty(t, incidents)

	incID := incidents[0].IncidentID
	err := engine.ResolveIncident(incID, "lead_sre", "PostgreSQL failover completed successfully")
	require.NoError(t, err)

	list := engine.ListIncidents()
	assert.Equal(t, IncidentStateResolved, list[0].Status)
	assert.Equal(t, "lead_sre", list[0].ResolvedBy)
}
