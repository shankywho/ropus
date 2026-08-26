package graphsage

import (
	"context"
	"testing"
	"time"
)

func TestPhase60_PrivacySanitizer_QuarantineProhibitedFields(t *testing.T) {
	sanitizer := NewPrivacySanitizer("test_salt")

	// Payload with raw PAN and CVV
	rawPayload := map[string]interface{}{
		"employee_id": "emp_01",
		"account_id":  "acct_01",
		"raw_pan":     "4111222233334444",
		"cvv":         "123",
		"notes":       "test notes",
	}

	sanitized, violations, isQuarantined := sanitizer.SanitizePayload(rawPayload)
	if !isQuarantined {
		t.Fatalf("Expected payload to be quarantined due to raw PAN/CVV")
	}
	if len(violations) < 2 {
		t.Fatalf("Expected at least 2 violations, got %d", len(violations))
	}
	if _, ok := sanitized["raw_pan"]; ok {
		t.Fatalf("Prohibited raw_pan leaked into sanitized payload")
	}
	if _, ok := sanitized["cvv"]; ok {
		t.Fatalf("Prohibited cvv leaked into sanitized payload")
	}
}

func TestPhase60_PassiveIngestion_DedupAndOutOfOrder(t *testing.T) {
	ctx := context.Background()
	store := NewTemporalHeteroGraphStore()
	adapter := NewPassiveShadowIngestionAdapter(store, false)

	now := time.Now().UTC()
	t1 := now.Add(-10 * time.Minute)
	t2 := now.Add(-5 * time.Minute)
	tPast := now.Add(-20 * time.Minute) // Out-of-order late arrival

	// 1. Consume normal event
	ok, _ := adapter.ConsumeEvent(ctx, "evt_1", "audit.events", map[string]interface{}{
		"employee_id": "emp_01",
		"account_id":  "acct_01",
	}, t1)
	if !ok {
		t.Fatalf("Expected evt_1 to be ingested")
	}

	// 2. Consume duplicate event
	okDup, msgDup := adapter.ConsumeEvent(ctx, "evt_1", "audit.events", map[string]interface{}{
		"employee_id": "emp_01",
		"account_id":  "acct_01",
	}, t2)
	if okDup || msgDup != "DUPLICATE_DROPPED" {
		t.Fatalf("Expected duplicate event to be dropped")
	}

	// 3. Consume out-of-order event
	okLate, _ := adapter.ConsumeEvent(ctx, "evt_2", "audit.events", map[string]interface{}{
		"employee_id": "emp_02",
		"account_id":  "acct_02",
	}, tPast)
	if !okLate {
		t.Fatalf("Expected late event to be ingested safely")
	}

	metrics := adapter.GetMetrics()
	if metrics.EventsConsumed != 3 {
		t.Fatalf("Expected 3 consumed events, got %d", metrics.EventsConsumed)
	}
	if metrics.EventsDuplicated != 1 {
		t.Fatalf("Expected 1 duplicate, got %d", metrics.EventsDuplicated)
	}
	if metrics.EventsOutOfOrder != 1 {
		t.Fatalf("Expected 1 out of order event, got %d", metrics.EventsOutOfOrder)
	}
}

func TestPhase60_LabelMaturationAndEvidenceLedger(t *testing.T) {
	matEngine := NewLabelMaturationEngine(90)
	now := time.Now().UTC()

	// Register fresh txn
	matEngine.RegisterTransaction("txn_001", now.AddDate(0, 0, -30))
	if matEngine.GetConfirmedCollusionCount() != 0 {
		t.Fatalf("Expected 0 confirmed collusion initially")
	}

	// Confirm as internal collusion
	matEngine.ConfirmLabel("txn_001", LabelConfirmedFraud, "INTERNAL_INVESTIGATION", now, true)
	if matEngine.GetConfirmedCollusionCount() != 1 {
		t.Fatalf("Expected 1 confirmed collusion after confirmation")
	}

	// Record dossier in evidence ledger
	ledger := NewShadowEvidenceLedger()
	dossier := &CollusionInvestigationDossier{
		DossierID:        "dos_001",
		OverallRiskScore: 0.88,
		RiskLevel:        RiskLevelHigh,
		EmployeeID:       "emp_01",
		EmployeeRole:     "customer_support",
	}
	entry := ledger.RecordDossier(dossier)
	if entry.DossierID != "dos_001" {
		t.Fatalf("Expected dossier ID dos_001 in ledger entry")
	}
	if ledger.TotalEntries() != 1 {
		t.Fatalf("Expected 1 entry in ledger")
	}
}
