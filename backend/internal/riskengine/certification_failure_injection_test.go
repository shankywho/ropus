package riskengine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 1. Crash During Training
// ---------------------------------------------------------------------------
func TestCert_FailureInjection_CrashDuringTraining(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "registry_state.json")
	store, err := NewFileStateStore(statePath)
	require.NoError(t, err)

	ctx := context.Background()
	cfg := DefaultRetrainingConfig()

	// Initial node crashes while running a training job
	c1 := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	c1.SetStateStore(store)

	c1.mu.Lock()
	c1.currentState = StateTraining
	c1.activeJob = &RetrainingJob{
		JobID:                 "job_crash_test_01",
		TriggerType:           "MANUAL_OPERATOR",
		State:                 StateTraining,
		TriggeredAt:           time.Now().UTC().Add(-2 * time.Minute),
		CandidateModelVersion: "fraud-xgb-25f-candidate-v1",
	}
	c1.persistStateLocked(ctx)
	c1.mu.Unlock()

	// Simulated reboot: DisasterRecoveryManager executes
	c2 := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	reg := NewModelRegistry()
	canary := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	dr := NewDisasterRecoveryManager(store, NewArtifactVerifier(), nil)

	report, err := dr.ExecuteRecovery(ctx, reg, c2, canary)
	require.NoError(t, err)
	require.NotNil(t, report)

	assert.Equal(t, StateTraining, report.PreviousState)
	assert.Equal(t, StateFailed, report.ReconciledState)
	assert.Equal(t, StateFailed, c2.currentState)
	assert.Nil(t, c2.activeJob)

	// Ensure inference model is valid and active
	pm, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.True(t, pm.IsProductionActive)
}

// ---------------------------------------------------------------------------
// 2. Crash During Canary Rollout
// ---------------------------------------------------------------------------
func TestCert_FailureInjection_CrashDuringCanary(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "registry_state.json")
	store, err := NewFileStateStore(statePath)
	require.NoError(t, err)

	ctx := context.Background()

	// Persist active canary state at 75%
	stateCanary := PersistedRetrainingState{
		ProductionModelVersion: "fraud-xgb-25f-v3.0",
		FallbackModelVersion:   "fraud-xgb-15f-v1.5",
		CurrentState:           StateCanary,
		CanaryStage:            75,
		SavedAt:                time.Now().UTC(),
	}
	require.NoError(t, store.Save(ctx, stateCanary))

	// Reboot and recover
	cfg := DefaultRetrainingConfig()
	c2 := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	reg := NewModelRegistry()
	canaryCfg := DefaultCanaryRouterConfig()
	canaryCfg.Enabled = true
	canaryCfg.Percentage = 75
	canaryCfg.CandidateModelVersion = "fraud-xgb-25f-candidate-v1"
	canary := NewCanaryRouter(canaryCfg, nil)
	dr := NewDisasterRecoveryManager(store, NewArtifactVerifier(), nil)

	report, err := dr.ExecuteRecovery(ctx, reg, c2, canary)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Invariant: Canary traffic must immediately reset to 0%
	assert.Equal(t, StateCanary, report.PreviousState)
	assert.Equal(t, StateIdle, report.ReconciledState)
	assert.Equal(t, 0, canary.GetPercentage())
	assert.Equal(t, StateIdle, c2.currentState)
}

