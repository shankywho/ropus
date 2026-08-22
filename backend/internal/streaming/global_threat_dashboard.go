package streaming

import (
	"time"
)

// GlobalThreatOverview summarizes collective intelligence, streaming alerts, and autonomous defense metrics.
type GlobalThreatOverview struct {
	Timestamp               time.Time `json:"timestamp"`
	ActiveCampaignsCount    int       `json:"active_campaigns_count"`
	AutonomousDefensesCount int       `json:"autonomous_defenses_count"`
	FederatedThreatsCount   int       `json:"federated_threats_count"`
	GlobalEntitiesMonitored int       `json:"global_entities_monitored"`
	NetworkThreatSeverity   string    `json:"network_threat_severity"` // "ELEVATED", "NORMAL", "CRITICAL"
	StreamingEngineStatus   string    `json:"streaming_engine_status"`
}

// GlobalThreatDashboard aggregates streaming telemetry.
type GlobalThreatDashboard struct {
	campaignDetector *CampaignDetector
	defenseEngine    *AutonomousDefenseEngine
	federatedMesh    *FederatedIntelligenceMesh
	globalGraph      *GlobalFraudGraph
}

// NewGlobalThreatDashboard initializes the global dashboard aggregator.
func NewGlobalThreatDashboard(
	cd *CampaignDetector,
	de *AutonomousDefenseEngine,
	fm *FederatedIntelligenceMesh,
	gg *GlobalFraudGraph,
) *GlobalThreatDashboard {
	return &GlobalThreatDashboard{
		campaignDetector: cd,
		defenseEngine:    de,
		federatedMesh:    fm,
		globalGraph:      gg,
	}
}

// GetOverview compiles real-time global defense telemetry.
func (d *GlobalThreatDashboard) GetOverview() *GlobalThreatOverview {
	campaigns := d.campaignDetector.ListCampaigns()
	severity := "NORMAL"
	if len(campaigns) > 0 {
		severity = "ELEVATED"
		for _, c := range campaigns {
			if c.Severity == "CRITICAL" {
				severity = "CRITICAL"
				break
			}
		}
	}

	return &GlobalThreatOverview{
		Timestamp:               time.Now().UTC(),
		ActiveCampaignsCount:    len(campaigns),
		AutonomousDefensesCount: len(d.defenseEngine.records),
		FederatedThreatsCount:   len(d.federatedMesh.signatures),
		GlobalEntitiesMonitored: len(d.globalGraph.nodes),
		NetworkThreatSeverity:   severity,
		StreamingEngineStatus:   "HEALTHY_STREAMING",
	}
}
