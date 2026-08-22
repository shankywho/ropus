package riskengine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateVersioning_EnvelopeAndChecksum(t *testing.T) {
	state := PersistedRetrainingState{
		ProductionModelVersion: "fraud-xgb-25f-v3.0",
		FallbackModelVersion:   "fraud-xgb-15f-v1.5",
		CurrentState:           StateIdle,
		MaintenanceMode:        false,
		ModelFrozen:            true,
		SavedAt:                time.Now().UTC(),
	}

	// 1. Wrap in envelope
	env, err := WrapState(state, 1)
	require.NoError(t, err)
	assert.Equal(t, CurrentSchemaVersion, env.SchemaVersion)
	assert.Equal(t, uint64(1), env.Generation)
	assert.NotEmpty(t, env.ChecksumSHA256)

	// 2. Validate envelope integrity
	err = env.ValidateEnvelope()
	assert.NoError(t, err)

	// 3. Corrupt checksum -> Validate fails
	env.ChecksumSHA256 = "corrupted_checksum_0000"
	err = env.ValidateEnvelope()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestFileStateStore_VersioningAndQuarantine(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "test_state.json")

	store, err := NewFileStateStore(statePath)
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Save state
	state1 := PersistedRetrainingState{
		ProductionModelVersion: "fraud-xgb-25f-v3.0",
		FallbackModelVersion:   "fraud-xgb-15f-v1.5",
		CurrentState:           StateIdle,
		ModelFrozen:            true,
	}
	err = store.Save(ctx, state1)
	require.NoError(t, err)

	// 2. Load state
	loaded, err := store.Load(ctx)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, "fraud-xgb-25f-v3.0", loaded.ProductionModelVersion)
	assert.True(t, loaded.ModelFrozen)

	// 3. Corrupt file content manually
	err = os.WriteFile(statePath, []byte("{ malformed corrupted json content "), 0644)
	require.NoError(t, err)

	// 4. Load should detect corruption, quarantine file, and return error
	corruptedLoad, err := store.Load(ctx)
	assert.Error(t, err)
	assert.Nil(t, corruptedLoad)
	assert.Contains(t, err.Error(), "corrupted state file quarantined")

	// Verify quarantine file was created
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	foundQuarantine := false
	for _, f := range files {
		if strings.Contains(f.Name(), ".corrupted.") {
			foundQuarantine = true
		}
	}
	assert.True(t, foundQuarantine, "Expected quarantined copy of corrupted state file")
}

func TestModelRegistry_ReconcileSelfHealing(t *testing.T) {
	reg := NewModelRegistry()
	verifier := NewArtifactVerifier()

	// 1. Clean registry reconciliation
	res, err := reg.Reconcile(verifier)
	require.NoError(t, err)
	assert.True(t, res.Valid)
	assert.Equal(t, "fraud-xgb-25f-v3.0", res.ProductionModelVersion)
	assert.Equal(t, "fraud-xgb-15f-v1.5", res.FallbackModelVersion)

	// 2. Inconsistency: Multiple active production models
	reg.mu.Lock()
	reg.models["fraud-xgb-15f-v1.5"].IsProductionActive = true
	reg.mu.Unlock()

	res2, err := reg.Reconcile(verifier)
	require.NoError(t, err)
	assert.Equal(t, 1, res2.RepairsMade)
	assert.Contains(t, res2.Repairs[0], "Deactivated extra production flag")

	// Verify only 1 active production remains
	models := reg.ListModels()
	prodCount := 0
	for _, m := range models {
		if m.IsProductionActive {
			prodCount++
		}
	}
	assert.Equal(t, 1, prodCount)

	// 3. Production Model Artifact Corrupted -> Failover to fallback
	reg.mu.Lock()
	reg.models["fraud-xgb-25f-v3.0"].ArtifactChecksum = "corrupted_checksum_invalid"
	reg.mu.Unlock()

	res3, err := reg.Reconcile(verifier)
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-15f-v1.5", res3.ProductionModelVersion)
	assert.Contains(t, res3.Repairs[0], "failed over to fallback")
}

