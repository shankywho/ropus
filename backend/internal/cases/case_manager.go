package cases

import (
	"fmt"
	"sync"
	"time"
)

// CaseManager maintains the active investigation case registry, state machine, and audit trail.
type CaseManager struct {
	mu         sync.RWMutex
	cases      map[string]*FraudCase
	prioritizer *CasePrioritizer
}

// NewCaseManager initializes the case repository.
func NewCaseManager(prioritizer *CasePrioritizer) *CaseManager {
	if prioritizer == nil {
		prioritizer = NewCasePrioritizer()
	}
	return &CaseManager{
		cases:       make(map[string]*FraudCase),
		prioritizer: prioritizer,
	}
}

// CreateCase initiates a new fraud case with automated prioritization.
func (m *CaseManager) CreateCase(
	tenantID, primaryUserID string,
	txnIDs []string,
	exposure float64,
	riskScore float64,
	connectedCount int,
	velocityCount int,
	threatMatch bool,
) (*FraudCase, error) {
	if tenantID == "" || primaryUserID == "" {
		return nil, fmt.Errorf("tenantID and primaryUserID are required")
	}

	priority, _ := m.prioritizer.CalculatePriority(riskScore, exposure, connectedCount, velocityCount, threatMatch)
	now := time.Now().UTC()
	caseID := fmt.Sprintf("case_%d_%s", now.UnixNano(), primaryUserID)

	c := &FraudCase{
		CaseID:         caseID,
		TenantID:       tenantID,
		TransactionIDs: txnIDs,
		PrimaryUserID:  primaryUserID,
		TotalExposure:  exposure,
		RiskScore:      riskScore,
		Priority:       priority,
		Status:         StatusOpen,
		Evidence:       make([]EvidenceItem, 0),
		Timeline: []TimelineEvent{
			{
				EventID:   fmt.Sprintf("evt_%d_init", now.UnixNano()),
				Actor:     "SYSTEM_FRAUD_OPS",
				Action:    "CASE_CREATED",
				Details:   fmt.Sprintf("Case opened automatically with %s priority (Risk Score: %.2f, Exposure: $%.2f)", priority, riskScore, exposure),
				Timestamp: now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.cases[caseID] = c
	return c, nil
}

// GetCase retrieves a fraud case by ID.
func (m *CaseManager) GetCase(caseID string) (*FraudCase, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, exists := m.cases[caseID]
	if !exists {
		return nil, fmt.Errorf("fraud case '%s' not found", caseID)
	}
	return c, nil
}

// AssignCase allocates a case to an investigator.
func (m *CaseManager) AssignCase(caseID, analystID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, exists := m.cases[caseID]
	if !exists {
		return fmt.Errorf("fraud case '%s' not found", caseID)
	}

	now := time.Now().UTC()
	c.AssignedAnalyst = analystID
	c.Status = StatusAssigned
	c.UpdatedAt = now
	c.Timeline = append(c.Timeline, TimelineEvent{
		EventID:   fmt.Sprintf("evt_%d_assign", now.UnixNano()),
		Actor:     "CASE_ROUTER",
		Action:    "CASE_ASSIGNED",
		Details:   fmt.Sprintf("Case assigned to analyst '%s'", analystID),
		Timestamp: now,
	})

	return nil
}

// AddEvidence attaches a forensic artifact to the case.
func (m *CaseManager) AddEvidence(caseID string, item EvidenceItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, exists := m.cases[caseID]
	if !exists {
		return fmt.Errorf("fraud case '%s' not found", caseID)
	}

	if item.CapturedAt.IsZero() {
		item.CapturedAt = time.Now().UTC()
	}
	c.Evidence = append(c.Evidence, item)
	c.UpdatedAt = time.Now().UTC()
	return nil
}

// AddTimelineEvent appends an audit event to the case history.
func (m *CaseManager) AddTimelineEvent(caseID, actor, action, details string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, exists := m.cases[caseID]
	if !exists {
		return fmt.Errorf("fraud case '%s' not found", caseID)
	}

	now := time.Now().UTC()
	c.Timeline = append(c.Timeline, TimelineEvent{
		EventID:   fmt.Sprintf("evt_%d", now.UnixNano()),
		Actor:     actor,
		Action:    action,
		Details:   details,
		Timestamp: now,
	})
	c.UpdatedAt = now
	return nil
}

// ResolveCase finalizes the case outcome and transitions state to CONFIRMED_FRAUD, FALSE_POSITIVE, or CLOSED.
func (m *CaseManager) ResolveCase(caseID string, status CaseStatus, notes string, resolverActor string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, exists := m.cases[caseID]
	if !exists {
		return fmt.Errorf("fraud case '%s' not found", caseID)
	}

	now := time.Now().UTC()
	c.Status = status
	c.ResolutionNotes = notes
	c.ResolvedAt = &now
	c.UpdatedAt = now

	c.Timeline = append(c.Timeline, TimelineEvent{
		EventID:   fmt.Sprintf("evt_%d_resolve", now.UnixNano()),
		Actor:     resolverActor,
		Action:    fmt.Sprintf("STATUS_CHANGED_TO_%s", status),
		Details:   notes,
		Timestamp: now,
	})

	return nil
}

// ListCases returns active or historic cases filtered by status or priority.
func (m *CaseManager) ListCases(status CaseStatus, priority CasePriority) []*FraudCase {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []*FraudCase
	for _, c := range m.cases {
		if (status == "" || c.Status == status) && (priority == "" || c.Priority == priority) {
			res = append(res, c)
		}
	}
	return res
}
