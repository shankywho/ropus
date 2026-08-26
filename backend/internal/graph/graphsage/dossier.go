package graphsage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateCollusionDossier creates a machine-readable investigation dossier in Go.
func GenerateCollusionDossier(
	empNode *HeteroNode,
	report *RelationshipIntelligenceReport,
	incidentEdges []*HeteroEdge,
	associatedNodes []*HeteroNode,
) *CollusionInvestigationDossier {
	b := make([]byte, 6)
	rand.Read(b)
	dossierID := fmt.Sprintf("dos_go_%s", hex.EncodeToString(b))

	now := time.Now().UTC()
	var affectedAccounts []string
	var affectedConsumers []string
	var sharedDevices []string
	var sharedIPs []string

	nodeMap := make(map[string]*HeteroNode)
	nodeMap[empNode.ID] = empNode

	for _, n := range associatedNodes {
		nodeMap[n.ID] = n
		switch n.Type {
		case NodeAccount:
			affectedAccounts = append(affectedAccounts, n.ID)
		case NodeConsumer:
			affectedConsumers = append(affectedConsumers, n.ID)
		case NodeDevice:
			sharedDevices = append(sharedDevices, n.ID)
		case NodeIP:
			sharedIPs = append(sharedIPs, n.ID)
		}
	}

	extractor := &PathExtractor{}
	paths := extractor.FindPaths(empNode.ID, "", nodeMap, incidentEdges, 3)
	if len(paths) == 0 {
		for _, e := range incidentEdges {
			paths = append(paths, fmt.Sprintf("(%s) -[%s]-> (%s)", e.SourceID, e.Type, e.TargetID))
		}
	}

	role := "unknown"
	if r, ok := empNode.Properties["role"].(string); ok {
		role = r
	}

	var supportingEvidence []string
	var counterEvidence []string

	if report.OverallRiskScore >= 0.70 {
		supportingEvidence = append(supportingEvidence, fmt.Sprintf("Elevated relationship risk score (%.2f) on cluster of %d accounts.", report.OverallRiskScore, len(affectedAccounts)))
	}
	if len(sharedDevices) > 0 {
		supportingEvidence = append(supportingEvidence, fmt.Sprintf("Shared device infrastructure detected across %d device(s).", len(sharedDevices)))
	}

	if role == "customer_support" || role == "fraud_analyst" {
		counterEvidence = append(counterEvidence, fmt.Sprintf("Employee role (%s) regularly performs legitimate customer service and account reviews.", role))
	}
	if len(affectedAccounts) > 10 {
		counterEvidence = append(counterEvidence, "High account dispersion consistent with standard customer queue processing.")
	}

	explanation := fmt.Sprintf(
		"Investigation Dossier for Employee %s (role: %s). Overall Risk: %.2f (Tier: %s). Supporting indicators: %d, Mitigating factors: %d.",
		empNode.ID, role, report.OverallRiskScore, report.RiskLevel, len(supportingEvidence), len(counterEvidence),
	)

	return &CollusionInvestigationDossier{
		DossierID:         dossierID,
		GeneratedAt:       now,
		OverallRiskScore:  report.OverallRiskScore,
		RiskLevel:         report.RiskLevel,
		EmployeeID:        empNode.ID,
		EmployeeRole:      role,
		AffectedAccounts:  affectedAccounts,
		AffectedConsumers: affectedConsumers,
		RelationshipPaths: paths,
		TemporalEvidence: map[string]interface{}{
			"total_interactions":  len(incidentEdges),
			"evaluated_at":        now.Format(time.RFC3339),
			"supporting_evidence": supportingEvidence,
			"counter_evidence":    counterEvidence,
		},
		TransactionSummary: map[string]interface{}{
			"affected_account_count": len(affectedAccounts),
		},
		SharedInfrastructure: map[string]interface{}{
			"devices": sharedDevices,
			"ips":     sharedIPs,
		},
		ConfidenceScore:  0.90,
		HumanExplanation: explanation,
		DataProvenance:   "shadow_graph_stream",
		ModelVersion:     report.ModelVersion,
		GovernanceNotice: "INVESTIGATION INTELLIGENCE ONLY - STRICTLY NON-ENFORCING",
	}
}
