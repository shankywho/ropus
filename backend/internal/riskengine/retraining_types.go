package riskengine

import (
	"fmt"
	"time"
)

// JobState represents the explicit state machine of a model retraining and deployment lifecycle.
type JobState string

const (
	StateIdle             JobState = "IDLE"
	StateTriggered        JobState = "TRIGGERED"
	StateQueued           JobState = "QUEUED"
	StateTraining         JobState = "TRAINING"
	StateValidating       JobState = "VALIDATING"
	StateShadowEvaluation JobState = "SHADOW_EVALUATION"
	StateAwaitingApproval JobState = "AWAITING_APPROVAL"
	StateCanary           JobState = "CANARY"
	StatePromoted         JobState = "PROMOTED"
	StateRejected         JobState = "REJECTED"
	StateFailed           JobState = "FAILED"
	StateRolledBack       JobState = "ROLLED_BACK"
)

// CanTransitionTo enforces strict, valid state machine transitions.
func (s JobState) CanTransitionTo(target JobState) bool {
	switch s {
	case StateIdle:
		return target == StateTriggered
	case StateTriggered:
		return target == StateQueued || target == StateFailed || target == StateIdle
	case StateQueued:
		return target == StateTraining || target == StateFailed || target == StateIdle
	case StateTraining:
		return target == StateValidating || target == StateFailed
	case StateValidating:
		return target == StateShadowEvaluation || target == StateRejected || target == StateFailed
	case StateShadowEvaluation:
		return target == StateAwaitingApproval || target == StateCanary || target == StateRejected || target == StateFailed
	case StateAwaitingApproval:
		return target == StateCanary || target == StateRejected || target == StateFailed
	case StateCanary:
		return target == StatePromoted || target == StateRolledBack || target == StateFailed
	case StatePromoted, StateRejected, StateFailed, StateRolledBack:
		// Terminal states can transition back to IDLE or TRIGGERED for new cycles
		return target == StateIdle || target == StateTriggered
	default:
		return false
	}
}

// RetrainingConfig defines configuration parameters for automated retraining and validation gates.
type RetrainingConfig struct {
	Enabled                    bool          `json:"enabled"`
	MinSamples                 uint32        `json:"min_samples"`
	DriftThreshold             float64       `json:"drift_threshold"`
	RequiredConsecutiveWindows int           `json:"required_consecutive_windows"`
	CooldownDuration           time.Duration `json:"cooldown_duration"`
	MaxConcurrentJobs          int           `json:"max_concurrent_jobs"`
	MinModelImprovement        float64       `json:"min_model_improvement"`
	MaxErrorRate               float64       `json:"max_error_rate"`
	MaxFallbackRate            float64       `json:"max_fallback_rate"`
	MaxLatencyRegressionMs     float64       `json:"max_latency_regression_ms"`
	MaxDecisionChangeRate      float64       `json:"max_decision_change_rate"`
	MaxAllowedFPR              float64       `json:"max_allowed_fpr"`
	MaxAllowedBrierScore       float64       `json:"max_allowed_brier_score"`
	MaxP95LatencyMs            float64       `json:"max_p95_latency_ms"`
	AutoApproveCanary          bool          `json:"auto_approve_canary"`
	ShadowMinEvaluations       uint32        `json:"shadow_min_evaluations"`
	CanaryObservationWindow    time.Duration `json:"canary_observation_window"`
}

// DefaultRetrainingConfig returns production defaults.
func DefaultRetrainingConfig() RetrainingConfig {
	return RetrainingConfig{
		Enabled:                    true,
		MinSamples:                 200,
		DriftThreshold:             0.20,
		RequiredConsecutiveWindows: 2,
		CooldownDuration:           30 * time.Minute,
		MaxConcurrentJobs:          1,
		MinModelImprovement:        -0.02, // Max allowed AUC regression: 0.02 (2%)
		MaxErrorRate:               0.01,  // 1% max error rate
		MaxFallbackRate:            0.01,  // 1% max fallback rate
		MaxLatencyRegressionMs:     5.0,   // 5ms max latency regression
		MaxDecisionChangeRate:      0.10,  // 10% max decision change rate in shadow
		MaxAllowedFPR:              0.05,  // 5% max false positive rate
		MaxAllowedBrierScore:       0.10,  // 0.10 max Brier score
		MaxP95LatencyMs:            15.0,  // 15ms max p95 latency
		AutoApproveCanary:          false, // Require explicit operator approval before Canary rollout
		ShadowMinEvaluations:       50,
		CanaryObservationWindow:    100 * time.Millisecond,
	}
}

