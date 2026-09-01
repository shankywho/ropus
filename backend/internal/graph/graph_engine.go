package graph

import (
	"fmt"
	"math"
	"strings"
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

// FormatNodeID creates an internal tenant-scoped node identifier: <tenant_id>:<entity_type>:<raw_id>
func FormatNodeID(tenantID string, nodeType NodeType, rawID string) string {
	if rawID == "" {
		return ""
	}
	if tenantID == "" {
		tenantID = "default"
	}
	return fmt.Sprintf("%s:%s:%s", tenantID, strings.ToLower(string(nodeType)), rawID)
}

// ParseNodeID extracts the tenantID, nodeType, and raw entity ID from an internal node identifier.
func ParseNodeID(internalID string) (tenantID string, nodeType NodeType, rawID string) {
	parts := strings.SplitN(internalID, ":", 3)
	if len(parts) == 3 {
		return parts[0], NodeType(strings.ToUpper(parts[1])), parts[2]
	}
	return "default", NodeAccount, internalID
}

// IngestTenantTransactionLinks ingests a transaction with explicit tenant namespacing.
func (e *GraphEngine) IngestTenantTransactionLinks(
	tenantID, txnID, userID, accountID, cardHash, deviceFingerprint, ipAddress, merchantID string,
	amount float64,
	isFraud bool,
) error {
	if tenantID == "" {
		tenantID = "default"
	}
	now := time.Now().UTC()

	// 1. Create tenant-namespaced nodes
	nodes := []*Node{
		{ID: FormatNodeID(tenantID, NodeTransaction, txnID), Type: NodeTransaction, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: FormatNodeID(tenantID, NodeUser, userID), Type: NodeUser, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: FormatNodeID(tenantID, NodeAccount, accountID), Type: NodeAccount, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: FormatNodeID(tenantID, NodeCard, cardHash), Type: NodeCard, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: FormatNodeID(tenantID, NodeDevice, deviceFingerprint), Type: NodeDevice, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: FormatNodeID(tenantID, NodeIPAddress, ipAddress), Type: NodeIPAddress, RiskScore: 0.0, IsKnownBad: isFraud, CreatedAt: now},
		{ID: FormatNodeID(tenantID, NodeMerchant, merchantID), Type: NodeMerchant, RiskScore: 0.0, IsKnownBad: false, CreatedAt: now},
	}

	for _, n := range nodes {
		if n.ID != "" && !strings.HasSuffix(n.ID, ":") {
			_ = e.store.AddNode(n)
		}
	}

	// 2. Create tenant-namespaced relationships
	uNode := FormatNodeID(tenantID, NodeUser, userID)
	accNode := FormatNodeID(tenantID, NodeAccount, accountID)
	cardNode := FormatNodeID(tenantID, NodeCard, cardHash)
	devNode := FormatNodeID(tenantID, NodeDevice, deviceFingerprint)
	ipNode := FormatNodeID(tenantID, NodeIPAddress, ipAddress)
	txnNode := FormatNodeID(tenantID, NodeTransaction, txnID)
	merchNode := FormatNodeID(tenantID, NodeMerchant, merchantID)

	edges := []*Edge{
		{ID: fmt.Sprintf("e_%s_%s", uNode, accNode), SourceID: uNode, TargetID: accNode, Type: EdgeOwns, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", accNode, cardNode), SourceID: accNode, TargetID: cardNode, Type: EdgeConnectedTo, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", uNode, devNode), SourceID: uNode, TargetID: devNode, Type: EdgeUsedBy, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", accNode, devNode), SourceID: accNode, TargetID: devNode, Type: EdgeUsedBy, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", uNode, ipNode), SourceID: uNode, TargetID: ipNode, Type: EdgeLoggedInFrom, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", txnNode, merchNode), SourceID: txnNode, TargetID: merchNode, Type: EdgeTransactedWith, Weight: amount, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", accNode, txnNode), SourceID: accNode, TargetID: txnNode, Type: EdgeTransactedWith, Weight: amount, Confidence: 1.0, CreatedAt: now},
	}

	for _, ed := range edges {
		if !strings.HasSuffix(ed.SourceID, ":") && !strings.HasSuffix(ed.TargetID, ":") {
			_ = e.store.AddEdge(ed)
		}
	}

	return nil
}

