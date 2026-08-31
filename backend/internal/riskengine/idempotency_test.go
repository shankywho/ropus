package riskengine

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIdempotency_SingleRequestAndReplay(t *testing.T) {
	store := NewIdempotencyStore(1*time.Minute, 1000)

	var executionCount int64
	handler := store.IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&executionCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"retraining_triggered","job_id":"job_123"}`))
	}))

	payload := []byte(`{"reason":"drift_detected"}`)

	// 1. Initial Request (Cache Miss)
	req1 := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", bytes.NewReader(payload))
	req1.Header.Set("X-Idempotency-Key", "idem_key_001")
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	assert.Equal(t, http.StatusAccepted, rr1.Code)
	assert.Contains(t, rr1.Body.String(), "job_123")
	assert.Equal(t, int64(1), atomic.LoadInt64(&executionCount))

	// 2. Exact Duplicate Request (Cache Hit)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", bytes.NewReader(payload))
	req2.Header.Set("X-Idempotency-Key", "idem_key_001")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	assert.Equal(t, http.StatusAccepted, rr2.Code)
	assert.Contains(t, rr2.Body.String(), "job_123")
	assert.Equal(t, "HIT", rr2.Header().Get("X-Cache-Lookup"))
	assert.Equal(t, "true", rr2.Header().Get("X-Idempotency-Replayed"))
	assert.Equal(t, int64(1), atomic.LoadInt64(&executionCount), "Underlying mutation MUST NOT execute a second time")

	// 3. Same Key with Different Payload (Conflict -> 409)
	diffPayload := []byte(`{"reason":"manual_trigger_by_sre"}`)
	req3 := httptest.NewRequest(http.MethodPost, "/v1/retraining/trigger", bytes.NewReader(diffPayload))
	req3.Header.Set("X-Idempotency-Key", "idem_key_001")
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	assert.Equal(t, http.StatusConflict, rr3.Code)
	assert.Contains(t, rr3.Body.String(), "idempotency_conflict")
	assert.Equal(t, int64(1), atomic.LoadInt64(&executionCount))

	hits, misses, conflicts := store.Stats()
	assert.Equal(t, int64(1), hits)
	assert.Equal(t, int64(1), misses)
	assert.Equal(t, int64(1), conflicts)
}

func TestIdempotency_ConcurrentDuplicateRequests(t *testing.T) {
	store := NewIdempotencyStore(1*time.Minute, 1000)

	var executionCount int64
	handler := store.IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate mutation processing delay
		atomic.AddInt64(&executionCount, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"success"}`))
	}))

	concurrency := 20
	var wg sync.WaitGroup
	results := make([]int, concurrency)

	payload := []byte(`{"percentage":50}`)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/v1/canary/control", bytes.NewReader(payload))
			req.Header.Set("X-Idempotency-Key", "shared_concurrent_key_999")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			results[idx] = rr.Code
		}(i)
	}

	wg.Wait()

	for _, code := range results {
		assert.Equal(t, http.StatusOK, code)
	}
	assert.Equal(t, int64(1), atomic.LoadInt64(&executionCount), "Mutation must execute exactly once even under heavy concurrency")
}

func TestIdempotency_TTLExpiration(t *testing.T) {
	store := NewIdempotencyStore(10*time.Millisecond, 1000)

	var executionCount int64
	handler := store.IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&executionCount, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"execution":%d}`, atomic.LoadInt64(&executionCount))))
	}))

	payload := []byte(`{"action":"test"}`)

	// First execution
	req1 := httptest.NewRequest(http.MethodPost, "/v1/test", bytes.NewReader(payload))
	req1.Header.Set("X-Idempotency-Key", "expiring_key")
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, int64(1), atomic.LoadInt64(&executionCount))

	// Wait for TTL expiration
	time.Sleep(20 * time.Millisecond)

	// Re-execution after expiration
	req2 := httptest.NewRequest(http.MethodPost, "/v1/test", bytes.NewReader(payload))
	req2.Header.Set("X-Idempotency-Key", "expiring_key")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, int64(2), atomic.LoadInt64(&executionCount), "Expired idempotency record should allow new execution")
}

func TestIdempotency_TenantIsolation(t *testing.T) {
	store := NewIdempotencyStore(1*time.Minute, 1000)

	var executionCount int64
	handler := store.IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&executionCount, 1)
		tenant := r.Header.Get("X-Tenant-ID")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"tenant":"%s"}`, tenant)))
	}))

	payload := []byte(`{"action":"evaluate"}`)
	sharedKey := "shared_key_between_tenants"

	// 1. Tenant Alpha Request
	reqA := httptest.NewRequest(http.MethodPost, "/v1/risk/evaluate", bytes.NewReader(payload))
	reqA.Header.Set("X-Tenant-ID", "tenant-alpha-001")
	reqA.Header.Set("X-Idempotency-Key", sharedKey)
	rrA := httptest.NewRecorder()
	handler.ServeHTTP(rrA, reqA)
	assert.Equal(t, http.StatusOK, rrA.Code)
	assert.Contains(t, rrA.Body.String(), "tenant-alpha-001")
	assert.Equal(t, int64(1), atomic.LoadInt64(&executionCount))

	// 2. Tenant Beta Request with SAME idempotency key (MUST NOT collision or replay Tenant Alpha's data)
	reqB := httptest.NewRequest(http.MethodPost, "/v1/risk/evaluate", bytes.NewReader(payload))
	reqB.Header.Set("X-Tenant-ID", "tenant-beta-002")
	reqB.Header.Set("X-Idempotency-Key", sharedKey)
	rrB := httptest.NewRecorder()
	handler.ServeHTTP(rrB, reqB)
	assert.Equal(t, http.StatusOK, rrB.Code)
	assert.Contains(t, rrB.Body.String(), "tenant-beta-002")
	assert.Equal(t, int64(2), atomic.LoadInt64(&executionCount), "Different tenants MUST NOT share or collide idempotency cache")
}

