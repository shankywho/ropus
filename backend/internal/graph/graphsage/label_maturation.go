package graphsage

import (
	"sync"
	"time"
)

// LabelMaturityRecord tracks the lifecycle of an individual entity label in Go.
type LabelMaturityRecord struct {
	RecordID               string             `json:"record_id"`
	EntityID               string             `json:"entity_id"`
	EventTimestamp         time.Time          `json:"event_timestamp"`
	State                  LabelMaturityState `json:"state"`
	InvestigationOpenedAt  time.Time          `json:"investigation_opened_at,omitempty"`
	InvestigationClosedAt  time.Time          `json:"investigation_closed_at,omitempty"`
	LabelTimestamp         time.Time          `json:"label_timestamp,omitempty"`
	LabelSource            string             `json:"label_source"`
	LabelConfidence        float64            `json:"label_confidence"`
	CaseID                 string             `json:"case_id,omitempty"`
	IsCollusion            bool               `json:"is_collusion"`
}

// LabelMaturationEngine coordinates dispute windows and ground-truth maturation in Go.
type LabelMaturationEngine struct {
	mu                 sync.RWMutex
	disputeWindowDays  int
	records            map[string]*LabelMaturityRecord
}

// NewLabelMaturationEngine initializes the maturation lifecycle engine.
func NewLabelMaturationEngine(disputeWindowDays int) *LabelMaturationEngine {
	if disputeWindowDays <= 0 {
		disputeWindowDays = 90
	}
	return &LabelMaturationEngine{
		disputeWindowDays: disputeWindowDays,
		records:           make(map[string]*LabelMaturityRecord),
	}
}

// RegisterTransaction records a fresh transaction entering dispute maturation.
func (e *LabelMaturationEngine) RegisterTransaction(entityID string, eventTS time.Time) *LabelMaturityRecord {
	e.mu.Lock()
	defer e.mu.Unlock()

	rec := &LabelMaturityRecord{
		RecordID:       "mat_" + entityID,
		EntityID:       entityID,
		EventTimestamp: eventTS,
		State:          LabelUnlabeled,
		LabelSource:    "PENDING",
	}
	e.records[entityID] = rec
	return rec
}

// ConfirmLabel updates the maturity state to mature confirmation.
func (e *LabelMaturationEngine) ConfirmLabel(
	entityID string,
	state LabelMaturityState,
	source string,
	confirmedAt time.Time,
	isCollusion bool,
) *LabelMaturityRecord {
	e.mu.Lock()
	defer e.mu.Unlock()

	rec, ok := e.records[entityID]
	if !ok {
		rec = &LabelMaturityRecord{
			RecordID:       "mat_" + entityID,
			EntityID:       entityID,
			EventTimestamp: confirmedAt.AddDate(0, 0, -e.disputeWindowDays),
		}
		e.records[entityID] = rec
	}

	rec.State = state
	rec.LabelSource = source
	rec.LabelTimestamp = confirmedAt
	rec.InvestigationClosedAt = confirmedAt
	rec.IsCollusion = isCollusion
	rec.LabelConfidence = 1.0
	return rec
}

// GetConfirmedCollusionCount returns the number of confirmed real internal collusion labels.
func (e *LabelMaturationEngine) GetConfirmedCollusionCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	count := 0
	for _, r := range e.records {
		if r.State == LabelConfirmedFraud && r.IsCollusion {
			count++
		}
	}
	return count
}
