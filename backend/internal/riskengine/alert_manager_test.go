package riskengine

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertManager_AsyncDispatch(t *testing.T) {
	sink := NewInMemoryAlertSink(20)
	mgr := NewAlertManager(sink)
	defer mgr.Close()

	for i := 0; i < 5; i++ {
		mgr.Emit(Alert{
			AlertID:       fmt.Sprintf("alt_%d", i),
			Timestamp:     time.Now().UTC(),
			Severity:      IncidentSeverityCritical,
			Title:         fmt.Sprintf("Critical Alert %d", i),
			Message:       "P99 latency threshold breached",
			IncidentID:    fmt.Sprintf("inc_%d", i),
			Subsystem:     "risk_engine",
			CorrelationID: "req_test_123",
		})
	}

	time.Sleep(50 * time.Millisecond)
	alerts := sink.GetRecentAlerts()
	require.Equal(t, 5, len(alerts))
	assert.Equal(t, "alt_4", alerts[0].AlertID) // Reverse chronological
}

func TestAlertManager_DropUnderExtremeLoad(t *testing.T) {
	sink := NewInMemoryAlertSink(500)
	mgr := NewAlertManager(sink)
	defer mgr.Close()

	// Rapidly emit 600 alerts (queue capacity 500)
	for i := 0; i < 600; i++ {
		mgr.Emit(Alert{
			AlertID:   fmt.Sprintf("alt_%d", i),
			Timestamp: time.Now().UTC(),
			Severity:  IncidentSeverityInfo,
			Title:     "Load Test Alert",
			Subsystem: "test",
		})
	}

	time.Sleep(100 * time.Millisecond)
	alerts := sink.GetRecentAlerts()
	assert.True(t, len(alerts) > 0)
}
