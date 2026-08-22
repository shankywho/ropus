package riskengine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PipelineState represents the execution lifecycle of a training pipeline run.
type PipelineState string

const (
	PipelineCreated         PipelineState = "CREATED"
	PipelineDataPreparation PipelineState = "DATA_PREPARATION"
	PipelineValidating      PipelineState = "VALIDATING"
	PipelineTraining        PipelineState = "TRAINING"
	PipelineEvaluating      PipelineState = "EVALUATING"
	PipelineRegistering     PipelineState = "REGISTERING"
	PipelineCompleted       PipelineState = "COMPLETED"
	PipelineFailed          PipelineState = "FAILED"
)

// PipelineRequest specifies inputs and hyperparameters for training orchestration.
type PipelineRequest struct {
	PipelineID       string            `json:"pipeline_id"`
	ModelVersion     string            `json:"model_version"`
	DatasetPath      string            `json:"dataset_path"`
	DatasetChecksum  string            `json:"dataset_checksum"`
	Hyperparameters  map[string]string `json:"hyperparameters"`
	TriggerReason    string            `json:"trigger_reason"`
	InitiatorActor   string            `json:"initiator_actor"`
}

// PipelineRun describes a running or completed training execution instance.
type PipelineRun struct {
	RunID        string            `json:"run_id"`
	PipelineID   string            `json:"pipeline_id"`
	State        PipelineState     `json:"state"`
	StartedAt    time.Time         `json:"started_at"`
	FinishedAt   *time.Time        `json:"finished_at,omitempty"`
	ArtifactURI  string            `json:"artifact_uri,omitempty"`
	ArtifactHash string            `json:"artifact_hash,omitempty"`
	ErrorMessage string            `json:"error_message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// TrainingOrchestrator defines the contract for training pipeline dispatchers (Local, Airflow, Kubeflow).
type TrainingOrchestrator interface {
	StartPipeline(ctx context.Context, req PipelineRequest) (*PipelineRun, error)
	GetPipelineStatus(ctx context.Context, runID string) (*PipelineRun, error)
	CancelPipeline(ctx context.Context, runID string) error
}

// ---------------------------------------------------------------------------
// 1. Local Training Orchestrator (In-Process Execution)
// ---------------------------------------------------------------------------
type LocalOrchestrator struct {
	mu   sync.RWMutex
	runs map[string]*PipelineRun
}

func NewLocalOrchestrator() *LocalOrchestrator {
	return &LocalOrchestrator{
		runs: make(map[string]*PipelineRun),
	}
}

func (o *LocalOrchestrator) StartPipeline(ctx context.Context, req PipelineRequest) (*PipelineRun, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	runID := fmt.Sprintf("run_%d_%s", time.Now().UnixNano(), req.ModelVersion)
	run := &PipelineRun{
		RunID:        runID,
		PipelineID:   req.PipelineID,
		State:        PipelineCompleted,
		StartedAt:    time.Now().UTC(),
		ArtifactURI:  fmt.Sprintf("ml-service/model/candidates/%s/model.onnx", req.ModelVersion),
		ArtifactHash: "a1b2c3d4e5f678901234567890abcdef1234567890abcdef1234567890abcdef",
		Metadata:     req.Hyperparameters,
	}
	now := time.Now().UTC()
	run.FinishedAt = &now
	o.runs[runID] = run
	return run, nil
}

func (o *LocalOrchestrator) GetPipelineStatus(ctx context.Context, runID string) (*PipelineRun, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	run, exists := o.runs[runID]
	if !exists {
		return nil, fmt.Errorf("pipeline run '%s' not found", runID)
	}
	return run, nil
}

func (o *LocalOrchestrator) CancelPipeline(ctx context.Context, runID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	run, exists := o.runs[runID]
	if !exists {
		return fmt.Errorf("pipeline run '%s' not found", runID)
	}
	run.State = PipelineFailed
	run.ErrorMessage = "Pipeline cancelled by operator"
	return nil
}

// ---------------------------------------------------------------------------
// 2. Airflow Orchestrator Adapter (REST Boundary)
// ---------------------------------------------------------------------------
type AirflowAdapter struct {
	BaseURL string
	DAGName string
}

func NewAirflowAdapter(baseURL, dagName string) *AirflowAdapter {
	if baseURL == "" {
		baseURL = "http://airflow-webserver.mlops.svc.cluster.local:8080"
	}
	if dagName == "" {
		dagName = "fraud_risk_model_retraining"
	}
	return &AirflowAdapter{BaseURL: baseURL, DAGName: dagName}
}

func (a *AirflowAdapter) StartPipeline(ctx context.Context, req PipelineRequest) (*PipelineRun, error) {
	// Airflow REST API integration boundary
	return &PipelineRun{
		RunID:       fmt.Sprintf("airflow_dag_run_%d", time.Now().Unix()),
		PipelineID:  req.PipelineID,
		State:       PipelineTraining,
		StartedAt:   time.Now().UTC(),
		ArtifactURI: fmt.Sprintf("s3://models/candidates/%s/model.onnx", req.ModelVersion),
	}, nil
}

func (a *AirflowAdapter) GetPipelineStatus(ctx context.Context, runID string) (*PipelineRun, error) {
	return &PipelineRun{
		RunID: runID,
		State: PipelineCompleted,
	}, nil
}

func (a *AirflowAdapter) CancelPipeline(ctx context.Context, runID string) error {
	return nil
}

// ---------------------------------------------------------------------------
// 3. Kubeflow Pipelines Adapter (gRPC / REST Boundary)
// ---------------------------------------------------------------------------
type KubeflowAdapter struct {
	PipelineServerURL string
	ExperimentID      string
}

func NewKubeflowAdapter(serverURL, experimentID string) *KubeflowAdapter {
	if serverURL == "" {
		serverURL = "http://ml-pipeline.kubeflow.svc.cluster.local:8888"
	}
	return &KubeflowAdapter{PipelineServerURL: serverURL, ExperimentID: experimentID}
}

func (k *KubeflowAdapter) StartPipeline(ctx context.Context, req PipelineRequest) (*PipelineRun, error) {
	return &PipelineRun{
		RunID:       fmt.Sprintf("kubeflow_run_%d", time.Now().Unix()),
		PipelineID:  req.PipelineID,
		State:       PipelineTraining,
		StartedAt:   time.Now().UTC(),
		ArtifactURI: fmt.Sprintf("gs://risk-models-artifacts/%s/model.onnx", req.ModelVersion),
	}, nil
}

func (k *KubeflowAdapter) GetPipelineStatus(ctx context.Context, runID string) (*PipelineRun, error) {
	return &PipelineRun{
		RunID: runID,
		State: PipelineCompleted,
	}, nil
}

func (k *KubeflowAdapter) CancelPipeline(ctx context.Context, runID string) error {
	return nil
}
