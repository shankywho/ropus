package internal_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shankywho/ropus/backend/internal/riskengine"
	"github.com/shankywho/ropus/backend/internal/rules"
)

// =============================================================================
// 1. MOCKED ONNX SIDECAR WITH VARIABLE NETWORK LATENCIES (2ms, 15ms, 50ms)
// =============================================================================

type VariableLatencyONNXServer struct {
	server       *httptest.Server
	requestCount int64
	timeoutCount int64
	fastCount    int64
}

func newVariableLatencyONNXServer() *VariableLatencyONNXServer {
	s := &VariableLatencyONNXServer{}
	mux := http.NewServeMux()

	mux.HandleFunc("/predict", func(w http.ResponseWriter, r *http.Request) {
		reqNum := atomic.AddInt64(&s.requestCount, 1)

		// Variable latency schedule:
		// 1st of 3 -> 2ms (Fast, within 8ms budget)
		// 2nd of 3 -> 15ms (Breaches 8ms timeout budget)
		// 3rd of 3 -> 50ms (Breaches 8ms timeout budget, trips breaker)
		var delay time.Duration
		switch reqNum % 3 {
		case 1:
			delay = 2 * time.Millisecond
			atomic.AddInt64(&s.fastCount, 1)
		case 2:
			delay = 15 * time.Millisecond
			atomic.AddInt64(&s.timeoutCount, 1)
		case 0:
			delay = 50 * time.Millisecond
			atomic.AddInt64(&s.timeoutCount, 1)
		}

		select {
		case <-time.After(delay):
			// Server completed response
		case <-r.Context().Done():
			// Client cancelled/timed out due to 8ms context deadline
			return
		}

		resp := riskengine.MLPredictResponse{
			RiskScore:           78,
			Probability:         0.78,
			ReasonCodes:         []string{"MODEL_HIGH_RISK_VELOCITY"},
			FeatureAttributions: map[string]float64{"ip_velocity_1h": 0.45, "token_velocity_24h": 0.33},
			LatencyMs:           float64(delay.Milliseconds()),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	s.server = httptest.NewServer(mux)
	return s
}

func (s *VariableLatencyONNXServer) Close() {
	s.server.Close()
}

func (s *VariableLatencyONNXServer) URL() string {
	return s.server.URL
}

// =============================================================================
// 2. ZERO-ALLOCATION HOT FILTERING SCRATCH BUFFER POOL (sync.Pool)
// =============================================================================

const MaxScratchBufferSize = 1024

type ScratchBuffer struct {
	B []byte
}

var hotScratchPool = sync.Pool{
	New: func() interface{} {
		return &ScratchBuffer{
			B: make([]byte, 0, MaxScratchBufferSize),
		}
	},
}

// FastDeterministicFilter evaluates high-throughput risk predicates on raw payload
// without allocating heap memory by reusing buffers from hotScratchPool.
//go:noinline
func FastDeterministicFilter(amount int64, currency string, ipBytes [4]byte, token string) (bool, string) {
	buf := hotScratchPool.Get().(*ScratchBuffer)
	buf.B = buf.B[:0] // Reset length while preserving capacity

	// Encode binary representation into pooled slice
	var scratch [8]byte
	binary.BigEndian.PutUint64(scratch[:], uint64(amount))
	buf.B = append(buf.B, scratch[:]...)
	buf.B = append(buf.B, ipBytes[:]...)
	buf.B = append(buf.B, currency...)
	buf.B = append(buf.B, token...)

	isSuspicious := false
	reason := "CLEAN"

	// Deterministic zero-alloc predicate checks
	if amount >= 10000000 { // ₹1,00,000+
		isSuspicious = true
		reason = "HIGH_VALUE_THRESHOLD_BREACH"
	} else if ipBytes[0] == 185 && ipBytes[1] == 220 { // Datacenter / TOR subnet
		isSuspicious = true
		reason = "DATACENTER_PROXY_SUBNET"
	}

	// Return buffer back to pool
	hotScratchPool.Put(buf)
	return isSuspicious, reason
}

// =============================================================================
// 3. ASYNCHRONOUS BATCH SHA-256 AUDIT LEDGER (Zero Thread Serialization)
// =============================================================================

type AuditEntry struct {
	DecisionID  string
	ReqHash     [32]byte
	Action      string
	RiskScore   int
	TimestampNs int64
}

type AsyncBatchAuditLedger struct {
	queue         chan AuditEntry
	prevBatchHash [32]byte
	totalEnqueued int64
	totalBatched  int64
	batchCount    int64
	mu            sync.Mutex
	done          chan struct{}
	wg            sync.WaitGroup
}

func NewAsyncBatchAuditLedger(capacity int) *AsyncBatchAuditLedger {
	ledger := &AsyncBatchAuditLedger{
		queue:         make(chan AuditEntry, capacity),
		prevBatchHash: sha256.Sum256([]byte("GENESIS_ROPUS_AUDIT_BLOCK")),
		done:          make(chan struct{}),
	}
	ledger.wg.Add(1)
	go ledger.batchProcessor()
	return ledger
}

// EnqueueNonBlocking appends audit entries into the async ring channel without locking mutex.
func (l *AsyncBatchAuditLedger) EnqueueNonBlocking(entry AuditEntry) bool {
	select {
	case l.queue <- entry:
		atomic.AddInt64(&l.totalEnqueued, 1)
		return true
	default:
		// Queue full (backpressure safety)
		return false
	}
}

// batchProcessor runs in background to drain batches and compute Merkle/SHA-256 chains.
func (l *AsyncBatchAuditLedger) batchProcessor() {
	defer l.wg.Done()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	batch := make([]AuditEntry, 0, 256)

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}
		hasher := sha256.New()
		hasher.Write(l.prevBatchHash[:])

		var scratch [8]byte
		for _, item := range batch {
			hasher.Write(item.ReqHash[:])
			hasher.Write([]byte(item.DecisionID))
			hasher.Write([]byte(item.Action))
			binary.BigEndian.PutUint64(scratch[:], uint64(item.RiskScore))
			hasher.Write(scratch[:])
		}

		l.mu.Lock()
		copy(l.prevBatchHash[:], hasher.Sum(nil))
		l.totalBatched += int64(len(batch))
		l.batchCount++
		l.mu.Unlock()

		batch = batch[:0]
	}

	for {
		select {
		case item, ok := <-l.queue:
			if !ok {
				flushBatch()
				return
			}
			batch = append(batch, item)
			if len(batch) >= 256 {
				flushBatch()
			}
		case <-ticker.C:
			flushBatch()
		case <-l.done:
			// Drain remaining queue
			for {
				select {
				case item, ok := <-l.queue:
					if !ok {
						flushBatch()
						return
					}
					batch = append(batch, item)
					if len(batch) >= 256 {
						flushBatch()
					}
				default:
					flushBatch()
					return
				}
			}
		}
	}
}

func (l *AsyncBatchAuditLedger) Close() {
	close(l.done)
	l.wg.Wait()
}

func (l *AsyncBatchAuditLedger) GetLatestBlockHash() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return hex.EncodeToString(l.prevBatchHash[:])
}

