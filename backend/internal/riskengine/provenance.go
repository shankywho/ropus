package riskengine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ModelProvenance encapsulates the end-to-end audit and reproducibility chain for a model.
// Every production candidate must be traceable through:
// Dataset -> Dataset Checksum -> Training Config -> Config Checksum -> Job ID ->
// Model Artifact -> Artifact Checksum -> Validation -> Shadow Evaluation -> Approval -> Canary -> Promotion.
type ModelProvenance struct {
	DatasetURI           string     `json:"dataset_uri"`
	DatasetChecksum      string     `json:"dataset_checksum"`
	DatasetVersion       string     `json:"dataset_version"`
	DatasetRowCount      int64      `json:"dataset_row_count"`
	TrainingConfigHash   string     `json:"training_config_hash"`
	TrainingJobID        string     `json:"training_job_id"`
	ParentModelVersion   string     `json:"parent_model_version"`
	CandidateVersion     string     `json:"candidate_version"`
	ArtifactURI          string     `json:"artifact_uri"`
	ArtifactChecksum     string     `json:"artifact_checksum"`
	ValidationReportHash string     `json:"validation_report_hash,omitempty"`
	ValidationPassed     bool       `json:"validation_passed"`
	ShadowEvaluationID   string     `json:"shadow_evaluation_id,omitempty"`
	ShadowPassed         bool       `json:"shadow_passed"`
	ApprovalActor        string     `json:"approval_actor,omitempty"`
	ApprovalReason       string     `json:"approval_reason,omitempty"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`
	CanaryStageCompleted int        `json:"canary_stage_completed,omitempty"`
	PromotedAt           *time.Time `json:"promoted_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// ComputeSHA256Checksum computes a lowercase hex SHA-256 hash of arbitrary byte slices.
func ComputeSHA256Checksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ComputeConfigChecksum produces a deterministic SHA-256 hash for any training configuration.
func ComputeConfigChecksum(cfg interface{}) (string, error) {
	bytes, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config for hashing: %w", err)
	}
	return ComputeSHA256Checksum(bytes), nil
}

// ValidateProvenanceChain validates that every required step in the reproducibility chain is complete.
func ValidateProvenanceChain(p *ModelProvenance) error {
	if p == nil {
		return errors.New("provenance record is nil")
	}
	if p.CandidateVersion == "" {
		return errors.New("provenance candidate version cannot be empty")
	}
	if p.DatasetChecksum == "" {
		return errors.New("provenance dataset checksum cannot be empty")
	}
	if p.TrainingConfigHash == "" {
		return errors.New("provenance training config hash cannot be empty")
	}
	if p.TrainingJobID == "" {
		return errors.New("provenance training job ID cannot be empty")
	}
	if p.ArtifactURI == "" {
		return errors.New("provenance artifact URI cannot be empty")
	}
	if p.ArtifactChecksum == "" {
		return errors.New("provenance artifact checksum cannot be empty")
	}
	if !p.ValidationPassed {
		return errors.New("candidate model did not pass offline validation")
	}
	if !p.ShadowPassed {
		return errors.New("candidate model did not pass shadow evaluation")
	}
	if p.ApprovalActor == "" {
		return errors.New("candidate model has not been approved by an authorized operator")
	}
	if p.ApprovedAt == nil {
		return errors.New("approval timestamp is missing")
	}
	return nil
}
