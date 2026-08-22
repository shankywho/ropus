package streaming

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// GlobalGraphNode stores privacy-preserved SHA-256 identifiers across financial institutions.
type GlobalGraphNode struct {
	HashedEntityID string    `json:"hashed_entity_id"`
	EntityType     string    `json:"entity_type"` // "HASHED_USER", "DEVICE_FINGERPRINT", "MERCHANT_ID", "IP_SUBNET"
	GlobalRiskScore float64  `json:"global_risk_score"`
	TenantsSeen    []string  `json:"tenants_seen"`
	IsKnownBad     bool      `json:"is_known_bad"`
	FirstSeenAt    time.Time `json:"first_seen_at"`
	LastSeenAt     time.Time `json:"last_seen_at"`
}

// GlobalFraudGraph provides cross-tenant collective intelligence without disclosing raw customer PII.
type GlobalFraudGraph struct {
	mu    sync.RWMutex
	nodes map[string]*GlobalGraphNode
}

// NewGlobalFraudGraph initializes the privacy-preserving cross-tenant graph.
func NewGlobalFraudGraph() *GlobalFraudGraph {
	return &GlobalFraudGraph{
		nodes: make(map[string]*GlobalGraphNode),
	}
}

// HashIdentifier produces a deterministic SHA-256 hash of raw entity strings.
func HashIdentifier(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// RecordGlobalSignal links cross-tenant risk indicators securely.
func (g *GlobalFraudGraph) RecordGlobalSignal(tenantID, entityType, rawEntityID string, isFraud bool, riskScore float64) *GlobalGraphNode {
	hashedID := HashIdentifier(rawEntityID)
	now := time.Now().UTC()

	g.mu.Lock()
	defer g.mu.Unlock()

	node, exists := g.nodes[hashedID]
	if !exists {
		node = &GlobalGraphNode{
			HashedEntityID:  hashedID,
			EntityType:      entityType,
			GlobalRiskScore: riskScore,
			TenantsSeen:     []string{tenantID},
			IsKnownBad:      isFraud,
			FirstSeenAt:     now,
			LastSeenAt:      now,
		}
		g.nodes[hashedID] = node
		return node
	}

	// Update existing record
	node.LastSeenAt = now
	if isFraud {
		node.IsKnownBad = true
		node.GlobalRiskScore = 0.99
	}
	tenantFound := false
	for _, t := range node.TenantsSeen {
		if t == tenantID {
			tenantFound = true
			break
		}
	}
	if !tenantFound {
		node.TenantsSeen = append(node.TenantsSeen, tenantID)
	}

	return node
}

// QueryGlobalReputation checks whether an entity is compromised across the global network.
func (g *GlobalFraudGraph) QueryGlobalReputation(rawEntityID string) (float64, bool, int) {
	hashedID := HashIdentifier(rawEntityID)

	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[hashedID]
	if !exists {
		return 0.05, false, 0
	}
	return node.GlobalRiskScore, node.IsKnownBad, len(node.TenantsSeen)
}
