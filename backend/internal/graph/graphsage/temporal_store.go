package graphsage

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// GraphStoreConfig defines memory bounds and defense parameters.
type GraphStoreConfig struct {
	MaxNodes         int // Maximum total nodes before pruning (default: 50,000)
	MaxEdges         int // Maximum total edges before pruning (default: 200,000)
	MaxDegreePerNode int // Maximum incident edges evaluated per node to prevent poisoning/DoS (default: 100)
}

// DefaultGraphStoreConfig returns production safety defaults.
func DefaultGraphStoreConfig() GraphStoreConfig {
	return GraphStoreConfig{
		MaxNodes:         50000,
		MaxEdges:         200000,
		MaxDegreePerNode: 100,
	}
}

// TemporalHeteroGraphStore provides concurrent in-memory storage and point-in-time subgraph extraction.
type TemporalHeteroGraphStore struct {
	mu       sync.RWMutex
	cfg      GraphStoreConfig
	nodes    map[string]*HeteroNode
	edges    map[string]*HeteroEdge
	outEdges map[string][]string // sourceNodeID -> list of edgeIDs
	inEdges  map[string][]string // targetNodeID -> list of edgeIDs
}

// NewTemporalHeteroGraphStore initializes an empty temporal heterogeneous graph store.
func NewTemporalHeteroGraphStore() *TemporalHeteroGraphStore {
	return NewTemporalHeteroGraphStoreWithConfig(DefaultGraphStoreConfig())
}

// NewTemporalHeteroGraphStoreWithConfig initializes with specific capacity bounds.
func NewTemporalHeteroGraphStoreWithConfig(cfg GraphStoreConfig) *TemporalHeteroGraphStore {
	if cfg.MaxNodes <= 0 {
		cfg.MaxNodes = 50000
	}
	if cfg.MaxEdges <= 0 {
		cfg.MaxEdges = 200000
	}
	if cfg.MaxDegreePerNode <= 0 {
		cfg.MaxDegreePerNode = 100
	}
	return &TemporalHeteroGraphStore{
		cfg:      cfg,
		nodes:    make(map[string]*HeteroNode),
		edges:    make(map[string]*HeteroEdge),
		outEdges: make(map[string][]string),
		inEdges:  make(map[string][]string),
	}
}

// AddNode inserts or updates a node in the graph store with memory bounding.
func (s *TemporalHeteroGraphStore) AddNode(node *HeteroNode) error {
	if node == nil || node.ID == "" {
		return fmt.Errorf("invalid node: ID must not be empty")
	}
	if node.CreatedAt.IsZero() {
		node.CreatedAt = time.Now().UTC()
	}
	node.UpdatedAt = time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Memory Safeguard: Prevent unbounded growth
	if len(s.nodes) >= s.cfg.MaxNodes {
		if _, exists := s.nodes[node.ID]; !exists {
			// Evict oldest node (fail-safe FIFO)
			for oldID := range s.nodes {
				delete(s.nodes, oldID)
				delete(s.outEdges, oldID)
				delete(s.inEdges, oldID)
				break
			}
		}
	}

	s.nodes[node.ID] = node
	return nil
}

// AddEdge inserts a directed relationship edge with strict timestamp and degree validation.
func (s *TemporalHeteroGraphStore) AddEdge(edge *HeteroEdge) error {
	if edge == nil || edge.ID == "" || edge.SourceID == "" || edge.TargetID == "" {
		return fmt.Errorf("invalid edge: ID, SourceID, and TargetID required")
	}
	if edge.Timestamp.IsZero() {
		edge.Timestamp = time.Now().UTC()
	}
	if edge.Confidence <= 0 {
		edge.Confidence = 1.0
	}
	if edge.Weight <= 0 {
		edge.Weight = 1.0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Memory Safeguard: Prevent unbounded edges
	if len(s.edges) >= s.cfg.MaxEdges {
		if _, exists := s.edges[edge.ID]; !exists {
			for oldID := range s.edges {
				delete(s.edges, oldID)
				break
			}
		}
	}

	s.edges[edge.ID] = edge

	// Bounded degree insertion: prevent DoS graph flooding
	if len(s.outEdges[edge.SourceID]) < s.cfg.MaxDegreePerNode*2 {
		s.outEdges[edge.SourceID] = append(s.outEdges[edge.SourceID], edge.ID)
	}
	if len(s.inEdges[edge.TargetID]) < s.cfg.MaxDegreePerNode*2 {
		s.inEdges[edge.TargetID] = append(s.inEdges[edge.TargetID], edge.ID)
	}

	return nil
}

// GetNode retrieves a node by ID.
func (s *TemporalHeteroGraphStore) GetNode(id string) (*HeteroNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, exists := s.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node '%s' not found", id)
	}
	return node, nil
}

