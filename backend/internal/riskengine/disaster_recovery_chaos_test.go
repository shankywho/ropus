package riskengine

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisasterRecoveryChaos_CrashDuringTraining(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "registry_state.json")
	store, err := NewFileStateStore(statePath)
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Setup initial coordinator in TRAINING state
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	coordinator.SetStateStore(store)

	coordinator.mu.Lock()
	coordinator.currentState = StateTraining
	coordinator.activeJob = &RetrainingJob{
		JobID:                 "job_chaos_training_01",
		TriggerType:           "MANUAL_OPERATOR",
		State:                 StateTraining,
		TriggeredAt:           time.Now().UTC().Add(-5 * time.Minute),
		CandidateModelVersion: "fraud-xgb-25f-candidate-v1",
	}
	coordinator.persistStateLocked(ctx)
	coordinator.mu.Unlock()

	// 2. Simulate Backend Crash & Restart: Initialize new DisasterRecoveryManager
	newCoordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	newRegistry := NewModelRegistry()
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	drManager := NewDisasterRecoveryManager(store, NewArtifactVerifier(), nil)

	report, err := drManager.ExecuteRecovery(ctx, newRegistry, newCoordinator, canaryRouter)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Assert: Interrupted training job was reconciled to FAILED
	assert.Equal(t, StateTraining, report.PreviousState)
	assert.Equal(t, StateFailed, report.ReconciledState)
	assert.Equal(t, 1, report.RepairsMade)
	assert.Equal(t, StateFailed, newCoordinator.currentState)
	assert.Nil(t, newCoordinator.activeJob)

	// Primary production model remains authoritative
	pm, err := newRegistry.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", pm.Version)
	assert.True(t, pm.IsProductionActive)
}

func TestDisasterRecoveryChaos_CrashDuringCanary(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "registry_state.json")
	store, err := NewFileStateStore(statePath)
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Simulate active Canary rollout at 50%
	stateCanary := PersistedRetrainingState{
		ProductionModelVersion: "fraud-xgb-25f-v3.0",
		FallbackModelVersion:   "fraud-xgb-15f-v1.5",
		CurrentState:           StateCanary,
		CanaryStage:            50,
		SavedAt:                time.Now().UTC(),
	}
	err = store.Save(ctx, stateCanary)
	require.NoError(t, err)

	// 2. Backend Crash & Restart: Run Disaster Recovery
	cfg := DefaultRetrainingConfig()
	newCoordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	newRegistry := NewModelRegistry()
	canaryCfg := DefaultCanaryRouterConfig()
	canaryCfg.Enabled = true
	canaryCfg.Percentage = 50
	canaryCfg.CandidateModelVersion = "fraud-xgb-25f-candidate-v1"
	canaryRouter := NewCanaryRouter(canaryCfg, nil)
	drManager := NewDisasterRecoveryManager(store, NewArtifactVerifier(), nil)

	report, err := drManager.ExecuteRecovery(ctx, newRegistry, newCoordinator, canaryRouter)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Assert: Canary safely reset to 0% traffic and IDLE state
	assert.Equal(t, StateCanary, report.PreviousState)
	assert.Equal(t, StateIdle, report.ReconciledState)
	assert.Equal(t, 0, canaryRouter.GetPercentage())
	assert.Equal(t, StateIdle, newCoordinator.currentState)
}

func TestDisasterRecoveryChaos_CorruptedStateFileRecovery(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "registry_state.json")

	// Write completely broken byte stream
	err := os.WriteFile(statePath, []byte("INVALID_BYTE_STREAM_CORRUPTION_DATA"), 0644)
	require.NoError(t, err)

	store, err := NewFileStateStore(statePath)
	require.NoError(t, err)

	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	registry := NewModelRegistry()
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	drManager := NewDisasterRecoveryManager(store, NewArtifactVerifier(), nil)

	report, err := drManager.ExecuteRecovery(context.Background(), registry, coordinator, canaryRouter)
	require.NoError(t, err)
	assert.Equal(t, "READY", report.Status)
	assert.Equal(t, StateIdle, report.ReconciledState)
	assert.Equal(t, "fraud-xgb-25f-v3.0", report.ProductionModelVersion)
}

func TestDisasterRecoveryChaos_RapidRestartLoop(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "registry_state.json")
	store, err := NewFileStateStore(statePath)
	require.NoError(t, err)

	ctx := context.Background()

	// Initial clean save
	err = store.Save(ctx, PersistedRetrainingState{
		ProductionModelVersion: "fraud-xgb-25f-v3.0",
		FallbackModelVersion:   "fraud-xgb-15f-v1.5",
		CurrentState:           StateIdle,
	})
	require.NoError(t, err)

	// Run 20 rapid simulated restarts
	for i := 0; i < 20; i++ {
		cfg := DefaultRetrainingConfig()
		coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
		coordinator.SetStateStore(store)
		registry := NewModelRegistry()
		canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
		drManager := NewDisasterRecoveryManager(store, NewArtifactVerifier(), nil)

		report, err := drManager.ExecuteRecovery(ctx, registry, coordinator, canaryRouter)
		require.NoError(t, err)
		assert.Equal(t, "READY", report.Status)

		// Mutate operational controls and save again
		_ = coordinator.SetMaintenanceMode(ctx, i%2 == 0, "chaos_tester", "toggling maintenance")
	}

	// Verify final state
	loaded, err := store.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", loaded.ProductionModelVersion)
}

func TestDisasterRecoveryChaos_ConcurrentRecoveryExecution(t *testing.T) {
	tempDir := t.TempDir()
	statePath := filepath.Join(tempDir, "registry_state.json")
	store, err := NewFileStateStore(statePath)
	require.NoError(t, err)

	ctx := context.Background()
	_ = store.Save(ctx, PersistedRetrainingState{
		ProductionModelVersion: "fraud-xgb-25f-v3.0",
		FallbackModelVersion:   "fraud-xgb-15f-v1.5",
		CurrentState:           StateIdle,
	})

	drManager := NewDisasterRecoveryManager(store, NewArtifactVerifier(), nil)
	cfg := DefaultRetrainingConfig()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
			registry := NewModelRegistry()
			canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
			rep, err := drManager.ExecuteRecovery(ctx, registry, coordinator, canaryRouter)
			assert.NoError(t, err)
			assert.NotNil(t, rep)
		}()
	}
	wg.Wait()
}
