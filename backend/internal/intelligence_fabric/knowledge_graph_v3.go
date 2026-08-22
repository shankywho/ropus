package intelligence_fabric

import (
	"fmt"
	"sync"
	"time"
)

// GraphV3EntityType classifies nodes in Knowledge Graph 3.0.
type GraphV3EntityType string

const (
	TypeThreatActor     GraphV3EntityType = "THREAT_ACTOR"
	TypeCampaign        GraphV3EntityType = "CAMPAIGN"
	TypeTechnique       GraphV3EntityType = "TECHNIQUE"
	TypeInfrastructure  GraphV3EntityType = "INFRASTRUCTURE"
	TypeVictimPattern   GraphV3EntityType = "VICTIM_PATTERN"
	TypeFinancialFlow   GraphV3EntityType = "FINANCIAL_FLOW"
	TypeDefensePattern  GraphV3EntityType = "DEFENSE_PATTERN"
	TypeRegulatoryEvent GraphV3EntityType = "REGULATORY_EVENT"
)

// KnowledgeNodeV3 represents an evolved intelligence node.
type KnowledgeNodeV3 struct {
	NodeID      string            `json:"node_id"`
	Type        GraphV3EntityType `json:"type"`
	HashedKey   string            `json:"hashed_key"`
	ThreatScore float64           `json:"threat_score"`
	Version     int               `json:"version"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ThreatKnowledgeGraphV3 provides an always-evolving dynamic knowledge graph.
type ThreatKnowledgeGraphV3 struct {
	mu    sync.RWMutex
	nodes map[string]*KnowledgeNodeV3
	edges map[string][]string // nodeID -> []connectedNodeIDs
}

// NewThreatKnowledgeGraphV3 initializes Knowledge Graph 3.0.
func NewThreatKnowledgeGraphV3() *ThreatKnowledgeGraphV3 {
	return &ThreatKnowledgeGraphV3{
		nodes: make(map[string]*KnowledgeNodeV3),
		edges: make(map[string][]string),
	}
}

// IngestSignalAndEvolve applies the 4-stage evolution lifecycle:
// NEW SIGNAL -> ENTITY UPDATE -> RELATIONSHIP CHANGE -> THREAT SCORE UPDATE.
func (g *ThreatKnowledgeGraphV3) IngestSignalAndEvolve(sig *IntelligenceSignal, nodeType GraphV3EntityType, connectedNodeID string) *KnowledgeNodeV3 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UTC()
	hashPrefix := sig.PrivacyHash
	if len(hashPrefix) > 12 {
		hashPrefix = hashPrefix[:12]
	}
	nodeKey := fmt.Sprintf("%s_%s", nodeType, hashPrefix)

	node, exists := g.nodes[nodeKey]
	if !exists {
		// Stage 1 & 2: New Signal -> Entity Creation
		node = &KnowledgeNodeV3{
			NodeID:      nodeKey,
			Type:        nodeType,
			HashedKey:   sig.PrivacyHash,
			ThreatScore: sig.Confidence * sig.ReliabilityScore,
			Version:     1,
			UpdatedAt:   now,
		}
		g.nodes[nodeKey] = node
	} else {
		// Stage 2: Entity Update
		node.Version++
		node.ThreatScore = (node.ThreatScore + (sig.Confidence * sig.ReliabilityScore)) / 2.0
		node.UpdatedAt = now
	}

	// Stage 3: Relationship Change
	if connectedNodeID != "" {
		g.edges[nodeKey] = appendUnique(g.edges[nodeKey], connectedNodeID)
		g.edges[connectedNodeID] = appendUnique(g.edges[connectedNodeID], nodeKey)
	}

	return node
}

// GetNode fetches a node.
func (g *ThreatKnowledgeGraphV3) GetNode(nodeKey string) (*KnowledgeNodeV3, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[nodeKey]
	return node, exists
}

// QueryConnected retrieves linked entities.
func (g *ThreatKnowledgeGraphV3) QueryConnected(nodeKey string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	conns := g.edges[nodeKey]
	res := make([]string, len(conns))
	copy(res, conns)
	return res
}

func appendUnique(slice []string, val string) []string {
	for _, s := range slice {
		if s == val {
			return slice
		}
	}
	return append(slice, val)
}