// ---------------------------------------------------------------------------
// 3. Active Model Artifact Corruption & Automatic Failover
// ---------------------------------------------------------------------------
func TestCert_FailureInjection_ActiveModelArtifactCorruption(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewLocalFilesystemArtifactStore(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	verifier := NewArtifactVerifier()

	// Store and register a real production model file
	prodModelID := "model_prod_real"
	prodURI, prodChecksum, err := store.StoreArtifact(ctx, prodModelID, "model.onnx", strings.NewReader("real production model content v3.0"))
	require.NoError(t, err)

	fallbackModelID := "model_fallback_real"
	fbURI, fbChecksum, err := store.StoreArtifact(ctx, fallbackModelID, "model.onnx", strings.NewReader("real fallback model content v1.5"))
	require.NoError(t, err)

	reg := &ModelRegistry{
		models:          make(map[string]*RegisteredModel),
		productionModel: "prod-v3",
		fallbackModel:   "fb-v1.5",
	}
	reg.models["prod-v3"] = &RegisteredModel{
		ModelID:            prodModelID,
		Version:            "prod-v3",
		ArtifactURI:        prodURI,
		ArtifactChecksum:   prodChecksum,
		IsProductionActive: true,
		LifecycleState:     LifecyclePromoted,
	}
	reg.models["fb-v1.5"] = &RegisteredModel{
		ModelID:            fallbackModelID,
		Version:            "fb-v1.5",
		ArtifactURI:        fbURI,
		ArtifactChecksum:   fbChecksum,
		IsProductionActive: false,
		LifecycleState:     LifecycleValidated,
	}

	// 1. Initially reconcile -> Valid
	r1, err := reg.Reconcile(verifier)
	require.NoError(t, err)
	assert.Equal(t, "prod-v3", r1.ProductionModelVersion)

	// 2. CORRUPT the production artifact on disk
	prodPath := strings.TrimPrefix(prodURI, "file://")
	err = os.WriteFile(prodPath, []byte("MALICIOUSLY TAMPERED BYTES"), 0644)
	require.NoError(t, err)

	// 3. Reconcile -> Must automatically detect corruption and failover to fallback
	r2, err := reg.Reconcile(verifier)
	require.NoError(t, err)
	assert.Equal(t, "fb-v1.5", r2.ProductionModelVersion)
	assert.Equal(t, 1, r2.RepairsMade)
	assert.Contains(t, r2.Repairs[0], "Production artifact corrupted; safely failed over to fallback")

	pm, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fb-v1.5", pm.Version)
	assert.True(t, pm.IsProductionActive)
}

// ---------------------------------------------------------------------------
// 4. State File Corruption & Forensic Quarantine
// ---------------------------------------------------------------------------
func TestCert_FailureInjection_StateFileCorruptionAndQuarantine(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "registry_state.json")

	// Write completely damaged non-JSON bytes
	err := os.WriteFile(statePath, []byte("%%% CORRUPTED RAW BYTES NOT JSON %%%"), 0644)
	require.NoError(t, err)

	store, err := NewFileStateStore(statePath)
	require.NoError(t, err)

	ctx := context.Background()
	cfg := DefaultRetrainingConfig()
	c := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	reg := NewModelRegistry()
	canary := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	dr := NewDisasterRecoveryManager(store, NewArtifactVerifier(), nil)

	// Execute disaster recovery -> Must quarantine corrupted file and restore clean state
	report, err := dr.ExecuteRecovery(ctx, reg, c, canary)
	require.NoError(t, err)
	assert.Equal(t, "READY", report.Status)
	assert.Equal(t, StateIdle, report.ReconciledState)

	// Verify quarantine artifact exists on filesystem
	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	hasQuarantine := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".corrupted.") {
			hasQuarantine = true
		}
	}
	assert.True(t, hasQuarantine, "Expected forensic quarantine copy of corrupted state")
}

// ---------------------------------------------------------------------------
// 5. Infrastructure Dependency Outages (ClickHouse, Redis, Postgres)
// ---------------------------------------------------------------------------
func TestCert_FailureInjection_InfrastructureDependencyOutage(t *testing.T) {
	matrix := NewDependencyMatrix()

	// ClickHouse Down: Synchronous traffic fails open; retraining dataset ingestion pauses
	chEval := matrix.Assess(DepClickHouse, DepStateUnavailable)
	assert.True(t, chEval.FailOpenForTraffic)
	assert.Contains(t, chEval.InferenceImpact, "Fail-open audit")
	assert.Contains(t, chEval.RetrainingImpact, "PAUSED")

	// Redis Down: Synchronous traffic continues using in-memory state; feature enrichment falls back
	redisEval := matrix.Assess(DepRedis, DepStateUnavailable)
	assert.True(t, redisEval.FailOpenForTraffic)
	assert.Contains(t, redisEval.InferenceImpact, "Device velocity features degrade")

	// Postgres Down: Synchronous traffic continues with cached tenant policies
	pgEval := matrix.Assess(DepPostgres, DepStateUnavailable)
	assert.True(t, pgEval.FailOpenForTraffic)
	assert.Contains(t, pgEval.InferenceImpact, "In-memory rule heuristics")
}

