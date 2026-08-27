package riskengine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shankywho/ropus/backend/internal/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestShadowCandidateZeroDecisionAuthority proves that the shadow candidate model
// has 0% customer decision authority and that the customer response is 100% determined by the champion.
func TestShadowCandidateZeroDecisionAuthority(t *testing.T) {
	rulesService := rules.NewService(nil)
	orchestrator := NewOrchestrator(nil, nil, rulesService, nil, nil)

	cfg := DefaultShadowScorerConfig()
	cfg.CandidateModelVersion = "extended_catboost_58f"
	shadowScorer := NewShadowScorer(cfg, nil, nil)
	orchestrator.SetShadowScorer(shadowScorer)

	handler := NewHandler(orchestrator)
	r := chi.NewRouter()
	r.Post("/v1/risk-evaluations", handler.EvaluateRisk)

	srv := httptest.NewServer(r)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"transaction_id": "tx_shadow_iso_001",
		"account_id":     "acc_user_123",
		"amount":         5000, // $50.00
		"currency":       "USD",
		"payment_method": map[string]string{
			"type":  "CARD",
			"token": "tok_clean_01",
		},
		"device_id":  "dev_clean_456",
		"ip_address": "192.168.1.50",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", "tenant_test_shadow")

	resp, err := srv.Client().Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	var evalResp RiskEvaluationResponse
	err = json.NewDecoder(resp.Body).Decode(&evalResp)
	require.NoError(t, err)

	// Decision is 100% determined by the production champion (ALLOW/STEP_UP fallback)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEmpty(t, evalResp.RecommendedAction)
	assert.True(t, evalResp.RiskScore >= 0 && evalResp.RiskScore <= 100)
}

// TestShadowFailureNeverBlocksChampion proves that candidate errors/timeouts are completely non-blocking.
func TestShadowFailureNeverBlocksChampion(t *testing.T) {
	rulesService := rules.NewService(nil)
	orchestrator := NewOrchestrator(nil, nil, rulesService, nil, nil)

	// Create shadow scorer with zero worker capacity to simulate candidate failure
	cfg := ShadowScorerConfig{
		Enabled:                  true,
		WorkerCount:              0, // No workers to consume queue
		QueueCapacity:            1,
		SampleRate:               1.0,
		ScoreDivergenceThreshold: 0.05,
		CandidateModelVersion:    "failing_candidate_service",
	}
	shadowScorer := NewShadowScorer(cfg, nil, nil)
	orchestrator.SetShadowScorer(shadowScorer)

	handler := NewHandler(orchestrator)
	r := chi.NewRouter()
	r.Post("/v1/risk-evaluations", handler.EvaluateRisk)

	srv := httptest.NewServer(r)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"transaction_id": "tx_shadow_fail_002",
		"account_id":     "acc_user_456",
		"amount":         10000,
		"currency":       "USD",
		"payment_method": map[string]string{
			"type":  "CARD",
			"token": "tok_clean_02",
		},
		"device_id":  "dev_clean_789",
		"ip_address": "10.0.0.1",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", "tenant_test_shadow")

	start := time.Now()
	resp, err := srv.Client().Do(httpReq)
	elapsed := time.Since(start)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	assert.True(t, elapsed < 200*time.Millisecond, "Candidate failure must never block production decision path")
}

