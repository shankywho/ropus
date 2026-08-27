package graph

import (
	"fmt"
	"math"
	"time"
)

// EntityGraphEvidence models the real BFS graph traversal evidence for an entity.
type EntityGraphEvidence struct {
	StartNodeID           string   `json:"start_node_id"`
	VisitedNodesCount     int      `json:"visited_nodes_count"`
	TraversedEdgesCount   int      `json:"traversed_edges_count"`
	MaxClusterDepth       int      `json:"max_cluster_depth"`
	ConnectedAccountCount int      `json:"connected_account_count"`
	SharedDeviceCount     int      `json:"shared_device_count"`
	FraudNodesCount       int      `json:"fraud_nodes_count"`
	FraudRingDetected     bool     `json:"fraud_ring_detected"`
	DegreeCentrality      int      `json:"degree_centrality"`
	PayoutDepots          []string `json:"payout_depots,omitempty"`
	GraphRiskContribution float64  `json:"graph_risk_contribution"`
	Matches               []string `json:"matches"`
}

// GraphEngine manages live knowledge graph operations and graph query algorithms.
type GraphEngine struct {
	store GraphStore
}

// NewGraphEngine initializes the graph intelligence engine.
func NewGraphEngine(store GraphStore) *GraphEngine {
	if store == nil {
		store = NewLocalGraphStore()
	}
	return &GraphEngine{store: store}
}

func (e *GraphEngine) Store() GraphStore {
	return e.store
}

// IngestTransactionLinks ingests a transaction and creates/links its associated entities.
func (e *GraphEngine) IngestTransactionLinks(
	txnID, userID, accountID, cardHash, deviceFingerprint, ipAddress, merchantID string,
	amount float64,
	isFraud bool,
) error {
	now := time.Now().UTC()

	// 1. Create or update nodes
	nodes := []*Node{
		{ID: txnID, Type: NodeTransaction, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: userID, Type: NodeUser, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: accountID, Type: NodeAccount, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: cardHash, Type: NodeCard, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: deviceFingerprint, Type: NodeDevice, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: ipAddress, Type: NodeIPAddress, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: merchantID, Type: NodeMerchant, RiskScore: 0.0, IsKnownBad: false, CreatedAt: now},
	}

	for _, n := range nodes {
		if n.ID != "" {
			_ = e.store.AddNode(n)
		}
	}

	// 2. Create relationships
	edges := []*Edge{
		{ID: fmt.Sprintf("e_%s_%s", userID, accountID), SourceID: userID, TargetID: accountID, Type: EdgeOwns, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", accountID, cardHash), SourceID: accountID, TargetID: cardHash, Type: EdgeConnectedTo, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", userID, deviceFingerprint), SourceID: userID, TargetID: deviceFingerprint, Type: EdgeUsedBy, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", accountID, deviceFingerprint), SourceID: accountID, TargetID: deviceFingerprint, Type: EdgeUsedBy, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", userID, ipAddress), SourceID: userID, TargetID: ipAddress, Type: EdgeLoggedInFrom, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", txnID, merchantID), SourceID: txnID, TargetID: merchantID, Type: EdgeTransactedWith, Weight: amount, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", accountID, txnID), SourceID: accountID, TargetID: txnID, Type: EdgeTransactedWith, Weight: amount, Confidence: 1.0, CreatedAt: now},
	}

	for _, ed := range edges {
		if ed.SourceID != "" && ed.TargetID != "" {
			_ = e.store.AddEdge(ed)
		}
	}

	return nil
}

// EvaluateEntityGraph executes a real 3-hop BFS expansion starting from primary entity identifiers.
func (e *GraphEngine) EvaluateEntityGraph(accountID, deviceFingerprint, cardHash, ipAddress string) *EntityGraphEvidence {
	evidence := &EntityGraphEvidence{
		Matches:      make([]string, 0),
		PayoutDepots: make([]string, 0),
	}

	startNodeID := accountID
	if startNodeID == "" {
		startNodeID = deviceFingerprint
	}
	if startNodeID == "" {
		startNodeID = ipAddress
	}
	if startNodeID == "" {
		return evidence
	}
	evidence.StartNodeID = startNodeID

	temporalEvidence, err := e.store.Traverse3HopTemporal(startNodeID, time.Now().UTC(), 72*time.Hour, 50)
	if err != nil || temporalEvidence == nil {
		return evidence
	}

	evidence.VisitedNodesCount = temporalEvidence.VisitedNodesCount
	evidence.TraversedEdgesCount = temporalEvidence.TraversedEdgesCount
	evidence.MaxClusterDepth = temporalEvidence.MaxClusterDepth
	evidence.FraudNodesCount = temporalEvidence.FraudNodesCount
	evidence.FraudRingDetected = temporalEvidence.FraudRingDetected

	// Calculate degree centrality and count entity types across reachable nodes
	connectedAccounts := 0
	sharedDevices := 0

	for _, nodeID := range temporalEvidence.ReachableNodeIDs {
		if node, err := e.store.GetNode(nodeID); err == nil && node != nil {
			if node.Type == NodeAccount {
				if node.ID != accountID {
					connectedAccounts++
				}
			} else if node.Type == NodeDevice {
				if node.ID != deviceFingerprint {
					sharedDevices++
				}
			}
		}
	}

	evidence.ConnectedAccountCount = connectedAccounts
	evidence.SharedDeviceCount = sharedDevices

	// Centrality metric counts external connected accounts and multi-device fan-outs
	degree := connectedAccounts*2 + sharedDevices*2 + (temporalEvidence.TraversedEdgesCount / 4)
	evidence.DegreeCentrality = degree

	risk := 0.0

	// Risk signal evaluation based on genuine graph metrics
	if evidence.FraudNodesCount > 0 {
		evidence.Matches = append(evidence.Matches, fmt.Sprintf("Graph traversal intersects %d confirmed fraudulent entities in 3-hop neighborhood", evidence.FraudNodesCount))
		risk = math.Max(risk, 0.85)
	}

	if connectedAccounts >= 3 {
		evidence.Matches = append(evidence.Matches, fmt.Sprintf("Entity linked across %d distinct accounts via shared hardware/network graph edges", connectedAccounts))
		risk = math.Max(risk, 0.40+float64(connectedAccounts)*0.04)
	}

	if degree >= 6 && connectedAccounts >= 2 {
		evidence.Matches = append(evidence.Matches, fmt.Sprintf("Elevated graph centrality degree (%d >= 6) indicates synthetic hub or mule clustering", degree))
		risk = math.Max(risk, 0.55)
	}

	if evidence.FraudRingDetected {
		evidence.Matches = append(evidence.Matches, "Cyclic or multi-way syndication pattern detected in 3-hop graph cluster")
		risk = math.Max(risk, 0.90)
	}

	if risk > 0.95 {
		risk = 0.95
	}
	evidence.GraphRiskContribution = math.Round(risk*100) / 100.0

	return evidence
}
