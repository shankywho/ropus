package graph

import (
	"math"
)

// GraphFeatureVector holds real-time network topology features extracted from the graph.
type GraphFeatureVector struct {
	ConnectedAccountsCount int     `json:"connected_accounts_count"`
	DeviceAccountCount     int     `json:"device_account_count"`
	IPAccountCount         int     `json:"ip_account_count"`
	FraudNeighborCount     int     `json:"fraud_neighbor_count"`
	SharedIdentifierScore  float64 `json:"shared_identifier_score"`
	GraphRiskScore         float64 `json:"graph_risk_score"`
}

// GraphFeatureExtractor extracts low-latency graph structural metrics for risk inference.
type GraphFeatureExtractor struct {
	engine *GraphEngine
}

// NewGraphFeatureExtractor initializes the graph feature extractor.
func NewGraphFeatureExtractor(engine *GraphEngine) *GraphFeatureExtractor {
	return &GraphFeatureExtractor{engine: engine}
}

// ExtractFeatures computes real-time graph attributes for an incoming transaction.
func (x *GraphFeatureExtractor) ExtractFeatures(userID, deviceFingerprint, ipAddress string) (*GraphFeatureVector, error) {
	var fraudNeighbors int
	var deviceAccounts int
	var ipAccounts int

	// 1. Check user neighbors for fraud links
	userNeighbors, _ := x.engine.Store().QueryNeighbors(userID, "")
	for _, n := range userNeighbors {
		if n.IsKnownBad || n.RiskScore > 0.80 {
			fraudNeighbors++
		}
	}

	// 2. Check device sharing fan-out
	if deviceFingerprint != "" {
		devNeighbors, _ := x.engine.Store().QueryNeighbors(deviceFingerprint, "")
		deviceAccounts = len(devNeighbors)
		for _, n := range devNeighbors {
			if n.IsKnownBad {
				fraudNeighbors++
			}
		}
	}

	// 3. Check IP address sharing fan-out
	if ipAddress != "" {
		ipNeighbors, _ := x.engine.Store().QueryNeighbors(ipAddress, "")
		ipAccounts = len(ipNeighbors)
		for _, n := range ipNeighbors {
			if n.IsKnownBad {
				fraudNeighbors++
			}
		}
	}

	// 4. Compute composite Graph Risk Score
	sharedScore := math.Min(1.0, float64(deviceAccounts)*0.10+float64(ipAccounts)*0.05)
	fraudImpact := math.Min(1.0, float64(fraudNeighbors)*0.35)

	graphRisk := math.Max(sharedScore*0.5, fraudImpact)
	if fraudNeighbors > 0 {
		graphRisk = math.Max(graphRisk, 0.85)
	}

	return &GraphFeatureVector{
		ConnectedAccountsCount: len(userNeighbors),
		DeviceAccountCount:     deviceAccounts,
		IPAccountCount:         ipAccounts,
		FraudNeighborCount:     fraudNeighbors,
		SharedIdentifierScore:  sharedScore,
		GraphRiskScore:         graphRisk,
	}, nil
}
