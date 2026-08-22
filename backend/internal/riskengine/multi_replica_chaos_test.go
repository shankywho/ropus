package riskengine

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ReplicaPod simulates an individual backend pod in a Kubernetes StatefulSet/Deployment.
type ReplicaPod struct {
	PodID        string
	Registry     *ModelRegistry
	CanaryRouter *CanaryRouter
	Coordinator  *RetrainingCoordinator
	LockManager  DistributedLock
	Metrics      *MetricsEngine
	SLO          *SLOEngine
	IsRunning    int32
}

func newReplicaPod(podID string, lockManager DistributedLock) *ReplicaPod {
	reg := NewModelRegistry()
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	cfg := DefaultRetrainingConfig()
	cfg.CooldownDuration = 5 * time.Millisecond
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	metrics := NewMetricsEngine()
	slo := NewSLOEngine(5 * time.Minute)

	coordinator.SetMetricsEngine(metrics)
	coordinator.SetSLOEngine(slo)

	return &ReplicaPod{
		PodID:        podID,
		Registry:     reg,
		CanaryRouter: canaryRouter,
		Coordinator:  coordinator,
		LockManager:  lockManager,
		Metrics:      metrics,
		SLO:          slo,
		IsRunning:    1,
	}
}

func TestMultiReplica_HorizontalScaleAndLeaderCoordination(t *testing.T) {
	runtime.GC()
	initialGoroutines := runtime.NumGoroutine()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sharedLock := NewLocalDistributedLock()
	numPods := 4
	pods := make([]*ReplicaPod, numPods)

	for i := 0; i < numPods; i++ {
		pods[i] = newReplicaPod(fmt.Sprintf("pod-risk-backend-%d", i), sharedLock)
	}

	var stopSignal int32
	var totalEvaluations int64
	var totalRetrainsTriggered int64
	var leaderLockAcquisitions int64
	var leaderLockFailures int64
	var podRestarts int64

	var wg sync.WaitGroup

	// -----------------------------------------------------------------------
	// Goroutines 1..4: Concurrent Synchronous Inference on All Pods
	// -----------------------------------------------------------------------
	for p := 0; p < numPods; p++ {
		wg.Add(1)
		go func(podIdx int) {
			defer wg.Done()
			pod := pods[podIdx]
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(podIdx)))

			for atomic.LoadInt32(&stopSignal) == 0 {
				if atomic.LoadInt32(&pod.IsRunning) == 0 {
					time.Sleep(10 * time.Millisecond)
					continue
				}

				txnID := fmt.Sprintf("txn_%s_%d", pod.PodID, r.Intn(100000))
				tenantID := fmt.Sprintf("tenant_%d", podIdx)

				reqStart := time.Now()
				route := pod.CanaryRouter.Route(tenantID, txnID)
				if route == RouteLegacy || route == RouteCandidate {
					atomic.AddInt64(&totalEvaluations, 1)
				}

				latMs := float64(time.Since(reqStart).Nanoseconds()) / 1e6
				pod.Metrics.RecordRequest(latMs, "ALLOW", r.Intn(100), true, false, false)
				pod.SLO.RecordEvaluation(latMs, true, false, false)
			}
		}(p)
	}

	// -----------------------------------------------------------------------
	// Goroutines 5..8: Competing Retraining Jobs with Distributed Leader Lock
	// -----------------------------------------------------------------------
	for p := 0; p < numPods; p++ {
		wg.Add(1)
		go func(podIdx int) {
			defer wg.Done()
			pod := pods[podIdx]
			ticker := time.NewTicker(20 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if atomic.LoadInt32(&stopSignal) == 1 {
						return
					}
					if atomic.LoadInt32(&pod.IsRunning) == 0 {
						continue
					}

					// Attempt to acquire distributed retraining leader lock
					lease, err := pod.LockManager.Acquire(ctx, "leader:retraining", pod.PodID, 30*time.Millisecond)
					if err == nil {
						atomic.AddInt64(&leaderLockAcquisitions, 1)
						atomic.AddInt64(&totalRetrainsTriggered, 1)

						// Simulate leader running drift check and retraining trigger
						m := &DriftMeasurement{
							MeasurementID: fmt.Sprintf("meas_%s", pod.PodID),
							Timestamp:     time.Now().UTC(),
							MaxPSI:        0.28,
							MaxJSD:        0.11,
							OverallStatus: DriftStatusCritical,
							ModelVersion:  "fraud-xgb-25f-v3.0",
						}
						pod.Coordinator.OnDriftEvaluated(ctx, m)

						time.Sleep(5 * time.Millisecond)
						_ = pod.LockManager.Release(ctx, lease)
					} else {
						atomic.AddInt64(&leaderLockFailures, 1)
					}
				case <-ctx.Done():
					return
				}
			}
		}(p)
	}

	// -----------------------------------------------------------------------
	// Goroutine 9: Kubernetes Rolling Restart Simulator (Killing & Resurrecting Pods)
	// -----------------------------------------------------------------------
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		targetPod := 0

		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopSignal) == 1 {
					return
				}
				pod := pods[targetPod%numPods]

				// Simulate SIGTERM / pod termination
				atomic.StoreInt32(&pod.IsRunning, 0)
				time.Sleep(15 * time.Millisecond)

				// Resurrect pod (Kubernetes replica replacement)
				atomic.StoreInt32(&pod.IsRunning, 1)
				atomic.AddInt64(&podRestarts, 1)
				targetPod++
			case <-ctx.Done():
				return
			}
		}
	}()

	// Run multi-replica test for 2 seconds
	time.Sleep(2000 * time.Millisecond)
	atomic.StoreInt32(&stopSignal, 1)
	wg.Wait()

	// -----------------------------------------------------------------------
	// Invariant Verification Across All Replicas
	// -----------------------------------------------------------------------
	evals := atomic.LoadInt64(&totalEvaluations)
	acquisitions := atomic.LoadInt64(&leaderLockAcquisitions)
	restarts := atomic.LoadInt64(&podRestarts)

	assert.Greater(t, evals, int64(10000), "Evaluations across pods must exceed 10,000 requests")
	assert.Greater(t, acquisitions, int64(10), "Leader locks must be acquired successfully across pods")
	assert.Greater(t, restarts, int64(5), "Simulated pod rolling restarts must have occurred")

	// Ensure no split brain: all active pods must report the identical production baseline
	for _, pod := range pods {
		prod, err := pod.Registry.GetProductionModel()
		require.NoError(t, err)
		assert.Equal(t, "fraud-xgb-25f-v3.0", prod.Version)
	}

	t.Logf("[MULTI-REPLICA TEST PASSED] TotalEvaluations: %d | LeaderLocksAcquired: %d | PodRestarts: %d",
		evals, acquisitions, restarts)

	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	finalGoroutines := runtime.NumGoroutine()
	assert.LessOrEqual(t, finalGoroutines-initialGoroutines, 3, "Goroutine leak check: delta must be <= 3")
}