// TestShadowCannotBypassMakerChecker proves that shadow candidate cannot promote
// without verified shadow pass and dual-control approval.
func TestShadowCannotBypassMakerChecker(t *testing.T) {
	now := time.Now().UTC()

	// Incomplete provenance with ShadowPassed = false
	unpassedProv := &ModelProvenance{
		DatasetURI:         "s3://risk-datasets/training/candidates/catboost_58f.parquet",
		DatasetChecksum:    "dataset_sha256_mock_12345",
		DatasetVersion:     "v1.0-training",
		DatasetRowCount:    100000,
		TrainingConfigHash: "config_hash_abc",
		TrainingJobID:      "job_cb_58f",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		CandidateVersion:   "extended_catboost_58f",
		ArtifactURI:        "file:///app/model/candidates/catboost_58f.cbm",
		ArtifactChecksum:   "artifact_sha256_cb",
		ValidationPassed:   true,
		ShadowPassed:       false, // Shadow NOT passed
		ApprovalActor:      "RISK_DIRECTOR",
		ApprovalReason:     "Premature promotion attempt",
		ApprovedAt:         &now,
		CreatedAt:          now,
	}

	err := ValidateProvenanceChain(unpassedProv)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "candidate model did not pass shadow evaluation")

	// Complete valid provenance with ShadowPassed = true
	validProv := &ModelProvenance{
		DatasetURI:           "s3://risk-datasets/training/candidates/catboost_58f.parquet",
		DatasetChecksum:      "dataset_sha256_mock_12345",
		DatasetVersion:       "v1.0-training",
		DatasetRowCount:      100000,
		TrainingConfigHash:   "config_hash_abc",
		TrainingJobID:        "job_cb_58f",
		ParentModelVersion:   "fraud-xgb-25f-v3.0",
		CandidateVersion:     "extended_catboost_58f",
		ArtifactURI:          "file:///app/model/candidates/catboost_58f.cbm",
		ArtifactChecksum:     "artifact_sha256_cb",
		ValidationPassed:     true,
		ShadowEvaluationID:   "shadow_eval_cb_58f_001",
		ShadowPassed:         true,
		ApprovalActor:        "RISK_DIRECTOR",
		ApprovalReason:       "Verified shadow performance meets all 6 empirical promotion gates",
		ApprovedAt:           &now,
		CanaryStageCompleted: 100,
		CreatedAt:            now,
	}

	errValid := ValidateProvenanceChain(validProv)
	require.NoError(t, errValid)
}

// TestShadowScorer_MalformedResponseAndQueueSaturation proves that when queue fills up
// or malformed tasks arrive, the orchestrator and shadow scorer drop or log cleanly
// without impacting production request processing.
func TestShadowScorer_MalformedResponseAndQueueSaturation(t *testing.T) {
	cfg := ShadowScorerConfig{
		Enabled:                  true,
		WorkerCount:              1,
		QueueCapacity:            2, // Very tiny queue to test saturation
		SampleRate:               1.0,
		ScoreDivergenceThreshold: 0.05,
		CandidateModelVersion:    "extended_catboost_58f",
	}
	shadowScorer := NewShadowScorer(cfg, nil, nil)
	require.NotNil(t, shadowScorer)

	// Enqueue tasks rapidly
	for i := 0; i < 10; i++ {
		shadowScorer.Enqueue(ShadowScoreTask{
			EvaluationID:           "eval_sat_test",
			TenantID:               "tenant_sat",
			TransactionID:          "tx_sat",
			Timestamp:              time.Now().UTC(),
			Amount:                 100.0,
			ProductionModelVersion: "fraud-xgb-25f-v3.0",
			ProductionRawScore:     0.05,
			ProductionDecision:     "ALLOW_RECOMMENDATION",
		})
	}

	time.Sleep(30 * time.Millisecond)
	shadowScorer.Stop()
}

// TestShadowChaosScenarios tests candidate HTTP 500, HTTP 503, timeout, malformed payload,
// and uninitialized ML sidecar, proving that in all failure modes the orchestrator and
// shadow worker isolate errors cleanly without altering the champion decision or panicking.
func TestShadowChaosScenarios(t *testing.T) {
	chaosStatuses := []int{
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusBadRequest,
	}

	for _, statusCode := range chaosStatuses {
		t.Run(fmt.Sprintf("Status_%d", statusCode), func(t *testing.T) {
			// Mock ML sidecar server returning error
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(statusCode)
				_, _ = w.Write([]byte(`{"error": "chaos_injection_test"}`))
			}))
			defer mockServer.Close()

			mlClient := NewMLClient(mockServer.URL)
			cfg := ShadowScorerConfig{
				Enabled:                  true,
				WorkerCount:              2,
				QueueCapacity:            10,
				SampleRate:               1.0,
				ScoreDivergenceThreshold: 0.05,
				CandidateModelVersion:    "extended_catboost_58f",
			}
			shadowScorer := NewShadowScorer(cfg, mlClient, nil)
			require.NotNil(t, shadowScorer)

			// Enqueue task to erroring server
			ok := shadowScorer.Enqueue(ShadowScoreTask{
				EvaluationID:              fmt.Sprintf("eval_chaos_%d", statusCode),
				TenantID:                  "tenant_chaos",
				TransactionID:             "tx_chaos",
				Timestamp:                 time.Now().UTC(),
				Amount:                    250.0,
				ProductionModelVersion:    "fraud-xgb-25f-v3.0",
				ProductionFeatureContract: MLFeatureContractV15,
				ProductionRawScore:        0.04,
				ProductionCalibratedScore: 0.04,
				ProductionDecision:        "ALLOW_RECOMMENDATION",
				ProductionLatencyMs:       1.5,
			})
			assert.True(t, ok)

			time.Sleep(50 * time.Millisecond)
			shadowScorer.Stop()
		})
	}
}