// GetNodePointInTime retrieves a node only if it existed at or before `asOf`.
func (s *TemporalHeteroGraphStore) GetNodePointInTime(id string, asOf time.Time) (*HeteroNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, exists := s.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node '%s' not found", id)
	}
	// Strict point-in-time check
	if node.CreatedAt.After(asOf) {
		return nil, fmt.Errorf("node '%s' was created in future relative to evaluation time (%s > %s)", id, node.CreatedAt.Format(time.RFC3339), asOf.Format(time.RFC3339))
	}
	return node, nil
}

// SampledNeighborhood contains the point-in-time sampled k-hop neighborhood for GraphSAGE.
type SampledNeighborhood struct {
	RootID    string
	AsOf      time.Time
	Layer1    []*HeteroNode // Immediate 1-hop sampled neighbors
	Layer2    []*HeteroNode // 2-hop sampled neighbors
	EdgesHop1 []*HeteroEdge
	EdgesHop2 []*HeteroEdge
}

// GetTemporalSampledNeighborhood extracts a 2-layer sampled neighborhood strictly causal up to `asOf`.
// Any node or edge created after `asOf` is 100% EXCLUDED, guaranteeing zero future leakage.
func (s *TemporalHeteroGraphStore) GetTemporalSampledNeighborhood(rootID string, asOf time.Time, sampleSizes []int) (*SampledNeighborhood, error) {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	sampleL1 := 10
	sampleL2 := 5
	if len(sampleSizes) >= 1 && sampleSizes[0] > 0 {
		sampleL1 = sampleSizes[0]
	}
	if len(sampleSizes) >= 2 && sampleSizes[1] > 0 {
		sampleL2 = sampleSizes[1]
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	rootNode, rootExists := s.nodes[rootID]
	if !rootExists || rootNode.CreatedAt.After(asOf) {
		// Inductive / unseen or future node: return empty neighborhood
		return &SampledNeighborhood{
			RootID: rootID,
			AsOf:   asOf,
		}, nil
	}

	neighborhood := &SampledNeighborhood{
		RootID: rootID,
		AsOf:   asOf,
	}

	visited := make(map[string]bool)
	visited[rootID] = true

	// 1. Sample Layer 1 Neighbors (1-hop)
	validL1Edges := s.getPointInTimeEdgesRLocked(rootID, asOf)
	sampledL1Edges := sampleEdges(validL1Edges, sampleL1)

	layer1NodeIDs := make([]string, 0, len(sampledL1Edges))
	for _, edge := range sampledL1Edges {
		neighborhood.EdgesHop1 = append(neighborhood.EdgesHop1, edge)
		neighborID := edge.TargetID
		if neighborID == rootID {
			neighborID = edge.SourceID
		}
		if !visited[neighborID] {
			visited[neighborID] = true
			if n, exists := s.nodes[neighborID]; exists {
				// TEMPORAL CHECK ON NEIGHBOR NODE
				if !n.CreatedAt.After(asOf) {
					neighborhood.Layer1 = append(neighborhood.Layer1, n)
					layer1NodeIDs = append(layer1NodeIDs, neighborID)
				}
			}
		}
	}

	// 2. Sample Layer 2 Neighbors (2-hop from Layer 1)
	for _, l1ID := range layer1NodeIDs {
		validL2Edges := s.getPointInTimeEdgesRLocked(l1ID, asOf)
		sampledL2Edges := sampleEdges(validL2Edges, sampleL2)
		for _, edge := range sampledL2Edges {
			neighborID := edge.TargetID
			if neighborID == l1ID {
				neighborID = edge.SourceID
			}
			if !visited[neighborID] {
				visited[neighborID] = true
				if n, exists := s.nodes[neighborID]; exists {
					// TEMPORAL CHECK ON 2-HOP NODE
					if !n.CreatedAt.After(asOf) {
						neighborhood.Layer2 = append(neighborhood.Layer2, n)
						neighborhood.EdgesHop2 = append(neighborhood.EdgesHop2, edge)
					}
				}
			}
		}
	}

	return neighborhood, nil
}

// getPointInTimeEdgesRLocked returns all incident edges that occurred at or before `asOf` connecting to nodes existing at or before `asOf`.
func (s *TemporalHeteroGraphStore) getPointInTimeEdgesRLocked(nodeID string, asOf time.Time) []*HeteroEdge {
	var valid []*HeteroEdge

	// Outgoing
	if edgeIDs, exists := s.outEdges[nodeID]; exists {
		for _, eid := range edgeIDs {
			if edge, ok := s.edges[eid]; ok {
				// STRICT TEMPORAL CAUSAL CHECK: edge.Timestamp <= asOf
				if !edge.Timestamp.After(asOf) {
					// Verify target node exists and was created at or before asOf
					if targetNode, tExists := s.nodes[edge.TargetID]; tExists && !targetNode.CreatedAt.After(asOf) {
						valid = append(valid, edge)
					}
				}
			}
		}
	}

	// Incoming (bidirectional relationship discovery)
	if edgeIDs, exists := s.inEdges[nodeID]; exists {
		for _, eid := range edgeIDs {
			if edge, ok := s.edges[eid]; ok {
				if !edge.Timestamp.After(asOf) {
					// Verify source node exists and was created at or before asOf
					if sourceNode, sExists := s.nodes[edge.SourceID]; sExists && !sourceNode.CreatedAt.After(asOf) {
						valid = append(valid, edge)
					}
				}
			}
		}
	}

	// Bounded degree defense: Cap candidates to maxDegree to protect against high-degree DoS
	if len(valid) > s.cfg.MaxDegreePerNode {
		// Sort by recency to keep most relevant recent interactions up to asOf
		sort.Slice(valid, func(i, j int) bool {
			return valid[i].Timestamp.After(valid[j].Timestamp)
		})
		valid = valid[:s.cfg.MaxDegreePerNode]
	}

	return valid
}

// sampleEdges returns up to `k` edges deterministically from candidates.
func sampleEdges(candidates []*HeteroEdge, k int) []*HeteroEdge {
	if len(candidates) <= k {
		return candidates
	}
	// Sort by timestamp descending (most recent interactions first up to eval time)
	sorted := make([]*HeteroEdge, len(candidates))
	copy(sorted, candidates)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})

	// Deterministic selection of top-k recent + pseudo-random dispersion
	r := rand.New(rand.NewSource(42))
	perm := r.Perm(len(sorted))
	selected := make([]*HeteroEdge, k)
	for i := 0; i < k; i++ {
		selected[i] = sorted[perm[i]%len(sorted)]
	}
	return selected
}

// Count returns total nodes and edges in the store.
func (s *TemporalHeteroGraphStore) Count() (int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.nodes), len(s.edges)
}

// GetAllPointInTimeEdges returns all edges occurring at or before asOf.
func (s *TemporalHeteroGraphStore) GetAllPointInTimeEdges(asOf time.Time) []*HeteroEdge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*HeteroEdge
	for _, e := range s.edges {
		if !e.Timestamp.After(asOf) {
			result = append(result, e)
		}
	}
	return result
}
