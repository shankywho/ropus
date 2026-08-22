package governance

import (
	"fmt"
	"sync"
	"time"
)

// ModelRiskTier defines the criticality tier of an AI risk model.
type ModelRiskTier string

const (
	RiskTier1High   ModelRiskTier = "TIER_1_HIGH"   // Direct payment blocking / automated financial loss prevention
	RiskTier2Medium ModelRiskTier = "TIER_2_MEDIUM" // Step-up authentication / manual review recommendations
	RiskTier3Low    ModelRiskTier = "TIER_3_LOW"    // Analytical / background scoring
)

// GovernanceLifecycleState tracks the formal compliance and audit status of an AI model.
type GovernanceLifecycleState string

const (
	StateDevelopment GovernanceLifecycleState = "DEVELOPMENT"
	StateValidation  GovernanceLifecycleState = "VALIDATION"
	StateApproved    GovernanceLifecycleState = "APPROVED"
	StateProduction  GovernanceLifecycleState = "PRODUCTION"
	StateMonitoring  GovernanceLifecycleState = "MONITORING"
	StateRetired     GovernanceLifecycleState = "RETIRED"
)

// GovernanceModelRecord stores immutable compliance, ownership, and risk classification metadata.
type GovernanceModelRecord struct {
	ModelID             string                   `json:"model_id"`
	Version             string                   `json:"version"`
	Owner               string                   `json:"owner"`
	Purpose             string                   `json:"purpose"`
	TrainingDataSource  string                   `json:"training_data_source"`
	FeatureDependencies []string                 `json:"feature_dependencies"`
	RiskTier            ModelRiskTier            `json:"risk_tier"`
	LifecycleState      GovernanceLifecycleState `json:"lifecycle_state"`
	ApprovalChain       ApprovalChainStatus      `json:"approval_chain"`
	CreatedAt           time.Time                `json:"created_at"`
	ApprovedAt          *time.Time               `json:"approved_at,omitempty"`
	RetiredAt           *time.Time               `json:"retired_at,omitempty"`
	Metadata            map[string]string        `json:"metadata,omitempty"`
}

// ApprovalChainStatus details the multi-stakeholder review gate sign-offs.
type ApprovalChainStatus struct {
	ValidationPassed     bool       `json:"validation_passed"`
	ExplainabilityPassed bool       `json:"explainability_passed"`
	FairnessPassed       bool       `json:"fairness_passed"`
	SecurityReviewPassed bool       `json:"security_review_passed"`
	RiskApproverActor    string     `json:"risk_approver_actor,omitempty"`
	RiskApprovedAt       *time.Time `json:"risk_approved_at,omitempty"`
	IsFullyApproved      bool       `json:"is_fully_approved"`
}

// ModelRiskManager maintains the enterprise inventory and lifecycle transitions.
type ModelRiskManager struct {
	mu     sync.RWMutex
	models map[string]*GovernanceModelRecord
}

// NewModelRiskManager initializes the Model Risk Management catalog.
func NewModelRiskManager() *ModelRiskManager {
	m := &ModelRiskManager{
		models: make(map[string]*GovernanceModelRecord),
	}
	m.registerBaselineModel()
	return m
}

func (m *ModelRiskManager) registerBaselineModel() {
	now := time.Now().UTC()
	baseline := &GovernanceModelRecord{
		ModelID:             "model_prod_25f_v3",
		Version:             "fraud-xgb-25f-v3.0",
		Owner:               "risk_ml_team@ropus.internal",
		Purpose:             "Synchronous real-time card and account fraud risk scoring",
		TrainingDataSource:  "s3://risk-datasets/gold/ieee_cis_fraud_2026.parquet",
		FeatureDependencies: []string{"amount", "user_txn_count_1h", "ip_txn_count_1h", "device_age_days"},
		RiskTier:            RiskTier1High,
		LifecycleState:      StateProduction,
		ApprovalChain: ApprovalChainStatus{
			ValidationPassed:     true,
			ExplainabilityPassed: true,
			FairnessPassed:       true,
			SecurityReviewPassed: true,
			RiskApproverActor:    "chief_risk_officer",
			RiskApprovedAt:       &now,
			IsFullyApproved:      true,
		},
		CreatedAt:  now.Add(-30 * 24 * time.Hour),
		ApprovedAt: &now,
	}
	m.models[baseline.Version] = baseline
}

// RegisterModel adds a new model into the governance inventory.
func (m *ModelRiskManager) RegisterModel(rec *GovernanceModelRecord) error {
	if rec == nil || rec.Version == "" {
		return fmt.Errorf("invalid governance model record")
	}
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = time.Now().UTC()
	}
	if rec.LifecycleState == "" {
		rec.LifecycleState = StateDevelopment
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.models[rec.Version]; exists {
		return fmt.Errorf("model version '%s' already exists in governance catalog", rec.Version)
	}

	m.models[rec.Version] = rec
	return nil
}

// GetModel fetches a governance model record by version.
func (m *ModelRiskManager) GetModel(version string) (*GovernanceModelRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rec, exists := m.models[version]
	if !exists {
		return nil, fmt.Errorf("model version '%s' not found in governance catalog", version)
	}
	return rec, nil
}

// ListModels returns all models in the governance catalog.
func (m *ModelRiskManager) ListModels() []*GovernanceModelRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]*GovernanceModelRecord, 0, len(m.models))
	for _, rec := range m.models {
		res = append(res, rec)
	}
	return res
}

// ApproveModel checks that all prerequisite validation/security gates passed and signs off the model.
func (m *ModelRiskManager) ApproveModel(version, approverActor string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, exists := m.models[version]
	if !exists {
		return fmt.Errorf("model version '%s' not found", version)
	}

	// Verify prerequisite gates before granting risk approval
	if !rec.ApprovalChain.ValidationPassed {
		return fmt.Errorf("cannot approve model %s: validation gate not passed", version)
	}
	if !rec.ApprovalChain.ExplainabilityPassed {
		return fmt.Errorf("cannot approve model %s: explainability gate not passed", version)
	}
	if !rec.ApprovalChain.FairnessPassed {
		return fmt.Errorf("cannot approve model %s: fairness gate not passed", version)
	}
	if !rec.ApprovalChain.SecurityReviewPassed {
		return fmt.Errorf("cannot approve model %s: security review not passed", version)
	}

	now := time.Now().UTC()
	rec.ApprovalChain.RiskApproverActor = approverActor
	rec.ApprovalChain.RiskApprovedAt = &now
	rec.ApprovalChain.IsFullyApproved = true
	rec.LifecycleState = StateApproved
	rec.ApprovedAt = &now

	return nil
}

// RetireModel moves an active or deprecated model into RETIRED state.
func (m *ModelRiskManager) RetireModel(version, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, exists := m.models[version]
	if !exists {
		return fmt.Errorf("model version '%s' not found", version)
	}

	now := time.Now().UTC()
	rec.LifecycleState = StateRetired
	rec.RetiredAt = &now
	if rec.Metadata == nil {
		rec.Metadata = make(map[string]string)
	}
	rec.Metadata["retirement_reason"] = reason

	return nil
}
