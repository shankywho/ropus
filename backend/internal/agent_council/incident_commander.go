package agent_council

import (
	"fmt"
	"sync"
	"time"
)

// IncidentLifecycleStage defines the active response phase during major coordinated attacks.
type IncidentLifecycleStage string

const (
	IncidentDetected   IncidentLifecycleStage = "DETECTED"
	IncidentAnalyzing  IncidentLifecycleStage = "ANALYZING"
	IncidentContaining IncidentLifecycleStage = "CONTAINING"
	IncidentMitigating IncidentLifecycleStage = "MITIGATING"
	IncidentRecovering IncidentLifecycleStage = "RECOVERING"
	IncidentLearning   IncidentLifecycleStage = "LEARNING"
)

// MajorFraudIncident tracks the command timeline of an enterprise-level threat event.
type MajorFraudIncident struct {
	IncidentID       string                 `json:"incident_id"`
	Title            string                 `json:"title"`
	Stage            IncidentLifecycleStage `json:"stage"`
	Severity         string                 `json:"severity"` // "SEV1", "SEV2", "SEV3"
	AssignedAgents   []string               `json:"assigned_agents"`
	ContainmentPlan  string                 `json:"containment_plan"`
	StartedAt        time.Time              `json:"started_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// AIIncidentCommander orchestrates end-to-end incident coordination during severe attacks.
type AIIncidentCommander struct {
	mu        sync.RWMutex
	incidents map[string]*MajorFraudIncident
}

// NewAIIncidentCommander initializes the incident commander.
func NewAIIncidentCommander() *AIIncidentCommander {
	return &AIIncidentCommander{
		incidents: make(map[string]*MajorFraudIncident),
	}
}

// DeclareMajorIncident initiates command protocol for an enterprise attack.
func (c *AIIncidentCommander) DeclareMajorIncident(title, severity string) *MajorFraudIncident {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UTC()
	incID := fmt.Sprintf("inc_maj_%d", now.UnixNano())

	inc := &MajorFraudIncident{
		IncidentID:      incID,
		Title:           title,
		Stage:           IncidentDetected,
		Severity:        severity,
		AssignedAgents:  []string{"InvestigatorAgent", "ThreatHunterAgent", "ResponseAgent"},
		ContainmentPlan: "Active multi-agent triage and automated blast-radius containment",
		StartedAt:       now,
		UpdatedAt:       now,
	}

	c.incidents[incID] = inc
	return inc
}

// AdvanceIncidentStage transitions the incident through containment to recovery.
func (c *AIIncidentCommander) AdvanceIncidentStage(incidentID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	inc, exists := c.incidents[incidentID]
	if !exists {
		return fmt.Errorf("incident '%s' not found", incidentID)
	}

	switch inc.Stage {
	case IncidentDetected:
		inc.Stage = IncidentAnalyzing
	case IncidentAnalyzing:
		inc.Stage = IncidentContaining
	case IncidentContaining:
		inc.Stage = IncidentMitigating
	case IncidentMitigating:
		inc.Stage = IncidentRecovering
	case IncidentRecovering:
		inc.Stage = IncidentLearning
	}
	inc.UpdatedAt = time.Now().UTC()
	return nil
}