func TestIdempotency_HighConcurrencyCoalescing(t *testing.T) {
	store := NewIdempotencyStore(1*time.Minute, 1000)

	var executionCount int64
	handler := store.IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Millisecond) // Simulate DB transaction + ML call
		atomic.AddInt64(&executionCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"decision_id":"dec_concurrent_123","action":"ALLOW"}`))
	}))

	concurrency := 50
	var wg sync.WaitGroup
	statusCodes := make([]int, concurrency)
	cacheLookups := make([]string, concurrency)

	payload := []byte(`{"transaction_id":"txn_shared_50","amount":5000}`)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(payload))
			req.Header.Set("X-Tenant-ID", "tenant_prod_1")
			req.Header.Set("X-Idempotency-Key", "idemp_high_concurrency_key")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			statusCodes[idx] = rr.Code
			cacheLookups[idx] = rr.Header().Get("X-Cache-Lookup")
		}(i)
	}

	wg.Wait()

	for _, code := range statusCodes {
		assert.Equal(t, http.StatusOK, code)
	}
	assert.Equal(t, int64(1), atomic.LoadInt64(&executionCount), "50 concurrent requests with identical key must execute downstream only once")

	hits, misses, conflicts := store.Stats()
	assert.Equal(t, int64(49), hits)
	assert.Equal(t, int64(1), misses)
	assert.Equal(t, int64(0), conflicts)
}

func TestIdempotency_GETBypass(t *testing.T) {
	store := NewIdempotencyStore(1*time.Minute, 1000)

	var executionCount int64
	handler := store.IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&executionCount, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	}))

	// GET requests should bypass idempotency middleware completely
	req := httptest.NewRequest(http.MethodGet, "/v1/canary/status", nil)
	req.Header.Set("X-Idempotency-Key", "get_key_ignored")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, int64(1), atomic.LoadInt64(&executionCount))
	assert.Equal(t, "", rr.Header().Get("X-Cache-Lookup"))

	hits, misses, conflicts := store.Stats()
	assert.Equal(t, int64(0), hits)
	assert.Equal(t, int64(0), misses)
	assert.Equal(t, int64(0), conflicts)
}

func TestIdempotency_TimeoutRetryDeterministicSemanticDecision(t *testing.T) {
	store := NewIdempotencyStore(1*time.Minute, 1000)

	var decisionExecutions int64
	handler := store.IdempotencyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&decisionExecutions, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"decision_id":"dec_retry_999","recommended_action":"ALLOW_RECOMMENDATION","risk_score":12}`))
	}))

	payload := []byte(`{"transaction_id":"txn_timeout_retry_01","amount":2500,"currency":"INR"}`)
	idempKey := "idemp_timeout_retry_key_01"
	tenantID := "tenant_merchant_test"

	// 1. Initial Request: Server computes and stores decision
	req1 := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(payload))
	req1.Header.Set("X-Tenant-ID", tenantID)
	req1.Header.Set("X-Idempotency-Key", idempKey)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	assert.Equal(t, http.StatusOK, rr1.Code)
	assert.Contains(t, rr1.Body.String(), "dec_retry_999")
	assert.Contains(t, rr1.Body.String(), "ALLOW_RECOMMENDATION")
	assert.Equal(t, int64(1), atomic.LoadInt64(&decisionExecutions))

	// 2. Client-side timeout occurs, client retries with SAME key, SAME tenant, SAME payload
	req2 := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(payload))
	req2.Header.Set("X-Tenant-ID", tenantID)
	req2.Header.Set("X-Idempotency-Key", idempKey)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	// Verifies exact semantic reproduction without re-executing pipeline
	assert.Equal(t, http.StatusOK, rr2.Code)
	assert.Equal(t, rr1.Body.String(), rr2.Body.String(), "Replayed response must match original semantic decision byte-for-byte")
	assert.Equal(t, "HIT", rr2.Header().Get("X-Cache-Lookup"))
	assert.Equal(t, "true", rr2.Header().Get("X-Idempotency-Replayed"))
	assert.Equal(t, int64(1), atomic.LoadInt64(&decisionExecutions), "Underlying decision pipeline must NOT execute again")

	// 3. Replaying SAME key with altered payload returns 409 Conflict
	tamperedPayload := []byte(`{"transaction_id":"txn_timeout_retry_01","amount":999999,"currency":"INR"}`)
	req3 := httptest.NewRequest(http.MethodPost, "/v1/risk-evaluations", bytes.NewReader(tamperedPayload))
	req3.Header.Set("X-Tenant-ID", tenantID)
	req3.Header.Set("X-Idempotency-Key", idempKey)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	assert.Equal(t, http.StatusConflict, rr3.Code)
	assert.Contains(t, rr3.Body.String(), "idempotency_conflict")
	assert.Equal(t, int64(1), atomic.LoadInt64(&decisionExecutions))
}
