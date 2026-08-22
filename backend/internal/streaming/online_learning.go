package streaming

import (
	"sync"
	"time"
)

// ProposedWeightUpdate captures an online learning proposal waiting for governance approval.
type ProposedWeightUpdate struct {
	UpdateID          string             `json:"update_id"`
	FeatureName       string             `json:"feature_name"`
	CurrentWeight     float64            `json:"current_weight"`
	ProposedWeight    float64            `json:"proposed_weight"`
	EvidenceCount     int                `json:"evidence_count"`
	ProposedAt        time.Time          `json:"proposed_at"`
	IsApproved        bool               `json:"is_approved"`
}

// OnlineLearningEngine collects real-time chargebacks and analyst verdicts to propose parameter calibrations.
type OnlineLearningEngine struct {
	mu            sync.RWMutex
	proposals     map[string]*ProposedWeightUpdate
	featureWeights map[string]float64
}

// NewOnlineLearningEngine initializes the online learning feedback engine.
func NewOnlineLearningEngine() *OnlineLearningEngine {
	weights := map[string]float64{
		"velocity_burst":   0.25,
		"device_anomaly":   0.20,
		"threat_intel_ioc": 0.35,
		"graph_risk":       0.20,
	}
	return &OnlineLearningEngine{
		proposals:      make(map[string]*ProposedWeightUpdate),
		featureWeights: weights,
	}
}

// IngestFeedback proposes weight calibrations safely without direct production overwrites.
func (e *OnlineLearningEngine) IngestFeedback(featureName string, isConfirmedFraud bool) *ProposedWeightUpdate {
	e.mu.Lock()
	defer e.mu.Unlock()

	curWeight := e.featureWeights[featureName]
	delta := 0.01
	if !isConfirmedFraud {
		delta = -0.01
	}

	propWeight := curWeight + delta
	if propWeight < 0.05 {
		propWeight = 0.05
	} else if propWeight > 0.60 {
		propWeight = 0.60
	}

	update := &ProposedWeightUpdate{
		UpdateID:       "prop_" + featureName,
		FeatureName:    featureName,
		CurrentWeight:  curWeight,
		ProposedWeight: propWeight,
		EvidenceCount:  1,
		ProposedAt:     time.Now().UTC(),
		IsApproved:     false,
	}
	e.proposals[update.UpdateID] = update
	return update
}

// ApproveProposal moves a proposed calibration into active weights following canary / governance sign-off.
func (e *OnlineLearningEngine) ApproveProposal(updateID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	p, exists := e.proposals[updateID]
	if !exists {
		return false
	}
	p.IsApproved = true
	e.featureWeights[p.FeatureName] = p.ProposedWeight
	return true
}

func (e *OnlineLearningEngine) GetWeight(featureName string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.featureWeights[featureName]
}
