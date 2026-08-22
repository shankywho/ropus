package intelligence_fabric

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// SignalSource categorizes the origin of an inbound intelligence signal.
type SignalSource string

const (
	SourceFraudEngine      SignalSource = "FRAUD_ENGINE"
	SourceThreatFeed       SignalSource = "THREAT_FEED"
	SourceGraphEvolution   SignalSource = "GRAPH_EVOLUTION"
	SourceAnalystFeedback  SignalSource = "ANALYST_FEEDBACK"
	SourceConsortiumPeer   SignalSource = "CONSORTIUM_PEER"
	SourceRedTeamSimulator SignalSource = "RED_TEAM_SIMULATOR"
)

// IntelligenceSignal represents an atomic, normalized telemetry signal in the intelligence fabric.
type IntelligenceSignal struct {
	SignalID         string                 `json:"signal_id"`
	Source           SignalSource           `json:"source"`
	Confidence       float64                `json:"confidence"`
	ReliabilityScore float64                `json:"reliability_score"` // 0.0 to 1.0 (Source trustworthiness)
	PrivacyHash      string                 `json:"privacy_hash"`      // SHA-256 hashed subject identifier
	RawTopic         string                 `json:"raw_topic"`
	Payload          map[string]interface{} `json:"payload"`
	Timestamp        time.Time              `json:"timestamp"`
}

// ComputePrivacyHash creates a deterministic SHA-256 hash.
func ComputePrivacyHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
