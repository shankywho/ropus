package riskengine

import (
	"context"
	"testing"
)

func BenchmarkDriftCollector_PushVector(b *testing.B) {
	featureNames := make([]string, 25)
	for i := 0; i < 25; i++ {
		featureNames[i] = Canonical25FeatureDefinitions[i].Name
	}
	collector := NewDriftCollector(10000, featureNames)

	vector := make(map[string]float64, 25)
	for _, name := range featureNames {
		vector[name] = 123.45
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		collector.PushVector(vector)
	}
}

func BenchmarkDriftDetector_IngestVector(b *testing.B) {
	cfg := DefaultDriftConfig()
	detector, err := NewDriftDetector(cfg, nil)
	if err != nil {
		b.Fatalf("failed to init detector: %v", err)
	}

	vector := make(map[string]float64, 25)
	for _, f := range Canonical25FeatureDefinitions {
		vector[f.Name] = f.DefaultValue
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		detector.IngestVector(vector)
	}
}

func BenchmarkDriftCalculator_CalculatePSI(b *testing.B) {
	actual := []float64{0.1, 0.2, 0.3, 0.2, 0.1, 0.05, 0.03, 0.01, 0.005, 0.005}
	expected := []float64{0.1, 0.1, 0.2, 0.2, 0.2, 0.1, 0.05, 0.03, 0.01, 0.01}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CalculatePSI(actual, expected, 1e-6)
	}
}

func BenchmarkDriftCalculator_CalculateJSDivergence(b *testing.B) {
	p := []float64{0.1, 0.2, 0.3, 0.2, 0.1, 0.05, 0.03, 0.01, 0.005, 0.005}
	q := []float64{0.1, 0.1, 0.2, 0.2, 0.2, 0.1, 0.05, 0.03, 0.01, 0.01}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CalculateJSDivergence(p, q)
	}
}

func BenchmarkDriftCalculator_CalculateFeatureDrift(b *testing.B) {
	cfg := DefaultDriftConfig()
	baseline := FeatureDistribution{
		Name:     "amount",
		DataType: "float64",
		Mean:     100.0,
		Std:      50.0,
		BinEdges: []float64{0.0, 50.0, 100.0, 150.0, 200.0, 250.0, 300.0, 350.0, 400.0, 500.0, 1000.0},
		BinProbs: []float64{0.1, 0.1, 0.2, 0.2, 0.15, 0.1, 0.05, 0.05, 0.03, 0.02},
	}

	values := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		values[i] = float64(i % 500)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CalculateFeatureDrift(values, baseline, cfg)
	}
}

func BenchmarkDriftDetector_EvaluateLiveDrift(b *testing.B) {
	cfg := DefaultDriftConfig()
	cfg.MaxWindowSize = 1000
	cfg.MinSamplesForDrift = 10
	detector, err := NewDriftDetector(cfg, nil)
	if err != nil {
		b.Fatalf("failed to init detector: %v", err)
	}

	for i := 0; i < 1000; i++ {
		vector := make(map[string]float64, 25)
		for _, f := range Canonical25FeatureDefinitions {
			vector[f.Name] = f.DefaultValue + float64(i%10)
		}
		detector.IngestVector(vector)
	}

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = detector.EvaluateLiveDrift(ctx)
	}
}