func TestArtifactHealthScanner_QuarantineAndScan(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewLocalFilesystemArtifactStore(tempDir)
	require.NoError(t, err)

	verifier := NewArtifactVerifier()
	scanner := NewArtifactHealthScanner(store, verifier)

	reg := NewModelRegistry()

	// Create valid artifact on disk
	validModelID := "model_valid_test"
	uri, checksum, err := store.StoreArtifact(context.Background(), validModelID, "model.onnx", strings.NewReader("dummy model content 12345"))
	require.NoError(t, err)

	now := time.Now().UTC()
	reg.models["model-valid-v1"] = &RegisteredModel{
		ModelID:          validModelID,
		Version:          "model-valid-v1",
		ArtifactURI:      uri,
		ArtifactChecksum: checksum,
		CreatedAt:        now,
		LifecycleState:   LifecycleValidated,
	}

	// 1. Scan health -> HEALTHY
	report, err := scanner.ScanHealth(context.Background(), reg)
	require.NoError(t, err)
	assert.Equal(t, "HEALTHY", report.Status)
	assert.GreaterOrEqual(t, report.Verified, 1)

	// 2. Corrupt artifact content -> Scan should detect, quarantine, and report DEGRADED
	cleanPath := strings.TrimPrefix(uri, "file://")
	_ = os.WriteFile(cleanPath, []byte("tampered artifact bytes"), 0644)

	report2, err := scanner.ScanHealth(context.Background(), reg)
	require.NoError(t, err)
	assert.Equal(t, "DEGRADED", report2.Status)
	assert.Equal(t, 1, report2.Corrupted)
	assert.Equal(t, 1, report2.Quarantined)
}

func TestCircuitBreaker_SelfTest(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker(cfg)

	err := cb.RunSelfTest()
	assert.NoError(t, err)
}

func TestSafetyAuditor_All14Invariants(t *testing.T) {
	reg := NewModelRegistry()
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	sloEngine := NewSLOEngine(5 * time.Minute)
	verifier := NewArtifactVerifier()

	auditor := NewSafetyAuditor(reg, coordinator, canaryRouter, sloEngine, verifier)

	rep := auditor.Audit(context.Background())
	assert.Equal(t, "SAFE", rep.Status)
	assert.Equal(t, 0, rep.FailedCount)
	assert.GreaterOrEqual(t, rep.PassedCount, 12)
	assert.Equal(t, "PASS", rep.Checks["production_model"])
	assert.Equal(t, "PASS", rep.Checks["fallback_model"])
	assert.Equal(t, "PASS", rep.Checks["registry_consistency"])
	assert.Equal(t, "PASS", rep.Checks["canary_traffic"])
	assert.Equal(t, "PASS", rep.Checks["circuit_breaker"])
	assert.Equal(t, "PASS", rep.Checks["retraining_state"])
}

func TestErrorBudgetPolicy_ModelFreezeAutomation(t *testing.T) {
	sloEngine := NewSLOEngine(5 * time.Minute)
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)

	policyEngine := NewErrorBudgetPolicyEngine(sloEngine, coordinator)

	// 1. Initial healthy state
	eval1 := policyEngine.Evaluate(context.Background())
	assert.Equal(t, ActionBudgetHealthy, eval1.Action)
	assert.True(t, eval1.PromotionPermitted)
	assert.False(t, eval1.AutoFrozenEnacted)

	// 2. Simulate severe SLO violations to exhaust error budget
	for i := 0; i < 500; i++ {
		sloEngine.RecordEvaluation(500.0, false, false, true)
	}

	// 3. Re-evaluate policy -> Should trigger emergency model freeze
	eval2 := policyEngine.Evaluate(context.Background())
	assert.Equal(t, ActionEmergencyModelFreeze, eval2.Action)
	assert.False(t, eval2.PromotionPermitted)
	assert.True(t, eval2.AutoFrozenEnacted)

	controls := coordinator.GetOperationalControls()
	assert.True(t, controls["model_frozen"].(bool))
}

func TestDependencyMatrix_AssessPolicies(t *testing.T) {
	matrix := NewDependencyMatrix()

	// 1. ClickHouse down -> Fail open for traffic, pause retraining dataset collection
	chAssess := matrix.Assess(DepClickHouse, DepStateUnavailable)
	assert.True(t, chAssess.FailOpenForTraffic)
	assert.Contains(t, chAssess.InferenceImpact, "Fail-open audit")
	assert.Contains(t, chAssess.RetrainingImpact, "PAUSED")

	// 2. ML Runtime down -> Route to fallback model
	mlAssess := matrix.Assess(DepMLRuntime, DepStateUnavailable)
	assert.True(t, mlAssess.FailOpenForTraffic)
	assert.Contains(t, mlAssess.InferenceImpact, "ROUTED_TO_FALLBACK")
	assert.Contains(t, mlAssess.RetrainingImpact, "BLOCKED")
}