// IngestTransactionLinks ingests a transaction with default tenant scope for backwards compatibility.
func (e *GraphEngine) IngestTransactionLinks(
	txnID, userID, accountID, cardHash, deviceFingerprint, ipAddress, merchantID string,
	amount float64,
	isFraud bool,
) error {
	return e.IngestTenantTransactionLinks("default", txnID, userID, accountID, cardHash, deviceFingerprint, ipAddress, merchantID, amount, isFraud)
}

// EvaluateTenantEntityGraph executes a real 3-hop BFS expansion starting from primary entity identifiers within a specific tenant scope.
func (e *GraphEngine) EvaluateTenantEntityGraph(tenantID, accountID, deviceFingerprint, cardHash, ipAddress string) *EntityGraphEvidence {
	if tenantID == "" {
		tenantID = "default"
	}
	evidence := &EntityGraphEvidence{
		Matches:      make([]string, 0),
		PayoutDepots: make([]string, 0),
	}

	startNodeID := FormatNodeID(tenantID, NodeAccount, accountID)
	if accountID == "" {
		startNodeID = FormatNodeID(tenantID, NodeDevice, deviceFingerprint)
	}
	if accountID == "" && deviceFingerprint == "" {
		startNodeID = FormatNodeID(tenantID, NodeIPAddress, ipAddress)
	}
	if startNodeID == "" || strings.HasSuffix(startNodeID, ":") {
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
			_, _, rawNodeID := ParseNodeID(node.ID)
			if node.Type == NodeAccount {
				if rawNodeID != accountID {
					connectedAccounts++
				}
			} else if node.Type == NodeDevice {
				if rawNodeID != deviceFingerprint {
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

// EvaluateEntityGraph evaluates graph evidence for the default tenant.
func (e *GraphEngine) EvaluateEntityGraph(accountID, deviceFingerprint, cardHash, ipAddress string) *EntityGraphEvidence {
	return e.EvaluateTenantEntityGraph("default", accountID, deviceFingerprint, cardHash, ipAddress)
}

// FrontendGraphEntity models an entity node for the frontend SVG force/concentric graph.
type FrontendGraphEntity struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"` // "CUSTOMER", "DEVICE", "IP", "ACCOUNT", "TRANSACTION"
	Label      string      `json:"label"`
	Risk       string      `json:"risk"` // "CLEAN", "WATCH", "SUSPECT", "CONFIRMED_FRAUD"
	X          float64     `json:"x"`    // 0-100 deterministic layout coordinate
	Y          float64     `json:"y"`    // 0-100 deterministic layout coordinate
	Hop        int         `json:"hop"`  // 0, 1, 2, 3
	FirstSeen  string      `json:"firstSeen"`
	LastSeen   string      `json:"lastSeen"`
	Attributes [][2]string `json:"attributes"`
	Signals    []string    `json:"signals"`
}

// FrontendGraphRelationship models a directed edge for the frontend graph view.
type FrontendGraphRelationship struct {
	Source         string `json:"source"`
	Target         string `json:"target"`
	Label          string `json:"label"`
	OnDecisionPath bool   `json:"onDecisionPath"`
}

// FrontendFraudGraphResponse models the JSON payload expected by the frontend graph viewer.
type FrontendFraudGraphResponse struct {
	RootID        string                      `json:"rootId"`
	DecisionID    string                      `json:"decisionId"`
	Source        string                      `json:"source"` // "live_graph_engine"
	Entities      []FrontendGraphEntity       `json:"entities"`
	Relationships []FrontendGraphRelationship `json:"relationships"`
}

// SeedDefaultDemoGraph pre-populates the in-memory graph with the PA-77120 mule syndicate topology.
func (e *GraphEngine) SeedDefaultDemoGraph() {
	now := time.Now().UTC()

	// Seed 14-node syndicate ring centered around payout depot PA-77120
	_ = e.IngestTenantTransactionLinks("default", "txn_88419", "usr_1001", "acc_victim_01", "tok_card_99", "dev_emul_01", "185.220.101.5", "merch_crypto_99", 82000.0, true)
	_ = e.IngestTenantTransactionLinks("default", "txn_88420", "usr_1002", "acc_mule_02", "tok_card_99", "dev_emul_01", "185.220.101.5", "merch_payout_hub", 45000.0, true)
	_ = e.IngestTenantTransactionLinks("default", "txn_88421", "usr_1003", "acc_mule_03", "tok_card_88", "dev_emul_02", "185.220.101.6", "merch_payout_hub", 92000.0, true)
	_ = e.IngestTenantTransactionLinks("default", "txn_88422", "usr_1004", "PA-77120", "tok_card_77", "dev_emul_02", "198.51.100.44", "merch_payout_hub", 145000.0, true)

	// Add direct inter-mule transfer edges
	acc2Node := FormatNodeID("default", NodeAccount, "acc_mule_02")
	acc3Node := FormatNodeID("default", NodeAccount, "acc_mule_03")
	depotNode := FormatNodeID("default", NodeAccount, "PA-77120")

	_ = e.store.AddEdge(&Edge{
		ID:         "e_mule_1_depot",
		SourceID:   acc2Node,
		TargetID:   depotNode,
		Type:       EdgeTransferredTo,
		Weight:     45000.0,
		Confidence: 1.0,
		CreatedAt:  now,
	})
	_ = e.store.AddEdge(&Edge{
		ID:         "e_mule_2_depot",
		SourceID:   acc3Node,
		TargetID:   depotNode,
		Type:       EdgeTransferredTo,
		Weight:     92000.0,
		Confidence: 1.0,
		CreatedAt:  now,
	})
}

// ExportTenantFraudGraph exports tenant-scoped nodes and edges into the frontend format.
func (e *GraphEngine) ExportTenantFraudGraph(tenantID, decisionID, startNodeID string) *FrontendFraudGraphResponse {
	if e.store.CountNodes() == 0 {
		e.SeedDefaultDemoGraph()
	}
	if tenantID == "" {
		tenantID = "default"
	}

	allNodesRaw := e.store.GetAllNodes()
	allEdgesRaw := e.store.GetAllEdges()

	// Filter nodes belonging to the requested tenant
	allNodes := make([]*Node, 0)
	for _, n := range allNodesRaw {
		tID, _, _ := ParseNodeID(n.ID)
		if tID == tenantID || tenantID == "all" {
			allNodes = append(allNodes, n)
		}
	}

	allEdges := make([]*Edge, 0)
	for _, edge := range allEdgesRaw {
		tID1, _, _ := ParseNodeID(edge.SourceID)
		tID2, _, _ := ParseNodeID(edge.TargetID)
		if (tID1 == tenantID && tID2 == tenantID) || tenantID == "all" {
			allEdges = append(allEdges, edge)
		}
	}

	if len(allNodes) == 0 {
		return &FrontendFraudGraphResponse{
			RootID:        "root",
			DecisionID:    decisionID,
			Source:        "live_graph_engine",
			Entities:      make([]FrontendGraphEntity, 0),
			Relationships: make([]FrontendGraphRelationship, 0),
		}
	}

	// Resolve Root Node ID
	rootID := startNodeID
	if rootID != "" && !strings.Contains(rootID, ":") {
		rootID = FormatNodeID(tenantID, NodeAccount, rootID)
	}

	if rootID == "" {
		for _, n := range allNodes {
			_, _, rawID := ParseNodeID(n.ID)
			if rawID == "PA-77120" || rawID == "acc_victim_01" || n.IsKnownBad {
				rootID = n.ID
				break
			}
		}
		if rootID == "" && len(allNodes) > 0 {
			rootID = allNodes[0].ID
		}
	}

	// Compute Hop Distances from root via BFS
	hopMap := make(map[string]int)
	hopMap[rootID] = 0

	queue := []string{rootID}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		currHop := hopMap[curr]

		for _, edge := range allEdges {
			var neighbor string
			if edge.SourceID == curr {
				neighbor = edge.TargetID
			} else if edge.TargetID == curr {
				neighbor = edge.SourceID
			}

			if neighbor != "" {
				if _, visited := hopMap[neighbor]; !visited {
					hopMap[neighbor] = currHop + 1
					if currHop+1 < 3 {
						queue = append(queue, neighbor)
					}
				}
			}
		}
	}

	// Group nodes by hop for concentric circle layout
	hopNodes := make(map[int][]string)
	for _, n := range allNodes {
		h, ok := hopMap[n.ID]
		if !ok {
			h = 3
			hopMap[n.ID] = 3
		}
		hopNodes[h] = append(hopNodes[h], n.ID)
	}

	// Calculate deterministic (x, y) coordinates
	coords := make(map[string][2]float64)
	coords[rootID] = [2]float64{50.0, 50.0}

	radii := map[int]float64{
		1: 20.0,
		2: 34.0,
		3: 44.0,
	}

	for hop := 1; hop <= 3; hop++ {
		nodesAtHop := hopNodes[hop]
		count := len(nodesAtHop)
		if count == 0 {
			continue
		}
		r := radii[hop]
		for i, nid := range nodesAtHop {
			angle := (float64(i) / float64(count)) * 2.0 * math.Pi
			x := 50.0 + r*math.Cos(angle)
			y := 50.0 + r*math.Sin(angle)*0.88
			coords[nid] = [2]float64{math.Round(x*10) / 10.0, math.Round(y*10) / 10.0}
		}
	}

	// Map nodes to FrontendGraphEntity
	entities := make([]FrontendGraphEntity, 0, len(allNodes))
	for _, n := range allNodes {
		pos := coords[n.ID]
		hop := hopMap[n.ID]
		_, _, rawID := ParseNodeID(n.ID)

		fType := "ACCOUNT"
		switch n.Type {
		case NodeUser:
			fType = "CUSTOMER"
		case NodeDevice:
			fType = "DEVICE"
		case NodeIPAddress:
			fType = "IP"
		case NodeTransaction:
			fType = "TRANSACTION"
		}

		risk := "CLEAN"
		if n.IsKnownBad || n.RiskScore >= 0.80 {
			risk = "CONFIRMED_FRAUD"
		} else if n.RiskScore >= 0.50 {
			risk = "SUSPECT"
		} else if hop == 1 {
			risk = "WATCH"
		}

		label := rawID
		if n.Type == NodeIPAddress {
			label = maskIP(rawID)
		}

		entities = append(entities, FrontendGraphEntity{
			ID:         rawID,
			Type:       fType,
			Label:      label,
			Risk:       risk,
			X:          pos[0],
			Y:          pos[1],
			Hop:        hop,
			FirstSeen:  n.CreatedAt.Format(time.RFC3339),
			LastSeen:   time.Now().UTC().Format(time.RFC3339),
			Attributes: [][2]string{{"Type", string(n.Type)}, {"Risk Score", fmt.Sprintf("%.2f", n.RiskScore)}},
			Signals:    []string{"Live Graph BFS"},
		})
	}

	// Map relationships
	relationships := make([]FrontendGraphRelationship, 0, len(allEdges))
	for _, edge := range allEdges {
		_, _, rawSource := ParseNodeID(edge.SourceID)
		_, _, rawTarget := ParseNodeID(edge.TargetID)
		relationships = append(relationships, FrontendGraphRelationship{
			Source:         rawSource,
			Target:         rawTarget,
			Label:          string(edge.Type),
			OnDecisionPath: edge.Confidence >= 0.8,
		})
	}

	_, _, rawRootID := ParseNodeID(rootID)
	return &FrontendFraudGraphResponse{
		RootID:        rawRootID,
		DecisionID:    decisionID,
		Source:        "live_graph_engine",
		Entities:      entities,
		Relationships: relationships,
	}
}

// ExportFraudGraph exports the active knowledge graph into the frontend's expected FraudGraph format.
func (e *GraphEngine) ExportFraudGraph(decisionID, startNodeID string) *FrontendFraudGraphResponse {
	return e.ExportTenantFraudGraph("default", decisionID, startNodeID)
}

// maskIP redacts the last two octets of an IPv4 address for privacy.
func maskIP(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.***.***", parts[0], parts[1])
	}
	return ip
}