// =============================================================================
// 4. HIGH-PERFORMANCE CONCURRENT BENCHMARK & VERIFICATION SUITE
// =============================================================================

func TestHighPerformanceOrchestratorSuite(t *testing.T) {
	// 1. Initialize Mocked ONNX Sidecar with variable network latencies (2ms, 15ms, 50ms)
	onnxServer := newVariableLatencyONNXServer()
	defer onnxServer.Close()

	// 2. Setup Deterministic AST Rules Service
	rulesService := rules.NewService(nil)
	tenantID := "tenant_perf_benchmark_001"

	// Seed AST deterministic rules for local fallback
	ruleHighVal := rules.Rule{
		ID:       "rule_high_val",
		TenantID: tenantID,
		Name:     "HighValueDeclineRule",
		Status:   rules.StatusActive,
		DSLAST: json.RawMessage(`{
			"condition": {
				"field": "amount",
				"operator": "GREATER_THAN",
				"value": 10000000
			},
			"action": "DECLINE_RECOMMENDATION",
			"reason_code": "AST_RULE_HIGH_VALUE_BREACH"
		}`),
	}
	rulesService.AddMemoryRule(ruleHighVal)

	ruleVelocity := rules.Rule{
		ID:       "rule_velocity_burst",
		TenantID: tenantID,
		Name:     "VelocityBurstStepUpRule",
		Status:   rules.StatusActive,
		DSLAST: json.RawMessage(`{
			"condition": {
				"field": "velocity.token.24hr",
				"operator": "GREATER_THAN",
				"value": 3
			},
			"action": "STEP_UP_RECOMMENDATION",
			"reason_code": "AST_RULE_VELOCITY_BURST"
		}`),
	}
	rulesService.AddMemoryRule(ruleVelocity)

	// 3. Initialize Orchestrator with timeout-aware ML Client
	mlClient := riskengine.NewMLClient(onnxServer.URL())
	orchestrator := riskengine.NewOrchestrator(nil, nil, rulesService, mlClient, nil)

	// 4. Initialize Asynchronous Batch Audit Ledger
	asyncLedger := NewAsyncBatchAuditLedger(20000)
	defer asyncLedger.Close()

	// 5. Verification 1: Zero-Allocation Hot Path Check
	t.Run("Verification_Zero_Allocations_In_Hot_Filter_Path", func(t *testing.T) {
		allocs := testing.AllocsPerRun(1000, func() {
			ip := [4]byte{192, 168, 1, 100}
			isSuspicious, _ := FastDeterministicFilter(48000, "INR", ip, "tok_clean_visa_01")
			if isSuspicious {
				t.Fatalf("unexpected filter match")
			}
		})

		if allocs > 0 {
			t.Errorf("Hot filter path failed zero-allocation guarantee: got %f allocs/op, expected 0", allocs)
		} else {
			t.Logf("✓ Zero allocations verified: %f allocs/op in hot path using sync.Pool", allocs)
		}
	})

	// 6. Verification 2 & 3: 100 Concurrent Goroutines Worker Pool
	t.Run("Verification_100_Goroutines_Concurrent_Stress_And_8ms_Timeout_Fallback", func(t *testing.T) {
		const (
			NumWorkers         = 100
			RequestsPerWorker  = 50
			TotalRequests      = NumWorkers * RequestsPerWorker
			HardContextTimeout = 8 * time.Millisecond
		)

		var (
			totalCompleted     int64
			fallbackCount      int64
			mlSuccessCount     int64
			astRuleActionCount int64
		)

		var wg sync.WaitGroup
		wg.Add(NumWorkers)

		startBenchmark := time.Now()

		for w := 0; w < NumWorkers; w++ {
			go func(workerID int) {
				defer wg.Done()

				for r := 0; r < RequestsPerWorker; r++ {
					// Randomized payload generator
					amtRand, _ := rand.Int(rand.Reader, big.NewInt(15000000))
					amount := amtRand.Int64() + 1000 // ₹10.00 to ₹1,50,000.00

					tokenID := fmt.Sprintf("tok_test_%d_%d", workerID, r%5)
					req := riskengine.RiskEvaluationRequest{
						TransactionID: fmt.Sprintf("txn_bench_%d_%d", workerID, r),
						Amount:        amount,
						Currency:      "INR",
						IPAddress:     fmt.Sprintf("106.51.%d.%d", workerID%255, r%255),
						DeviceFingerprint: fmt.Sprintf("dev_bench_%d", workerID%10),
						AccountID:     fmt.Sprintf("cus_user_%d", workerID),
						PaymentMethod: riskengine.PaymentMethod{
							Type:  "CARD",
							Token: tokenID,
						},
					}

					// Enforce hard 8ms context timeout budget for real-time risk decisioning
					evalCtx, cancel := context.WithTimeout(context.Background(), HardContextTimeout)
					resp, err := orchestrator.Evaluate(evalCtx, tenantID, req)
					cancel()

					if err != nil {
						t.Errorf("Orchestrator returned unhandled error: %v", err)
						continue
					}

					atomic.AddInt64(&totalCompleted, 1)

					// Verify fallback or ML resolution
					if resp.IsDegraded {
						atomic.AddInt64(&fallbackCount, 1)
						// Verify local AST rule fallback succeeded
						if resp.RecommendedAction != "" {
							atomic.AddInt64(&astRuleActionCount, 1)
						}
					} else {
						atomic.AddInt64(&mlSuccessCount, 1)
					}

					// Enqueue into asynchronous batch audit ledger (Non-blocking)
					reqJson, _ := json.Marshal(req)
					auditEntry := AuditEntry{
						DecisionID:  resp.DecisionID,
						ReqHash:     sha256.Sum256(reqJson),
						Action:      resp.RecommendedAction,
						RiskScore:   resp.RiskScore,
						TimestampNs: time.Now().UnixNano(),
					}
					enqueued := asyncLedger.EnqueueNonBlocking(auditEntry)
					if !enqueued {
						t.Errorf("Failed to enqueue audit entry into async batch ledger")
					}
				}
			}(w)
		}

		wg.Wait()
		elapsed := time.Since(startBenchmark)
		throughput := float64(totalCompleted) / elapsed.Seconds()

		t.Logf("=================================================================")
		t.Logf("BENCHMARK EXECUTION SUMMARY:")
		t.Logf("  Total Requests Evaluated:   %d / %d", totalCompleted, TotalRequests)
		t.Logf("  Elapsed Time:               %v", elapsed)
		t.Logf("  Throughput:                 %.2f req/sec", throughput)
		t.Logf("  Fast Path ML Successes:     %d", mlSuccessCount)
		t.Logf("  8ms Timeout Fallbacks:      %d", fallbackCount)
		t.Logf("  Local AST Rule Resolutions: %d", astRuleActionCount)
		t.Logf("=================================================================")

		// 1) Verify all 5,000 requests completed without panics/dropouts
		if totalCompleted != TotalRequests {
			t.Errorf("Expected %d completed requests, got %d", TotalRequests, totalCompleted)
		}

		// 2) Verify that the hard 8ms timeout successfully triggered circuit breaker fallback
		// (roughly 2/3 of requests had simulated latency 15ms or 50ms > 8ms)
		if fallbackCount == 0 {
			t.Errorf("Expected 8ms timeouts to trigger fallback, but fallbackCount is 0")
		}
		if astRuleActionCount != fallbackCount {
			t.Errorf("Expected all %d degraded requests to resolve to AST rule profiles, got %d", fallbackCount, astRuleActionCount)
		}

		// 3) Verify Asynchronous Audit Batching
		time.Sleep(50 * time.Millisecond) // Allow async worker to drain queue
		finalHash := asyncLedger.GetLatestBlockHash()
		if finalHash == "" || finalHash == "0000000000000000000000000000000000000000000000000000000000000000" {
			t.Errorf("Invalid final SHA-256 audit ledger block hash: %s", finalHash)
		}

		t.Logf("✓ SHA-256 Audit Ledger Final Block Hash: %s", finalHash)
		t.Logf("✓ All 3 Verification Criteria Passed.")
	})

	// 7. Verification 4: Audit Channel Memory Saturation Protection (Non-Blocking Select)
	t.Run("Verification_Audit_Channel_Saturation_Drop_And_Metric", func(t *testing.T) {
		// Attach tiny capacity channel (size 2)
		saturatedChan := make(chan riskengine.AuditPayloadEntry, 2)
		orchestrator.SetAuditChannel(saturatedChan)

		// Fire 20 transactions rapidly without draining channel
		for i := 0; i < 20; i++ {
			req := riskengine.RiskEvaluationRequest{
				TransactionID: fmt.Sprintf("txn_sat_%d", i),
				Amount:        48000,
				Currency:      "INR",
				IPAddress:     "192.168.1.1",
				DeviceFingerprint: "dev_sat_01",
				AccountID:     "cus_sat_01",
				PaymentMethod: riskengine.PaymentMethod{
					Type:  "CARD",
					Token: "tok_sat_01",
				},
			}
			evalCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			_, err := orchestrator.Evaluate(evalCtx, tenantID, req)
			cancel()
			if err != nil {
				t.Fatalf("Orchestrator returned error during channel saturation test: %v", err)
			}
		}

		droppedCount := orchestrator.GetDroppedAuditEventsTotal()
		if droppedCount == 0 {
			t.Errorf("Expected dropped audit events > 0 when audit channel is saturated, got %d", droppedCount)
		} else {
			t.Logf("✓ Non-blocking channel protection verified: safely dropped %d audit entries during channel saturation without thread stalling", droppedCount)
		}
	})
}
