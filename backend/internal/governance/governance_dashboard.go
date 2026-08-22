package governance

import (
	"time"
)

// DashboardOverview summarizes key governance, fairness, and compliance metrics for executive dashboards.
type DashboardOverview struct {
	Timestamp                    time.Time `json:"timestamp"`
	ActiveModelsCount            int       `json:"active_models_count"`
	ModelsPendingApproval        int       `json:"models_pending_approval"`
	TotalAuditedDecisions        int       `json:"total_audited_decisions"`
	OpenManualReviews            int       `json:"open_manual_reviews"`
	FairnessCompliant            bool      `json:"fairness_compliant"`
	AuditChainIntegrityVerified  bool      `json:"audit_chain_integrity_verified"`
	ActivePolicyVersion          string    `json:"active_policy_version"`
	RegulatoryReadinessStatus    string    `json:"regulatory_readiness_status"`
}

// GovernanceDashboardAggregator compiles real-time compliance status across all governance subsystems.
type GovernanceDashboardAggregator struct {
	modelManager *ModelRiskManager
	auditTrail   *DecisionAuditTrail
	fairness     *FairnessMonitor
	reviews      *HumanReviewSystem
	policyEngine *PolicyEngine
}

// NewGovernanceDashboardAggregator initializes the unified dashboard aggregator.
func NewGovernanceDashboardAggregator(
	mm *ModelRiskManager,
	at *DecisionAuditTrail,
	fm *FairnessMonitor,
	hrs *HumanReviewSystem,
	pe *PolicyEngine,
) *GovernanceDashboardAggregator {
	if mm == nil {
		mm = NewModelRiskManager()
	}
	if at == nil {
		at = NewDecisionAuditTrail()
	}
	if fm == nil {
		fm = NewFairnessMonitor()
	}
	if hrs == nil {
		hrs = NewHumanReviewSystem()
	}
	if pe == nil {
		pe = NewPolicyEngine()
	}
	return &GovernanceDashboardAggregator{
		modelManager: mm,
		auditTrail:   at,
		fairness:     fm,
		reviews:      hrs,
		policyEngine: pe,
	}
}

// GetOverview compiles real-time system governance health.
func (d *GovernanceDashboardAggregator) GetOverview() *DashboardOverview {
	models := d.modelManager.ListModels()
	activeCount := 0
	pendingCount := 0
	for _, m := range models {
		if m.LifecycleState == StateProduction {
			activeCount++
		} else if m.LifecycleState == StateValidation || (m.LifecycleState == StateDevelopment && !m.ApprovalChain.IsFullyApproved) {
			pendingCount++
		}
	}

	fairnessReport := d.fairness.GenerateAuditReport("production_fleet")
	auditVerified, _ := d.auditTrail.VerifyIntegrity()
	openReviews := len(d.reviews.ListReviews(ReviewPending)) + len(d.reviews.ListReviews(ReviewAssigned))

	return &DashboardOverview{
		Timestamp:                   time.Now().UTC(),
		ActiveModelsCount:           activeCount,
		ModelsPendingApproval:       pendingCount,
		TotalAuditedDecisions:       d.auditTrail.Count(),
		OpenManualReviews:           openReviews,
		FairnessCompliant:           fairnessReport.CompliantWithPolicy,
		AuditChainIntegrityVerified: auditVerified,
		ActivePolicyVersion:         d.policyEngine.activeVersion,
		RegulatoryReadinessStatus:   "COMPLIANT_READY",
	}
}
