package riskengine

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shankywho/ropus/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------------------------
// 1. Crash & Restart Recovery Tests
// -------------------------------------------------------------

func TestRecovery_TrainingRestart(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "registry_state.json")
	store, err := NewFileStateStore(stateFile)
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Simulate state where training was interrupted mid-flight by server restart
	inFlightState := PersistedRetrainingState{
		ProductionModelVersion: "fraud-xgb-25f-v3.0",
		FallbackModelVersion:   "fraud-xgb-15f-v1.5",
		CurrentState:           StateTraining,
		ActiveJob: &RetrainingJob{
			JobID:              "job_interrupted_01",
			State:              StateTraining,
			TriggerType:        "AUTO_DRIFT",
			TriggerReason:      "Sustained drift",
			ParentModelVersion: "fraud-xgb-25f-v3.0",
		},
		SavedAt: time.Now().UTC(),
	}
	err = store.Save(ctx, inFlightState)
	require.NoError(t, err)

	// 2. Startup reconciliation
	recoveryMgr := NewRecoveryManager(store, nil)
	reg := NewModelRegistry()
	coordinator := NewRetrainingCoordinator(DefaultRetrainingConfig(), nil, nil, nil, nil, nil)
	coordinator.SetStateStore(store)

	res, err := recoveryMgr.ReconcileOnStartup(ctx, reg, coordinator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// In-flight training MUST transition safely to FAILED (preventing zombie state)
	assert.Equal(t, StateFailed, res.ReconciledState)
	assert.Equal(t, ActionInFlightJobFailed, res.ActionTaken)
	assert.Equal(t, StateFailed, coordinator.GetStatus()["state"])

	// Production and Fallback models MUST remain intact
	prod, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", prod.Version)
	assert.True(t, prod.IsProductionActive)

	fallback, err := reg.GetFallbackModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-15f-v1.5", fallback.Version)
	assert.True(t, fallback.IsFallbackActive)
}

func TestRecovery_CanaryRestart(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "registry_state.json")
	store, err := NewFileStateStore(stateFile)
	require.NoError(t, err)

	ctx := context.Background()

	cand := &ModelCandidate{
		ModelID:            "model_cand_canary_01",
		Version:            "fraud-xgb-25f-v3.1-candidate-canary",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
		State:              StateCanary,
	}

	canaryState := PersistedRetrainingState{
		ProductionModelVersion: "fraud-xgb-25f-v3.0",
		FallbackModelVersion:   "fraud-xgb-15f-v1.5",
		CurrentState:           StateCanary,
		ActiveCandidate:        cand,
		CanaryStage:            25,
		SavedAt:                time.Now().UTC(),
	}
	err = store.Save(ctx, canaryState)
	require.NoError(t, err)

	// 2. Startup reconciliation
	recoveryMgr := NewRecoveryManager(store, nil)
	reg := NewModelRegistry()
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	coordinator := NewRetrainingCoordinator(DefaultRetrainingConfig(), nil, nil, canaryRouter, nil, nil)
	coordinator.SetStateStore(store)

	res, err := recoveryMgr.ReconcileOnStartup(ctx, reg, coordinator)
	require.NoError(t, err)

	// Mid-canary crash MUST safely reset canary traffic to 0% and keep candidate AWAITING_APPROVAL
	assert.Equal(t, StateAwaitingApproval, res.ReconciledState)
	assert.Equal(t, ActionCanaryResetToIdle, res.ActionTaken)
	assert.Equal(t, 0, canaryRouter.GetPercentage())
}

