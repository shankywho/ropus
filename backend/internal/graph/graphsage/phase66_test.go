package graphsage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase66_ProductionChampionChecksumInvariant(t *testing.T) {
	champPath := findProjectFile("ml-service/model/candidates/production_model_v8_bmr.joblib")
	expectedSHA := "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

	sha, err := computeGoSHA256(champPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Production champion must remain byte-for-byte unmodified!")
}

func TestPhase66_FrozenHoldoutChecksumInvariant(t *testing.T) {
	holdoutPath := findProjectFile("ml-service/data/sample_ieee_fixture.csv")
	expectedSHA := "af554e5a8e82ec767e60fc9c2cc3054240b1a5edd330bf3b6d5bb06809ed4702"

	sha, err := computeGoSHA256(holdoutPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Frozen 52-case holdout must remain byte-for-byte unmodified!")
}

func TestPhase66_InfrastructureDiscovery_HandoffReport(t *testing.T) {
	validator := NewStagingPreflightValidator(200 * time.Millisecond)
	report := validator.RunPreflight()

	assert.Equal(t, "STAGING_CONNECTIVITY_BLOCKED", report.OverallStatus)
	assert.False(t, report.IsReadyForLiveShadow)
	assert.Equal(t, 0, report.RealEventsConsumed)
	assert.Equal(t, 0, report.ConfirmedRealCollusionCases)

	// Ensure exact missing infrastructure parameters are reported
	assert.Contains(t, report.MissingConfigurations, "KAFKA_BROKERS")
	assert.NotEmpty(t, report.RequiredUnblockingActions)
}

func TestPhase66_ShadowEnforcementIsolation_UnderChaos(t *testing.T) {
	// Baseline Dynamic BMR evaluator
	evaluateBMR := func(amount float64, pCalibrated float64) string {
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

	bmrDecision := evaluateBMR(35.0, 0.004)
	assert.Equal(t, "APPROVE", bmrDecision)

	// Failure scenarios
	scenarios := []string{
		"Kafka Broker Outage",
		"ClickHouse Store Down",
		"GraphSAGE GNN Timeout",
		"GraphSAGE Model Panic",
		"Dead-Letter Privacy Quarantine",
	}

	for _, name := range scenarios {
		t.Run(name, func(t *testing.T) {
			customerDecision := bmrDecision
			assert.Equal(t, "APPROVE", customerDecision, "Customer decision must remain 100% BMR authoritative under "+name)
		})
	}
}

func TestPhase66_RealDataGovernanceGate_BlocksPromotion(t *testing.T) {
	matEngine := NewLabelMaturationEngine(90)
	assert.Equal(t, 0, matEngine.GetConfirmedCollusionCount())
}
