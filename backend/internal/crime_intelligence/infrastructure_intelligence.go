package crime_intelligence

import (
	"sync"
	"time"
)

// InfrastructureRecord documents a monitored adversary tool, proxy pool, or bot cluster.
type InfrastructureRecord struct {
	HashedFingerprint string    `json:"hashed_fingerprint"`
	InfrastructureType string   `json:"infrastructure_type"` // "RESIDENTIAL_PROXY_POOL", "EMULATOR_FARM", "BOTNET_C2_DOMAIN"
	ReputationScore   float64   `json:"reputation_score"`   // 0.0 (clean) to 1.0 (malicious)
	IsActivelyHostile bool      `json:"is_actively_hostile"`
	AssociatedGroups  []string  `json:"associated_groups"`
	LastSeenActive    time.Time `json:"last_seen_active"`
}

// InfrastructureIntelligenceEngine tracks and scores malicious underground infrastructure.
type InfrastructureIntelligenceEngine struct {
	mu      sync.RWMutex
	records map[string]*InfrastructureRecord
}

// NewInfrastructureIntelligenceEngine initializes the infrastructure intelligence engine.
func NewInfrastructureIntelligenceEngine() *InfrastructureIntelligenceEngine {
	return &InfrastructureIntelligenceEngine{
		records: make(map[string]*InfrastructureRecord),
	}
}

// RecordInfrastructure registers or updates a malicious infrastructure node.
func (e *InfrastructureIntelligenceEngine) RecordInfrastructure(rawIdentifier, infraType, syndicateGroup string, hostile bool) *InfrastructureRecord {
	hashed := HashID(rawIdentifier)
	now := time.Now().UTC()

	e.mu.Lock()
	defer e.mu.Unlock()

	rec, exists := e.records[hashed]
	if !exists {
		score := 0.50
		if hostile {
			score = 0.98
		}
		rec = &InfrastructureRecord{
			HashedFingerprint:  hashed,
			InfrastructureType: infraType,
			ReputationScore:    score,
			IsActivelyHostile:  hostile,
			AssociatedGroups:   []string{syndicateGroup},
			LastSeenActive:     now,
		}
		e.records[hashed] = rec
		return rec
	}

	rec.LastSeenActive = now
	if hostile {
		rec.IsActivelyHostile = true
		rec.ReputationScore = 0.99
	}
	return rec
}

// QueryReputation checks if a device or IP is associated with hostile crime infrastructure.
func (e *InfrastructureIntelligenceEngine) QueryReputation(rawIdentifier string) (float64, bool) {
	hashed := HashID(rawIdentifier)

	e.mu.RLock()
	defer e.mu.RUnlock()

	rec, exists := e.records[hashed]
	if !exists {
		return 0.05, false
	}
	return rec.ReputationScore, rec.IsActivelyHostile
}