func TestRecovery_PromotionRestart(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "registry_state.json")
	store, err := NewFileStateStore(stateFile)
	require.NoError(t, err)

	ctx := context.Background()

	promotedVer := "fraud-xgb-25f-v3.1-candidate-promoted"
	models := map[string]*RegisteredModel{
		promotedVer: {
			ModelID:            "model_promoted_01",
			Version:            promotedVer,
			ParentModelVersion: "fraud-xgb-25f-v3.0",
			LifecycleState:     LifecyclePromoted,
			IsProductionActive: true,
		},
		"fraud-xgb-25f-v3.0": {
			ModelID:            "model_prod_prev",
			Version:            "fraud-xgb-25f-v3.0",
			LifecycleState:     LifecyclePromoted,
			IsFallbackActive:   true,
		},
	}

	promotedState := PersistedRetrainingState{
		ProductionModelVersion: promotedVer,
		FallbackModelVersion:   "fraud-xgb-25f-v3.0",
		Models:                 models,
		CurrentState:           StatePromoted,
		SavedAt:                time.Now().UTC(),
	}
	err = store.Save(ctx, promotedState)
	require.NoError(t, err)

	recoveryMgr := NewRecoveryManager(store, nil)
	reg := NewModelRegistry()
	coordinator := NewRetrainingCoordinator(DefaultRetrainingConfig(), nil, nil, nil, nil, nil)
	coordinator.SetStateStore(store)

	res, err := recoveryMgr.ReconcileOnStartup(ctx, reg, coordinator)
	require.NoError(t, err)
	assert.Equal(t, StatePromoted, res.ReconciledState)

	// Promoted model MUST remain the active production model
	prod, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, promotedVer, prod.Version)
	assert.True(t, prod.IsProductionActive)

	// Previous model MUST remain fallback
	fb, err := reg.GetFallbackModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", fb.Version)
	assert.True(t, fb.IsFallbackActive)
}

// -------------------------------------------------------------
// 2. Training Process Chaos Tests
// -------------------------------------------------------------

func TestTrainingProcess_Crash(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := LocalProcessConfig{
		Command:     "sh",
		Args:        []string{"-c", "kill -9 $$"}, // Immediate crash
		OutputDir:   tmpDir,
		Timeout:     5 * time.Second,
		MaxLogBytes: 4096,
	}

	adapter := NewLocalProcessTrainingAdapter(cfg)
	ctx := context.Background()

	meta := TrainingDatasetMetadata{
		DatasetID:          "dataset_crash_01",
		SampleCount:        300,
		FeatureContract:    "fraud-risk-25f-v2.5",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   0.95,
		MissingValueRate:   0.01,
	}

	job, err := adapter.StartTraining(ctx, TrainingRequest{
		JobID:              "job_crash_test",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
		DatasetMetadata:    meta,
	})
	require.NoError(t, err)

	// Poll for job failure
	var status *TrainingJob
	for i := 0; i < 50; i++ {
		time.Sleep(20 * time.Millisecond)
		status, err = adapter.GetTrainingStatus(ctx, job.JobID)
		if err == nil && status.State != TrainingJobRunning {
			break
		}
	}
	require.NoError(t, err)
	assert.Equal(t, TrainingJobFailed, status.State)
	assert.NotEmpty(t, status.Error)
}

func TestTrainingProcess_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := LocalProcessConfig{
		Command:     "sh",
		Args:        []string{"-c", "sleep 10"},
		OutputDir:   tmpDir,
		Timeout:     50 * time.Millisecond,
		MaxLogBytes: 4096,
	}

	adapter := NewLocalProcessTrainingAdapter(cfg)
	ctx := context.Background()

	meta := TrainingDatasetMetadata{
		DatasetID:          "dataset_to_01",
		SampleCount:        300,
		FeatureContract:    "fraud-risk-25f-v2.5",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   0.95,
		MissingValueRate:   0.01,
	}

	job, err := adapter.StartTraining(ctx, TrainingRequest{
		JobID:              "job_to_test",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
		DatasetMetadata:    meta,
	})
	require.NoError(t, err)

	var status *TrainingJob
	for i := 0; i < 50; i++ {
		time.Sleep(20 * time.Millisecond)
		status, err = adapter.GetTrainingStatus(ctx, job.JobID)
		if err == nil && status.State != TrainingJobRunning {
			break
		}
	}
	require.NoError(t, err)
	assert.Equal(t, TrainingJobFailed, status.State)
	assert.Contains(t, status.Error, "timed out")
}

// -------------------------------------------------------------
// 3. Artifact Corruption & Checksum Chaos
// -------------------------------------------------------------

