package intelligence_fabric

import (
	"fmt"
	"sync"
	"time"
)

// HumanAIInteractionRecord documents collaboration between analysts and AI copilot agents.
type HumanAIInteractionRecord struct {
	InteractionID    string    `json:"interaction_id"`
	CaseID           string    `json:"case_id"`
	AnalystID        string    `json:"analyst_id"`
	AnalystQuery     string    `json:"analyst_query"`
	AIRecommendation string    `json:"ai_recommendation"`
	AIConfidence     float64   `json:"ai_confidence"`
	HumanApproved    bool      `json:"human_approved"`
	AnalystFeedback  string    `json:"analyst_feedback"`
	Timestamp        time.Time `json:"timestamp"`
}

// HumanAIWorkspace maintains interactive investigation sessions and captures ground-truth analyst feedback.
type HumanAIWorkspace struct {
	mu      sync.RWMutex
	history map[string]*HumanAIInteractionRecord
}

// NewHumanAIWorkspace initializes the collaboration workspace.
func NewHumanAIWorkspace() *HumanAIWorkspace {
	return &HumanAIWorkspace{
		history: make(map[string]*HumanAIInteractionRecord),
	}
}

// RecordInteraction logs an analyst interaction and determination.
func (w *HumanAIWorkspace) RecordInteraction(caseID, analystID, query, aiRec string, aiConf float64, approved bool, feedback string) *HumanAIInteractionRecord {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now().UTC()
	id := fmt.Sprintf("hai_%d_%s", now.UnixNano(), caseID)

	rec := &HumanAIInteractionRecord{
		InteractionID:    id,
		CaseID:           caseID,
		AnalystID:        analystID,
		AnalystQuery:     query,
		AIRecommendation: aiRec,
		AIConfidence:     aiConf,
		HumanApproved:    approved,
		AnalystFeedback:  feedback,
		Timestamp:        now,
	}

	w.history[id] = rec
	return rec
}

// GetInteraction retrieves an interaction record.
func (w *HumanAIWorkspace) GetInteraction(id string) (*HumanAIInteractionRecord, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	rec, exists := w.history[id]
	return rec, exists
}
