package governance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// DecisionAuditRecord represents a single immutable decision entry in the cryptographic audit chain.
type DecisionAuditRecord struct {
	Index           int64     `json:"index"`
	DecisionID      string    `json:"decision_id"`
	RequestHash     string    `json:"request_hash"`
	ModelVersion    string    `json:"model_version"`
	FeatureContract string    `json:"feature_contract"`
	RiskScore       float64   `json:"risk_score"`
	Decision        string    `json:"decision"`
	ExplanationID   string    `json:"explanation_id"`
	PolicyVersion   string    `json:"policy_version"`
	Timestamp       time.Time `json:"timestamp"`
	PreviousHash    string    `json:"previous_hash"`
	RecordHash      string    `json:"record_hash"`
}

// ComputeDecisionHash calculates SHA256(prevHash + ":" + reqHash + ":" + payloadString).
func ComputeDecisionHash(prevHash, reqHash, decisionID, decision string, score float64, ts time.Time) string {
	raw := fmt.Sprintf("%s:%s:%s:%s:%.4f:%d", prevHash, reqHash, decisionID, decision, score, ts.UnixNano())
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// DecisionAuditTrail maintains a tamper-evident sequential hash chain of all risk decisions.
type DecisionAuditTrail struct {
	mu          sync.RWMutex
	records     []*DecisionAuditRecord
	currentHash string
}

// NewDecisionAuditTrail initializes the cryptographic audit ledger with a genesis block.
func NewDecisionAuditTrail() *DecisionAuditTrail {
	genesisHash := "0000000000000000000000000000000000000000000000000000000000000000"
	return &DecisionAuditTrail{
		records:     make([]*DecisionAuditRecord, 0),
		currentHash: genesisHash,
	}
}

// AppendDecision securely writes a decision into the hash chain.
func (t *DecisionAuditTrail) AppendDecision(decisionID, reqHash, modelVer, featureContract, decision, explanationID, policyVer string, score float64) *DecisionAuditRecord {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now().UTC()
	prev := t.currentHash
	hash := ComputeDecisionHash(prev, reqHash, decisionID, decision, score, now)

	rec := &DecisionAuditRecord{
		Index:           int64(len(t.records)),
		DecisionID:      decisionID,
		RequestHash:     reqHash,
		ModelVersion:    modelVer,
		FeatureContract: featureContract,
		RiskScore:       score,
		Decision:        decision,
		ExplanationID:   explanationID,
		PolicyVersion:   policyVer,
		Timestamp:       now,
		PreviousHash:    prev,
		RecordHash:      hash,
	}

	t.records = append(t.records, rec)
	t.currentHash = hash
	return rec
}

// VerifyIntegrity traverses the entire audit trail and ensures no records have been altered or deleted.
func (t *DecisionAuditTrail) VerifyIntegrity() (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.records) == 0 {
		return true, nil
	}

	expectedPrev := "0000000000000000000000000000000000000000000000000000000000000000"
	for i, rec := range t.records {
		if rec.PreviousHash != expectedPrev {
			return false, fmt.Errorf("audit chain broken at index %d: expected previous hash %s, got %s", i, expectedPrev, rec.PreviousHash)
		}

		recomputed := ComputeDecisionHash(rec.PreviousHash, rec.RequestHash, rec.DecisionID, rec.Decision, rec.RiskScore, rec.Timestamp)
		if recomputed != rec.RecordHash {
			return false, fmt.Errorf("audit record tampered at index %d: expected hash %s, got %s", i, recomputed, rec.RecordHash)
		}

		expectedPrev = rec.RecordHash
	}

	return true, nil
}

// Count returns the number of audited records in the chain.
func (t *DecisionAuditTrail) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.records)
}