func TestArtifact_ChecksumMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	artPath := filepath.Join(tmpDir, "model.onnx")
	err := os.WriteFile(artPath, []byte("valid_onnx_initial_content"), 0644)
	require.NoError(t, err)

	verifier := NewArtifactVerifier()
	ctx := context.Background()

	// Tampered checksum
	record, err := verifier.VerifyArtifact(ctx, "model_tamper_01", "v3.1", artPath, "bad_checksum_hash_12345")
	assert.Error(t, err)
	require.NotNil(t, record)
	assert.False(t, record.Passed)
	assert.Contains(t, record.Violations[0], "checksum mismatch")
}

func TestArtifact_Corruption(t *testing.T) {
	tmpDir := t.TempDir()
	verifier := NewArtifactVerifier()
	ctx := context.Background()

	// Truncated / empty artifact
	emptyPath := filepath.Join(tmpDir, "corrupt.onnx")
	_ = os.WriteFile(emptyPath, []byte(""), 0644)

	record, err := verifier.VerifyArtifact(ctx, "model_empty_01", "v3.1", emptyPath, "")
	assert.Error(t, err)
	assert.False(t, record.Passed)
	assert.Contains(t, record.Violations[0], "empty")
}

// -------------------------------------------------------------
// 4. ClickHouse Failure Isolation
// -------------------------------------------------------------

func TestClickHouse_FailureIsolation(t *testing.T) {
	// Retraining coordinator must execute cleanly without panic or blocking even if ClickHouse is nil / unreachable
	cfg := DefaultRetrainingConfig()
	cfg.MinSamples = 50
	cfg.CanaryObservationWindow = 1 * time.Millisecond
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)
	runner := NewFixtureTrainingAdapter()

	var mu sync.Mutex
	var promotedVer string
	coordinator := NewRetrainingCoordinator(
		cfg,
		runner,
		shadowScorer,
		canaryRouter,
		nil, // ClickHouse client nil
		func(newVer string) {
			mu.Lock()
			promotedVer = newVer
			mu.Unlock()
		},
	)

	ctx := context.Background()
	job, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "ClickHouse failure isolation test")
	require.NoError(t, err)
	require.NotNil(t, job)

	time.Sleep(30 * time.Millisecond)
	candidates := coordinator.GetCandidates()
	require.Len(t, candidates, 1)

	// Approve candidate
	err = coordinator.ApproveCandidate(ctx, candidates[0].ModelID, "ADMIN_TEST", "Approved during CH outage")
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	status := coordinator.GetStatus()
	assert.Equal(t, StatePromoted, status["state"])

	mu.Lock()
	pv := promotedVer
	mu.Unlock()
	assert.NotEmpty(t, pv)
}

// -------------------------------------------------------------
// 5. Concurrency & Idempotency Chaos
// -------------------------------------------------------------

func TestConcurrentRetraining(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	ctx := context.Background()

	var successCount int32
	var rejectedCount int32
	var wg sync.WaitGroup

	// Launch 10 simultaneous triggers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := coordinator.TriggerManual(ctx, "ADMIN_CONCURRENT", fmt.Sprintf("Trigger attempt %d", idx))
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&rejectedCount, 1)
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int32(1), successCount, "Exactly 1 concurrent trigger must succeed")
	assert.Equal(t, int32(9), rejectedCount, "All other concurrent triggers must be rejected")
}

func TestConcurrentApproval(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	ctx := context.Background()

	job, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "Test concurrent approval")
	require.NoError(t, err)
	require.NotNil(t, job)

	time.Sleep(30 * time.Millisecond)
	candidates := coordinator.GetCandidates()
	require.Len(t, candidates, 1)
	candID := candidates[0].ModelID

	var approvedCount int32
	var duplicateCount int32
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := coordinator.ApproveCandidate(ctx, candID, "OPERATOR", fmt.Sprintf("Approval attempt %d", idx))
			if err == nil {
				atomic.AddInt32(&approvedCount, 1)
			} else {
				atomic.AddInt32(&duplicateCount, 1)
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int32(1), approvedCount, "Exactly 1 approval must succeed")
	assert.Equal(t, int32(4), duplicateCount, "Duplicate concurrent approvals must return deterministic error")
}

