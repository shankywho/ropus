package riskengine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PersistedRetrainingState encapsulates all persistent model registry and retraining state.
type PersistedRetrainingState struct {
	ProductionModelVersion string                      `json:"production_model_version"`
	FallbackModelVersion   string                      `json:"fallback_model_version"`
	Models                 map[string]*RegisteredModel `json:"models"`
	ActiveJob              *RetrainingJob              `json:"active_job,omitempty"`
	ActiveCandidate        *ModelCandidate             `json:"active_candidate,omitempty"`
	CurrentState           JobState                    `json:"current_state"`
	JobHistory             []RetrainingJob             `json:"job_history"`
	CanaryStage            int                         `json:"canary_stage"`
	MaintenanceMode        bool                        `json:"maintenance_mode,omitempty"`
	ModelFrozen            bool                        `json:"model_frozen,omitempty"`
	RetrainingPaused       bool                        `json:"retraining_paused,omitempty"`
	CanaryPaused           bool                        `json:"canary_paused,omitempty"`
	SavedAt                time.Time                   `json:"saved_at"`
	ChecksumSHA256         string                      `json:"checksum_sha256,omitempty"`
}

// StateStore defines the persistence contract for retraining and registry lifecycle state.
type StateStore interface {
	Save(ctx context.Context, state PersistedRetrainingState) error
	Load(ctx context.Context) (*PersistedRetrainingState, error)
	StateExists(ctx context.Context) bool
}

// FileStateStore persists state atomically to the local filesystem.
// FileStateStore persists state atomically to the local filesystem with versioning and forensic quarantine.
type FileStateStore struct {
	mu                sync.Mutex
	filePath          string
	currentGeneration uint64
}

// NewFileStateStore initializes a file-based state store at the specified path.
func NewFileStateStore(filePath string) (*FileStateStore, error) {
	if filePath == "" {
		filePath = "ml-service/model/registry_state.json"
	}
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for state store %s: %w", dir, err)
	}
	return &FileStateStore{filePath: filePath}, nil
}

// Save writes state atomically wrapped in a versioned envelope to a temporary file, syncs to disk, and renames.
func (s *FileStateStore) Save(ctx context.Context, state PersistedRetrainingState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state.SavedAt = time.Now().UTC()
	s.currentGeneration++

	envelope, err := WrapState(state, s.currentGeneration)
	if err != nil {
		return fmt.Errorf("failed to wrap state envelope: %w", err)
	}

	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state envelope: %w", err)
	}

	tmpPath := fmt.Sprintf("%s.tmp.%d", s.filePath, time.Now().UnixNano())
	tmpFile, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp state file: %w", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to write state data: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to sync state file to disk: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close state file: %w", err)
	}

	if err := os.Rename(tmpPath, s.filePath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to commit atomic state file: %w", err)
	}

	return nil
}

// Load reads, verifies, and migrates persisted state from disk.
// If corruption is detected, the corrupted file is automatically quarantined for forensic analysis.
func (s *FileStateStore) Load(ctx context.Context) (*PersistedRetrainingState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil, nil // No previous state exists
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	envelope, err := ParseAndMigrateState(data)
	if err != nil {
		// FORENSIC QUARANTINE: Preserve damaged state file before proceeding
		quarantinePath := fmt.Sprintf("%s.corrupted.%d", s.filePath, time.Now().UnixNano())
		_ = os.Rename(s.filePath, quarantinePath)
		log.Printf("[DISASTER_RECOVERY] Corrupted state file detected (%v). Quarantined to: %s", err, quarantinePath)
		return nil, fmt.Errorf("corrupted state file quarantined to %s: %w", quarantinePath, err)
	}

	if envelope != nil {
		s.currentGeneration = envelope.Generation
		return &envelope.Payload, nil
	}

	return nil, nil
}

// StateExists returns true if a persisted state file exists.
func (s *FileStateStore) StateExists(ctx context.Context) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := os.Stat(s.filePath)
	return err == nil
}

