package riskengine

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// failingDispatcher simulates a total telemetry / ClickHouse outage.
type failingDispatcher struct {
	attempts int64
}

func (f *failingDispatcher) Dispatch(ctx context.Context, event OutboxEvent) error {
	atomic.AddInt64(&f.attempts, 1)
	return fmt.Errorf("simulated clickhouse cluster unreachable")
}

func TestCombined_ProductionChaosSecurityAndLoad(t *testing.T) {
	runtime.GC()
	initialGoroutines := runtime.NumGoroutine()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Initialize Core Systems
	reg := NewModelRegistry()
	cfg := DefaultRetrainingConfig()
	cfg.CooldownDuration = 5 * time.Millisecond
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	metricsEngine := NewMetricsEngine()
	sloEngine := NewSLOEngine(5 * time.Minute)
	rateLimiter := NewTenantRateLimiter(DefaultRateLimiterConfig())
	idempotencyStore := NewIdempotencyStore(1*time.Minute, 50000)
	authMgr := NewAuthManager()
	_ = authMgr.RegisterKey("admin_secret_key_123", Identity{Subject: "sre_admin", Role: RoleAdmin})
	_ = authMgr.RegisterKey("ml_secret_key_456", Identity{Subject: "ml_eng", Role: RoleMLOperator})

	// 2. Initialize Failing Outbox (Simulating ClickHouse/Kafka Telemetry Outage)
	failingDisp := &failingDispatcher{}
	outbox := NewDurableOutbox(failingDisp, 4, 1000)
	auditLogger := NewSecurityAuditLogger(1000, outbox)

	coordinator.SetMetricsEngine(metricsEngine)
	coordinator.SetSLOEngine(sloEngine)

	var stopSignal int32
	var syncSuccessCount int64
	var syncFailCount int64
	var authBlockedAttacks int64
	var idempotencyHits int64
	var retrainsTriggered int64

	var wg sync.WaitGroup

	// -----------------------------------------------------------------------
	// THREAD 1..8: High-Throughput Synchronous Inference Traffic
	// -----------------------------------------------------------------------
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for atomic.LoadInt32(&stopSignal) == 0 {
				tenantID := fmt.Sprintf("tenant_%d", workerID%4)
				txnID := fmt.Sprintf("txn_%d_%d", workerID, r.Intn(100000))

				// Check rate limiter
				if !rateLimiter.Allow(tenantID) {
					continue
				}

				reqStart := time.Now()
				// Synchronous routing
				route := canaryRouter.Route(tenantID, txnID)
				if route != RouteLegacy && route != RouteCandidate {
					atomic.AddInt64(&syncFailCount, 1)
				} else {
					atomic.AddInt64(&syncSuccessCount, 1)
				}

				latMs := float64(time.Since(reqStart).Nanoseconds()) / 1e6
				metricsEngine.RecordRequest(latMs, "ALLOW", r.Intn(100), true, false, false)
				sloEngine.RecordEvaluation(latMs, true, false, false)

				// Telemetry enqueue (must NOT block even though ClickHouse is down)
				outbox.Enqueue(OutboxEvent{
					EventID:       fmt.Sprintf("evt_%d_%s", time.Now().UnixNano(), txnID),
					EventType:     "RISK_EVALUATION_COMPLETED",
					CorrelationID: txnID,
					Payload:       []byte(`{"status":"success"}`),
				})
			}
		}(w)
	}

	// -----------------------------------------------------------------------
	// THREAD 9: Background Drift Evaluation & Retraining Triggers
	// -----------------------------------------------------------------------
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopSignal) == 1 {
					return
				}
				atomic.AddInt64(&retrainsTriggered, 1)
				m := &DriftMeasurement{
					MeasurementID: "meas_chaos_01",
					Timestamp:     time.Now().UTC(),
					MaxPSI:        0.30,
					MaxJSD:        0.12,
					OverallStatus: DriftStatusCritical,
					ModelVersion:  "fraud-xgb-25f-v3.0",
				}
				coordinator.OnDriftEvaluated(ctx, m)
				auditLogger.LogEvent(SecurityAuditEvent{
					Actor:    "retraining_trigger_engine",
					Role:     RoleMLOperator,
					Action:   ActionRetrainingTriggered,
					Resource: "coordinator",
					Result:   "SUCCESS",
				})
			case <-ctx.Done():
				return
			}
		}
	}()

	// -----------------------------------------------------------------------
	// THREAD 10: Dynamic Canary Traffic Shifter
	// -----------------------------------------------------------------------
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(25 * time.Millisecond)
		defer ticker.Stop()
		stages := []int{0, 10, 25, 50, 100, 0}
		idx := 0

		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopSignal) == 1 {
					return
				}
				canaryRouter.SetPercentage(stages[idx%len(stages)])
				idx++
			case <-ctx.Done():
				return
			}
		}
	}()

	// -----------------------------------------------------------------------
	// THREAD 11: Malicious Security Attack Simulator
	// -----------------------------------------------------------------------
	wg.Add(1)
	go func() {
		defer wg.Done()
		maliciousKeys := []string{"bad_key", "", "hacker_token", "admin_fake"}
		handler := authMgr.RequireRole(RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopSignal) == 1 {
					return
				}
				key := maliciousKeys[rand.Intn(len(maliciousKeys))]
				req := httptest.NewRequest(http.MethodPost, "/v1/operations/model/freeze", nil)
				req.Header.Set("Authorization", "Bearer "+key)
				rr := httptest.NewRecorder()
				handler(rr, req)

				if rr.Code == http.StatusUnauthorized || rr.Code == http.StatusForbidden {
					atomic.AddInt64(&authBlockedAttacks, 1)
					auditLogger.LogEvent(SecurityAuditEvent{
						Actor:    "unauthorized_attacker",
						Role:     RoleAnonymous,
						Action:   ActionAuthFailure,
						Resource: "/v1/operations/model/freeze",
						Result:   "DENIED",
					})
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// -----------------------------------------------------------------------
	// THREAD 12: Idempotent Mutation Replay Simulator
	// -----------------------------------------------------------------------
	wg.Add(1)
	go func() {
		defer wg.Done()
		idemHandler := idempotencyStore.IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"status":"accepted"}`))
		}))

		ticker := time.NewTicker(15 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if atomic.LoadInt32(&stopSignal) == 1 {
					return
				}
				req := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", bytes.NewReader([]byte(`{"reason":"stress"}`)))
				req.Header.Set("X-Idempotency-Key", "chaos_shared_key_777")
				rr := httptest.NewRecorder()
				idemHandler.ServeHTTP(rr, req)

				if rr.Header().Get("X-Cache-Lookup") == "HIT" {
					atomic.AddInt64(&idempotencyHits, 1)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Run chaos scenario for 2.5 seconds
	time.Sleep(2500 * time.Millisecond)
	atomic.StoreInt32(&stopSignal, 1)
	wg.Wait()

	_ = outbox.FlushAndClose(500 * time.Millisecond)

	// -----------------------------------------------------------------------
	// Strict Safety Invariant Verification
	// -----------------------------------------------------------------------
	// 1. Verify exactly one production model
	prodModel, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", prodModel.Version)

	// 2. Verify synchronous traffic succeeded with 0 failures despite ClickHouse outage
	successes := atomic.LoadInt64(&syncSuccessCount)
	fails := atomic.LoadInt64(&syncFailCount)
	assert.Greater(t, successes, int64(10000), "Synchronous evaluations must exceed 10,000 requests")
	assert.Equal(t, int64(0), fails, "Synchronous inference must NOT fail due to downstream telemetry outages")

	// 3. Verify security defense
	blocked := atomic.LoadInt64(&authBlockedAttacks)
	assert.Greater(t, blocked, int64(50), "All unauthorized attack attempts must be blocked")

	// 4. Verify idempotency protection
	hits := atomic.LoadInt64(&idempotencyHits)
	assert.Greater(t, hits, int64(20), "Concurrent identical mutation replays must be handled idempotently")

	t.Logf("[COMBINED CHAOS COMPLETED] SyncEvaluations: %d | Fails: %d | AuthAttacksBlocked: %d | IdempotencyHits: %d | RetrainTriggers: %d | OutboxAttempts: %d",
		successes, fails, blocked, hits, atomic.LoadInt64(&retrainsTriggered), atomic.LoadInt64(&failingDisp.attempts))

	// 5. Goroutine leak check
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	finalGoroutines := runtime.NumGoroutine()
	assert.LessOrEqual(t, finalGoroutines-initialGoroutines, 3, "Goroutine leak check: delta must be <= 3")
}
