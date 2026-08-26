package graphsage

import (
	"context"
	"testing"
	"time"
)

func TestReadinessCheckerAndGovernanceGate(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	// 1. Synthetic dataset source -> must be blocked from promotion
	synthSource := NewMockRealGraphDataSource(DatasetSynthetic)
	synthSource.AddNode(&HeteroNode{ID: "emp_01", Type: NodeEmployee, CreatedAt: now})
	synthSource.AddNode(&HeteroNode{ID: "acct_01", Type: NodeAccount, CreatedAt: now})
	synthSource.AddEdge(&HeteroEdge{ID: "e_01", SourceID: "emp_01", TargetID: "acct_01", Type: EdgeAccessesAccount, Timestamp: now})

	checkerSynth := NewReadinessChecker(synthSource)
	reportSynth := checkerSynth.Audit(ctx, now)

	if reportSynth.IsReadyForRealValidation {
		t.Fatalf("Synthetic dataset MUST NOT be ready for production validation")
	}
	if reportSynth.GovernanceState != StateRealDataRequired {
		t.Fatalf("Expected StateRealDataRequired, got %v", reportSynth.GovernanceState)
	}

	// 2. Real dataset with insufficient labels (< 50) -> must be blocked
	realSource := NewMockRealGraphDataSource(DatasetRealShadow)
	realSource.AddNode(&HeteroNode{ID: "emp_01", Type: NodeEmployee, CreatedAt: now})
	realSource.AddNode(&HeteroNode{ID: "acct_01", Type: NodeAccount, CreatedAt: now})
	realSource.AddEdge(&HeteroEdge{ID: "e_01", SourceID: "emp_01", TargetID: "acct_01", Type: EdgeAccessesAccount, Timestamp: now})
	realSource.AddLabel(map[string]interface{}{
		"entity_id":      "acct_01",
		"maturity_state": LabelConfirmedFraud,
		"is_collusion":   true,
	})

	checkerReal := NewReadinessChecker(realSource)
	reportReal := checkerReal.Audit(ctx, now)

	if reportReal.IsReadyForRealValidation {
		t.Fatalf("Real dataset with only 1 label MUST NOT pass the gate")
	}
	if reportReal.ConfirmedFraudLabelsCount != 1 {
		t.Fatalf("Expected 1 confirmed label, got %d", reportReal.ConfirmedFraudLabelsCount)
	}
}

func TestCollusionDossierGeneration(t *testing.T) {
	now := time.Now().UTC()
	empNode := &HeteroNode{
		ID:        "emp_suspicious_01",
		Type:      NodeEmployee,
		CreatedAt: now,
		Properties: map[string]interface{}{
			"role": "customer_support",
		},
	}
	report := &RelationshipIntelligenceReport{
		RootNodeID:       "emp_suspicious_01",
		OverallRiskScore: 0.85,
		RiskLevel:        RiskLevelHigh,
		ModelVersion:     "graphsage-v1.0-shadow",
	}
	incidentEdges := []*HeteroEdge{
		{ID: "e_1", SourceID: "emp_suspicious_01", TargetID: "acct_mule_1", Type: EdgeAccessesAccount, Timestamp: now},
	}
	associatedNodes := []*HeteroNode{
		{ID: "acct_mule_1", Type: NodeAccount, CreatedAt: now},
		{ID: "usr_victim_1", Type: NodeConsumer, CreatedAt: now},
	}

	dossier := GenerateCollusionDossier(empNode, report, incidentEdges, associatedNodes)
	if dossier == nil {
		t.Fatalf("Dossier generation failed")
	}
	if dossier.EmployeeID != "emp_suspicious_01" {
		t.Fatalf("Expected employee ID emp_suspicious_01, got %s", dossier.EmployeeID)
	}
	if len(dossier.AffectedAccounts) != 1 || dossier.AffectedAccounts[0] != "acct_mule_1" {
		t.Fatalf("Expected affected account acct_mule_1, got %v", dossier.AffectedAccounts)
	}
	if dossier.GovernanceNotice != "INVESTIGATION INTELLIGENCE ONLY - STRICTLY NON-ENFORCING" {
		t.Fatalf("Expected non-enforcing notice")
	}
}