// ReconciliationAction describes the action taken during startup state recovery.
type ReconciliationAction string

const (
	ActionNone                 ReconciliationAction = "NONE"
	ActionRestoredClean        ReconciliationAction = "RESTORED_CLEAN"
	ActionInFlightJobFailed    ReconciliationAction = "IN_FLIGHT_JOB_FAILED"
	ActionCanaryResetToIdle    ReconciliationAction = "CANARY_RESET_TO_IDLE"
	ActionCandidatePreserved   ReconciliationAction = "CANDIDATE_PRESERVED"
	ActionFallbackModelRestored ReconciliationAction = "FALLBACK_MODEL_RESTORED"
)

// ReconciliationResult contains details of the startup state reconciliation.
type ReconciliationResult struct {
	PreviousState          JobState             `json:"previous_state"`
	ReconciledState        JobState             `json:"reconciled_state"`
	ProductionModelVersion string               `json:"production_model_version"`
	FallbackModelVersion   string               `json:"fallback_model_version"`
	ActionTaken            ReconciliationAction `json:"action_taken"`
	Message                string               `json:"message"`
	ActiveCandidateID      string               `json:"active_candidate_id,omitempty"`
}

// RecoveryManager reconciles persisted state on startup and ensures crash recovery safety.
type RecoveryManager struct {
	store    StateStore
	verifier *ArtifactVerifier
}

// NewRecoveryManager initializes the recovery manager with the given state store and verifier.
func NewRecoveryManager(store StateStore, verifier *ArtifactVerifier) *RecoveryManager {
	if verifier == nil {
		verifier = NewArtifactVerifier()
	}
	return &RecoveryManager{
		store:    store,
		verifier: verifier,
	}
}

