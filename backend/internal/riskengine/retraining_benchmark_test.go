package riskengine

import (
	"testing"
)

func BenchmarkRetrainingTriggerEngine_EvaluateDrift(b *testing.B) {
	cfg := DefaultRetrainingConfig()
	engine := NewRetrainingTriggerEngine(cfg)

	meas := &DriftMeasurement{
		SampleCount:   500,
		OverallStatus: DriftStatusDegraded,
		MaxPSI:        0.25,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = engine.EvaluateDrift(meas, CircuitStateHealthy)
	}
}

func BenchmarkOfflineValidator_ValidateCandidate(b *testing.B) {
	cfg := DefaultRetrainingConfig()
	validator := NewOfflineValidator(cfg)
	baseline := DefaultProductionBaselineMetrics()

	candidate := ModelCandidate{
		ModelID:            "model_bench_01",
		Version:            "fraud-xgb-25f-v3.1-candidate",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
	}

	metrics := ValidationMetrics{
		ROCAUC:           0.9150,
		PRAUC:            0.7400,
		Precision:        0.8300,
		Recall:           0.7700,
		FPR:              0.0210,
		BrierScore:       0.0390,
		CalibrationError: 0.0130,
		P95LatencyMs:     6.10,
		InferenceErrors:  0,
		NaNCount:         0,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = validator.ValidateCandidate(candidate, metrics, baseline)
	}
}

func BenchmarkRetrainingCoordinator_GetStatus(b *testing.B) {
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, nil, nil, nil, nil, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = coordinator.GetStatus()
	}
}
