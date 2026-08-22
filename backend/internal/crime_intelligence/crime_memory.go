package crime_intelligence

import (
	"sync"
	"time"
)

// CrimeEpisodicMemory stores longitudinal intelligence on threat actor maneuvers and counter-defenses.
type CrimeEpisodicMemory struct {
	MemoryID       string    `json:"memory_id"`
	SyndicateAlias string    `json:"syndicate_alias"`
	TacticObserved string    `json:"tactic_observed"`
	CounterDefense string    `json:"counter_defense"`
	OutcomeScore   float64   `json:"outcome_score"` // 1.0 = attacker completely thwarted
	RecordedAt     time.Time `json:"recorded_at"`
}

// CrimeKnowledgeMemory maintains privacy-preserving historical criminal tactics and lessons.
type CrimeKnowledgeMemory struct {
	mu      sync.RWMutex
	records map[string]*CrimeEpisodicMemory
}

// NewCrimeKnowledgeMemory initializes the crime knowledge memory.
func NewCrimeKnowledgeMemory() *CrimeKnowledgeMemory {
	return &CrimeKnowledgeMemory{
		records: make(map[string]*CrimeEpisodicMemory),
	}
}

// StoreMemory logs a tactical outcome for ongoing learning.
func (m *CrimeKnowledgeMemory) StoreMemory(rawSyndicate, tactic, defense string, outcomeScore float64) *CrimeEpisodicMemory {
	hashed := HashID(rawSyndicate)
	now := time.Now().UTC()

	mem := &CrimeEpisodicMemory{
		MemoryID:       "cmem_" + hashed[:12],
		SyndicateAlias: "Adversary_" + hashed[:8],
		TacticObserved: tactic,
		CounterDefense: defense,
		OutcomeScore:   outcomeScore,
		RecordedAt:     now,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[hashed] = mem
	return mem
}

// QueryMemory retrieves historical tactical precedent on a syndicate.
func (m *CrimeKnowledgeMemory) QueryMemory(rawSyndicate string) (*CrimeEpisodicMemory, bool) {
	hashed := HashID(rawSyndicate)

	m.mu.RLock()
	defer m.mu.RUnlock()

	mem, exists := m.records[hashed]
	return mem, exists
}
