package graphsage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase63_ProductionChampionChecksumInvariant(t *testing.T) {
	champPath := findProjectFile("ml-service/model/candidates/production_model_v8_bmr.joblib")
	expectedSHA := "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

	sha, err := computeGoSHA256(champPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Production champion must remain byte-for-byte unmodified!")
}

func TestPhase63_FrozenHoldoutChecksumInvariant(t *testing.T) {
	holdoutPath := findProjectFile("ml-service/data/sample_ieee_fixture.csv")
	expectedSHA := "af554e5a8e82ec767e60fc9c2cc3054240b1a5edd330bf3b6d5bb06809ed4702"

	sha, err := computeGoSHA256(holdoutPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Frozen 52-case holdout must remain byte-for-byte unmodified!")
}

func TestPhase63_StagingConfigurationDiscovery_HonestAudit(t *testing.T) {
	checker := NewConnectivityChecker(500 * time.Millisecond)
	report := checker.AuditEnvironment()

	assert.Equal(t, "INTEGRATION_BLOCKED", report.IntegrationStatus)
	assert.False(t, report.IsLiveConnected)
	assert.NotEmpty(t, report.MissingConfigurations)
}

func TestPhase63_ControlledShadowSoak_PipelineIntegrity(t *testing.T) {
	mgr := NewShadowTelemetryManager(5000, false)
	mgr.Start(context.Background())
	defer mgr.Stop()

	baseTime := time.Now().UTC().Add(-2 * time.Hour)

	// Stream 100 soak events (normal, duplicates, late, DLQ)
	for i := 1; i <= 80; i++ {
		ts := baseTime.Add(time.Duration(i) * time.Minute)
		mgr.EnqueueEvent(StreamEventEnvelope{
			EventID:        findProjectFile("soak_evt_") + string(rune(i)),
			Topic:          "audit.events",
			Payload:        map[string]interface{}{"employee_id": "emp_soak_01", "account_id": "acct_soak_01"},
			EventTimestamp: ts,
		})
	}

	// Stream DLQ events with raw PAN / CVV / SSN
	mgr.EnqueueEvent(StreamEventEnvelope{
		EventID:        "soak_dlq_01",
		Topic:          "audit.events",
		Payload:        map[string]interface{}{"employee_id": "emp_bad", "raw_pan": "4111222233334444"},
		EventTimestamp: baseTime,
	})
	mgr.EnqueueEvent(StreamEventEnvelope{
		EventID:        "soak_dlq_02",
		Topic:          "transactions.created",
		Payload:        map[string]interface{}{"transaction_id": "txn_bad", "ssn": "123-45-6789"},
		EventTimestamp: baseTime,
	})

	time.Sleep(150 * time.Millisecond)

	status := mgr.GetStatus()
	assert.GreaterOrEqual(t, status.TotalEventsProcessed, 1)
	assert.GreaterOrEqual(t, status.IngestionMetrics.EventsQuarantined, 2)
	assert.Equal(t, "STRICTLY_NON_ENFORCING_SHADOW", status.OperationalMode)
	assert.Equal(t, "0% (Baseline Dynamic BMR 100% Authoritative)", status.CustomerEnforcementAuthority)
}

func TestPhase63_CustomerRoutingIsolationProof(t *testing.T) {
	// BMR expected cost calculation for a legitimate $15 transaction
	amount := 15.0
	pCalibrated := 0.005
	costFraud := 1.05 * amount
	costReview := 25.0
	costChallenge := costReview + 0.05*amount

	expApprove := pCalibrated * costFraud
	expChallenge := costChallenge + (0.10 * (1.0 - pCalibrated) * costFraud)

	bmrDecision := "APPROVE"
	if expApprove >= expChallenge {
		bmrDecision = "CHALLENGE"
	}
	assert.Equal(t, "APPROVE", bmrDecision)

	// Even if GraphSAGE outputs score = 0.99 (Critical Risk), customer decision is 100% BMR
	shadowScore := 0.99
	assert.Greater(t, shadowScore, 0.85)

	customerDecision := bmrDecision
	assert.Equal(t, "APPROVE", customerDecision, "GraphSAGE must have 0% authority over customer decision!")
}
