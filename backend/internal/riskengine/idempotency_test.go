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
