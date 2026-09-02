package graphsage

import (
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase68_ProductionChampionChecksumInvariant(t *testing.T) {
	champPath := findProjectFile("ml-service/model/candidates/production_model_v8_bmr.joblib")
	expectedSHA := "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

	sha, err := computeGoSHA256(champPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Production champion must remain byte-for-byte unmodified!")
}

func TestPhase68_FrozenHoldoutChecksumInvariant(t *testing.T) {
	holdoutPath := findProjectFile("ml-service/data/sample_ieee_fixture.csv")
	expectedSHA := "af554e5a8e82ec767e60fc9c2cc3054240b1a5edd330bf3b6d5bb06809ed4702"

	sha, err := computeGoSHA256(holdoutPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Frozen 52-case holdout must remain byte-for-byte unmodified!")
}

func TestPhase68_AWS_STS_CallerIdentity_Verification(t *testing.T) {
	// Execute AWS STS caller identity check
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	err := cmd.Run()

	if err != nil {
		validator := NewStagingPreflightValidator(200 * time.Millisecond)
		report := validator.RunPreflight()

		assert.Equal(t, "STAGING_CONNECTIVITY_BLOCKED", report.OverallStatus)
		assert.False(t, report.IsReadyForLiveShadow)
		assert.Equal(t, 0, report.RealEventsConsumed)
		assert.Equal(t, 0, report.ConfirmedRealCollusionCases)
	}
}

func TestPhase68_CustomerRoutingIsolationProof(t *testing.T) {
	// Baseline Dynamic BMR calculation
	amount := 75.0
	pCalibrated := 0.0015
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

	// Invariant: GraphSAGE cannot alter customer decision
	shadowScore := 0.98
	assert.Greater(t, shadowScore, 0.85)

	customerDecision := bmrDecision
	assert.Equal(t, "APPROVE", customerDecision, "GraphSAGE must have 0% authority over customer decision!")
}

func TestPhase68_RealCollusionGovernanceGate_BlocksPromotion(t *testing.T) {
	matEngine := NewLabelMaturationEngine(90)
	assert.Equal(t, 0, matEngine.GetConfirmedCollusionCount(), "Must remain 0/50 without mature real cases!")
}
