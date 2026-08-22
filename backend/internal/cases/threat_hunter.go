package cases

import (
	"fmt"
	"time"
)

// ThreatFinding represents an anomalous pattern detected during threat hunting scans.
type ThreatFinding struct {
	FindingID     string    `json:"finding_id"`
	PatternType   string    `json:"pattern_type"` // "HIDDEN_CLUSTER", "RAPID_GRAPH_GROWTH", "SUSPICIOUS_MERCHANT_SPIKE"
	Description   string    `json:"description"`
	Severity      string    `json:"severity"` // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	InvolvedNodes []string  `json:"involved_nodes"`
	DiscoveredAt  time.Time `json:"discovered_at"`
}

// ThreatHuntReport summarizes the results of an autonomous threat hunting sweep.
type ThreatHuntReport struct {
	HuntID       string          `json:"hunt_id"`
	StartedAt    time.Time       `json:"started_at"`
	CompletedAt  time.Time       `json:"completed_at"`
	Findings     []ThreatFinding `json:"findings"`
	TotalThreats int             `json:"total_threats"`
	Status       string          `json:"status"`
}

// ThreatHunter scans knowledge graphs and transaction streams for emergent fraud campaigns.
type ThreatHunter struct{}

// NewThreatHunter initializes the threat hunter.
func NewThreatHunter() *ThreatHunter {
	return &ThreatHunter{}
}

// RunHuntSweep executes an automated threat hunting sweep across recent activity.
func (h *ThreatHunter) RunHuntSweep(graphNodeCount int, recentTransactionsCount int) *ThreatHuntReport {
	now := time.Now().UTC()
	var findings []ThreatFinding

	// Simulate detection of abnormal graph growth or emerging syndicates
	if graphNodeCount > 50 {
		findings = append(findings, ThreatFinding{
			FindingID:     fmt.Sprintf("fnd_%d_growth", now.UnixNano()),
			PatternType:   "RAPID_GRAPH_GROWTH",
			Description:   fmt.Sprintf("Knowledge graph expanded by %d nodes in last monitoring window", graphNodeCount),
			Severity:      "MEDIUM",
			InvolvedNodes: []string{"cluster_hub_alpha", "cluster_hub_beta"},
			DiscoveredAt:  now,
		})
	}

	return &ThreatHuntReport{
		HuntID:       fmt.Sprintf("hunt_%d", now.UnixNano()),
		StartedAt:    now.Add(-5 * time.Second),
		CompletedAt:  now,
		Findings:     findings,
		TotalThreats: len(findings),
		Status:       "COMPLETED",
	}
}
