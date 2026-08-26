package graphsage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// EvidenceLedgerEntry holds an immutable record of an investigation dossier in Go.
type EvidenceLedgerEntry struct {
	EntryID           string                 `json:"entry_id"`
	DossierID         string                 `json:"dossier_id"`
	RecordedAt        time.Time              `json:"recorded_at"`
	EmployeeID        string                 `json:"employee_id"`
	EmployeeRole      string                 `json:"employee_role"`
	OverallRiskScore  float64                `json:"overall_risk_score"`
	RiskLevel         RelationshipRiskLevel  `json:"risk_level"`
	Payload           map[string]interface{} `json:"payload"`
	ProvenanceHash    string                 `json:"provenance_hash"`
}

// ShadowEvidenceLedger provides an append-only thread-safe evidence ledger in Go.
type ShadowEvidenceLedger struct {
	mu      sync.RWMutex
	entries []*EvidenceLedgerEntry
	index   map[string]*EvidenceLedgerEntry
}

// NewShadowEvidenceLedger creates a new evidence ledger.
func NewShadowEvidenceLedger() *ShadowEvidenceLedger {
	return &ShadowEvidenceLedger{
		entries: make([]*EvidenceLedgerEntry, 0),
		index:   make(map[string]*EvidenceLedgerEntry),
	}
}

// RecordDossier appends a dossier to the audit ledger.
func (l *ShadowEvidenceLedger) RecordDossier(dossier *CollusionInvestigationDossier) *EvidenceLedgerEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	b, _ := json.Marshal(dossier)
	h := sha256.Sum256(b)
	provHash := hex.EncodeToString(h[:])

	entryID := fmt.Sprintf("ev_go_%06d_%s", len(l.entries)+1, provHash[:8])

	entry := &EvidenceLedgerEntry{
		EntryID:          entryID,
		DossierID:        dossier.DossierID,
		RecordedAt:       now,
		EmployeeID:       dossier.EmployeeID,
		EmployeeRole:     dossier.EmployeeRole,
		OverallRiskScore: dossier.OverallRiskScore,
		RiskLevel:        dossier.RiskLevel,
		ProvenanceHash:   provHash,
	}

	l.entries = append(l.entries, entry)
	l.index[dossier.DossierID] = entry
	return entry
}

// TotalEntries returns count of recorded dossiers.
func (l *ShadowEvidenceLedger) TotalEntries() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}
