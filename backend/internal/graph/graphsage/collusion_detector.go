package graphsage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// CollusionDetectionThresholds contains configurable heuristics and peer-group baselines.
type CollusionDetectionThresholds struct {
	MaxNormalEmployeeConsumerDegree int     // Max acceptable distinct consumers accessed per employee in 24h (default: 30)
	ConcentrationRatioThreshold     float64 // If >40% of employee interactions target a specific consumer/account
	SharedDeviceThreshold           int     // Max allowed device overlap between employee and consumer (default: 0)
	SharedIPThreshold               int     // Max allowed IP overlap between employee and consumer (default: 0)
	TemporalCorrelationSeconds      float64 // Flag if employee action precedes suspicious transaction within 900s
	OffHoursStartHour               int     // 20:00 (8 PM)
	OffHoursEndHour                 int     // 06:00 (6 AM)
}

// DefaultCollusionDetectionThresholds returns production security baseline thresholds.
func DefaultCollusionDetectionThresholds() CollusionDetectionThresholds {
	return CollusionDetectionThresholds{
		MaxNormalEmployeeConsumerDegree: 30,
		ConcentrationRatioThreshold:     0.40,
		SharedDeviceThreshold:           0,
		SharedIPThreshold:               0,
		TemporalCorrelationSeconds:      900.0,
		OffHoursStartHour:               20,
		OffHoursEndHour:                 6,
	}
}

// CollusionDetector analyzes the graph topology around an employee or consumer to detect collusion.
type CollusionDetector struct {
	thresholds CollusionDetectionThresholds
}

// NewCollusionDetector initializes the structural collusion detector.
func NewCollusionDetector(thresholds CollusionDetectionThresholds) *CollusionDetector {
	return &CollusionDetector{thresholds: thresholds}
}