// ---------------------------------------------------------------------------
// 6. ML Runtime Timeouts and Fallback
// ---------------------------------------------------------------------------
func TestCert_FailureInjection_MLRuntimeTimeoutAndFallback(t *testing.T) {
	// Setup a mock slow/unresponsive ML service
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Exceeds 20ms timeout
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"risk_score": 99, "probability": 0.99}`))
	}))
	defer slowServer.Close()

	mlClient := NewMLClient(slowServer.URL)

	// Call ML service with tight deadline
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	resp, err := mlClient.Predict(ctx, MLPredictRequest{
		Amount:           100.0,
		TokenVelocity24h: 1.0,
		IsNewDevice:      0,
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// ---------------------------------------------------------------------------
// 7. Circuit Breaker Sustained Failure Trip
// ---------------------------------------------------------------------------
func TestCert_FailureInjection_CircuitBreakerSustainedFailure(t *testing.T) {
	cbCfg := CircuitBreakerConfig{
		MaxErrorRate:    0.02,
		MaxP95LatencyMs: 20.0,
		MinSampleCount:  50,
		FailureWindow:   3,
		CooldownSeconds: 300,
	}
	cb := NewCircuitBreaker(cbCfg)

	failingMetrics := map[string]interface{}{
		"candidate_requests_total": int64(100),
		"candidate_error_rate":     0.10, // 10% > 2% max
		"candidate_p95_latency_ms": 15.0,
	}

	// 1. First failure -> FAILED state, no trip
	tripped1, st1, _, _ := cb.EvaluateAndCheckTrip(failingMetrics)
	assert.False(t, tripped1)
	assert.Equal(t, CircuitStateFailed, st1)

	// 2. Second failure -> FAILED state, no trip
	tripped2, st2, _, _ := cb.EvaluateAndCheckTrip(failingMetrics)
	assert.False(t, tripped2)
	assert.Equal(t, CircuitStateFailed, st2)

	// 3. Third failure (Window = 3) -> ROLLED_BACK state, Tripped = true
	tripped3, st3, reason, _ := cb.EvaluateAndCheckTrip(failingMetrics)
	assert.True(t, tripped3)
	assert.Equal(t, CircuitStateRolledBack, st3)
	assert.Contains(t, reason, "Sustained safety breach")
}

// ---------------------------------------------------------------------------
// 8. Error Budget Depletion Lock
// ---------------------------------------------------------------------------
func TestCert_FailureInjection_ErrorBudgetDepletionLock(t *testing.T) {
	slo := NewSLOEngine(5 * time.Minute)
	cfg := DefaultRetrainingConfig()
	c := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	policy := NewErrorBudgetPolicyEngine(slo, c)

	// Initially healthy
	e1 := policy.Evaluate(context.Background())
	assert.Equal(t, ActionBudgetHealthy, e1.Action)
	assert.True(t, e1.PromotionPermitted)

	// Inject 1000 failing high-latency measurements to completely exhaust error budgets
	for i := 0; i < 1000; i++ {
		slo.RecordEvaluation(800.0, false, false, true)
	}

	// Policy evaluation must enforce EMERGENCY_MODEL_FREEZE
	e2 := policy.Evaluate(context.Background())
	assert.Equal(t, ActionEmergencyModelFreeze, e2.Action)
	assert.False(t, e2.PromotionPermitted)
	assert.True(t, e2.AutoFrozenEnacted)

	// Operational controls verify model is frozen
	controls := c.GetOperationalControls()
	assert.True(t, controls["model_frozen"].(bool))
}

// ---------------------------------------------------------------------------
// 9. Compound Multi-Failure Chaos (ClickHouse + Redis Down + Tampered Artifact + Crash)
// ---------------------------------------------------------------------------
func TestCert_FailureInjection_SimultaneousCompoundChaos(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "registry_state.json")
	artifactStore, err := NewLocalFilesystemArtifactStore(filepath.Join(tempDir, "artifacts"))
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Seed valid artifacts
	prodURI, prodChecksum, err := artifactStore.StoreArtifact(ctx, "prod_id", "model.onnx", strings.NewReader("clean production model"))
	require.NoError(t, err)
	fbURI, fbChecksum, err := artifactStore.StoreArtifact(ctx, "fb_id", "model.onnx", strings.NewReader("clean fallback model"))
	require.NoError(t, err)

	// 2. Setup state with 50% canary
	state := PersistedRetrainingState{
		ProductionModelVersion: "prod_v1",
		FallbackModelVersion:   "fb_v1",
		CurrentState:           StateCanary,
		CanaryStage:            50,
	}
	stateStore, err := NewFileStateStore(statePath)
	require.NoError(t, err)
	require.NoError(t, stateStore.Save(ctx, state))

	// 3. INJECT COMPOUND FAILURES:
	// Failure A: Corrupt the primary production artifact
	_ = os.WriteFile(strings.TrimPrefix(prodURI, "file://"), []byte("CORRUPTED_BYTES"), 0644)
	// Failure B: Corrupt the state file
	_ = os.WriteFile(statePath, []byte("NOT_VALID_JSON"), 0644)

	// 4. Initialize recovery under compound failure
	reg := &ModelRegistry{
		models:          make(map[string]*RegisteredModel),
		productionModel: "prod_v1",
		fallbackModel:   "fb_v1",
	}
	reg.models["prod_v1"] = &RegisteredModel{
		ModelID:            "prod_id",
		Version:            "prod_v1",
		ArtifactURI:        prodURI,
		ArtifactChecksum:   prodChecksum,
		IsProductionActive: true,
	}
	reg.models["fb_v1"] = &RegisteredModel{
		ModelID:            "fb_id",
		Version:            "fb_v1",
		ArtifactURI:        fbURI,
		ArtifactChecksum:   fbChecksum,
		IsProductionActive: false,
	}

	canaryCfg := DefaultCanaryRouterConfig()
	canaryCfg.Enabled = true
	canaryCfg.Percentage = 50
	canary := NewCanaryRouter(canaryCfg, nil)

	cfg := DefaultRetrainingConfig()
	c := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	verifier := NewArtifactVerifier()
	dr := NewDisasterRecoveryManager(stateStore, verifier, nil)

	// 5. Execute Recovery
	rep, err := dr.ExecuteRecovery(ctx, reg, c, canary)
	require.NoError(t, err)
	require.NotNil(t, rep)

	// Assertions:
	// A: Canary reset to 0%
	assert.Equal(t, 0, canary.GetPercentage())
	// B: State transitioned to IDLE
	assert.Equal(t, StateIdle, rep.ReconciledState)
	// C: Failed over to healthy fallback model
	pm, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fb_v1", pm.Version)
	assert.True(t, pm.IsProductionActive)
}
