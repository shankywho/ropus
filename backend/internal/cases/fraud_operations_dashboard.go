package cases

import (
	"time"
)

// OperationsDashboardSummary aggregates high-level metrics for fraud operation centers (FOC).
type OperationsDashboardSummary struct {
	Timestamp                    time.Time `json:"timestamp"`
	OpenCasesCount               int       `json:"open_cases_count"`
	CriticalCasesCount           int       `json:"critical_cases_count"`
	TotalExposureUnderReview     float64   `json:"total_exposure_under_review"`
	FraudLossPrevented           float64   `json:"fraud_loss_prevented"`
	FalsePositiveRate            float64   `json:"false_positive_rate"`
	AverageInvestigationSeconds  float64   `json:"average_investigation_seconds"`
	ActiveAlertsCount            int       `json:"active_alerts_count"`
	AnalystWorkloadCount         int       `json:"analyst_workload_count"`
}

// FraudOperationsDashboard compiles operational metrics.
type FraudOperationsDashboard struct {
	caseManager *CaseManager
	alertEngine *FraudAlertEngine
}

// NewFraudOperationsDashboard initializes the operations dashboard aggregator.
func NewFraudOperationsDashboard(cm *CaseManager, ae *FraudAlertEngine) *FraudOperationsDashboard {
	return &FraudOperationsDashboard{
		caseManager: cm,
		alertEngine: ae,
	}
}

// GetSummary compiles the real-time operational dashboard.
func (d *FraudOperationsDashboard) GetSummary() *OperationsDashboardSummary {
	openCases := d.caseManager.ListCases(StatusOpen, "")
	assignedCases := d.caseManager.ListCases(StatusAssigned, "")
	criticalCases := d.caseManager.ListCases("", PriorityCritical)
	resolvedFraud := d.caseManager.ListCases(StatusConfirmedFraud, "")
	falsePositives := d.caseManager.ListCases(StatusFalsePositive, "")

	totalExposure := 0.0
	for _, c := range append(openCases, assignedCases...) {
		totalExposure += c.TotalExposure
	}

	fraudLossPrevented := 0.0
	for _, c := range resolvedFraud {
		fraudLossPrevented += c.TotalExposure
	}

	totalResolved := len(resolvedFraud) + len(falsePositives)
	fpRate := 0.0
	if totalResolved > 0 {
		fpRate = float64(len(falsePositives)) / float64(totalResolved)
	}

	alerts := d.alertEngine.ListAlerts()

	return &OperationsDashboardSummary{
		Timestamp:                   time.Now().UTC(),
		OpenCasesCount:              len(openCases) + len(assignedCases),
		CriticalCasesCount:          len(criticalCases),
		TotalExposureUnderReview:    totalExposure,
		FraudLossPrevented:          fraudLossPrevented,
		FalsePositiveRate:           fpRate,
		AverageInvestigationSeconds: 45.0,
		ActiveAlertsCount:           len(alerts),
		AnalystWorkloadCount:        len(assignedCases),
	}
}
