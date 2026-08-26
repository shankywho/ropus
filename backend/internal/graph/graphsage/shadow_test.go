package graphsage

import (
	"context"
	"testing"
	"time"
)

func TestShadowSafety_BMRImmutability(t *testing.T) {
	// Shadow safety invariant test: GraphSAGE outputs must NEVER alter BMR customer decisions
	ctx := context.Background()
	engine := NewRelationshipIntelligenceEngine(nil)

	scenarios := []struct {
		name       string
		rootID     string
		isDegraded bool
	}{
		{"Extreme High Risk", "emp_rogue_999", false},
		{"Zero Risk", "emp_clean_001", false},
		{"Non-existent Entity", "emp_ghost_404", true},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			report := engine.EvaluateRelationship(ctx, sc.rootID, time.Now())
			if report == nil {
				t.Fatalf("Expected valid non-nil report")
			}
			if report.OperationalMode != "NON_ENFORCING" {
				t.Fatalf("Expected NON_ENFORCING, got %s", report.OperationalMode)
			}
			if report.IntendedUse != "INVESTIGATION_ONLY" {
				t.Fatalf("Expected INVESTIGATION_ONLY, got %s", report.IntendedUse)
			}
		})
	}
}

func TestEmployeeBehavioralBaseline(t *testing.T) {
	now := time.Now().UTC()
	empID := "emp_baseline_01"
	engine := &EmployeeBaselineEngine{}

	// Historical edges: normal hours (10:00 AM)
	var hist []*HeteroEdge
	for day := 10; day >= 1; day-- {
		ts := now.AddDate(0, 0, -day).Truncate(24 * time.Hour).Add(10 * time.Hour)
		hist = append(hist, &HeteroEdge{
			ID:        "e_hist",
			SourceID:  empID,
			TargetID:  "acct_hist",
			Type:      EdgeAccessesAccount,
			Timestamp: ts,
		})
	}

	// Current event: off-hours (02:00 AM)
	currTS := now.Truncate(24 * time.Hour).Add(2 * time.Hour)
	curr := []*HeteroEdge{
		{
			ID:        "e_curr",
			SourceID:  empID,
			TargetID:  "acct_curr",
			Type:      EdgeAccessesAccount,
			Timestamp: currTS,
		},
	}

	profile := engine.ComputeBaselineAndDeviation(empID, curr, hist, now)
	if !profile.HasHistoricalBaseline {
		t.Fatalf("Expected valid baseline")
	}
	if !profile.IsUnusualHour {
		t.Fatalf("Expected unusual hour detection for 02:00 AM access")
	}
	if profile.PersonalAnomalyDelta < 0.30 {
		t.Fatalf("Expected elevated anomaly delta, got %.2f", profile.PersonalAnomalyDelta)
	}
}

func TestMultiHopPathExtraction(t *testing.T) {
	now := time.Now().UTC()
	nodes := map[string]*HeteroNode{
		"emp_1":  {ID: "emp_1", Type: NodeEmployee, CreatedAt: now},
		"acct_1": {ID: "acct_1", Type: NodeAccount, CreatedAt: now},
		"usr_1":  {ID: "usr_1", Type: NodeConsumer, CreatedAt: now},
	}
	edges := []*HeteroEdge{
		{ID: "e1", SourceID: "emp_1", TargetID: "acct_1", Type: EdgeAccessesAccount, Timestamp: now},
		{ID: "e2", SourceID: "usr_1", TargetID: "acct_1", Type: EdgeOwns, Timestamp: now},
	}

	extractor := &PathExtractor{}
	paths := extractor.FindPaths("emp_1", "usr_1", nodes, edges, 3)
	if len(paths) == 0 {
		t.Fatalf("Expected multi-hop path from emp_1 to usr_1")
	}
}

func TestGraphDriftMonitoring(t *testing.T) {
	now := time.Now().UTC()
	nodes := []*HeteroNode{
		{ID: "n1", Type: NodeEmployee, CreatedAt: now},
		{ID: "n2", Type: NodeAccount, CreatedAt: now},
	}
	edges := []*HeteroEdge{
		{ID: "e1", SourceID: "n1", TargetID: "n2", Type: EdgeAccessesAccount, Timestamp: now},
	}
	scores := []float64{0.10, 0.85}

	monitor := &GraphDriftMonitor{}
	report := monitor.EvaluateDrift(nodes, edges, scores)

	if report.AverageDegree != 1.0 {
		t.Fatalf("Expected avg degree 1.0, got %.2f", report.AverageDegree)
	}
	if report.DataDriftStatus != "STABLE" {
		t.Fatalf("Expected STABLE data drift, got %s", report.DataDriftStatus)
	}
}