// Validate ensures retraining configuration is within safe bounds.
func (cfg *RetrainingConfig) Validate() {
	if cfg.MinSamples < 50 {
		cfg.MinSamples = 50
	}
	if cfg.DriftThreshold <= 0.0 {
		cfg.DriftThreshold = 0.20
	}
	if cfg.RequiredConsecutiveWindows < 1 {
		cfg.RequiredConsecutiveWindows = 1
	}
	if cfg.CooldownDuration <= 0 {
		cfg.CooldownDuration = 30 * time.Minute
	}
	if cfg.MaxConcurrentJobs <= 0 {
		cfg.MaxConcurrentJobs = 1
	}
	if cfg.MaxErrorRate <= 0.0 {
		cfg.MaxErrorRate = 0.01
	}
	if cfg.MaxFallbackRate <= 0.0 {
		cfg.MaxFallbackRate = 0.01
	}
	if cfg.MaxAllowedFPR <= 0.0 {
		cfg.MaxAllowedFPR = 0.05
	}
	if cfg.MaxAllowedBrierScore <= 0.0 {
		cfg.MaxAllowedBrierScore = 0.10
	}
	if cfg.MaxP95LatencyMs <= 0.0 {
		cfg.MaxP95LatencyMs = 15.0
	}
	if cfg.ShadowMinEvaluations < 10 {
		cfg.ShadowMinEvaluations = 10
	}
}

// RetrainingJob encapsulates a single model retraining execution.
type RetrainingJob struct {
	JobID                 string    `json:"job_id"`
	TriggeredAt           time.Time `json:"triggered_at"`
	State                 JobState  `json:"state"`
	TriggerType           string    `json:"trigger_type"`   // "DRIFT_SUSTAINED" or "MANUAL_OPERATOR"
	TriggerReason         string    `json:"trigger_reason"` // Context details
	ParentModelVersion    string    `json:"parent_model_version"`
	CandidateModelVersion string    `json:"candidate_model_version"`
	DatasetID             string    `json:"dataset_id"`
	SampleCount           uint32    `json:"sample_count"`
	CompletedAt           time.Time `json:"completed_at,omitempty"`
	DurationMs            float64   `json:"duration_ms"`
	Error                 string    `json:"error,omitempty"`
}

// TrainingDatasetMetadata encapsulates data quality metrics and metadata for training data.
type TrainingDatasetMetadata struct {
	DatasetID              string    `json:"dataset_id"`
	TimeWindowStart        time.Time `json:"time_window_start"`
	TimeWindowEnd          time.Time `json:"time_window_end"`
	SampleCount            uint32    `json:"sample_count"`
	FeatureContract        string    `json:"feature_contract"`
	ParentModelVersion     string    `json:"parent_model_version"`
	FraudRate              float64   `json:"fraud_rate"`
	DataQualityScore       float64   `json:"data_quality_score"` // 0.0 to 1.0
	MissingValueRate       float64   `json:"missing_value_rate"`
	SchemaCompatible       bool      `json:"schema_compatible"`
	ZeroLabelsDetected     bool      `json:"zero_labels_detected"`
	DuplicateRate          float64   `json:"duplicate_rate"`
	LabelDistributionRatio float64   `json:"label_distribution_ratio"`
}

// ValidationMetrics encapsulates offline evaluation performance metrics.
type ValidationMetrics struct {
	ROCAUC           float64 `json:"roc_auc"`
	PRAUC            float64 `json:"pr_auc"`
	Precision        float64 `json:"precision"`
	Recall           float64 `json:"recall"`
	F1Score          float64 `json:"f1_score"`
	FPR              float64 `json:"fpr"`
	FNR              float64 `json:"fnr"`
	BrierScore       float64 `json:"brier_score"`
	CalibrationError float64 `json:"calibration_error"`
	P95LatencyMs     float64 `json:"p95_latency_ms"`
	InferenceErrors  int     `json:"inference_errors"`
	NaNCount         int     `json:"nan_count"`
}

// ValidationGateResult represents the outcome of offline validation against the production baseline.
type ValidationGateResult struct {
	ValidationID       string            `json:"validation_id"`
	Timestamp          time.Time         `json:"timestamp"`
	ModelID            string            `json:"model_id"`
	ModelVersion       string            `json:"model_version"`
	ParentModelVersion string            `json:"parent_model_version"`
	Passed             bool              `json:"passed"`
	Violations         []string          `json:"violations"`
	Warnings           []string          `json:"warnings"`
	CandidateMetrics   ValidationMetrics `json:"candidate_metrics"`
	BaselineMetrics    ValidationMetrics `json:"baseline_metrics"`
	GateDetails        string            `json:"gate_details"`
}

