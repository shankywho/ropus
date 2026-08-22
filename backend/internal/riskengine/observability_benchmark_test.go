package riskengine

import (
	"context"
	"testing"
	"time"
)

func BenchmarkMetrics_RecordRequest(b *testing.B) {
	metrics := NewMetricsEngine()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			metrics.RecordRequest(12.5, "ALLOW", 15, true, false, false)
		}
	})
}

func BenchmarkSLO_RecordEvaluation(b *testing.B) {
	engine := NewSLOEngine(5 * time.Minute)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			engine.RecordEvaluation(14.0, true, false, false)
		}
	})
}

func BenchmarkSLO_Evaluate(b *testing.B) {
	engine := NewSLOEngine(5 * time.Minute)
	for i := 0; i < 1000; i++ {
		engine.RecordEvaluation(float64(i%50), true, false, false)
	}
	now := time.Now().UTC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.Evaluate(now)
	}
}

func BenchmarkHealthAggregator_GetHealthReport(b *testing.B) {
	agg := NewHealthAggregator(nil, nil, nil, nil, nil, nil, nil, nil)
	defer agg.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = agg.GetHealthReport()
	}
}

func BenchmarkIncidentEngine_Evaluate(b *testing.B) {
	incidentEngine := NewIncidentEngine(nil, nil)
	ctx := context.Background()
	health := SystemHealthReport{
		OverallStatus: HealthStatusHealthy,
		Components:    make(map[string]ComponentHealth),
	}
	slo := SLOSummary{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = incidentEngine.Evaluate(ctx, health, slo, CircuitStateHealthy, DriftStatusHealthy, StateIdle, "v3.0")
	}
}

func BenchmarkMetrics_ExportPrometheus(b *testing.B) {
	metrics := NewMetricsEngine()
	metrics.RecordRequest(10.0, "ALLOW", 20, true, false, false)
	sloEngine := NewSLOEngine(5 * time.Minute)
	slo := sloEngine.Evaluate(time.Now().UTC())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metrics.ExportPrometheus(&slo, "fraud-xgb-25f-v3.0", "fraud-xgb-15f-v1.5", DriftStatusHealthy, 0.05, 0.02, 10, CircuitStateHealthy, false)
	}
}