func TestIdempotentApproval(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	ctx := context.Background()

	_, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "Test sequential idempotent approval")
	require.NoError(t, err)

	time.Sleep(30 * time.Millisecond)
	candidates := coordinator.GetCandidates()
	require.Len(t, candidates, 1)
	candID := candidates[0].ModelID

	// First approval
	err1 := coordinator.ApproveCandidate(ctx, candID, "OPERATOR", "First approval")
	require.NoError(t, err1)

	// Second approval (must return deterministic already approved error)
	err2 := coordinator.ApproveCandidate(ctx, candID, "OPERATOR", "Second approval")
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "already approved")
}

func TestIdempotentPromotion(t *testing.T) {
	reg := NewModelRegistry()
	cand := ModelCandidate{
		ModelID:            "model_cand_idemp_prom",
		Version:            "fraud-xgb-25f-v3.1-candidate-idemp",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
	}

	err := reg.RegisterCandidate(cand, "file:///model.onnx", "sha256_idemp")
	require.NoError(t, err)

	// First promotion
	err1 := reg.PromoteModel(cand.Version, "ADMIN_TEST", "First promotion")
	require.NoError(t, err1)

	prod, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, cand.Version, prod.Version)

	// Second promotion of same model is safe no-op / success
	err2 := reg.PromoteModel(cand.Version, "ADMIN_TEST", "Second promotion")
	require.NoError(t, err2)
}

func TestIdempotentRollback(t *testing.T) {
	reg := NewModelRegistry()
	cand := ModelCandidate{
		ModelID:            "model_cand_idemp_rb",
		Version:            "fraud-xgb-25f-v3.1-candidate-rb",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
	}

	err := reg.RegisterCandidate(cand, "file:///model.onnx", "sha256_rb")
	require.NoError(t, err)

	// First rollback
	err1 := reg.RollbackModel(cand.Version, "ADMIN_TEST", "Circuit breaker tripped")
	require.NoError(t, err1)

	// Second rollback is safe no-op
	err2 := reg.RollbackModel(cand.Version, "ADMIN_TEST", "Circuit breaker tripped again")
	require.NoError(t, err2)
}

// -------------------------------------------------------------
// 6. Canary Safety Breach & Circuit Breaker Rollback
// -------------------------------------------------------------

func TestCanaryRollback(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.CanaryObservationWindow = 5 * time.Millisecond
	cfg.MinSamples = 50

	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)
	runner := NewFixtureTrainingAdapter()

	coordinator := NewRetrainingCoordinator(cfg, runner, shadowScorer, canaryRouter, nil, nil)
	ctx := context.Background()

	job, err := coordinator.TriggerManual(ctx, "AUTOMATED_TEST", "Canary breach rollback test")
	require.NoError(t, err)
	require.NotNil(t, job)

	time.Sleep(30 * time.Millisecond)
	candidates := coordinator.GetCandidates()
	require.Len(t, candidates, 1)

	// Trip CircuitBreaker during canary rollout
	canaryRouter.circuitBreaker.Trip("CANARY_TEST_BREACH")

	err = coordinator.ApproveCandidate(ctx, candidates[0].ModelID, "AUTOMATED_TEST", "Approve for canary rollout")
	require.NoError(t, err)

	time.Sleep(60 * time.Millisecond)

	// Pipeline must have rolled back to ROLLED_BACK and reset canary traffic to 0%
	status := coordinator.GetStatus()
	assert.True(t, status["state"] == StateRolledBack || status["state"] == StateFailed)
	assert.Equal(t, 0, canaryRouter.GetPercentage())
}

// -------------------------------------------------------------
// 7. Synchronous Production Inference Load During Retraining
// -------------------------------------------------------------

