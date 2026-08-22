package agent_council

import (
	"sync"
	"time"
)

// EpisodicMemoryEntry stores post-incident findings and learning signals for agent evolution.
type EpisodicMemoryEntry struct {
	EntryID         string    `json:"entry_id"`
	IncidentID      string    `json:"incident_id"`
	OutcomeType     string    `json:"outcome_type"` // "SUCCESSFUL_DEFENSE", "FAILED_DEFENSE", "ANALYST_CORRECTION", "POLICY_OVERRIDE"
	HashedEntities  []string  `json:"hashed_entities"`
	ActionTaken     string    `json:"action_taken"`
	LessonsLearned  string    `json:"lessons_learned"`
	RecordedAt      time.Time `json:"recorded_at"`
}

// CollectiveAgentMemory maintains organizational learning across all participating AI agents.
type CollectiveAgentMemory struct {
	mu      sync.RWMutex
	history map[string]*EpisodicMemoryEntry
}

// NewCollectiveAgentMemory initializes the collective memory system.
func NewCollectiveAgentMemory() *CollectiveAgentMemory {
	return &CollectiveAgentMemory{
		history: make(map[string]*EpisodicMemoryEntry),
	}
}

// StoreEpisodicOutcome logs an investigation outcome into shared organizational memory.
func (m *CollectiveAgentMemory) StoreEpisodicOutcome(incidentID, outcomeType, actionTaken, lessonsLearned string, rawEntities []string) *EpisodicMemoryEntry {
	var hashed []string
	for _, e := range rawEntities {
		hashed = append(hashed, HashKey(e))
	}

	now := time.Now().UTC()
	entry := &EpisodicMemoryEntry{
		EntryID:        "mem_" + HashKey(incidentID)[:12],
		IncidentID:     incidentID,
		OutcomeType:    outcomeType,
		HashedEntities: hashed,
		ActionTaken:    actionTaken,
		LessonsLearned: lessonsLearned,
		RecordedAt:     now,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.history[incidentID] = entry
	return entry
}

// QueryPastIncident retrieves retrospective lessons from historical incidents.
func (m *CollectiveAgentMemory) QueryPastIncident(incidentID string) (*EpisodicMemoryEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.history[incidentID]
	return entry, exists
}
