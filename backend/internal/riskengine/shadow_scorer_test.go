package riskengine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// makeTestVec returns an immutable *MLFeatureVector suitable for shadow scoring tests.
func makeTestVec(amount float64) *MLFeatureVector {
	fm := make(map[string]float64, 25)
	for _, def := range Canonical25FeatureDefinitions {
		fm[def.Name] = def.DefaultValue
	}
	fm["amount"] = amount
	return &MLFeatureVector{
		Version:    MLFeatureContractV25,
		FeatureMap: fm,
	}
}

func TestShadowScorer_Disabled(t *testing.T) {
	cfg := ShadowScorerConfig{Enabled: false}
	scorer := NewShadowScorer(cfg, nil, nil)
	defer scorer.Close(1 * time.Second)

	task := ShadowScoreTask{
		EvaluationID:              "eval_test_01",
		TenantID:                  "tenant_01",
		Canonical25Vector:         makeTestVec(100.0),
		ProductionModelVersion:    "xgb-ieee-canonical-v2-calibrated",
		ProductionFeatureContract: MLFeatureContractV15,
		ProductionCalibratedScore: 0.03,
		ProductionDecision:        "ALLOW_RECOMMENDATION",
	}
	if scorer.Enqueue(task) {
		t.Errorf("Expected Enqueue to return false when shadow scorer is disabled, got true")
	}
	status := scorer.GetStatus()
	if status["enabled"] != false {
		t.Errorf("Expected status enabled to be false")
	}
}

func TestShadowScorer_EnabledAndAsyncExecution(t *testing.T) {
	var processedTask sync.WaitGroup
	processedTask.Add(1)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/predict/shadow" {
			var req MLShadowPredictRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			resp := MLShadowPredictResponse{
				ModelVersion:           "fraud-xgb-25f-candidate-v1",
				FeatureContractVersion: "v2.5",
				FeatureCount:           25,
				RawProbability:         0.04,
				CalibratedProbability:  0.035,
				RiskScore:              4,
				ShadowDecision:         "ALLOW_RECOMMENDATION",
				LatencyMs:              0.5,
				Runtime:                "mock",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			processedTask.Done()
		}
	}))
	defer mockServer.Close()

	mlClient := NewMLClient(mockServer.URL)
	cfg := ShadowScorerConfig{
		Enabled:                  true,
		WorkerCount:              2,
		QueueCapacity:            50,
		SampleRate:               1.0,
		ScoreDivergenceThreshold: 0.05,
		CandidateModelVersion:    "fraud-xgb-25f-candidate-v1",
		CandidateFeatureContract: MLFeatureContractV25,
	}
	scorer := NewShadowScorer(cfg, mlClient, nil)
	defer scorer.Close(2 * time.Second)

	task := ShadowScoreTask{
		EvaluationID:              "eval_test_02",
		TenantID:                  "tenant_01",
		TransactionID:             "txn_123",
		Timestamp:                 time.Now().UTC(),
		Amount:                    250.0,
		Canonical25Vector:         makeTestVec(250.0),
		ProductionModelVersion:    "xgb-ieee-canonical-v2-calibrated",
		ProductionFeatureContract: MLFeatureContractV15,
		ProductionCalibratedScore: 0.03,
		ProductionDecision:        "ALLOW_RECOMMENDATION",
	}
	if !scorer.Enqueue(task) {
		t.Fatalf("Expected Enqueue to return true")
	}
	processedTask.Wait()
	// Give the worker goroutine time to finish incrementing success_total after the HTTP round-trip.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		st := scorer.GetStatus()
		if st["metrics"].(map[string]interface{})["success_total"].(int64) >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	status := scorer.GetStatus()
	metrics := status["metrics"].(map[string]interface{})
	if metrics["success_total"].(int64) < 1 {
		t.Errorf("Expected success_total >= 1, got %v", metrics["success_total"])
	}
}

func TestShadowScorer_QueueSaturationNonBlockingDrop(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(MLShadowPredictResponse{ShadowDecision: "ALLOW_RECOMMENDATION"})
	}))
	defer mockServer.Close()

	mlClient := NewMLClient(mockServer.URL)
	cfg := ShadowScorerConfig{
		Enabled:       true,
		WorkerCount:   1,
		QueueCapacity: 2,
		SampleRate:    1.0,
	}
	scorer := NewShadowScorer(cfg, mlClient, nil)
	defer scorer.Close(2 * time.Second)

	drops := 0
	for i := 0; i < 10; i++ {
		if !scorer.Enqueue(ShadowScoreTask{EvaluationID: "eval_sat", Canonical25Vector: makeTestVec(50.0)}) {
			drops++
		}
	}
	if drops == 0 {
		t.Errorf("Expected some tasks to be dropped due to queue saturation")
	}
	status := scorer.GetStatus()
	metrics := status["metrics"].(map[string]interface{})
	if metrics["queue_dropped_total"].(int64) == 0 {
		t.Errorf("Expected queue_dropped_total > 0, got %v", metrics["queue_dropped_total"])
	}
}

