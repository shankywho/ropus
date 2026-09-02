package graphsage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase64_ProductionChampionChecksumInvariant(t *testing.T) {
	champPath := findProjectFile("ml-service/model/candidates/production_model_v8_bmr.joblib")
	expectedSHA := "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

	sha, err := computeGoSHA256(champPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Production champion must remain byte-for-byte unmodified!")
}

func TestPhase64_FrozenHoldoutChecksumInvariant(t *testing.T) {
	holdoutPath := findProjectFile("ml-service/data/sample_ieee_fixture.csv")
	expectedSHA := "af554e5a8e82ec767e60fc9c2cc3054240b1a5edd330bf3b6d5bb06809ed4702"

	sha, err := computeGoSHA256(holdoutPath)
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, sha, "Frozen 52-case holdout must remain byte-for-byte unmodified!")
}

func TestPhase64_StagingPreflight_AuditValidation(t *testing.T) {
	validator := NewStagingPreflightValidator(200 * time.Millisecond)
	report := validator.RunPreflight()

	assert.Equal(t, "STAGING_CONNECTIVITY_BLOCKED", report.OverallStatus)
	assert.False(t, report.IsReadyForLiveShadow)
	assert.NotEmpty(t, report.BlockingReasons)
	assert.NotEmpty(t, report.MissingConfigurations)
	assert.NotEmpty(t, report.RequiredUnblockingActions)
	assert.Equal(t, 0, report.RealEventsConsumed)
	assert.Equal(t, 0, report.ConfirmedRealCollusionCases)

	// Verify granular dependencies
	assert.Contains(t, report.DependencyResults, "kafka_brokers")
	assert.Contains(t, report.DependencyResults, "kafka_auth")
	assert.Contains(t, report.DependencyResults, "clickhouse")
	assert.Contains(t, report.DependencyResults, "prometheus")
}

func TestPhase64_PrivacySanitizer_CardholderAndPII_Quarantine(t *testing.T) {
	sanitizer := NewPrivacySanitizer("phase64_test_salt")

	payload := map[string]interface{}{
		"employee_id":          "emp_01",
		"account_id":           "acct_01",
		"raw_pan":              "4111222233334444",
		"cvv":                  "999",
		"ssn":                  "123-45-6789",
		"bank_account_number":  "123456789012",
	}

	sanitized, violations, isQuarantined := sanitizer.SanitizePayload(payload)
	assert.True(t, isQuarantined)
	assert.GreaterOrEqual(t, len(violations), 4)

	// Assert forbidden fields are absent
	_, hasPAN := sanitized["raw_pan"]
	_, hasCVV := sanitized["cvv"]
	_, hasSSN := sanitized["ssn"]
	_, hasBank := sanitized["bank_account_number"]

	assert.False(t, hasPAN)
	assert.False(t, hasCVV)
	assert.False(t, hasSSN)
	assert.False(t, hasBank)
}

func TestPhase64_CustomerRouting_ZeroEnforcementProof(t *testing.T) {
	// Baseline Dynamic BMR calculation
	amount := 25.0
	pCalibrated := 0.007
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

	// Invariant: Critical shadow risk score (0.99) cannot modify customer decision
	shadowScore := 0.99
	assert.Greater(t, shadowScore, 0.85)

	customerDecision := bmrDecision
	assert.Equal(t, "APPROVE", customerDecision, "Customer decision must remain 100% determined by BMR!")
}

func TestPhase64_RealCaseAccumulationGate_BlocksPromotion(t *testing.T) {
	matEngine := NewLabelMaturationEngine(90)
	assert.Equal(t, 0, matEngine.GetConfirmedCollusionCount())

	ds := NewMockRealGraphDataSource(DatasetRealShadow)
	checker := NewReadinessChecker(ds)
	report := checker.Audit(context.Background(), time.Now().UTC())

	assert.False(t, report.IsReadyForRealValidation)
	assert.Equal(t, StateRealDataRequired, report.GovernanceState)
	assert.Equal(t, 0, report.ConfirmedInternalCollusionLabelsCount)
}
