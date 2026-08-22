package intelligence_fabric

import (
	"math"
	"time"
)

// UnifiedThreatPicture represents the holistic synthesized intelligence assessment.
type UnifiedThreatPicture struct {
	ThreatLevel         string    `json:"threat_level"` // "CRITICAL", "ELEVATED", "GUARDED", "NOMINAL"
	ActiveCampaigns     []string  `json:"active_campaigns"`
	CompositeConfidence float64   `json:"composite_confidence"`
	AffectedEntities    []string  `json:"affected_entities"`
	RecommendedActions  []string  `json:"recommended_actions"`
	SignalsFusedCount   int       `json:"signals_fused_count"`
	FusedAt             time.Time `json:"fused_at"`
}

// IntelligenceFusionEngine combines multi-source signals into a single actionable operational picture.
type IntelligenceFusionEngine struct{}

// NewIntelligenceFusionEngine initializes the fusion engine.
func NewIntelligenceFusionEngine() *IntelligenceFusionEngine {
	return &IntelligenceFusionEngine{}
}

// FuseTelemetry synthesizes raw signals, graph updates, and analyst feedback into the Unified Threat Picture.
func (f *IntelligenceFusionEngine) FuseTelemetry(signals []*IntelligenceSignal) *UnifiedThreatPicture {
	now := time.Now().UTC()

	if len(signals) == 0 {
		return &UnifiedThreatPicture{
			ThreatLevel:         "NOMINAL",
			CompositeConfidence: 0.99,
			FusedAt:             now,
		}
	}

	confidenceSum := 0.0
	highThreatSignals := 0
	var entities []string
	var campaigns []string

	for _, s := range signals {
		confidenceSum += s.Confidence * s.ReliabilityScore
		if s.Confidence >= 0.85 {
			highThreatSignals++
		}
		hashKey := s.PrivacyHash
		if len(hashKey) > 12 {
			hashKey = hashKey[:12]
		}
		entities = appendUnique(entities, hashKey)
		if s.RawTopic != "" {
			campaigns = appendUnique(campaigns, s.RawTopic)
		}
	}

	avgConf := confidenceSum / float64(len(signals))
	threatLevel := "GUARDED"
	var actions []string

	if highThreatSignals >= 5 || avgConf >= 0.88 {
		threatLevel = "CRITICAL"
		actions = []string{
			"Activate autonomous proxy subnet containment",
			"Mandate hardware attestation MFA on high-value checkout",
			"Broadcast anonymized threat signatures to consortium peers",
		}
	} else if highThreatSignals >= 2 || avgConf >= 0.60 {
		threatLevel = "ELEVATED"
		actions = []string{
			"Increase velocity sampling resolution",
			"Deploy shadow evaluation rules for suspicious device clusters",
		}
	} else {
		actions = []string{"Maintain standard surveillance baselines"}
	}

	return &UnifiedThreatPicture{
		ThreatLevel:         threatLevel,
		ActiveCampaigns:     campaigns,
		CompositeConfidence: math.Min(0.99, avgConf),
		AffectedEntities:    entities,
		RecommendedActions:  actions,
		SignalsFusedCount:   len(signals),
		FusedAt:             now,
	}
}
