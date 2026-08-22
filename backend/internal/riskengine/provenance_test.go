package riskengine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvenance_ValidationCompleteChain(t *testing.T) {
	now := time.Now().UTC()
	prov := &ModelProvenance{
		DatasetURI:           "s3://risk-datasets/training/candidates/cand_v1.parquet",
		DatasetChecksum:      "dataset_sha256_mock_12345",
		DatasetVersion:       "v1.0-training",
		DatasetRowCount:      100000,
		TrainingConfigHash:   "config_hash_abc",
		TrainingJobID:        "job_123",
		ParentModelVersion:   "fraud-xgb-25f-v3.0",
		CandidateVersion:     "fraud-xgb-25f-candidate-v1",
		ArtifactURI:          "file:///app/model/candidates/cand_v1.onnx",
		ArtifactChecksum:     "artifact_sha256_cand_v1",
		ValidationPassed:     true,
		ShadowPassed:         true,
		ApprovalActor:        "ADMIN_OPERATOR",
		ApprovalReason:       "Verified model meets 1% FPR and AUC improvement",
		ApprovedAt:           &now,
		CanaryStageCompleted: 100,
		CreatedAt:            now,
	}

	err := ValidateProvenanceChain(prov)
	assert.NoError(t, err)
}

func TestProvenance_ValidationIncompleteRejection(t *testing.T) {
	now := time.Now().UTC()

	// Missing Approval Actor
	provMissingApproval := &ModelProvenance{
		DatasetChecksum:    "dataset_sha256_mock",
		TrainingConfigHash: "config_hash_mock",
		TrainingJobID:      "job_123",
		CandidateVersion:   "cand_v1",
		ArtifactURI:        "file:///test.onnx",
		ArtifactChecksum:   "sha256_mock",
		ValidationPassed:   true,
		ShadowPassed:       true,
		ApprovalActor:      "", // Missing
		ApprovedAt:         &now,
	}
	err := ValidateProvenanceChain(provMissingApproval)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not been approved")

	// Missing Validation
	provFailedValidation := &ModelProvenance{
		DatasetChecksum:    "dataset_sha256_mock",
		TrainingConfigHash: "config_hash_mock",
		TrainingJobID:      "job_123",
		CandidateVersion:   "cand_v1",
		ArtifactURI:        "file:///test.onnx",
		ArtifactChecksum:   "sha256_mock",
		ValidationPassed:   false, // Failed
		ShadowPassed:       true,
		ApprovalActor:      "ADMIN",
		ApprovedAt:         &now,
	}
	err = ValidateProvenanceChain(provFailedValidation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "did not pass offline validation")
}

func TestProvenance_RegistryPromotionEnforcement(t *testing.T) {
	reg := NewModelRegistry()

	// Candidate with broken provenance cannot be promoted
	brokenCand := ModelCandidate{
		ModelID: "cand_broken",
		Version: "fraud-xgb-25f-broken",
	}
	err := reg.RegisterCandidate(brokenCand, "file:///app/cand_broken.onnx", "sha256_cand_broken")
	require.NoError(t, err)

	brokenProv := &ModelProvenance{
		CandidateVersion:   "fraud-xgb-25f-broken",
		DatasetChecksum:    "dataset_sha256_mock",
		TrainingConfigHash: "config_hash_mock",
		TrainingJobID:      "job_123",
		ArtifactURI:        "file:///app/cand_broken.onnx",
		ArtifactChecksum:   "sha256_cand_broken",
		ValidationPassed:   false, // broken validation
		ShadowPassed:       false,
		ApprovalActor:      "",
		ApprovedAt:         nil,
	}
	_ = reg.AttachProvenance("fraud-xgb-25f-broken", brokenProv)

	// Attempt promotion
	err = reg.PromoteModel("fraud-xgb-25f-broken", "ADMIN", "Test promote broken")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "promotion blocked: candidate 'fraud-xgb-25f-broken' has incomplete provenance")

	// Ensure active production model was untouched
	prod, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", prod.Version)

	// Get Baseline Provenance
	baseProv, err := reg.GetModelProvenance("fraud-xgb-25f-v3.0")
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", baseProv.CandidateVersion)
	assert.NotEmpty(t, baseProv.DatasetChecksum)
}
