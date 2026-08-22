package incident

import (
	"fmt"
	"sync"
	"time"
)

// IncidentSeverity categorizes operational issues.
type IncidentSeverity string

const (
	SeverityP0Critical IncidentSeverity = "P0_CRITICAL" // Outage or data loss risk
	SeverityP1High     IncidentSeverity = "P1_HIGH"     // Degraded latency or partial failover
	SeverityP2Medium   IncidentSeverity = "P2_MEDIUM"   // Model drift or non-critical worker lag
)

// IncidentState defines the incident response lifecycle.
type IncidentState string

const (
	StateDetected     IncidentState = "DETECTED"
	StateAcknowledged IncidentState = "ACKNOWLEDGED"
	StateMitigated    IncidentState = "MITIGATED"
	StateResolved     IncidentState = "RESOLVED"
	StatePostmortem   IncidentState = "POSTMORTEM"
)

// OperationalIncident represents an active or historical incident ticket.
type OperationalIncident struct {
	IncidentID   string           `json:"incident_id"`
	Title        string           `json:"title"`
	Severity     IncidentSeverity `json:"severity"`
	State        IncidentState    `json:"state"`
	RootCause    string           `json:"root_cause"`
	Mitigation   string           `json:"mitigation"`
	DetectedAt   time.Time        `json:"detected_at"`
	ResolvedAt   *time.Time       `json:"resolved_at,omitempty"`
}

// IncidentManager coordinates automated incident detection and recovery tracking.
type IncidentManager struct {
	mu        sync.RWMutex
	incidents map[string]*OperationalIncident
}

// NewIncidentManager initializes the incident response manager.
func NewIncidentManager() *IncidentManager {
	return &IncidentManager{
		incidents: make(map[string]*OperationalIncident),
	}
}

// TriggerIncident logs an automated incident and notifies on-call systems.
func (m *IncidentManager) TriggerIncident(title string, sev IncidentSeverity, rootCause string) *OperationalIncident {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	id := fmt.Sprintf("inc_%d", now.UnixNano())

	inc := &OperationalIncident{
		IncidentID: id,
		Title:      title,
		Severity:   sev,
		State:      StateDetected,
		RootCause:  rootCause,
		DetectedAt: now,
	}

	m.incidents[id] = inc
	return inc
}

// TransitionState updates the incident lifecycle state.
func (m *IncidentManager) TransitionState(incidentID string, newState IncidentState, notes string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inc, exists := m.incidents[incidentID]
	if !exists {
		return fmt.Errorf("incident '%s' not found", incidentID)
	}

	inc.State = newState
	if newState == StateMitigated {
		inc.Mitigation = notes
	} else if newState == StateResolved {
		now := time.Now().UTC()
		inc.ResolvedAt = &now
	}
	return nil
}

// ListIncidents returns all active and resolved incidents.
func (m *IncidentManager) ListIncidents() []*OperationalIncident {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]*OperationalIncident, 0, len(m.incidents))
	for _, v := range m.incidents {
		res = append(res, v)
	}
	return res
}
