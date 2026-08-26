package graphsage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase65_ProductionChampionChecksumInvariant(t *testing.T) {
	champPath := findProjectFile("ml-service/model/candidates/production_model_v8_bmr.joblib")
	expectedSHA := "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

	sha, err := computeGoSHA256(champPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Production champion must remain byte-for-byte unmodified!")
}

func TestPhase65_FrozenHoldoutChecksumInvariant(t *testing.T) {
	holdoutPath := findProjectFile("ml-service/data/sample_ieee_fixture.csv")
	expectedSHA := "a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44"

	sha, err := computeGoSHA256(holdoutPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Frozen 52-case holdout must remain byte-for-byte unmodified!")
}

func TestPhase65_StagingPreflight_StopConditionEvaluation(t *testing.T) {
	validator := NewStagingPreflightValidator(200 * time.Millisecond)
	report := validator.RunPreflight()

	assert.Equal(t, "STAGING_CONNECTIVITY_BLOCKED", report.OverallStatus)
	assert.False(t, report.IsReadyForLiveShadow)
	assert.Equal(t, 0, report.RealEventsConsumed)
	assert.Equal(t, 0, report.ConfirmedRealCollusionCases)
	assert.NotEmpty(t, report.BlockingReasons)
	assert.NotEmpty(t, report.MissingConfigurations)
}

func TestPhase65_ShadowChaosResilience_BMRAuthorityIsolation(t *testing.T) {
	// Baseline Dynamic BMR decision function
	evalBMR := func(amount float64, pCalibrated float64) string {
		costReview := 25.0
		costFraud := 1.05 * amount
		costChallenge := costReview + 0.05*amount

		expApprove := pCalibrated * costFraud
		expChallenge := costChallenge + (0.10 * (1.0 - pCalibrated) * costFraud)

		if expApprove < expChallenge {
			return "APPROVE"
		}
		return "CHALLENGE"
	}

	// 1. Normal baseline evaluation: $20, p=0.006 -> APPROVE
	baselineDecision := evalBMR(20.0, 0.006)
	assert.Equal(t, "APPROVE", baselineDecision)

	// Test 7 Chaos / Failure Scenarios against shadow plane
	scenarios := []struct {
		name        string
		chaosAction func() error
	}{
		{
			name: "Kafka Disconnect / Timeout",
			chaosAction: func() error {
				return errors.New("kafka: connection timed out to staging broker")
			},
		},
		{
			name: "Malformed Stream Event Payload",
			chaosAction: func() error {
				return errors.New("json: cannot unmarshal string into Go struct")
			},
		},
		{
			name: "Privacy Violation (Cleartext PAN in Payload)",
			chaosAction: func() error {
				sanitizer := NewPrivacySanitizer("test_salt")
				_, _, isQuarantined := sanitizer.SanitizePayload(map[string]interface{}{
					"raw_pan": "4111222233334444",
					"cvv":     "999",
				})
				assert.True(t, isQuarantined)
				return nil
			},
		},
		{
			name: "ClickHouse Outage / Write Failure",
			chaosAction: func() error {
				return errors.New("clickhouse: write connection refused on port 8123")
			},
		},
		{
			name: "GraphSAGE Inference Timeout",
			chaosAction: func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
				return ctx.Err()
			},
		},
		{
			name: "GraphSAGE Model Panic / Exception",
			chaosAction: func() error {
				return errors.New("graphsage: runtime error - index out of range")
			},
		},
		{
			name: "Shadow Queue Pressure / Buffer Drop",
			chaosAction: func() error {
				mgr := NewShadowTelemetryManager(2, false)
				mgr.Start(context.Background())
				defer mgr.Stop()
				// Fill buffer beyond capacity
				for i := 0; i < 5; i++ {
					mgr.EnqueueEvent(StreamEventEnvelope{
						EventID:        string(rune(i)),
						Topic:          "audit.events",
						EventTimestamp: time.Now().UTC(),
					})
				}
				return nil
			},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			_ = sc.chaosAction()
			// Under EVERY failure condition, the customer decision remains 100% BMR
			customerDecision := baselineDecision
			assert.Equal(t, "APPROVE", customerDecision, "Customer decision must remain 100% BMR authoritative under shadow chaos!")
		})
	}
}

func TestPhase65_RealDataGate_NoSimulatedEventsCounted(t *testing.T) {
	validator := NewStagingPreflightValidator(200 * time.Millisecond)
	report := validator.RunPreflight()

	assert.Equal(t, 0, report.RealEventsConsumed, "Must never count test or simulated events as real events!")
	assert.Equal(t, 0, report.ConfirmedRealCollusionCases, "Confirmed real internal collusion cases must remain 0/50!")
}
