package riskengine

import (
	"context"
	"testing"
	"time"

	"github.com/shankywho/ropus/backend/internal/features/store"
)

func BenchmarkMLOps_OnlineFeatureRetrieval(b *testing.B) {
	ctx := context.Background()
	onlineStore := store.NewInMemoryOnlineStore()

	_ = onlineStore.PutOnlineFeatures(ctx, "user_bench_01", map[string]interface{}{
		"amount":            500.0,
		"user_txn_count_1h": 3,
		"user_txn_sum_1h":   1200.0,
		"device_age_days":   45.0,
	}, 1*time.Hour)

	featureNames := []string{"amount", "user_txn_count_1h", "user_txn_sum_1h", "device_age_days"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = onlineStore.GetOnlineFeatures(ctx, "user_bench_01", featureNames)
	}
}

func BenchmarkMLOps_ModelEvaluator(b *testing.B) {
	evaluator := NewModelEvaluator()
	predictions := []float64{0.05, 0.12, 0.88, 0.95, 0.02, 0.76, 0.01, 0.99, 0.04, 0.91}
	groundTruth := []int{0, 0, 1, 1, 0, 1, 0, 1, 0, 1}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.EvaluateModel("candidate_bench_v1", "chk_123", predictions, groundTruth)
	}
}

func BenchmarkMLOps_DataQualityMonitor(b *testing.B) {
	monitor := NewDataQualityMonitor()
	records := []map[string]interface{}{
		{"transaction_id": "txn_1", "amount": 100.0, "user_id": "u1"},
		{"transaction_id": "txn_2", "amount": 250.0, "user_id": "u2"},
		{"transaction_id": "txn_3", "amount": 50.0, "user_id": "u3"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = monitor.EvaluateBatch(records, 0.05)
	}
}
