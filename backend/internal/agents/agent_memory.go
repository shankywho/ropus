package agents

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// MemoryFact represents an individual granular fact or forensic proof item in agent memory.
type MemoryFact struct {
	FactID     string    `json:"fact_id"`
	Subject    string    `json:"subject"` // Hashed entity ID
	Predicate  string    `json:"predicate"` // e.g. "CONNECTED_TO_FRAUD_RING", "UNRECOGNIZED_DEVICE"
	Object     string    `json:"object"`
	Confidence float64   `json:"confidence"`
	LearnedAt  time.Time `json:"learned_at"`
}

// AgentMemory maintains short-term working context and long-term historical fraud patterns.
type AgentMemory struct {
	mu              sync.RWMutex
	shortTermFacts  map[string][]MemoryFact // traceID -> facts
	longTermPatterns map[string]float64     // hashedPattern -> confidence score
}

// NewAgentMemory initializes the dual-tiered agent memory system.
func NewAgentMemory() *AgentMemory {
	return &AgentMemory{
		shortTermFacts:   make(map[string][]MemoryFact),
		longTermPatterns: make(map[string]float64),
	}
}

// HashIdentifier safely transforms raw PII strings into SHA-256 hashes.
func HashIdentifier(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// RememberShortTerm records a fact into an ongoing investigation's short-term working memory.
func (m *AgentMemory) RememberShortTerm(traceID, subjectRaw, predicate, object string, confidence float64) {
	hashedSubject := HashIdentifier(subjectRaw)
	fact := MemoryFact{
		FactID:     "fct_" + hashedSubject[:12],
		Subject:    hashedSubject,
		Predicate:  predicate,
		Object:     object,
		Confidence: confidence,
		LearnedAt:  time.Now().UTC(),
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.shortTermFacts[traceID] = append(m.shortTermFacts[traceID], fact)
}

// RecallShortTerm retrieves all active working facts for an investigation trace.
func (m *AgentMemory) RecallShortTerm(traceID string) []MemoryFact {
	m.mu.RLock()
	defer m.mu.RUnlock()

	facts := m.shortTermFacts[traceID]
	res := make([]MemoryFact, len(facts))
	copy(res, facts)
	return res
}

// RememberLongTerm stores a verified fraud attack pattern for collective recall across all agents.
func (m *AgentMemory) RememberLongTerm(patternKey string, confidence float64) {
	hashedKey := HashIdentifier(patternKey)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.longTermPatterns[hashedKey] = confidence
}

// RecallLongTerm checks if a pattern has been historically observed and verified.
func (m *AgentMemory) RecallLongTerm(patternKey string) (float64, bool) {
	hashedKey := HashIdentifier(patternKey)

	m.mu.RLock()
	defer m.mu.RUnlock()
	conf, exists := m.longTermPatterns[hashedKey]
	return conf, exists
}
