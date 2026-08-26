package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shankywho/ropus/backend/internal/auth/api_keys"
	"github.com/shankywho/ropus/backend/internal/product_api"
	"github.com/shankywho/ropus/backend/internal/saas"
)

func TestIntegrationBenchmark_EndToEndLatency(t *testing.T) {
	keyService := api_keys.NewAPIKeyService()
	meter := saas.NewUsageMeterEngine()
	pipeline := product_api.NewUnifiedRiskPipeline(keyService, meter, nil)

	keyResp, err := keyService.GenerateKey("org_benchmark_corp", "Benchmark Key", "live", []string{"risk:evaluate"}, 90)
	require.NoError(t, err)

	ctx := context.Background()
	req := product_api.CanonicalRiskRequest{
		TransactionID: "tx_bench_001",
		CustomerID:    "usr_bench_clean",
		Amount:        120.0,
		Currency:      "USD",
		MerchantID:    "TargetStore",
		DeviceID:      "dev_bench_01",
		IPAddress:     "192.0.2.1",
		Country:       "US",
		Timestamp:     time.Now().UTC(),
	}

	iterations := 100
	var totalDuration time.Duration

	for i := 0; i < iterations; i++ {
		req.TransactionID = fmt.Sprintf("tx_bench_%03d", i)
		req.CustomerID = fmt.Sprintf("usr_bench_%03d", i)
		req.DeviceID = fmt.Sprintf("dev_bench_%03d", i)
		req.IPAddress = fmt.Sprintf("192.0.2.%d", (i%200)+1)
		start := time.Now()
		res, err := pipeline.EvaluateRisk(ctx, keyResp.PlaintextKey, req)
		dur := time.Since(start)
		totalDuration += dur

		require.NoError(t, err)
		if res.Decision != "APPROVE" {
			t.Fatalf("Decision was %s (score: %.4f, reasons: %v, factors: %+v)", res.Decision, res.RiskScore, res.Reasons, res.RiskFactors)
		}
		assert.Equal(t, "APPROVE", res.Decision)
	}

	avgLatencyMs := (float64(totalDuration.Microseconds()) / float64(iterations)) / 1000.0
	t.Logf("Measured End-to-End Pipeline Latency: %.3f ms per request across %d sequential iterations", avgLatencyMs, iterations)

	// Assert integration SLA: average evaluation latency < 10ms
	assert.Less(t, avgLatencyMs, 10.0, "End-to-End decision pipeline latency must be under 10ms")
}