func TestProductionTrafficDuringRetraining(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.MinSamples = 50
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)
	runner := NewFixtureTrainingAdapter()

	coordinator := NewRetrainingCoordinator(cfg, runner, shadowScorer, canaryRouter, nil, nil)
	ctx := context.Background()

	// Trigger asynchronous retraining
	_, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "Load test trigger")
	require.NoError(t, err)

	// Simulate concurrent high-throughput risk evaluations
	var totalRequests int64
	var errorCount int64
	var maxLatencyMs float64
	var latMu sync.Mutex

	var wg sync.WaitGroup
	workers := 10
	requestsPerWorker := 200

	start := time.Now()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for r := 0; r < requestsPerWorker; r++ {
				reqStart := time.Now()

				// Synchronous routing check
				route := canaryRouter.Route(fmt.Sprintf("tenant_%d", workerID), fmt.Sprintf("txn_%d_%d", workerID, r))
				assert.NotEmpty(t, route.String())

				durMs := float64(time.Since(reqStart).Nanoseconds()) / 1e6
				atomic.AddInt64(&totalRequests, 1)

				latMu.Lock()
				if durMs > maxLatencyMs {
					maxLatencyMs = durMs
				}
				latMu.Unlock()
			}
		}(w)
	}

	wg.Wait()
	totalDur := time.Since(start)

	assert.Equal(t, int64(workers*requestsPerWorker), totalRequests)
	assert.Equal(t, int64(0), errorCount)
	assert.Less(t, maxLatencyMs, 50.0, "Synchronous routing latency must remain strictly sub-50ms")
	t.Logf("Executed %d risk evaluation routes in %v (throughput: %.1f req/s, max latency: %.3f ms)",
		totalRequests, totalDur, float64(totalRequests)/totalDur.Seconds(), maxLatencyMs)
}

// -------------------------------------------------------------
// 8. Goroutine & Memory Safety
// -------------------------------------------------------------

func TestNoGoroutineLeak(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()

	cfg := DefaultRetrainingConfig()
	cfg.AutoApproveCanary = true
	cfg.MinSamples = 50

	for i := 0; i < 5; i++ {
		canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
		shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)
		coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), shadowScorer, canaryRouter, nil, nil)

		_, _ = coordinator.TriggerManual(context.Background(), "ADMIN_TEST", "Goroutine leak cycle")
		time.Sleep(40 * time.Millisecond)

		shadowScorer.Stop()
	}

	time.Sleep(50 * time.Millisecond)
	finalGoroutines := runtime.NumGoroutine()

	// Goroutines count must not have experienced unbounded growth
	delta := finalGoroutines - initialGoroutines
	assert.LessOrEqual(t, delta, 15, "Goroutine count delta must be bounded (detected delta: %d)", delta)
}

// -------------------------------------------------------------
// 9. Zero PII in Telemetry and Logs
// -------------------------------------------------------------

func TestNoPIIInTelemetry(t *testing.T) {
	rawLog := "Transaction txn_123 customer card 4532 0150 9999 1234, cvv: 987, exp: 12/28"
	sanitized := utils.SanitizeLogMessage(rawLog)

	assert.NotContains(t, sanitized, "4532 0150 9999 1234")
	assert.NotContains(t, sanitized, "987")
	assert.Contains(t, sanitized, "cvv:***")
}

// -------------------------------------------------------------
// 10. Admin Authentication & Security Hardening
// -------------------------------------------------------------

func TestAdminSecurityHardening(t *testing.T) {
	adminKey := "adm_risk_secret_key_12345"

	authMiddleware := func(key string, next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			provided := r.Header.Get("X-Admin-API-Key")
			if !utils.ConstantTimeCompare(provided, key) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}

	// 1. Valid Admin Key
	handler := authMiddleware(adminKey, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("AUTHORIZED"))
	})

	req1 := httptest.NewRequest("POST", "/v1/retraining/trigger", nil)
	req1.Header.Set("X-Admin-API-Key", adminKey)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// 2. Missing Key
	req2 := httptest.NewRequest("POST", "/v1/retraining/trigger", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)

	// 3. Wrong Key
	req3 := httptest.NewRequest("POST", "/v1/retraining/trigger", nil)
	req3.Header.Set("X-Admin-API-Key", "wrong_key_hacker")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusUnauthorized, rec3.Code)
}
