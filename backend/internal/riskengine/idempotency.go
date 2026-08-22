package riskengine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// IdempotencyRecord stores cached execution outcomes for a specific idempotency key.
type IdempotencyRecord struct {
	RequestHash        string
	ResponseStatusCode int
	ResponseBody       []byte
	ResponseHeaders    http.Header
	CreatedAt          time.Time
	ExpiresAt          time.Time
	InFlight           bool
	mu                 sync.Mutex
}

// IdempotencyStore manages thread-safe idempotent request caching and collision detection.
type IdempotencyStore struct {
	mu        sync.RWMutex
	records   map[string]*IdempotencyRecord
	ttl       time.Duration
	maxSize   int
	hits      int64
	misses    int64
	conflicts int64
}

// NewIdempotencyStore initializes an in-memory IdempotencyStore with bounded retention.
func NewIdempotencyStore(ttl time.Duration, maxSize int) *IdempotencyStore {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	if maxSize <= 0 {
		maxSize = 50000
	}
	return &IdempotencyStore{
		records: make(map[string]*IdempotencyRecord),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// ComputeRequestHash generates a canonical SHA-256 fingerprint from the HTTP method, path, and body.
func ComputeRequestHash(method, path string, body []byte) string {
	hasher := sha256.New()
	hasher.Write([]byte(method))
	hasher.Write([]byte(":"))
	hasher.Write([]byte(path))
	hasher.Write([]byte(":"))
	hasher.Write(body)
	return hex.EncodeToString(hasher.Sum(nil))
}

// Stats returns current idempotency operational metrics.
func (s *IdempotencyStore) Stats() (hits, misses, conflicts int64) {
	return atomic.LoadInt64(&s.hits), atomic.LoadInt64(&s.misses), atomic.LoadInt64(&s.conflicts)
}

// responseRecorder captures the response status code and body for idempotency caching.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	bodyBuf    bytes.Buffer
}

func (rw *responseRecorder) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	rw.bodyBuf.Write(b)
	return rw.ResponseWriter.Write(b)
}

// IdempotencyMiddleware intercepts mutation requests carrying X-Idempotency-Key.
func (s *IdempotencyStore) IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only enforce idempotency for mutation methods (POST, PUT, PATCH, DELETE)
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch && r.Method != http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("X-Idempotency-Key")
		if key == "" {
			// No key provided -> execute normally without idempotency caching
			next.ServeHTTP(w, r)
			return
		}

		// Read and buffer the request body
		var bodyBytes []byte
		if r.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(r.Body)
			if err != nil {
				writeJSONAuthError(w, http.StatusBadRequest, "bad_request", "Failed to read request body for idempotency check")
				return
			}
			// Restore request body for downstream handlers
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		reqHash := ComputeRequestHash(r.Method, r.URL.Path, bodyBytes)
		now := time.Now().UTC()

		s.mu.Lock()
		// Clean up store if maxSize exceeded
		if len(s.records) >= s.maxSize {
			s.evictExpiredLocked(now)
		}

		rec, exists := s.records[key]
		if exists && now.After(rec.ExpiresAt) {
			delete(s.records, key)
			exists = false
		}

		if exists {
			// Key exists: check for payload conflict
			if rec.RequestHash != reqHash {
				s.mu.Unlock()
				atomic.AddInt64(&s.conflicts, 1)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"type":   "https://errors.ropus.io/idempotency_conflict",
					"title":  "Idempotency Conflict",
					"status": http.StatusConflict,
					"detail": "Idempotency key was previously used with a different request payload or endpoint",
					"code":   "idempotency_conflict",
				})
				return
			}

			// Key exists with identical payload
			rec.mu.Lock()
			s.mu.Unlock()
			defer rec.mu.Unlock()

			atomic.AddInt64(&s.hits, 1)
			w.Header().Set("X-Cache-Lookup", "HIT")
			w.Header().Set("X-Idempotency-Replayed", "true")
			for k, v := range rec.ResponseHeaders {
				w.Header()[k] = v
			}
			w.WriteHeader(rec.ResponseStatusCode)
			_, _ = w.Write(rec.ResponseBody)
			return
		}

		// First time seeing this key: create in-flight record
		atomic.AddInt64(&s.misses, 1)
		rec = &IdempotencyRecord{
			RequestHash: reqHash,
			CreatedAt:   now,
			ExpiresAt:   now.Add(s.ttl),
			InFlight:    true,
		}
		rec.mu.Lock()
		s.records[key] = rec
		s.mu.Unlock()

		// Execute downstream handler with response capture
		recWriter := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recWriter, r)

		// Cache final response outcome
		rec.ResponseStatusCode = recWriter.statusCode
		rec.ResponseBody = recWriter.bodyBuf.Bytes()
		rec.ResponseHeaders = recWriter.Header().Clone()
		rec.InFlight = false
		rec.mu.Unlock()
	})
}

// evictExpiredLocked removes expired records to enforce memory bounds.
func (s *IdempotencyStore) evictExpiredLocked(now time.Time) {
	for k, v := range s.records {
		if now.After(v.ExpiresAt) {
			delete(s.records, k)
		}
	}
	// If still full, drop a small fraction of oldest
	if len(s.records) >= s.maxSize {
		count := 0
		for k := range s.records {
			delete(s.records, k)
			count++
			if count > s.maxSize/10 {
				break
			}
		}
	}
}
