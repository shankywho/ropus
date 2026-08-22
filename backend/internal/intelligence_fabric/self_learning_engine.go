package intelligence_fabric

import (
	"sync"
	"time"
)

// DefenseOutcomeRecord documents the actual measured effectiveness of an applied containment control.
type DefenseOutcomeRecord struct {
	IncidentID          string    `json:"incident_id"`
	LossPreventedUSD    float64   `json:"loss_prevented_usd"`
	IsFalsePositive     bool      `json:"is_false_positive"`
	StrategyAdaptationDelta float64 `json:"strategy_adaptation_delta"`
	MeasuredAt          time.Time `json:"measured_at"`
}

// SelfLearningDefenseLoop implements closed-loop adaptation to continuously improve future defense models.
type SelfLearningDefenseLoop struct {
	mu                  sync.RWMutex
	outcomes            map[string]*DefenseOutcomeRecord
	cumulativeLossPrevented float64
	falsePositiveCount  int
	truePositiveCount   int
}

// NewSelfLearningDefenseLoop initializes the self-learning defense loop.
func NewSelfLearningDefenseLoop() *SelfLearningDefenseLoop {
	return &SelfLearningDefenseLoop{
		outcomes: make(map[string]*DefenseOutcomeRecord),
	}
}

// RecordDefenseOutcome evaluates the outcome of an applied defense and updates systemic learning metrics.
func (l *SelfLearningDefenseLoop) RecordDefenseOutcome(incidentID string, lossPreventedUSD float64, isFalsePositive bool) *DefenseOutcomeRecord {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	delta := 0.05
	if isFalsePositive {
		l.falsePositiveCount++
		delta = -0.05
	} else {
		l.truePositiveCount++
		l.cumulativeLossPrevented += lossPreventedUSD
	}

	rec := &DefenseOutcomeRecord{
		IncidentID:              incidentID,
		LossPreventedUSD:        lossPreventedUSD,
		IsFalsePositive:         isFalsePositive,
		StrategyAdaptationDelta: delta,
		MeasuredAt:              now,
	}

	l.outcomes[incidentID] = rec
	return rec
}

// GetLearningMetrics returns the current system accuracy and learning stats.
func (l *SelfLearningDefenseLoop) GetLearningMetrics() (float64, float64, int, int) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	total := l.truePositiveCount + l.falsePositiveCount
	fpRate := 0.0
	if total > 0 {
		fpRate = float64(l.falsePositiveCount) / float64(total)
	}

	return l.cumulativeLossPrevented, fpRate, l.truePositiveCount, l.falsePositiveCount
}