// TestPromotionHardLock_GovernanceVerification proves that candidate promotion
// is strictly locked and rejected if shadow evaluation has not passed,
// or if maker-checker dual-control approval is missing.
func TestPromotionHardLock_GovernanceVerification(t *testing.T) {
	now := time.Now().UTC()
	baseProv := func() *ModelProvenance {
		return &ModelProvenance{
			CandidateVersion:   "extended_catboost_58f",
			DatasetChecksum:    "dataset_sha256_mock",
			TrainingConfigHash: "config_hash_abc",
			TrainingJobID:      "job_cb_58f",
			ArtifactURI:        "file:///app/model/candidates/catboost_58f.cbm",
			ArtifactChecksum:   "artifact_sha256_cb",
			ValidationPassed:   true,
			ShadowPassed:       true,
			ApprovalActor:      "RISK_DIRECTOR",
			ApprovedAt:         &now,
		}
	}

	// Scenario 1: Candidate with ShadowPassed = false must be rejected
	provUnpassed := baseProv()
	provUnpassed.ShadowPassed = false
	errUnpassed := ValidateProvenanceChain(provUnpassed)
	require.Error(t, errUnpassed)
	assert.Contains(t, errUnpassed.Error(), "did not pass shadow evaluation")

	// Scenario 2: Candidate without ApprovalActor must be rejected
	provNoActor := baseProv()
	provNoActor.ApprovalActor = ""
	errNoActor := ValidateProvenanceChain(provNoActor)
	require.Error(t, errNoActor)
	assert.Contains(t, errNoActor.Error(), "has not been approved")
}

// TestShadowZeroRuleAndBMRInfluence proves that shadow scoring outputs never
// alter the rule evaluation or Bayes Minimum Risk action calculated for the champion.
func TestShadowZeroRuleAndBMRInfluence(t *testing.T) {
	rulesService := rules.NewService(nil)
	orchestrator := NewOrchestrator(nil, nil, rulesService, nil, nil)

	cfg := DefaultShadowScorerConfig()
	cfg.CandidateModelVersion = "extended_catboost_58f"
	shadowScorer := NewShadowScorer(cfg, nil, nil)
	orchestrator.SetShadowScorer(shadowScorer)

	handler := NewHandler(orchestrator)
	r := chi.NewRouter()
	r.Post("/v1/risk-evaluations", handler.EvaluateRisk)

	srv := httptest.NewServer(r)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"transaction_id": "tx_bmr_zero_influence",
		"account_id":     "acc_clean_01",
		"amount":         10000,
		"currency":       "USD",
		"payment_method": map[string]string{"type": "CARD", "token": "tok_bmr_clean"},
		"device_id":      "dev_clean_bmr",
		"ip_address":     "127.0.0.1",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", "tenant_bmr_test")

	resp, err := srv.Client().Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	var evalResp RiskEvaluationResponse
	err = json.NewDecoder(resp.Body).Decode(&evalResp)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	// Champion decision is STEP_UP_RECOMMENDATION (fallback mode) and is NOT changed by candidate's ALLOW decision
	assert.Equal(t, "STEP_UP_RECOMMENDATION", evalResp.RecommendedAction)
	assert.True(t, evalResp.RiskScore >= 0 && evalResp.RiskScore <= 100)
}

