package graph

import (
	"math"
)

// GraphMLAdapter defines the contract for Graph Neural Network (GNN) and embedding inferences.
type GraphMLAdapter interface {
	PredictGraphRisk(nodeID string, embedding []float64, neighborIDs []string) float64
}

// LocalGNNAdapter simulates GNN graph convolutional message passing for relational link prediction.
type LocalGNNAdapter struct{}

func NewLocalGNNAdapter() *LocalGNNAdapter {
	return &LocalGNNAdapter{}
}

func (a *LocalGNNAdapter) PredictGraphRisk(nodeID string, embedding []float64, neighborIDs []string) float64 {
	if len(neighborIDs) == 0 {
		return 0.05
	}

	// GNN degree and link embedding aggregation simulation
	degreeWeight := math.Min(1.0, float64(len(neighborIDs))*0.08)
	return degreeWeight
}
