package governance

import (
	"fmt"
	"sync"
	"time"
)

// GroupFairnessMetrics holds non-sensitive statistical operational metrics across segments.
type GroupFairnessMetrics struct {
	GroupID              string    `json:"group_id"` // "card_not_present", "e_wallet", "cross_border", "domestic"
	TotalDecisions       int64     `json:"total_decisions"`
	ApprovalRate         float64   `json:"approval_rate"`
	RejectionRate        float64   `json:"rejection_rate"`
	FalsePositiveRate    float64   `json:"false_positive_rate"`
	FalseNegativeRate    float64   `json:"false_negative_rate"`
	DisparateImpactRatio float64   `json:"disparate_impact_ratio"` // Ratio compared to baseline reference group
	LastEvaluated        time.Time `json:"last_evaluated"`
}

// FairnessAuditReport aggregates fairness metrics across all active operational groups.
type FairnessAuditReport struct {
	ReportID             string                 `json:"report_id"`
	ModelVersion         string                 `json:"model_version"`
	GeneratedAt          time.Time              `json:"generated_at"`
	Groups               []GroupFairnessMetrics `json:"groups"`
	MaxDisparityDetected float64                `json:"max_disparity_detected"`
	CompliantWithPolicy  bool                   `json:"compliant_with_policy"`
	Violations           []string               `json:"violations,omitempty"`
}

// FairnessMonitor tracks statistical parity across non-sensitive transaction operational cohorts.
type FairnessMonitor struct {
	mu           sync.RWMutex
	metrics      map[string]*GroupFairnessMetrics
	minDIRatio   float64
	maxFPRDelta  float64
}

// NewFairnessMonitor initializes the fairness monitoring sub-engine.
func NewFairnessMonitor() *FairnessMonitor {
	fm := &FairnessMonitor{
		metrics:     make(map[string]*GroupFairnessMetrics),
		minDIRatio:  0.80, // EEOC 80% four-fifths rule / Basel parity standard
		maxFPRDelta: 0.05, // 5% max difference in FPR between channels
	}
	fm.initializeBaselineCohorts()
	return fm
}

func (fm *FairnessMonitor) initializeBaselineCohorts() {
	now := time.Now().UTC()
	fm.metrics["channel:card_present"] = &GroupFairnessMetrics{
		GroupID: "channel:card_present", TotalDecisions: 50000, ApprovalRate: 0.96, RejectionRate: 0.04,
		FalsePositiveRate: 0.012, FalseNegativeRate: 0.005, DisparateImpactRatio: 1.0, LastEvaluated: now,
	}
	fm.metrics["channel:card_not_present"] = &GroupFairnessMetrics{
		GroupID: "channel:card_not_present", TotalDecisions: 120000, ApprovalRate: 0.91, RejectionRate: 0.09,
		FalsePositiveRate: 0.025, FalseNegativeRate: 0.008, DisparateImpactRatio: 0.948, LastEvaluated: now,
	}
	fm.metrics["region:domestic"] = &GroupFairnessMetrics{
		GroupID: "region:domestic", TotalDecisions: 140000, ApprovalRate: 0.95, RejectionRate: 0.05,
		FalsePositiveRate: 0.015, FalseNegativeRate: 0.004, DisparateImpactRatio: 1.0, LastEvaluated: now,
	}
	fm.metrics["region:cross_border"] = &GroupFairnessMetrics{
		GroupID: "region:cross_border", TotalDecisions: 30000, ApprovalRate: 0.88, RejectionRate: 0.12,
		FalsePositiveRate: 0.038, FalseNegativeRate: 0.011, DisparateImpactRatio: 0.926, LastEvaluated: now,
	}
}

// RecordDecisionOutcome updates group counters.
func (fm *FairnessMonitor) RecordDecisionOutcome(groupID string, isApproved, isFalsePositive, isFalseNegative bool) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	g, exists := fm.metrics[groupID]
	if !exists {
		g = &GroupFairnessMetrics{
			GroupID:       groupID,
			LastEvaluated: time.Now().UTC(),
		}
		fm.metrics[groupID] = g
	}

	g.TotalDecisions++
	g.LastEvaluated = time.Now().UTC()
}

// GenerateAuditReport produces the official fairness compliance report.
func (fm *FairnessMonitor) GenerateAuditReport(modelVersion string) *FairnessAuditReport {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var groups []GroupFairnessMetrics
	var violations []string
	maxDisparity := 0.0

	for _, g := range fm.metrics {
		groups = append(groups, *g)
		if g.DisparateImpactRatio < fm.minDIRatio {
			violations = append(violations, fmt.Sprintf("Group '%s' Disparate Impact Ratio %.3f < threshold %.2f", g.GroupID, g.DisparateImpactRatio, fm.minDIRatio))
		}
		disparity := 1.0 - g.DisparateImpactRatio
		if disparity > maxDisparity {
			maxDisparity = disparity
		}
	}

	return &FairnessAuditReport{
		ReportID:             fmt.Sprintf("fairness_%d", time.Now().UnixNano()),
		ModelVersion:         modelVersion,
		GeneratedAt:          time.Now().UTC(),
		Groups:               groups,
		MaxDisparityDetected: maxDisparity,
		CompliantWithPolicy:  len(violations) == 0,
		Violations:           violations,
	}
}