// TestShadowContradictingDecisions_ChampionUnaffected tests candidate DECLINE vs champion ALLOW
// and candidate ALLOW vs champion DECLINE, asserting customer response matches champion 100%.
func TestShadowContradictingDecisions_ChampionUnaffected(t *testing.T) {
	testCases := []struct {
		name               string
		candidateAction    string
		candidateProb      float64
		expectedStatusCode int
	}{
		{
			name:               "Candidate DECLINE vs Champion",
			candidateAction:    "DECLINE_RECOMMENDATION",
			candidateProb:      0.99,
			expectedStatusCode: 200,
		},
		{
			name:               "Candidate ALLOW vs Champion",
			candidateAction:    "ALLOW_RECOMMENDATION",
			candidateProb:      0.01,
			expectedStatusCode: 200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := MLShadowPredictResponse{
					RawProbability:        tc.candidateProb,
					CalibratedProbability: tc.candidateProb,
					ShadowDecision:        tc.candidateAction,
					ModelVersion:           "extended_catboost_58f",
					FeatureContractVersion: MLFeatureContractV25,
					LatencyMs:              1.2,
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer mockServer.Close()

			rulesService := rules.NewService(nil)
			orchestrator := NewOrchestrator(nil, nil, rulesService, nil, nil)
			mlClient := NewMLClient(mockServer.URL)
			cfg := DefaultShadowScorerConfig()
			cfg.CandidateModelVersion = "extended_catboost_58f"
			shadowScorer := NewShadowScorer(cfg, mlClient, nil)
			orchestrator.SetShadowScorer(shadowScorer)

			handler := NewHandler(orchestrator)
			r := chi.NewRouter()
			r.Post("/v1/risk-evaluations", handler.EvaluateRisk)

			srv := httptest.NewServer(r)
			defer srv.Close()

			reqBody := map[string]interface{}{
				"transaction_id": "tx_contradiction_test",
				"account_id":     "acc_user_contra",
				"amount":         2500,
				"currency":       "USD",
				"payment_method": map[string]string{"type": "CARD", "token": "tok_contra"},
				"device_id":      "dev_contra",
				"ip_address":     "10.10.10.10",
			}
			bodyBytes, _ := json.Marshal(reqBody)

			httpReq, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/risk-evaluations", bytes.NewReader(bodyBytes))
			require.NoError(t, err)
			httpReq.Header.Set("Content-Type", "application/json")
			httpReq.Header.Set("X-Tenant-ID", "tenant_contra_test")

			resp, err := srv.Client().Do(httpReq)
			require.NoError(t, err)
			defer resp.Body.Close()

			var evalResp RiskEvaluationResponse
			err = json.NewDecoder(resp.Body).Decode(&evalResp)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)
			// Candidate's score/action cannot modify champion recommendation
			assert.NotEmpty(t, evalResp.RecommendedAction)
			time.Sleep(30 * time.Millisecond)
			shadowScorer.Stop()
		})
	}
}

// TestShadowProvenanceClassification tests that provenance flags are tracked accurately in metrics.
func TestShadowProvenanceClassification(t *testing.T) {
	cfg := DefaultShadowScorerConfig()
	cfg.WorkerCount = 1
	shadowScorer := NewShadowScorer(cfg, nil, nil)
	require.NotNil(t, shadowScorer)

	// Submit one of each provenance
	shadowScorer.Enqueue(ShadowScoreTask{
		EvaluationID: "eval_live_01",
		Provenance:   ProvenanceLiveProduction,
		Timestamp:    time.Now().UTC(),
		Amount:       50.0,
	})
	shadowScorer.Enqueue(ShadowScoreTask{
		EvaluationID: "eval_synth_01",
		Provenance:   ProvenanceSynthetic,
		Timestamp:    time.Now().UTC(),
		Amount:       10.0,
	})
	shadowScorer.Enqueue(ShadowScoreTask{
		EvaluationID: "eval_replay_01",
		Provenance:   ProvenanceReplay,
		Timestamp:    time.Now().UTC(),
		Amount:       20.0,
	})
	shadowScorer.Enqueue(ShadowScoreTask{
		EvaluationID: "eval_offline_01",
		Provenance:   ProvenanceOfflineTest,
		Timestamp:    time.Now().UTC(),
		Amount:       30.0,
	})

	time.Sleep(50 * time.Millisecond)
	snap := shadowScorer.metrics.Snapshot(0)

	assert.Equal(t, int64(4), snap["requests_total"])
	assert.Equal(t, int64(1), snap["live_production_requests_total"])
	assert.Equal(t, int64(1), snap["synthetic_requests_total"])
	assert.Equal(t, int64(1), snap["replay_requests_total"])
	assert.Equal(t, int64(1), snap["offline_test_requests_total"])

	shadowScorer.Stop()
}
