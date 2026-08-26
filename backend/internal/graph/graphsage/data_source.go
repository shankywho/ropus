package graphsage

import (
	"context"
	"time"
)

// RealGraphDataSource defines the ingestion adapter contract for real and shadow graph feeds.
type RealGraphDataSource interface {
	GetDatasetType() DatasetType
	GetNodes(ctx context.Context, nodeTypes []HeteroNodeType, asOf time.Time) ([]*HeteroNode, error)
	GetEdges(ctx context.Context, edgeTypes []HeteroEdgeType, asOf time.Time) ([]*HeteroEdge, error)
	GetTransactions(ctx context.Context, startTime, endTime time.Time) ([]map[string]interface{}, error)
	GetLabels(ctx context.Context, asOf time.Time, matureOnly bool) ([]map[string]interface{}, error)
}

// MockRealGraphDataSource provides a safe offline fixture adapter.
type MockRealGraphDataSource struct {
	datasetType  DatasetType
	nodes        []*HeteroNode
	edges        []*HeteroEdge
	transactions []map[string]interface{}
	labels       []map[string]interface{}
}

// NewMockRealGraphDataSource initializes an in-memory offline mock adapter.
func NewMockRealGraphDataSource(datasetType DatasetType) *MockRealGraphDataSource {
	return &MockRealGraphDataSource{
		datasetType:  datasetType,
		nodes:        make([]*HeteroNode, 0),
		edges:        make([]*HeteroEdge, 0),
		transactions: make([]map[string]interface{}, 0),
		labels:       make([]map[string]interface{}, 0),
	}
}

func (m *MockRealGraphDataSource) AddNode(n *HeteroNode) {
	m.nodes = append(m.nodes, n)
}

func (m *MockRealGraphDataSource) AddEdge(e *HeteroEdge) {
	m.edges = append(m.edges, e)
}

func (m *MockRealGraphDataSource) AddLabel(l map[string]interface{}) {
	m.labels = append(m.labels, l)
}

func (m *MockRealGraphDataSource) GetDatasetType() DatasetType {
	return m.datasetType
}

func (m *MockRealGraphDataSource) GetNodes(ctx context.Context, nodeTypes []HeteroNodeType, asOf time.Time) ([]*HeteroNode, error) {
	typeMap := make(map[HeteroNodeType]bool)
	for _, t := range nodeTypes {
		typeMap[t] = true
	}
	res := make([]*HeteroNode, 0)
	for _, n := range m.nodes {
		if len(typeMap) > 0 && !typeMap[n.Type] {
			continue
		}
		if !asOf.IsZero() && n.CreatedAt.After(asOf) {
			continue
		}
		res = append(res, n)
	}
	return res, nil
}

func (m *MockRealGraphDataSource) GetEdges(ctx context.Context, edgeTypes []HeteroEdgeType, asOf time.Time) ([]*HeteroEdge, error) {
	typeMap := make(map[HeteroEdgeType]bool)
	for _, t := range edgeTypes {
		typeMap[t] = true
	}
	res := make([]*HeteroEdge, 0)
	for _, e := range m.edges {
		if len(typeMap) > 0 && !typeMap[e.Type] {
			continue
		}
		if !asOf.IsZero() && e.Timestamp.After(asOf) {
			continue
		}
		res = append(res, e)
	}
	return res, nil
}

func (m *MockRealGraphDataSource) GetTransactions(ctx context.Context, startTime, endTime time.Time) ([]map[string]interface{}, error) {
	return m.transactions, nil
}

func (m *MockRealGraphDataSource) GetLabels(ctx context.Context, asOf time.Time, matureOnly bool) ([]map[string]interface{}, error) {
	res := make([]map[string]interface{}, 0)
	for _, l := range m.labels {
		if matureOnly {
			mat, ok := l["maturity_state"].(LabelMaturityState)
			if !ok || (mat != LabelConfirmedFraud && mat != LabelConfirmedLegitimate) {
				continue
			}
		}
		res = append(res, l)
	}
	return res, nil
}