// ModelCandidate represents an immutable candidate model in the registry.
type ModelCandidate struct {
	ModelID            string                `json:"model_id"`
	Version            string                `json:"version"`
	ParentModelVersion string                `json:"parent_model_version"`
	FeatureContract    string                `json:"feature_contract"`
	CalibrationVersion string                `json:"calibration_version"`
	TrainingJobID      string                `json:"training_job_id"`
	DatasetID          string                `json:"dataset_id"`
	CreatedAt          time.Time             `json:"created_at"`
	ArtifactChecksum   string                `json:"artifact_checksum"`
	ConfigHash         string                `json:"config_hash"`
	State              JobState              `json:"state"`
	ValidationResult   *ValidationGateResult `json:"validation_result,omitempty"`
	ShadowResult       *ShadowGateResult     `json:"shadow_result,omitempty"`
	CanaryResult       *CanaryGateResult     `json:"canary_result,omitempty"`
}

// ShadowGateResult represents the outcome of live shadow evaluation.
type ShadowGateResult struct {
	EvaluationID            string    `json:"evaluation_id"`
	Timestamp               time.Time `json:"timestamp"`
	CandidateModelVersion   string    `json:"candidate_model_version"`
	ProductionModelVersion  string    `json:"production_model_version"`
	SamplesEvaluated        uint32    `json:"samples_evaluated"`
	ScoreDivergenceRate     float64   `json:"score_divergence_rate"`
	DecisionChangeRate      float64   `json:"decision_change_rate"`
	ErrorRate               float64   `json:"error_rate"`
	FallbackRate            float64   `json:"fallback_rate"`
	AvgScoreDelta           float64   `json:"avg_score_delta"`
	P95LatencyMs            float64   `json:"p95_latency_ms"`
	Passed                  bool      `json:"passed"`
	Violations              []string  `json:"violations"`
	GateDetails             string    `json:"gate_details"`
}

// CanaryGateResult represents candidate safety gate evaluations across canary stages.
type CanaryGateResult struct {
	RolloutID             string    `json:"rollout_id"`
	CandidateModelVersion string    `json:"candidate_model_version"`
	StagePercentage       int       `json:"stage_percentage"`
	Passed                bool      `json:"passed"`
	Violations            []string  `json:"violations"`
	Warnings              []string  `json:"warnings"`
	ErrorRate             float64   `json:"error_rate"`
	FallbackRate          float64   `json:"fallback_rate"`
	P95LatencyMs          float64   `json:"p95_latency_ms"`
	P99LatencyMs          float64   `json:"p99_latency_ms"`
	DecisionChangeRate    float64   `json:"decision_change_rate"`
	ActionTaken           string    `json:"action_taken"` // "STAGE_ADVANCED", "PROMOTED", "ROLLED_BACK"
}

// RetrainingStatusSummary provides a clean overview for GET /v1/system/status.
type RetrainingStatusSummary struct {
	Enabled               bool      `json:"enabled"`
	State                 JobState  `json:"state"`
	ActiveJobID           *string   `json:"active_job_id"`
	CandidateModel        *string   `json:"candidate_model"`
	LastTrigger           string    `json:"last_trigger"`
	LastSuccessfulTrain   string    `json:"last_successful_training"`
	LastFailure           string    `json:"last_failure"`
	CooldownRemainingSec  int64     `json:"cooldown_remaining_sec"`
	TrainingAdapterStatus string    `json:"training_adapter_status"`
}

// ModelLifecycleEvent tracks auditable state transitions.
type ModelLifecycleEvent struct {
	EventID       string    `json:"event_id"`
	Timestamp     time.Time `json:"timestamp"`
	ModelID       string    `json:"model_id"`
	ModelVersion  string    `json:"model_version"`
	PreviousState JobState  `json:"previous_state"`
	NewState      JobState  `json:"new_state"`
	Trigger       string    `json:"trigger"`
	Actor         string    `json:"actor"`
	Reason        string    `json:"reason"`
}

// String provides formatted event description.
func (e ModelLifecycleEvent) String() string {
	return fmt.Sprintf("[%s] Model %s transitioned %s -> %s (trigger: %s, actor: %s): %s",
		e.Timestamp.Format(time.RFC3339), e.ModelVersion, e.PreviousState, e.NewState, e.Trigger, e.Actor, e.Reason)
}
