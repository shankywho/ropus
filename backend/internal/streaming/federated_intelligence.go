package streaming

import (
	"sync"
	"time"
)

// FederatedThreatSignature represents an anonymized attack signature shared across participants.
type FederatedThreatSignature struct {
	SignatureID            string    `json:"signature_id"`
	HashedPattern          string    `json:"hashed_pattern"`
	ThreatType             string    `json:"threat_type"` // "EMULATOR_FARM", "BOT_HEADER_ANOMALY", "PROXY_SUBNET"
	Confidence             float64   `json:"confidence"`
	ContributingPeersCount int       `json:"contributing_peers_count"`
	DisseminatedAt         time.Time `json:"disseminated_at"`
}

// FederatedIntelligenceMesh enables privacy-preserving collaborative threat intelligence exchange.
type FederatedIntelligenceMesh struct {
	mu         sync.RWMutex
	signatures map[string]*FederatedThreatSignature
}

// NewFederatedIntelligenceMesh initializes the federated sharing mesh.
func NewFederatedIntelligenceMesh() *FederatedIntelligenceMesh {
	return &FederatedIntelligenceMesh{
		signatures: make(map[string]*FederatedThreatSignature),
	}
}

// BroadcastSignature shares an anonymized threat signature across the network.
func (m *FederatedIntelligenceMesh) BroadcastSignature(rawPattern, threatType string, confidence float64) *FederatedThreatSignature {
	hashed := HashIdentifier(rawPattern)
	now := time.Now().UTC()

	m.mu.Lock()
	defer m.mu.Unlock()

	sig, exists := m.signatures[hashed]
	if !exists {
		sig = &FederatedThreatSignature{
			SignatureID:            "sig_" + hashed[:16],
			HashedPattern:          hashed,
			ThreatType:             threatType,
			Confidence:             confidence,
			ContributingPeersCount: 1,
			DisseminatedAt:         now,
		}
		m.signatures[hashed] = sig
		return sig
	}

	sig.ContributingPeersCount++
	sig.Confidence = (sig.Confidence + confidence) / 2.0
	sig.DisseminatedAt = now
	return sig
}

// QuerySignature checks whether an incoming transaction pattern is recognized in the federated network.
func (m *FederatedIntelligenceMesh) QuerySignature(rawPattern string) (*FederatedThreatSignature, bool) {
	hashed := HashIdentifier(rawPattern)

	m.mu.RLock()
	defer m.mu.RUnlock()

	sig, exists := m.signatures[hashed]
	return sig, exists
}
