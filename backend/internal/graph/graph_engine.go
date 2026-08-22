package graph

import (
	"fmt"
	"time"
)

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
		_ = e.store.AddNode(n)
	}

	// 2. Create relationships
	edges := []*Edge{
		{ID: fmt.Sprintf("e_%s_%s", userID, accountID), SourceID: userID, TargetID: accountID, Type: EdgeOwns, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", accountID, cardHash), SourceID: accountID, TargetID: cardHash, Type: EdgeConnectedTo, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", userID, deviceFingerprint), SourceID: userID, TargetID: deviceFingerprint, Type: EdgeUsedBy, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", userID, ipAddress), SourceID: userID, TargetID: ipAddress, Type: EdgeLoggedInFrom, Weight: 1.0, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", txnID, merchantID), SourceID: txnID, TargetID: merchantID, Type: EdgeTransactedWith, Weight: amount, Confidence: 1.0, CreatedAt: now},
		{ID: fmt.Sprintf("e_%s_%s", accountID, txnID), SourceID: accountID, TargetID: txnID, Type: EdgeTransactedWith, Weight: amount, Confidence: 1.0, CreatedAt: now},
	}

	for _, ed := range edges {
		_ = e.store.AddEdge(ed)
	}

	return nil
}
