package riskengine

import (
	"fmt"
	"strings"
	"time"
)

// OfflineValidator evaluates candidate model performance against production baseline safety thresholds.
type OfflineValidator struct {
	config RetrainingConfig
}

// NewOfflineValidator initializes an offline model validator.
func NewOfflineValidator(config RetrainingConfig) *OfflineValidator {
	config.Validate()
	return &OfflineValidator{
		config: config,
	}
}

// DefaultProductionBaselineMetrics returns the verified production baseline performance metrics for fraud-xgb-25f-v3.0.
func DefaultProductionBaselineMetrics() ValidationMetrics {
	return ValidationMetrics{
		ROCAUC:           0.8950,
		PRAUC:            0.7100,
		Precision:        0.8100,
		Recall:           0.7600,
		F1Score:          0.7842,
		FPR:              0.0250,
		FNR:              0.2400,
		BrierScore:       0.0420,
		CalibrationError: 0.0150,
		P95LatencyMs:     6.50,
		InferenceErrors:  0,
		NaNCount:         0,
	}
}

// ValidateCandidate executes comprehensive offline validation gates comparing candidate vs baseline.
func (v *OfflineValidator) ValidateCandidate(
	candidate ModelCandidate,
	candidateMetrics ValidationMetrics,
	baselineMetrics ValidationMetrics,
) *ValidationGateResult {
	timestamp := time.Now().UTC()
	validationID := fmt.Sprintf("val_%s_%d", candidate.ModelID, timestamp.UnixNano())

	var violations []string
	var warnings []string

	// 1. Check ROC-AUC regression gate
	minRequiredAUC := baselineMetrics.ROCAUC + v.config.MinModelImprovement
	if candidateMetrics.ROCAUC < minRequiredAUC {
		violations = append(violations, fmt.Sprintf("Candidate ROC-AUC (%.4f) below minimum threshold (%.4f) vs baseline (%.4f)",
			candidateMetrics.ROCAUC, minRequiredAUC, baselineMetrics.ROCAUC))
	}

	// 2. Check False Positive Rate (FPR) gate
	if candidateMetrics.FPR > v.config.MaxAllowedFPR {
		violations = append(violations, fmt.Sprintf("Candidate FPR (%.4f) exceeds maximum allowed FPR (%.4f)",
			candidateMetrics.FPR, v.config.MaxAllowedFPR))
	}

	// 3. Check Brier Score / Calibration Quality gate
	if candidateMetrics.BrierScore > v.config.MaxAllowedBrierScore {
		violations = append(violations, fmt.Sprintf("Candidate Brier score (%.4f) exceeds maximum threshold (%.4f)",
			candidateMetrics.BrierScore, v.config.MaxAllowedBrierScore))
	}

	// 4. Check Latency regression gate
	if candidateMetrics.P95LatencyMs > v.config.MaxP95LatencyMs {
		violations = append(violations, fmt.Sprintf("Candidate P95 latency (%.2f ms) exceeds max latency limit (%.2f ms)",
			candidateMetrics.P95LatencyMs, v.config.MaxP95LatencyMs))
	}

	// 5. Check inference reliability & arithmetic integrity
	if candidateMetrics.InferenceErrors > 0 {
		violations = append(violations, fmt.Sprintf("Candidate encountered %d inference errors during validation", candidateMetrics.InferenceErrors))
	}
	if candidateMetrics.NaNCount > 0 {
		violations = append(violations, fmt.Sprintf("Candidate produced %d NaN/Inf output values during validation", candidateMetrics.NaNCount))
	}

	// Warnings
	if candidateMetrics.CalibrationError > baselineMetrics.CalibrationError*1.2 {
		warnings = append(warnings, fmt.Sprintf("Candidate calibration error (%.4f) is 20%% higher than baseline (%.4f)",
			candidateMetrics.CalibrationError, baselineMetrics.CalibrationError))
	}

	passed := len(violations) == 0

	gateDetails := "ALL_GATES_PASSED"
	if !passed {
		gateDetails = fmt.Sprintf("FAILED: %s", strings.Join(violations, "; "))
	} else if len(warnings) > 0 {
		gateDetails = fmt.Sprintf("PASSED_WITH_WARNINGS: %s", strings.Join(warnings, "; "))
	}

	return &ValidationGateResult{
		ValidationID:       validationID,
		Timestamp:          timestamp,
		ModelID:            candidate.ModelID,
		ModelVersion:       candidate.Version,
		ParentModelVersion: candidate.ParentModelVersion,
		Passed:             passed,
		Violations:         violations,
		Warnings:           warnings,
		CandidateMetrics:   candidateMetrics,
		BaselineMetrics:    baselineMetrics,
		GateDetails:        gateDetails,
	}
}
