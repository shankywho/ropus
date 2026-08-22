package riskengine

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// runProgressiveLoadStage executes N requests across concurrent workers while background ML work runs.
func runProgressiveLoadStage(t *testing.T, totalRequests int, concurrency int) (throughput float64, p50, p95, p99, maxLat float64, errorRate float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := DefaultRetrainingConfig()
	cfg.CooldownDuration = 5 * time.Millisecond
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	metricsEngine := NewMetricsEngine()
	sloEngine := NewSLOEngine(5 * time.Minute)

	coordinator.SetMetricsEngine(metricsEngine)
	coordinator.SetSLOEngine(sloEngine)

	// Background ML Worker: Continuous Drift & Retraining (MUST NOT BLOCK INFERENCE)
	var bgStop int32
	var bgRetrainCount int64
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&bgStop) == 1 {
					return
				}
				atomic.AddInt64(&bgRetrainCount, 1)
				m := &DriftMeasurement{
					MeasurementID: "meas_load_01",
					Timestamp:     time.Now().UTC(),
					MaxPSI:        0.35,
					MaxJSD:        0.15,
					OverallStatus: DriftStatusCritical,
					ModelVersion:  "fraud-xgb-25f-v3.0",
				}
				coordinator.OnDriftEvaluated(ctx, m)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Background Canary Shifter
	go func() {
		ticker := time.NewTicker(15 * time.Millisecond)
		defer ticker.Stop()
		stages := []int{0, 10, 25, 50, 100, 0}
		idx := 0
		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&bgStop) == 1 {
					return
				}
				canaryRouter.SetPercentage(stages[idx%len(stages)])
				idx++
			case <-ctx.Done():
				return
			}
		}
	}()

	// Execute Synchronous Evaluation Load
	reqsPerWorker := totalRequests / concurrency
	latencies := make([]float64, totalRequests)
	var successCount int64
	var failCount int64
	var collectedIdx int64

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for i := 0; i < reqsPerWorker; i++ {
				reqStart := time.Now()

				// Synchronous routing and policy check
				route := canaryRouter.Route(fmt.Sprintf("tenant_%d", workerID%5), fmt.Sprintf("txn_%d_%d", workerID, i))
				if route != RouteLegacy && route != RouteCandidate {
					atomic.AddInt64(&failCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}

				latMs := float64(time.Since(reqStart).Nanoseconds()) / 1e6
				idx := atomic.AddInt64(&collectedIdx, 1) - 1
				if idx < int64(len(latencies)) {
					latencies[idx] = latMs
				}

				metricsEngine.RecordRequest(latMs, "ALLOW", r.Intn(100), true, false, false)
				sloEngine.RecordEvaluation(latMs, true, false, false)
			}
		}(w)
	}

	wg.Wait()
	duration := time.Since(start)
	atomic.StoreInt32(&bgStop, 1)

	// Calculate Percentiles
	sort.Float64s(latencies)
	validCount := int(collectedIdx)
	if validCount > len(latencies) {
		validCount = len(latencies)
	}
	sorted := latencies[:validCount]

	if len(sorted) > 0 {
		p50 = sorted[int(float64(len(sorted))*0.50)]
		p95 = sorted[int(float64(len(sorted))*0.95)]
		p99 = sorted[int(float64(len(sorted))*0.99)]
		maxLat = sorted[len(sorted)-1]
	}

	throughput = float64(validCount) / duration.Seconds()
	errorRate = float64(failCount) / float64(validCount)

	t.Logf("[LOAD STAGE: %d requests] Throughput: %.1f req/sec | P50: %.4fms | P95: %.4fms | P99: %.4fms | Max: %.4fms | Errors: %.2f%% | RetrainTriggers: %d",
		totalRequests, throughput, p50, p95, p99, maxLat, errorRate*100, atomic.LoadInt64(&bgRetrainCount))

	return throughput, p50, p95, p99, maxLat, errorRate
}

func TestLoad_ProgressiveScaleWithConcurrentML(t *testing.T) {
	runtime.GC()
	initialGoroutines := runtime.NumGoroutine()

	stages := []int{1000, 5000, 10000, 25000, 50000}

	for _, totalReqs := range stages {
		t.Run(fmt.Sprintf("Load_%d_Requests", totalReqs), func(t *testing.T) {
			throughput, p50, p95, p99, _, errorRate := runProgressiveLoadStage(t, totalReqs, 16)

			assert.Greater(t, throughput, 10000.0, "Throughput must exceed 10,000 req/sec")
			assert.Less(t, p50, 1.0, "P50 latency must be under 1ms")
			assert.Less(t, p95, 5.0, "P95 latency must be under 5ms")
			assert.Less(t, p99, 15.0, "P99 latency must be under 15ms")
			assert.Equal(t, float64(0), errorRate, "Error rate must be zero")
			_ = p99
		})
	}

	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	finalGoroutines := runtime.NumGoroutine()

	assert.LessOrEqual(t, finalGoroutines-initialGoroutines, 3, "Goroutine leak check: delta must be <= 3")
}