// ReconcileOnStartup inspects persisted state, verifies model registry consistency,
// reconciles any in-flight interrupted state machine transitions, and returns a safe state.
func (rm *RecoveryManager) ReconcileOnStartup(
	ctx context.Context,
	registry *ModelRegistry,
	coordinator *RetrainingCoordinator,
) (*ReconciliationResult, error) {
	if rm.store == nil {
		return &ReconciliationResult{
			PreviousState:   StateIdle,
			ReconciledState: StateIdle,
			ActionTaken:     ActionNone,
			Message:         "No state store configured; using default baseline models",
		}, nil
	}

	state, err := rm.store.Load(ctx)
	if err != nil {
		log.Printf("[RECOVERY_WARN] Corrupted state file encountered (%v). Falling back to verified baseline.", err)
		return &ReconciliationResult{
			PreviousState:   StateFailed,
			ReconciledState: StateIdle,
			ActionTaken:     ActionFallbackModelRestored,
			Message:         fmt.Sprintf("Corrupted state file ignored: %v", err),
		}, nil
	}

	if state == nil {
		return &ReconciliationResult{
			PreviousState:   StateIdle,
			ReconciledState: StateIdle,
			ActionTaken:     ActionNone,
			Message:         "Clean first startup; initial baselines active",
		}, nil
	}

	res := &ReconciliationResult{
		PreviousState:          state.CurrentState,
		ReconciledState:        state.CurrentState,
		ProductionModelVersion: state.ProductionModelVersion,
		FallbackModelVersion:   state.FallbackModelVersion,
		ActionTaken:            ActionRestoredClean,
	}

	// 1. Reconcile Model Registry models
	if registry != nil {
		registry.mu.Lock()
		if len(state.Models) > 0 {
			for k, v := range state.Models {
				registry.models[k] = v
			}
		}
		if state.ProductionModelVersion != "" && registry.models[state.ProductionModelVersion] != nil {
			registry.productionModel = state.ProductionModelVersion
			registry.models[state.ProductionModelVersion].IsProductionActive = true
		}
		if state.FallbackModelVersion != "" && registry.models[state.FallbackModelVersion] != nil {
			registry.fallbackModel = state.FallbackModelVersion
			registry.models[state.FallbackModelVersion].IsFallbackActive = true
		}
		registry.mu.Unlock()
	}

	// 2. Reconcile Retraining Coordinator state machine
	if coordinator != nil {
		coordinator.mu.Lock()
		defer coordinator.mu.Unlock()

		// Restore candidates map
		if state.ActiveCandidate != nil {
			coordinator.candidates[state.ActiveCandidate.ModelID] = state.ActiveCandidate
			coordinator.activeCandidate = state.ActiveCandidate
			res.ActiveCandidateID = state.ActiveCandidate.ModelID
		}

		// Restore operational controls
		coordinator.maintenanceMode = state.MaintenanceMode
		coordinator.modelFrozen = state.ModelFrozen
		coordinator.retrainingPaused = state.RetrainingPaused
		coordinator.canaryPaused = state.CanaryPaused

		// State-specific reconciliation rules
		switch state.CurrentState {
		case StateTriggered, StateQueued, StateTraining, StateValidating, StateShadowEvaluation:
			// Process died while asynchronously training or validating.
			// Mark active job as failed to prevent orphaned/zombie states and allow future triggers.
			coordinator.currentState = StateFailed
			res.ReconciledState = StateFailed
			res.ActionTaken = ActionInFlightJobFailed
			res.Message = fmt.Sprintf("Interrupted during %s when process restarted; safely transitioned to FAILED", state.CurrentState)

			if state.ActiveJob != nil {
				job := *state.ActiveJob
				job.State = StateFailed
				job.Error = "Retraining process interrupted by server restart"
				job.CompletedAt = time.Now().UTC()
				coordinator.activeJob = nil
				coordinator.jobHistory = append(coordinator.jobHistory, job)
				coordinator.triggerEngine.ClearActiveJob(false)
			}

		case StateAwaitingApproval:
			// Candidate is awaiting operator approval.
			// If candidate metadata and scorecards are intact, keep AWAITING_APPROVAL.
			if state.ActiveCandidate != nil {
				coordinator.currentState = StateAwaitingApproval
				coordinator.activeCandidate = state.ActiveCandidate
				coordinator.activeJob = state.ActiveJob
				res.ReconciledState = StateAwaitingApproval
				res.ActionTaken = ActionCandidatePreserved
				res.Message = "Candidate model in AWAITING_APPROVAL successfully recovered and awaiting operator review"
			} else {
				coordinator.currentState = StateIdle
				res.ReconciledState = StateIdle
				res.ActionTaken = ActionNone
			}

		case StateCanary:
			// Server restarted mid-canary rollout.
			// For absolute safety, reset canary traffic to 0% and keep candidate AWAITING_APPROVAL for operator inspection.
			coordinator.currentState = StateAwaitingApproval
			res.ReconciledState = StateAwaitingApproval
			res.ActionTaken = ActionCanaryResetToIdle
			res.Message = "Server restarted during Canary rollout; canary traffic safely reset to 0%, candidate returned to AWAITING_APPROVAL"
			if coordinator.canaryRouter != nil {
				_ = coordinator.canaryRouter.UpdateConfig(false, 0, "SYSTEM_RECOVERY", "Canary traffic reset to 0% following server restart")
			}

		case StatePromoted, StateIdle, StateRejected, StateFailed, StateRolledBack:
			// Terminal or idle states require no state mutation
			coordinator.currentState = state.CurrentState
			coordinator.activeJob = nil
			coordinator.triggerEngine.ClearActiveJob(false)
			res.ReconciledState = state.CurrentState
			res.ActionTaken = ActionRestoredClean
			res.Message = fmt.Sprintf("Clean state %s restored successfully", state.CurrentState)
		}
	}

	// Persist reconciled clean state
	if coordinator != nil {
		persisted := coordinator.buildPersistedStateLocked()
		_ = rm.store.Save(ctx, persisted)
	}

	return res, nil
}