func TestShadowScorer_Sampling(t *testing.T) {
	cfg0 := ShadowScorerConfig{Enabled: true, SampleRate: 0.0}
	scorer0 := NewShadowScorer(cfg0, nil, nil)
	defer scorer0.Close(1 * time.Second)
	if scorer0.Enqueue(ShadowScoreTask{EvaluationID: "eval_0", Canonical25Vector: makeTestVec(0)}) {
		t.Errorf("Expected 0%% sample rate to reject all enqueues")
	}

	cfg100 := ShadowScorerConfig{Enabled: true, SampleRate: 1.0, QueueCapacity: 10}
	scorer100 := NewShadowScorer(cfg100, nil, nil)
	defer scorer100.Close(1 * time.Second)
	if !scorer100.Enqueue(ShadowScoreTask{EvaluationID: "eval_100", Canonical25Vector: makeTestVec(0)}) {
		t.Errorf("Expected 100%% sample rate to accept enqueue")
	}
}

func TestShadowScorer_DivergenceCategorization(t *testing.T) {
	scorer := &ShadowScorer{
		config: ShadowScorerConfig{ScoreDivergenceThreshold: 0.05},
	}
	tests := []struct {
		prodAction   string
		shadowAction string
		absDelta     float64
		expectedCat  string
	}{
		{"ALLOW_RECOMMENDATION", "ALLOW_RECOMMENDATION", 0.01, "FULL_AGREEMENT"},
		{"ALLOW_RECOMMENDATION", "ALLOW_RECOMMENDATION", 0.08, "SCORE_DIVERGENCE_DECISION_AGREEMENT"},
		{"ALLOW_RECOMMENDATION", "MANUAL_REVIEW", 0.06, "ALLOW_TO_REVIEW"},
		{"ALLOW_RECOMMENDATION", "DECLINE_RECOMMENDATION", 0.40, "ALLOW_TO_DECLINE"},
		{"MANUAL_REVIEW", "ALLOW_RECOMMENDATION", 0.06, "REVIEW_TO_ALLOW"},
		{"MANUAL_REVIEW", "DECLINE_RECOMMENDATION", 0.20, "REVIEW_TO_DECLINE"},
		{"DECLINE_RECOMMENDATION", "ALLOW_RECOMMENDATION", 0.50, "DECLINE_TO_ALLOW"},
		{"DECLINE_RECOMMENDATION", "MANUAL_REVIEW", 0.20, "DECLINE_TO_REVIEW"},
	}
	for _, tt := range tests {
		cat := scorer.categorizeDivergence(tt.prodAction, tt.shadowAction, tt.absDelta)
		if cat != tt.expectedCat {
			t.Errorf("categorizeDivergence(%s, %s, %v) = %s; want %s",
				tt.prodAction, tt.shadowAction, tt.absDelta, cat, tt.expectedCat)
		}
	}
}

func TestShadowScorer_ProductionIndependenceOnMLError(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"detail":"internal server error"}`))
		wg.Done()
	}))
	defer mockServer.Close()

	mlClient := NewMLClient(mockServer.URL)
	cfg := ShadowScorerConfig{Enabled: true, WorkerCount: 2, QueueCapacity: 50, SampleRate: 1.0}
	scorer := NewShadowScorer(cfg, mlClient, nil)
	defer scorer.Close(2 * time.Second)

	task := ShadowScoreTask{
		EvaluationID:              "eval_err",
		Canonical25Vector:         makeTestVec(100.0),
		ProductionCalibratedScore: 0.02,
		ProductionDecision:        "ALLOW_RECOMMENDATION",
	}
	if !scorer.Enqueue(task) {
		t.Fatalf("Enqueue failed")
	}
	wg.Wait()
	// Give the worker goroutine time to finish incrementing errors_total after the HTTP round-trip.
	deadline2 := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline2) {
		st := scorer.GetStatus()
		if st["metrics"].(map[string]interface{})["errors_total"].(int64) >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	status := scorer.GetStatus()
	metrics := status["metrics"].(map[string]interface{})
	if metrics["errors_total"].(int64) < 1 {
		t.Errorf("Expected errors_total >= 1, got %v", metrics["errors_total"])
	}
}

func TestShadowScorer_ConcurrencyStressAndRace(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(MLShadowPredictResponse{
			RawProbability: 0.03, CalibratedProbability: 0.025,
			ShadowDecision: "ALLOW_RECOMMENDATION", LatencyMs: 0.1,
		})
	}))
	defer mockServer.Close()

	mlClient := NewMLClient(mockServer.URL)
	cfg := ShadowScorerConfig{Enabled: true, WorkerCount: 4, QueueCapacity: 200, SampleRate: 1.0}
	scorer := NewShadowScorer(cfg, mlClient, nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				scorer.Enqueue(ShadowScoreTask{
					EvaluationID:              "eval_concurrent",
					TenantID:                  "tenant_conc",
					Canonical25Vector:         makeTestVec(100.0),
					ProductionCalibratedScore: 0.02,
					ProductionDecision:        "ALLOW_RECOMMENDATION",
				})
			}
		}()
	}
	wg.Wait()
	_ = scorer.Close(3 * time.Second)
}