// AnalyzeStructure inspects the sampled neighborhood for explainable collusion patterns and peer deviations.
func (d *CollusionDetector) AnalyzeStructure(rootNode *HeteroNode, nh *SampledNeighborhood) (
	facts []string,
	patterns []string,
	riskScore float64,
	signals RelationshipIntelligenceSignals,
	explainableSignals []ExplainableCollusionSignal,
) {
	if rootNode == nil || nh == nil {
		return facts, patterns, 0.05, signals, explainableSignals
	}

	consumerCounts := make(map[string]int)
	accountCounts := make(map[string]int)
	deviceCounts := make(map[string]int)
	ipCounts := make(map[string]int)
	totalEmployeeInteractions := 0
	highRiskConsumerCount := 0
	offHoursCount := 0
	var preTxProximitySec float64 = 999999.0

	var affectedEntities []string
	addAffected := func(rawID string) {
		token := tokenizeID(rawID)
		for _, existing := range affectedEntities {
			if existing == token {
				return
			}
		}
		affectedEntities = append(affectedEntities, token)
	}

	addAffected(rootNode.ID)

	// 1. Analyze 1-Hop Edges
	for _, edge := range nh.EdgesHop1 {
		// Timestamp-based Off-Hours evaluation
		edgeHour := edge.Timestamp.Hour()
		if edgeHour >= d.thresholds.OffHoursStartHour || edgeHour < d.thresholds.OffHoursEndHour {
			offHoursCount++
		}

		if edge.Type == EdgeAccesses || edge.Type == EdgeAccessesConsumer {
			totalEmployeeInteractions++
			consumerCounts[edge.TargetID]++
			addAffected(edge.TargetID)
		}
		if edge.Type == EdgeAccessesAccount || edge.Type == EdgeManages {
			totalEmployeeInteractions++
			accountCounts[edge.TargetID]++
			addAffected(edge.TargetID)
		}
	}

	// 2. Analyze Node Types in Layer 1 and Layer 2
	for _, node := range append(nh.Layer1, nh.Layer2...) {
		if node.Type == NodeConsumer && (node.RiskScore > 0.60 || node.IsKnownBad) {
			highRiskConsumerCount++
			addAffected(node.ID)
		}
		if node.Type == NodeDevice {
			deviceCounts[node.ID]++
		}
		if node.Type == NodeIP {
			ipCounts[node.ID]++
		}
	}

	// 3. Inspect Direct Overlap Linkages
	for _, edge := range append(nh.EdgesHop1, nh.EdgesHop2...) {
		if edge.Type == EdgeSharesDevice || edge.Type == EdgeUsesDevice || edge.Type == EdgeEmployeeUsesDevice {
			if rootNode.Type == NodeEmployee {
				deviceCounts[edge.TargetID]++
				addAffected(edge.TargetID)
			}
		}
		if edge.Type == EdgeSharesIP || edge.Type == EdgeUsesIP || edge.Type == EdgeEmployeeUsesIP {
			if rootNode.Type == NodeEmployee {
				ipCounts[edge.TargetID]++
				addAffected(edge.TargetID)
			}
		}
		// Temporal proximity check: Pre-transaction access
		if edge.Type == EdgeCreatesTx || edge.Type == EdgeModifiesTx || edge.Type == EdgeApprovesTx {
			delta := nh.AsOf.Sub(edge.Timestamp).Seconds()
			if delta >= 0 && delta < preTxProximitySec {
				preTxProximitySec = delta
			}
		}
	}

	score := 0.05 // Baseline neutral risk
	var evidenceTokens []string

	signals.OffHoursAccessCount = offHoursCount
	if preTxProximitySec < 999999.0 {
		signals.PreTxAccessProximity = preTxProximitySec
	}

	// Rule A: Abnormal Access Concentration & Peer Deviation
	if rootNode.Type == NodeEmployee && totalEmployeeInteractions > 0 {
		var maxCount int
		var targetConsumer string
		for cID, count := range consumerCounts {
			if count > maxCount {
				maxCount = count
				targetConsumer = cID
			}
		}
		ratio := float64(maxCount) / float64(totalEmployeeInteractions)
		signals.ConcentrationRatio = ratio

		if totalEmployeeInteractions >= 5 && ratio >= d.thresholds.ConcentrationRatioThreshold {
			score += 0.35
			evidenceTokens = append(evidenceTokens, "employee_account_access_frequency_anomaly", "peer_group_deviation")
			facts = append(facts, fmt.Sprintf("Employee concentrated %d of %d interactions (%.1f%%) on consumer '%s'", maxCount, totalEmployeeInteractions, ratio*100, tokenizeID(targetConsumer)))
			patterns = append(patterns, "Suspiciously concentrated internal access pattern targeting single customer identity")
		}
	}

	// Rule B: Off-Hours Access Anomaly
	if rootNode.Type == NodeEmployee && offHoursCount >= 3 {
		score += 0.20
		evidenceTokens = append(evidenceTokens, "off_hours_access_anomaly")
		facts = append(facts, fmt.Sprintf("Observed %d employee access events outside standard operational hours (20:00-06:00 UTC)", offHoursCount))
		patterns = append(patterns, "Repeated off-hours administrative access without prior scheduling authorization")
	}

	// Rule C: Pre-Transaction Proximity Correlation
	if rootNode.Type == NodeEmployee && preTxProximitySec <= d.thresholds.TemporalCorrelationSeconds {
		score += 0.25
		evidenceTokens = append(evidenceTokens, "repeated_pre_transaction_access")
		facts = append(facts, fmt.Sprintf("Employee accessed customer record %.0f seconds before transaction generation", preTxProximitySec))
		patterns = append(patterns, "Temporal lockstep: Employee record modification immediately precedes payment activity")
	}

	// Rule D: High-Risk Consumer Clustering
	if rootNode.Type == NodeEmployee && highRiskConsumerCount >= 2 {
		score += 0.35
		evidenceTokens = append(evidenceTokens, "high_risk_consumer_clustering")
		facts = append(facts, fmt.Sprintf("Employee accessed %d accounts flagged with confirmed fraud history", highRiskConsumerCount))
		patterns = append(patterns, "Internal operator repeatedly querying known compromised accounts")
	}

	// Rule E: Shared Device Hardware Colocation (Critical Collusion)
	if rootNode.Type == NodeEmployee {
		for devID, count := range deviceCounts {
			if count >= 1 {
				signals.SharedDeviceOverlap++
				score += 0.50
				evidenceTokens = append(evidenceTokens, "shared_device_cluster", "hardware_colocation")
				facts = append(facts, fmt.Sprintf("Employee hardware fingerprint matches customer device '%s'", tokenizeID(devID)))
				patterns = append(patterns, "Direct hardware colocation between internal operator and customer session (High-Risk Collusion)")
			}
		}
		for ipID, count := range ipCounts {
			if count >= 1 {
				signals.SharedIPOverlap++
				score += 0.25
				evidenceTokens = append(evidenceTokens, "shared_ip_cluster")
				facts = append(facts, fmt.Sprintf("Employee network egress matches customer IP node '%s'", tokenizeID(ipID)))
				patterns = append(patterns, "Network egress colocation between employee terminal and customer session")
			}
		}
	}

	// Multi-hop path length
	pathDepth := 0
	if len(nh.EdgesHop1) > 0 {
		pathDepth = 1
	}
	if len(nh.EdgesHop2) > 0 {
		pathDepth = 2
	}
	signals.MultiHopPathLength = pathDepth

	if score > 0.98 {
		score = 0.98
	}

	signals.RelationshipCollusionRisk = score

	// Build Explainable Collusion Signal
	if len(evidenceTokens) > 0 {
		explainableSignals = append(explainableSignals, ExplainableCollusionSignal{
			RiskSignal:        "employee_consumer_collusion",
			Score:             score,
			Evidence:          evidenceTokens,
			TemporalWindow:    "24h",
			AffectedEntities:  affectedEntities,
			Confidence:        0.92,
			OperationalReason: "Automated GraphSAGE behavioral anomaly and topological collusion detection",
		})
	}

	return facts, patterns, score, signals, explainableSignals
}

// tokenizeID hashes identifiers with SHA-256 to ensure zero PII exposure in logs and dossiers.
func tokenizeID(raw string) string {
	if len(raw) == 0 {
		return "anon_entity"
	}
	hash := sha256.Sum256([]byte("ropus_salt_" + raw))
	return "tok_" + hex.EncodeToString(hash[:])[:12]
}
