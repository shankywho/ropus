package graph

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// GraphStore defines the contract for fraud graph database backends.
type GraphStore interface {
	AddNode(node *Node) error
	AddEdge(edge *Edge) error
	GetNode(id string) (*Node, error)
	GetEdge(id string) (*Edge, error)
	QueryNeighbors(nodeID string, edgeType EdgeType) ([]*Node, error)
	FindPaths(sourceID, targetID string, maxDepth int) ([]*Path, error)
	GetOutgoingEdges(nodeID string) ([]*Edge, error)
	CountNodes() int
	CountEdges() int
}

// ---------------------------------------------------------------------------
// 1. Local In-Memory Graph Store (High Performance Concurrent Engine)
// ---------------------------------------------------------------------------
type LocalGraphStore struct {
	mu       sync.RWMutex
	nodes    map[string]*Node
	edges    map[string]*Edge
	outEdges map[string][]string // nodeID -> []edgeID
	inEdges  map[string][]string // nodeID -> []edgeID
}

// NewLocalGraphStore initializes the in-memory graph repository.
func NewLocalGraphStore() *LocalGraphStore {
	return &LocalGraphStore{
		nodes:    make(map[string]*Node),
		edges:    make(map[string]*Edge),
		outEdges: make(map[string][]string),
		inEdges:  make(map[string][]string),
	}
}

func (s *LocalGraphStore) AddNode(node *Node) error {
	if node == nil || node.ID == "" {
		return fmt.Errorf("invalid node")
	}
	if node.CreatedAt.IsZero() {
		node.CreatedAt = time.Now().UTC()
	}
	node.UpdatedAt = time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nodes[node.ID] = node
	return nil
}

func (s *LocalGraphStore) AddEdge(edge *Edge) error {
	if edge == nil || edge.ID == "" || edge.SourceID == "" || edge.TargetID == "" {
		return fmt.Errorf("invalid edge")
	}
	if edge.CreatedAt.IsZero() {
		edge.CreatedAt = time.Now().UTC()
	}
	edge.UpdatedAt = time.Now().UTC()
	if edge.Confidence <= 0 {
		edge.Confidence = 1.0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.edges[edge.ID] = edge
	s.outEdges[edge.SourceID] = append(s.outEdges[edge.SourceID], edge.ID)
	s.inEdges[edge.TargetID] = append(s.inEdges[edge.TargetID], edge.ID)
	return nil
}

func (s *LocalGraphStore) GetNode(id string) (*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, exists := s.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node '%s' not found", id)
	}
	return node, nil
}

func (s *LocalGraphStore) GetEdge(id string) (*Edge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	edge, exists := s.edges[id]
	if !exists {
		return nil, fmt.Errorf("edge '%s' not found", id)
	}
	return edge, nil
}

func (s *LocalGraphStore) GetOutgoingEdges(nodeID string) ([]*Edge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	edgeIDs := s.outEdges[nodeID]
	res := make([]*Edge, 0, len(edgeIDs))
	for _, eid := range edgeIDs {
		if e, exists := s.edges[eid]; exists {
			res = append(res, e)
		}
	}
	return res, nil
}

func (s *LocalGraphStore) QueryNeighbors(nodeID string, edgeType EdgeType) ([]*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	seen := make(map[string]bool)
	var neighbors []*Node

	// Check outgoing edges
	for _, eid := range s.outEdges[nodeID] {
		edge, exists := s.edges[eid]
		if !exists {
			continue
		}
		if edgeType == "" || edge.Type == edgeType {
			if !seen[edge.TargetID] {
				seen[edge.TargetID] = true
				if targetNode, found := s.nodes[edge.TargetID]; found {
					neighbors = append(neighbors, targetNode)
				}
			}
		}
	}

	// Check incoming edges
	for _, eid := range s.inEdges[nodeID] {
		edge, exists := s.edges[eid]
		if !exists {
			continue
		}
		if edgeType == "" || edge.Type == edgeType {
			if !seen[edge.SourceID] {
				seen[edge.SourceID] = true
				if srcNode, found := s.nodes[edge.SourceID]; found {
					neighbors = append(neighbors, srcNode)
				}
			}
		}
	}

	return neighbors, nil
}

func (s *LocalGraphStore) FindPaths(sourceID, targetID string, maxDepth int) ([]*Path, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if maxDepth <= 0 {
		maxDepth = 3
	}

	var results []*Path
	visited := make(map[string]bool)

	var dfs func(currentID string, currentPath []*Node, currentEdges []*Edge, depth int)
	dfs = func(currentID string, currentPath []*Node, currentEdges []*Edge, depth int) {
		if depth > maxDepth {
			return
		}

		currNode, exists := s.nodes[currentID]
		if !exists {
			return
		}

		pathNodes := append(currentPath, currNode)
		visited[currentID] = true

		if currentID == targetID && len(currentEdges) > 0 {
			p := &Path{
				Nodes:  pathNodes,
				Edges:  currentEdges,
				Length: len(currentEdges),
			}
			results = append(results, p)
			visited[currentID] = false
			return
		}

		for _, eid := range s.outEdges[currentID] {
			edge, edgeExists := s.edges[eid]
			if !edgeExists {
				continue
			}
			if !visited[edge.TargetID] {
				dfs(edge.TargetID, pathNodes, append(currentEdges, edge), depth+1)
			}
		}

		visited[currentID] = false
	}

	dfs(sourceID, nil, nil, 0)
	return results, nil
}

func (s *LocalGraphStore) CountNodes() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.nodes)
}

func (s *LocalGraphStore) CountEdges() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.edges)
}

// ---------------------------------------------------------------------------
// 2. Neo4j Cypher Adapter Boundary
// ---------------------------------------------------------------------------
type Neo4jAdapter struct {
	URI      string
	Database string
}

func NewNeo4jAdapter(uri, database string) *Neo4jAdapter {
	if uri == "" {
		uri = "bolt://neo4j-cluster.graph.svc.cluster.local:7687"
	}
	return &Neo4jAdapter{URI: uri, Database: database}
}

func (a *Neo4jAdapter) QueryNeighborsCypher(ctx context.Context, nodeID string) ([]*Node, error) {
	// Cypher query boundary: MATCH (n {id: $nodeID})-[r]->(m) RETURN m
	return nil, nil
}

// ---------------------------------------------------------------------------
// 3. RedisGraph / FalkorDB Adapter Boundary
// ---------------------------------------------------------------------------
type RedisGraphAdapter struct {
	Host string
	Port int
	Key  string
}

func NewRedisGraphAdapter(host string, port int, key string) *RedisGraphAdapter {
	if host == "" {
		host = "risk-redis-master"
		port = 6379
	}
	if key == "" {
		key = "fraud_graph"
	}
	return &RedisGraphAdapter{Host: host, Port: port, Key: key}
}
