package product_api

import (
	"testing"
)

func BenchmarkProduct_RiskEvaluation(b *testing.B) {
	evaluator := NewProductRiskEvaluator()
	req := &EvaluateRiskRequest{
		TransactionID: "tx_bench_01",
		UserID:        "usr_bench_01",
		Amount:        1250.0,
		Currency:      "USD",
		Merchant:      "TechRetailDirect",
		Device: DeviceDetails{
			DeviceFingerprint: "fp_bench_device",
			IPAddress:         "198.51.100.1",
			IsEmulator:        false,
			IsVPN:             false,
		},
		Location: LocationDetails{
			Country: "US",
			City:    "New York",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = evaluator.EvaluateTransaction(req)
	}
}
