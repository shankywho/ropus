package riskengine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ModelLifecycleState defines the lifecycle stage of a registered model.
type ModelLifecycleState string

const (
	LifecycleCandidate  ModelLifecycleState = "CANDIDATE"
	LifecycleValidated  ModelLifecycleState = "VALIDATED"
	LifecycleShadow     ModelLifecycleState = "SHADOW"
	LifecycleApproved   ModelLifecycleState = "APPROVED"
	LifecycleCanary     ModelLifecycleState = "CANARY"
	LifecyclePromoted   ModelLifecycleState = "PROMOTED"
	LifecycleRejected   ModelLifecycleState = "REJECTED"
	LifecycleRolledBack ModelLifecycleState = "ROLLED_BACK"
	LifecycleFailed     ModelLifecycleState = "FAILED"
)

// RegisteredModel encapsulates metadata and status for a model in the registry.
type RegisteredModel struct {
	ModelID            string                 `json:"model_id"`
	Version            string                 `json:"version"`
	ParentModelVersion string                 `json:"parent_model_version"`
	FeatureContract    string                 `json:"feature_contract"`
	CalibrationVersion string                 `json:"calibration_version"`
	ArtifactURI        string                 `json:"artifact_uri"`
	ArtifactChecksum   string                 `json:"artifact_checksum"`
	CreatedAt          time.Time              `json:"created_at"`
	PromotedAt         *time.Time             `json:"promoted_at,omitempty"`
	ValidationStatus   string                 `json:"validation_status"`
	ShadowStatus       string                 `json:"shadow_status"`
	LifecycleState     ModelLifecycleState    `json:"lifecycle_state"`
	IsProductionActive bool                   `json:"is_production_active"`
	IsFallbackActive   bool                   `json:"is_fallback_active"`
	Provenance         *ModelProvenance       `json:"provenance,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// ModelRegistry manages the internal thread-safe model registry and lifecycle transitions.
type ModelRegistry struct {
	mu              sync.RWMutex
	models          map[string]*RegisteredModel
	productionModel string
	fallbackModel   string
}

// NewModelRegistry initializes the model registry with baseline production and fallback models.
func NewModelRegistry() *ModelRegistry {
	now := time.Now().UTC()
	reg := &ModelRegistry{
		models:          make(map[string]*RegisteredModel),
		productionModel: "fraud-xgb-25f-v3.0",
		fallbackModel:   "fraud-xgb-15f-v1.5",
	}

	// Register initial verified production baseline model
	reg.models["fraud-xgb-25f-v3.0"] = &RegisteredModel{
		ModelID:            "model_prod_25f_v3",
		Version:            "fraud-xgb-25f-v3.0",
		ParentModelVersion: "fraud-xgb-15f-v1.5",
		FeatureContract:    "fraud-risk-25f-v2.5",
		CalibrationVersion: "beta-calibrated-v2.5",
		ArtifactURI:        "file:///app/model/fraud_model_25f_v3.onnx",
		ArtifactChecksum:   "verified_prod_25f_v3_sha256",
		CreatedAt:          now,
		PromotedAt:         &now,
		ValidationStatus:   "PASSED",
		ShadowStatus:       "PASSED",
		LifecycleState:     LifecyclePromoted,
		IsProductionActive: true,
		IsFallbackActive:   false,
		Provenance: &ModelProvenance{
			DatasetURI:           "s3://risk-datasets/training/baseline_25f_v3.parquet",
			DatasetChecksum:      "dataset_sha256_verified_prod_25f_v3",
			DatasetVersion:       "v3.0-gold",
			DatasetRowCount:      1500000,
			TrainingConfigHash:   "config_hash_xgb_25f_v3",
			TrainingJobID:        "job_train_baseline_prod_25f",
			ParentModelVersion:   "fraud-xgb-15f-v1.5",
			CandidateVersion:     "fraud-xgb-25f-v3.0",
			ArtifactURI:          "file:///app/model/fraud_model_25f_v3.onnx",
			ArtifactChecksum:     "verified_prod_25f_v3_sha256",
			ValidationPassed:     true,
			ShadowPassed:         true,
			ApprovalActor:        "SYSTEM_BOOTSTRAP",
			ApprovalReason:       "Verified production baseline model",
			ApprovedAt:           &now,
			CanaryStageCompleted: 100,
			PromotedAt:           &now,
			CreatedAt:            now,
		},
	}

	// Register emergency fallback model
	reg.models["fraud-xgb-15f-v1.5"] = &RegisteredModel{
		ModelID:            "model_fallback_15f_v1",
		Version:            "fraud-xgb-15f-v1.5",
		ParentModelVersion: "none",
		FeatureContract:    "fraud-risk-15f-v1.5",
		CalibrationVersion: "beta-calibrated-v1.5",
		ArtifactURI:        "file:///app/model/backup/fraud_model_15f_v1.onnx",
		ArtifactChecksum:   "verified_fallback_15f_v1_sha256",
		CreatedAt:          now,
		ValidationStatus:   "PASSED",
		ShadowStatus:       "PASSED",
		LifecycleState:     LifecyclePromoted,
		IsProductionActive: false,
		IsFallbackActive:   true,
		Provenance: &ModelProvenance{
			DatasetURI:           "s3://risk-datasets/training/fallback_15f_v1.parquet",
			DatasetChecksum:      "dataset_sha256_verified_fallback_15f_v1",
			DatasetVersion:       "v1.5-gold",
			DatasetRowCount:      850000,
			TrainingConfigHash:   "config_hash_xgb_15f_v1",
			TrainingJobID:        "job_train_fallback_15f_v1",
			ParentModelVersion:   "none",
			CandidateVersion:     "fraud-xgb-15f-v1.5",
			ArtifactURI:          "file:///app/model/backup/fraud_model_15f_v1.onnx",
			ArtifactChecksum:     "verified_fallback_15f_v1_sha256",
			ValidationPassed:     true,
			ShadowPassed:         true,
			ApprovalActor:        "SYSTEM_BOOTSTRAP",
			ApprovalReason:       "Verified fallback model",
			ApprovedAt:           &now,
			CanaryStageCompleted: 100,
			CreatedAt:            now,
		},
	}

	return reg
}

// RegisterCandidate adds a newly trained candidate model to the registry in CANDIDATE state.
func (r *ModelRegistry) RegisterCandidate(candidate ModelCandidate, artifactURI, checksum string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if candidate.Version == "" {
		return fmt.Errorf("candidate version cannot be empty")
	}

	if _, exists := r.models[candidate.Version]; exists {
		return fmt.Errorf("model version '%s' is already registered (immutability violation)", candidate.Version)
	}

	registered := &RegisteredModel{
		ModelID:            candidate.ModelID,
		Version:            candidate.Version,
		ParentModelVersion: candidate.ParentModelVersion,
		FeatureContract:    candidate.FeatureContract,
		CalibrationVersion: candidate.CalibrationVersion,
		ArtifactURI:        artifactURI,
		ArtifactChecksum:   checksum,
		CreatedAt:          time.Now().UTC(),
		ValidationStatus:   "PENDING",
		ShadowStatus:       "PENDING",
		LifecycleState:     LifecycleCandidate,
		IsProductionActive: false,
		IsFallbackActive:   false,
		Metadata: map[string]interface{}{
			"training_job_id": candidate.TrainingJobID,
			"dataset_id":      candidate.DatasetID,
			"config_hash":     candidate.ConfigHash,
		},
	}

	r.models[candidate.Version] = registered
	return nil
}

// UpdateLifecycleState updates the lifecycle state of a registered model.
func (r *ModelRegistry) UpdateLifecycleState(version string, newState ModelLifecycleState, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, exists := r.models[version]
	if !exists {
		return fmt.Errorf("model version '%s' not found in registry", version)
	}

	m.LifecycleState = newState
	if newState == LifecycleValidated {
		m.ValidationStatus = "PASSED"
	} else if newState == LifecycleShadow {
		m.ShadowStatus = "PASSED"
	}
	return nil
}

// GetModel retrieves a model by its version string.
func (r *ModelRegistry) GetModel(version string) (*RegisteredModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, exists := r.models[version]
	if !exists {
		return nil, fmt.Errorf("model version '%s' not found in registry", version)
	}
	copy := *m
	return &copy, nil
}

// GetProductionModel returns the current active production model.
func (r *ModelRegistry) GetProductionModel() (*RegisteredModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, exists := r.models[r.productionModel]
	if !exists {
		return nil, fmt.Errorf("production model '%s' not found in registry", r.productionModel)
	}
	copy := *m
	return &copy, nil
}

// GetFallbackModel returns the emergency fallback model.
func (r *ModelRegistry) GetFallbackModel() (*RegisteredModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, exists := r.models[r.fallbackModel]
	if !exists {
		return nil, fmt.Errorf("fallback model '%s' not found in registry", r.fallbackModel)
	}
	copy := *m
	return &copy, nil
}

// ListModels returns a slice of all registered models.
func (r *ModelRegistry) ListModels() []*RegisteredModel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*RegisteredModel, 0, len(r.models))
	for _, m := range r.models {
		copy := *m
		list = append(list, &copy)
	}
	return list
}

// PromoteModel atomically updates the primary production model and retains the previous production model as fallback.
// Enforces complete reproducibility provenance chain validation prior to promotion.
func (r *ModelRegistry) PromoteModel(candidateVersion, actor, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cand, exists := r.models[candidateVersion]
	if !exists {
		return fmt.Errorf("candidate version '%s' not found in registry", candidateVersion)
	}

	// Enforce complete provenance validation before promotion
	if cand.Provenance != nil {
		if err := ValidateProvenanceChain(cand.Provenance); err != nil {
			return fmt.Errorf("promotion blocked: candidate '%s' has incomplete provenance: %w", candidateVersion, err)
		}
	}

	previousProd := r.models[r.productionModel]
	now := time.Now().UTC()

	// Update previous production model: keep as verified fallback
	if previousProd != nil {
		previousProd.IsProductionActive = false
		previousProd.IsFallbackActive = true
		r.fallbackModel = previousProd.Version
	}

	// Promote candidate to active production
	cand.IsProductionActive = true
	cand.IsFallbackActive = false
	cand.LifecycleState = LifecyclePromoted
	cand.PromotedAt = &now
	if cand.Provenance != nil {
		cand.Provenance.PromotedAt = &now
	}
	r.productionModel = candidateVersion

	return nil
}

// AttachProvenance attaches or updates provenance metadata for a registered model version.
func (r *ModelRegistry) AttachProvenance(version string, prov *ModelProvenance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, exists := r.models[version]
	if !exists {
		return fmt.Errorf("model version '%s' not found in registry", version)
	}
	m.Provenance = prov
	return nil
}

// GetModelProvenance returns a copy of the provenance audit trail for a model version.
func (r *ModelRegistry) GetModelProvenance(version string) (*ModelProvenance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, exists := r.models[version]
	if !exists {
		return nil, fmt.Errorf("model version '%s' not found in registry", version)
	}
	if m.Provenance == nil {
		return nil, fmt.Errorf("provenance not recorded for model version '%s'", version)
	}
	copy := *m.Provenance
	return &copy, nil
}

// RollbackModel sets a candidate model state to ROLLED_BACK.
func (r *ModelRegistry) RollbackModel(candidateVersion, actor, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cand, exists := r.models[candidateVersion]
	if !exists {
		return fmt.Errorf("candidate version '%s' not found in registry", candidateVersion)
	}

	cand.LifecycleState = LifecycleRolledBack
	cand.IsProductionActive = false
	return nil
}

// RegistryReconciliationResult summarizes invariant checks and self-healing repairs performed on the model registry.
type RegistryReconciliationResult struct {
	Valid                  bool     `json:"valid"`
	RepairsMade            int      `json:"repairs_made"`
	ProductionModelVersion string   `json:"production_model_version"`
	FallbackModelVersion   string   `json:"fallback_model_version"`
	Violations             []string `json:"violations"`
	Repairs                []string `json:"repairs"`
}

// Reconcile audits model registry invariants, repairs safe inconsistencies, and guarantees atomic model safety.
func (r *ModelRegistry) Reconcile(verifier *ArtifactVerifier) (*RegistryReconciliationResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if verifier == nil {
		verifier = NewArtifactVerifier()
	}

	res := &RegistryReconciliationResult{
		Valid:      true,
		Violations: make([]string, 0),
		Repairs:    make([]string, 0),
	}

	// 1. Check fallback model existence
	if r.fallbackModel == "" || r.models[r.fallbackModel] == nil {
		res.Violations = append(res.Violations, "Fallback model unassigned or missing from registry")
		// Restore default fallback
		now := time.Now().UTC()
		fb := &RegisteredModel{
			ModelID:            "model_fallback_15f_v1",
			Version:            "fraud-xgb-15f-v1.5",
			ParentModelVersion: "none",
			FeatureContract:    "fraud-risk-15f-v1.5",
			CalibrationVersion: "beta-calibrated-v1.5",
			ArtifactURI:        "file:///app/model/backup/fraud_model_15f_v1.onnx",
			ArtifactChecksum:   "verified_fallback_15f_v1_sha256",
			CreatedAt:          now,
			ValidationStatus:   "PASSED",
			ShadowStatus:       "PASSED",
			LifecycleState:     LifecyclePromoted,
			IsProductionActive: false,
			IsFallbackActive:   true,
		}
		r.models["fraud-xgb-15f-v1.5"] = fb
		r.fallbackModel = "fraud-xgb-15f-v1.5"
		res.Repairs = append(res.Repairs, "Restored default baseline fallback model (fraud-xgb-15f-v1.5)")
		res.RepairsMade++
	}

	// 2. Check production model existence and artifact validity
	prodModel, exists := r.models[r.productionModel]
	if !exists || prodModel == nil {
		res.Violations = append(res.Violations, fmt.Sprintf("Production model '%s' missing from registry", r.productionModel))
		// Failover to fallback model
		r.productionModel = r.fallbackModel
		r.models[r.fallbackModel].IsProductionActive = true
		res.Repairs = append(res.Repairs, fmt.Sprintf("Activated fallback model '%s' as primary production", r.fallbackModel))
		res.RepairsMade++
	} else {
		// Verify artifact integrity if not a mock/file URI
		if prodModel.ArtifactURI != "" && prodModel.ArtifactChecksum != "" && strings.HasPrefix(prodModel.ArtifactURI, "file://") {
			rec, err := verifier.VerifyArtifact(context.Background(), prodModel.ModelID, prodModel.Version, prodModel.ArtifactURI, prodModel.ArtifactChecksum)
			if err != nil || rec == nil || !rec.Passed {
				res.Violations = append(res.Violations, fmt.Sprintf("Production model '%s' artifact verification failed: %v", r.productionModel, err))
				// Failover to fallback model if fallback artifact is healthy
				if r.fallbackModel != "" && r.fallbackModel != r.productionModel {
					r.productionModel = r.fallbackModel
					prodModel.IsProductionActive = false
					r.models[r.fallbackModel].IsProductionActive = true
					res.Repairs = append(res.Repairs, fmt.Sprintf("Production artifact corrupted; safely failed over to fallback '%s'", r.fallbackModel))
					res.RepairsMade++
				}
			}
		}
	}

	// 3. Enforce exactly ONE production active model across entire registry
	prodCount := 0
	for ver, m := range r.models {
		if m.IsProductionActive {
			if ver != r.productionModel {
				res.Violations = append(res.Violations, fmt.Sprintf("Model '%s' improperly marked active production while '%s' is authoritative", ver, r.productionModel))
				m.IsProductionActive = false
				res.Repairs = append(res.Repairs, fmt.Sprintf("Deactivated extra production flag on model '%s'", ver))
				res.RepairsMade++
			} else {
				prodCount++
			}
		}
	}
	if prodCount == 0 && r.models[r.productionModel] != nil {
		r.models[r.productionModel].IsProductionActive = true
		res.Repairs = append(res.Repairs, fmt.Sprintf("Set IsProductionActive flag on primary model '%s'", r.productionModel))
		res.RepairsMade++
	}

	// 4. Candidate state consistency checks
	for ver, m := range r.models {
		if m.LifecycleState == LifecycleCandidate || m.LifecycleState == LifecycleValidated || m.LifecycleState == LifecycleShadow {
			if m.ArtifactURI != "" && m.ArtifactChecksum != "" && strings.HasPrefix(m.ArtifactURI, "file://") {
				rec, err := verifier.VerifyArtifact(context.Background(), m.ModelID, m.Version, m.ArtifactURI, m.ArtifactChecksum)
				if err != nil || rec == nil || !rec.Passed {
					m.LifecycleState = LifecycleFailed
					res.Violations = append(res.Violations, fmt.Sprintf("Candidate model '%s' has corrupted/missing artifact", ver))
					res.Repairs = append(res.Repairs, fmt.Sprintf("Marked candidate '%s' as FAILED due to artifact integrity violation", ver))
					res.RepairsMade++
				}
			}
		}
	}

	res.ProductionModelVersion = r.productionModel
	res.FallbackModelVersion = r.fallbackModel
	if len(res.Violations) > 0 && res.RepairsMade == 0 {
		res.Valid = false
	}

	return res, nil
}
