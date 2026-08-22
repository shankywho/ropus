package crime_intelligence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// CrimeIntelligenceGraph maintains global threat actor networks, campaign genealogies, and money flows.
type CrimeIntelligenceGraph struct {
	mu       sync.RWMutex
	nodes    map[string]*CrimeNode
	outEdges map[string][]*CrimeEdge
	inEdges  map[string][]*CrimeEdge
}

// NewCrimeIntelligenceGraph initializes the crime intelligence graph.
func NewCrimeIntelligenceGraph() *CrimeIntelligenceGraph {
	return &CrimeIntelligenceGraph{
		nodes:    make(map[string]*CrimeNode),
		outEdges: make(map[string][]*CrimeEdge),
		inEdges:  make(map[string][]*CrimeEdge),
	}
}

// HashID produces a deterministic SHA-256 hash string.
func HashID(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// AddNode registers a crime entity in the knowledge graph.
func (g *CrimeIntelligenceGraph) AddNode(node *CrimeNode) error {
	if node == nil || node.EntityID == "" {
		return fmt.Errorf("node or entityID cannot be empty")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if node.FirstSeenAt.IsZero() {
		node.FirstSeenAt = time.Now().UTC()
	}
	if node.LastSeenAt.IsZero() {
		node.LastSeenAt = time.Now().UTC()
	}

	g.nodes[node.EntityID] = node
	return nil
}

// AddEdge records an intelligence linkage between two crime entities.
func (g *CrimeIntelligenceGraph) AddEdge(edge *CrimeEdge) error {
	if edge == nil || edge.SourceID == "" || edge.TargetID == "" {
		return fmt.Errorf("invalid edge endpoints")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if edge.EdgeID == "" {
		edge.EdgeID = fmt.Sprintf("edge_%s_%s_%s", edge.SourceID, edge.Type, edge.TargetID)
	}
	if edge.ObservedAt.IsZero() {
		edge.ObservedAt = time.Now().UTC()
	}

	g.outEdges[edge.SourceID] = append(g.outEdges[edge.SourceID], edge)
	g.inEdges[edge.TargetID] = append(g.inEdges[edge.TargetID], edge)
	return nil
}

// GetNode retrieves a crime entity by ID.
func (g *CrimeIntelligenceGraph) GetNode(entityID string) (*CrimeNode, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[entityID]
	return node, exists
}

// QueryNeighbors retrieves adjacent threat nodes matching an optional relationship filter.
func (g *CrimeIntelligenceGraph) QueryNeighbors(entityID string, relType CrimeRelationshipType) []*CrimeNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var res []*CrimeNode
	seen := make(map[string]bool)

	// Traverse outgoing edges
	for _, e := range g.outEdges[entityID] {
		if relType == "" || e.Type == relType {
			if target, exists := g.nodes[e.TargetID]; exists && !seen[target.EntityID] {
				seen[target.EntityID] = true
				res = append(res, target)
			}
		}
	}

	// Traverse incoming edges
	for _, e := range g.inEdges[entityID] {
		if relType == "" || e.Type == relType {
			if src, exists := g.nodes[e.SourceID]; exists && !seen[src.EntityID] {
				seen[src.EntityID] = true
				res = append(res, src)
			}
		}
	}

	return res
}

// GetCampaignLineage traces the ancestor and descendant evolution path of a campaign.
func (g *CrimeIntelligenceGraph) GetCampaignLineage(campaignID string) []*CrimeNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var lineage []*CrimeNode
	visited := make(map[string]bool)

	var dfs func(currID string)
	dfs = func(currID string) {
		if visited[currID] {
			return
		}
		visited[currID] = true
		if node, exists := g.nodes[currID]; exists {
			lineage = append(lineage, node)
		}
		// Follow EVOLVED_FROM edges
		for _, e := range g.inEdges[currID] {
			if e.Type == RelEvolvedFrom {
				dfs(e.SourceID)
			}
		}
		for _, e := range g.outEdges[currID] {
			if e.Type == RelEvolvedFrom {
				dfs(e.TargetID)
			}
		}
	}

	dfs(campaignID)
	return lineage
}
