package riskengine

import (
	"fmt"
	"sync"
	"time"
)

// ExperimentRun records hyperparameters, feature versions, and benchmark metrics for an ML training experiment.
type ExperimentRun struct {
	ExperimentID      string            `json:"experiment_id"`
	RunID             string            `json:"run_id"`
	ModelVersion      string            `json:"model_version"`
	DatasetChecksum   string            `json:"dataset_checksum"`
	FeatureSetVersion int               `json:"feature_set_version"`
	Parameters        map[string]string `json:"parameters"`
	Metrics           EvaluationMetrics `json:"metrics"`
	StartedAt         time.Time         `json:"started_at"`
	CompletedAt       time.Time         `json:"completed_at"`
	DurationSeconds   float64           `json:"duration_seconds"`
	Status            string            `json:"status"` // "SUCCESS", "FAILED"
	ArtifactURI       string            `json:"artifact_uri"`
	ArtifactHash      string            `json:"artifact_hash"`
}

// ExperimentTracker manages experiment lineage and metric comparisons.
type ExperimentTracker struct {
	mu   sync.RWMutex
	runs map[string]*ExperimentRun
}

// NewExperimentTracker initializes the experiment tracking catalog.
func NewExperimentTracker() *ExperimentTracker {
	return &ExperimentTracker{
		runs: make(map[string]*ExperimentRun),
	}
}

// LogRun records a completed experiment run.
func (t *ExperimentTracker) LogRun(run *ExperimentRun) error {
	if run == nil {
		return fmt.Errorf("experiment run is nil")
	}
	if run.RunID == "" {
		run.RunID = fmt.Sprintf("exp_run_%d", time.Now().UnixNano())
	}
	if run.CompletedAt.IsZero() {
		run.CompletedAt = time.Now().UTC()
	}
	if !run.StartedAt.IsZero() {
		run.DurationSeconds = run.CompletedAt.Sub(run.StartedAt).Seconds()
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.runs[run.RunID] = run
	return nil
}

// GetRun fetches an experiment by run ID.
func (t *ExperimentTracker) GetRun(runID string) (*ExperimentRun, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	run, exists := t.runs[runID]
	if !exists {
		return nil, fmt.Errorf("experiment run '%s' not found", runID)
	}
	return run, nil
}

// ListRuns returns all runs for a specific experiment ID or all runs if experimentID is empty.
func (t *ExperimentTracker) ListRuns(experimentID string) []*ExperimentRun {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []*ExperimentRun
	for _, run := range t.runs {
		if experimentID == "" || run.ExperimentID == experimentID {
			result = append(result, run)
		}
	}
	return result
}

// GetBestRun returns the highest ROC-AUC run for an experiment.
func (t *ExperimentTracker) GetBestRun(experimentID string) (*ExperimentRun, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var best *ExperimentRun
	for _, run := range t.runs {
		if experimentID == "" || run.ExperimentID == experimentID {
			if run.Status == "SUCCESS" {
				if best == nil || run.Metrics.ROCAUC > best.Metrics.ROCAUC {
					best = run
				}
			}
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no successful runs found for experiment '%s'", experimentID)
	}
	return best, nil
}
