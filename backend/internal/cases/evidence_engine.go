package cases

import (
	"fmt"
	"time"
)

// EvidenceEngine aggregates and normalizes forensic evidence from distributed intelligence feeds.
type EvidenceEngine struct{}

// NewEvidenceEngine initializes the evidence collector.
func NewEvidenceEngine() *EvidenceEngine {
	return &EvidenceEngine{}
}

// BuildEvidencePackage compiles multi-vector forensic indicators into structured evidence artifacts.
func (e *EvidenceEngine) BuildEvidencePackage(
	txnID string,
	amount float64,
	graphFraudNeighbors int,
	behaviorAnomalies []string,
	threatMatches []string,
	modelScore float64,
) []EvidenceItem {
	var items []EvidenceItem
	now := time.Now().UTC()

	// 1. Transaction Signal
	items = append(items, EvidenceItem{
		EvidenceID: fmt.Sprintf("ev_txn_%s", txnID),
		Type:       "TRANSACTION",
		Summary:    fmt.Sprintf("Transaction for $%.2f scored at ML risk %.2f", amount, modelScore),
		Details:    fmt.Sprintf("Transaction ID %s evaluated by primary ML model with score %.4f", txnID, modelScore),
		RiskWeight: modelScore,
		CapturedAt: now,
	})

	// 2. Graph Intelligence Evidence
	if graphFraudNeighbors > 0 {
		items = append(items, EvidenceItem{
			EvidenceID: fmt.Sprintf("ev_grp_%s", txnID),
			Type:       "GRAPH",
			Summary:    fmt.Sprintf("Direct graph linkage to %d confirmed fraudulent entities", graphFraudNeighbors),
			Details:    fmt.Sprintf("Entity resolution identified %d high-risk neighbor connections across shared device/IP nodes", graphFraudNeighbors),
			RiskWeight: 0.90,
			CapturedAt: now,
		})
	}

	// 3. Behavioral Anomaly Evidence
	for i, anom := range behaviorAnomalies {
		items = append(items, EvidenceItem{
			EvidenceID: fmt.Sprintf("ev_beh_%s_%d", txnID, i),
			Type:       "BEHAVIOR",
			Summary:    anom,
			Details:    fmt.Sprintf("Behavioral profiling anomaly: %s", anom),
			RiskWeight: 0.75,
			CapturedAt: now,
		})
	}

	// 4. Threat Intelligence Evidence
	for i, threat := range threatMatches {
		items = append(items, EvidenceItem{
			EvidenceID: fmt.Sprintf("ev_thr_%s_%d", txnID, i),
			Type:       "THREAT_INTEL",
			Summary:    threat,
			Details:    fmt.Sprintf("Threat intelligence feed match: %s", threat),
			RiskWeight: 0.95,
			CapturedAt: now,
		})
	}

	return items
}
