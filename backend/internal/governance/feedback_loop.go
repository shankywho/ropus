package governance

import (
	"fmt"
	"sync"
	"time"
)

// FeedbackOutcome aggregates ground truth outcomes for active learning.
type FeedbackOutcome struct {
	FeedbackID        string    `json:"feedback_id"`
	TransactionID     string    `json:"transaction_id"`
	OriginalScore     float64   `json:"original_score"`
	OriginalDecision  string    `json:"original_decision"`
	AnalystDecision   string    `json:"analyst_decision,omitempty"`
	ChargebackOutcome bool      `json:"chargeback_outcome"`
	IsFalsePositive   bool      `json:"is_false_positive"`
	IsFalseNegative   bool      `json:"is_false_negative"`
	ConfirmedFraud    bool      `json:"confirmed_fraud"`
	RecordedAt        time.Time `json:"recorded_at"`
}

// FeedbackLearningLoop stores verified labels to continuously curate gold retraining datasets.
type FeedbackLearningLoop struct {
	mu       sync.RWMutex
	outcomes map[string]*FeedbackOutcome
}

// NewFeedbackLearningLoop initializes the ground truth feedback repository.
func NewFeedbackLearningLoop() *FeedbackLearningLoop {
	return &FeedbackLearningLoop{
		outcomes: make(map[string]*FeedbackOutcome),
	}
}

// IngestAnalystFeedback converts a resolved manual review into a verified learning sample.
func (f *FeedbackLearningLoop) IngestAnalystFeedback(txnID string, origScore float64, origDecision, analystDecision string) *FeedbackOutcome {
	f.mu.Lock()
	defer f.mu.Unlock()

	confirmedFraud := analystDecision == "CONFIRM_FRAUD"
	isFP := origDecision == "BLOCK" && analystDecision == "FALSE_POSITIVE"
	isFN := origDecision == "ALLOW" && analystDecision == "CONFIRM_FRAUD"

	outcome := &FeedbackOutcome{
		FeedbackID:        fmt.Sprintf("fb_%d_%s", time.Now().UnixNano(), txnID),
		TransactionID:     txnID,
		OriginalScore:     origScore,
		OriginalDecision:  origDecision,
		AnalystDecision:   analystDecision,
		ChargebackOutcome: confirmedFraud,
		IsFalsePositive:   isFP,
		IsFalseNegative:   isFN,
		ConfirmedFraud:    confirmedFraud,
		RecordedAt:        time.Now().UTC(),
	}

	f.outcomes[txnID] = outcome
	return outcome
}

// GetGoldTrainingSamples returns structured ground truth data for the feature store retraining loop.
func (f *FeedbackLearningLoop) GetGoldTrainingSamples() []*FeedbackOutcome {
	f.mu.RLock()
	defer f.mu.RUnlock()

	res := make([]*FeedbackOutcome, 0, len(f.outcomes))
	for _, o := range f.outcomes {
		res = append(res, o)
	}
	return res
}
