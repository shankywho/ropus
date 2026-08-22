package incident

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncident_LifecycleTracking(t *testing.T) {
	mgr := NewIncidentManager()

	// 1. Trigger Incident
	inc := mgr.TriggerIncident("Kafka Consumer Lag Spike", SeverityP1High, "Partition rebalancing delay")
	assert.NotEmpty(t, inc.IncidentID)
	assert.Equal(t, StateDetected, inc.State)

	// 2. Transition to Acknowledged -> Mitigated -> Resolved
	require.NoError(t, mgr.TransitionState(inc.IncidentID, StateAcknowledged, "On-call engineer paged"))
	require.NoError(t, mgr.TransitionState(inc.IncidentID, StateMitigated, "Scaled consumer replicas from 4 to 12"))
	require.NoError(t, mgr.TransitionState(inc.IncidentID, StateResolved, "Lag restored to normal"))

	incidents := mgr.ListIncidents()
	assert.Equal(t, 1, len(incidents))
	assert.Equal(t, StateResolved, incidents[0].State)
	assert.NotNil(t, incidents[0].ResolvedAt)
}
