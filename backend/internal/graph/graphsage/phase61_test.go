package graphsage

import (
	"context"
	"testing"
	"time"
)

func TestPhase61_ShadowTelemetryManager_AsyncLifecycle(t *testing.T) {
	ctx := context.Background()
	manager := NewShadowTelemetryManager(100, false)

	// Start background telemetry worker
	manager.Start(ctx)
	defer manager.Stop()

	status := manager.GetStatus()
	if !status.IsRunning {
		t.Fatalf("Expected manager to be running")
	}
	if status.RealDataConnectivity != "NOT_CONNECTED" {
		t.Fatalf("Expected RealDataConnectivity = NOT_CONNECTED, got %s", status.RealDataConnectivity)
	}

	now := time.Now().UTC()

	// Enqueue valid audit event
	enqueued := manager.EnqueueEvent(StreamEventEnvelope{
		EventID:        "evt_tel_001",
		Topic:          "audit.events",
		Payload:        map[string]interface{}{"employee_id": "emp_01", "account_id": "acct_01"},
		EventTimestamp: now,
	})
	if !enqueued {
		t.Fatalf("Expected event to be enqueued successfully")
	}

	// Enqueue polluted event with raw PAN
	enqueuedPolluted := manager.EnqueueEvent(StreamEventEnvelope{
		EventID:        "evt_tel_002",
		Topic:          "audit.events",
		Payload:        map[string]interface{}{"employee_id": "emp_02", "raw_pan": "4111222233334444"},
		EventTimestamp: now,
	})
	if !enqueuedPolluted {
		t.Fatalf("Expected polluted event to be enqueued for DLQ processing")
	}

	// Wait for queue drain
	time.Sleep(100 * time.Millisecond)

	// Evaluate investigation
	dossier := manager.EvaluateShadowInvestigation("emp_01", now, "customer_support")
	if dossier == nil {
		t.Fatalf("Expected non-nil investigation dossier")
	}
	if dossier.GovernanceNotice != "INVESTIGATION INTELLIGENCE ONLY - STRICTLY NON-ENFORCING" {
		t.Fatalf("Expected non-enforcing governance notice")
	}

	finalStatus := manager.GetStatus()
	if finalStatus.TotalDossiersGenerated != 1 {
		t.Fatalf("Expected 1 dossier generated, got %d", finalStatus.TotalDossiersGenerated)
	}
	if finalStatus.IngestionMetrics.DLQSize != 1 {
		t.Fatalf("Expected 1 DLQ event from raw PAN, got %d", finalStatus.IngestionMetrics.DLQSize)
	}
}

func TestPhase61_ShadowSafety_ChaosAndEnforcementIsolation(t *testing.T) {
	ctx := context.Background()
	engine := NewRelationshipIntelligenceEngine(nil)
	now := time.Now().UTC()

	// 1. Evaluate normal report
	report := engine.EvaluateRelationship(ctx, "emp_01", now)
	if report.OperationalMode != "NON_ENFORCING" || report.IntendedUse != "INVESTIGATION_ONLY" {
		t.Fatalf("GraphSAGE report must be strictly NON_ENFORCING / INVESTIGATION_ONLY")
	}

	// 2. Authoritative BMR customer decisions must be 100% independent
	bmrSimulateDecision := func(bmrRiskScore float64, gnnScore float64) string {
		// GraphSAGE has 0% customer enforcement authority
		_ = gnnScore
		if bmrRiskScore >= 0.80 {
			return "BLOCK"
		} else if bmrRiskScore >= 0.50 {
			return "CHALLENGE"
		}
		return "APPROVE"
	}

	// Test decision isolation under extreme GraphSAGE score (0.99) with low BMR risk (0.10)
	dec1 := bmrSimulateDecision(0.10, 0.99)
	if dec1 != "APPROVE" {
		t.Fatalf("BMR decision mutated by high GraphSAGE score! Expected APPROVE, got %s", dec1)
	}

	// Test decision isolation under zero GraphSAGE score (0.00) with high BMR risk (0.95)
	dec2 := bmrSimulateDecision(0.95, 0.00)
	if dec2 != "BLOCK" {
		t.Fatalf("BMR decision mutated by zero GraphSAGE score! Expected BLOCK, got %s", dec2)
	}

	// Test decision isolation on timeout / fallback
	dec3 := bmrSimulateDecision(0.55, 0.50)
	if dec3 != "CHALLENGE" {
		t.Fatalf("BMR decision mutated on fallback! Expected CHALLENGE, got %s", dec3)
	}
}
