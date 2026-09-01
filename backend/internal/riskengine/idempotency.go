package riskengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shankywho/ropus/backend/internal/tenant"
)

// StaleInFlightLeaseDuration is the duration after which an in-flight claim by a crashed pod is considered stale.
const StaleInFlightLeaseDuration = 10 * time.Second

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

type dbIdempotencyRecord struct {
	RequestHash        string
	ResponseStatusCode int
	ResponseBody       []byte
	ResponseHeaders    http.Header
	InFlight           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ExpiresAt          time.Time
}

// IdempotencyStore manages thread-safe idempotent request caching, cross-pod DB persistence, crash recovery, and collision detection.
type IdempotencyStore struct {
	mu        sync.RWMutex
	records   map[string]*IdempotencyRecord
	ttl       time.Duration
	maxSize   int
	hits      int64
	misses    int64
	conflicts int64
	db        *pgxpool.Pool
}

// NewIdempotencyStore initializes an in-memory IdempotencyStore with bounded retention.
func NewIdempotencyStore(ttl time.Duration, maxSize int) *IdempotencyStore {
	return NewIdempotencyStoreWithDB(nil, ttl, maxSize)
}

// NewIdempotencyStoreWithDB initializes a multi-pod durable IdempotencyStore backed by PostgreSQL.
func NewIdempotencyStoreWithDB(db *pgxpool.Pool, ttl time.Duration, maxSize int) *IdempotencyStore {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	if maxSize <= 0 {
		maxSize = 50000
	}
	store := &IdempotencyStore{
		records: make(map[string]*IdempotencyRecord),
		ttl:     ttl,
		maxSize: maxSize,
		db:      db,
	}
	if db != nil {
		_ = store.ensureSchema(context.Background())
	}
	return store
}

// ensureSchema creates the idempotency_records table if it does not already exist.
func (s *IdempotencyStore) ensureSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS idempotency_records (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id VARCHAR(255) NOT NULL,
			idempotency_key VARCHAR(255) NOT NULL,
			request_hash VARCHAR(64) NOT NULL,
			response_status INT,
			response_headers JSONB,
			response_body BYTEA,
			in_flight BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL,
			CONSTRAINT uq_tenant_idempotency_key UNIQUE (tenant_id, idempotency_key)
		);
		CREATE INDEX IF NOT EXISTS idx_idempotency_lookup ON idempotency_records(tenant_id, idempotency_key);
		CREATE INDEX IF NOT EXISTS idx_idempotency_expires_at ON idempotency_records(expires_at);
	`
	_, err := s.db.Exec(ctx, query)
	return err
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

func writeConflictError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"type":   "https://errors.ropus.io/idempotency_conflict",
		"title":  "Idempotency Conflict",
		"status": http.StatusConflict,
		"detail": "Idempotency key was previously used with a different request payload or endpoint",
		"code":   "idempotency_conflict",
	})
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
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "Failed to read request body for idempotency check",
				})
				return
			}
			// Restore request body for downstream handlers
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		tenantIdentity := tenant.ResolveTenantOrFallback(r)
		tenantID := tenantIdentity.TenantID
		storeKey := tenantID + ":" + key

		reqHash := ComputeRequestHash(r.Method, r.URL.Path, bodyBytes)
		now := time.Now().UTC()

		// -------------------------------------------------------------
		// LAYER 1: Pod-local in-memory fast-path (sub-millisecond replay)
		// -------------------------------------------------------------
		s.mu.Lock()
		if len(s.records) >= s.maxSize {
			s.evictExpiredLocked(now)
		}

		rec, exists := s.records[storeKey]
		if exists && now.After(rec.ExpiresAt) {
			delete(s.records, storeKey)
			exists = false
		}

		if exists {
			// Key exists in local memory
			if rec.RequestHash != reqHash {
				s.mu.Unlock()
				atomic.AddInt64(&s.conflicts, 1)
				writeConflictError(w)
				return
			}

			// Key exists with identical payload -> wait for in-flight or replay
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

		// First time seeing this key locally: create and register in-flight record under lock
		atomic.AddInt64(&s.misses, 1)
		rec = &IdempotencyRecord{
			RequestHash: reqHash,
			CreatedAt:   now,
			ExpiresAt:   now.Add(s.ttl),
			InFlight:    true,
		}
		rec.mu.Lock()
		s.records[storeKey] = rec
		s.mu.Unlock()

		// -------------------------------------------------------------
		// LAYER 2: PostgreSQL Durable Cross-Pod Idempotency & Crash Recovery
		// -------------------------------------------------------------
		if s.db != nil {
			dbRec, found, err := s.fetchDBRecord(r.Context(), tenantID, key)
			if err == nil && found && now.Before(dbRec.ExpiresAt) {
				// Record exists in PostgreSQL across pods
				if dbRec.RequestHash != reqHash {
					rec.mu.Unlock()
					s.mu.Lock()
					delete(s.records, storeKey)
					s.mu.Unlock()
					atomic.AddInt64(&s.conflicts, 1)
					writeConflictError(w)
					return
				}

				if !dbRec.InFlight {
					// Execution already completed: populate local cache and replay
					rec.ResponseStatusCode = dbRec.ResponseStatusCode
					rec.ResponseBody = dbRec.ResponseBody
					rec.ResponseHeaders = dbRec.ResponseHeaders
					rec.InFlight = false
					rec.mu.Unlock()

					atomic.AddInt64(&s.hits, 1)
					w.Header().Set("X-Cache-Lookup", "HIT")
					w.Header().Set("X-Idempotency-Replayed", "true")
					for k, v := range dbRec.ResponseHeaders {
						w.Header()[k] = v
					}
					w.WriteHeader(dbRec.ResponseStatusCode)
					_, _ = w.Write(dbRec.ResponseBody)
					return
				}

				// If in-flight is stale (>10s since last update), previous pod crashed -> reclaim lease
				if now.Sub(dbRec.UpdatedAt) >= StaleInFlightLeaseDuration {
					reclaimed, err := s.reclaimStaleDBInFlight(r.Context(), tenantID, key, reqHash)
					if err == nil && reclaimed {
						// Successfully reclaimed stale crash lease -> proceed to execute
						goto executeDownstream
					}
				}

				// Another pod is currently executing in-flight: wait for DB completion
				dbCompletedRec, ok := s.waitForDBCompletion(r.Context(), tenantID, key, reqHash)
				if ok && dbCompletedRec != nil {
					if dbCompletedRec.RequestHash != reqHash {
						rec.mu.Unlock()
						s.mu.Lock()
						delete(s.records, storeKey)
						s.mu.Unlock()
						atomic.AddInt64(&s.conflicts, 1)
						writeConflictError(w)
						return
					}
					rec.ResponseStatusCode = dbCompletedRec.ResponseStatusCode
					rec.ResponseBody = dbCompletedRec.ResponseBody
					rec.ResponseHeaders = dbCompletedRec.ResponseHeaders
					rec.InFlight = false
					rec.mu.Unlock()

					atomic.AddInt64(&s.hits, 1)
					w.Header().Set("X-Cache-Lookup", "HIT")
					w.Header().Set("X-Idempotency-Replayed", "true")
					for k, v := range dbCompletedRec.ResponseHeaders {
						w.Header()[k] = v
					}
					w.WriteHeader(dbCompletedRec.ResponseStatusCode)
					_, _ = w.Write(dbCompletedRec.ResponseBody)
					return
				}
			}

			// Attempt atomic in-flight claim in DB (First pod wins)
			claimed, err := s.claimDBInFlight(r.Context(), tenantID, key, reqHash, now.Add(s.ttl))
			if err == nil && !claimed {
				// Lost race against another pod: check if stale or wait for completion
				dbRec, found, err := s.fetchDBRecord(r.Context(), tenantID, key)
				if err == nil && found && dbRec.InFlight && now.Sub(dbRec.UpdatedAt) >= StaleInFlightLeaseDuration {
					reclaimed, _ := s.reclaimStaleDBInFlight(r.Context(), tenantID, key, reqHash)
					if reclaimed {
						goto executeDownstream
					}
				}

				dbCompletedRec, ok := s.waitForDBCompletion(r.Context(), tenantID, key, reqHash)
				if ok && dbCompletedRec != nil {
					if dbCompletedRec.RequestHash != reqHash {
						rec.mu.Unlock()
						s.mu.Lock()
						delete(s.records, storeKey)
						s.mu.Unlock()
						atomic.AddInt64(&s.conflicts, 1)
						writeConflictError(w)
						return
					}
					rec.ResponseStatusCode = dbCompletedRec.ResponseStatusCode
					rec.ResponseBody = dbCompletedRec.ResponseBody
					rec.ResponseHeaders = dbCompletedRec.ResponseHeaders
					rec.InFlight = false
					rec.mu.Unlock()

					atomic.AddInt64(&s.hits, 1)
					w.Header().Set("X-Cache-Lookup", "HIT")
					w.Header().Set("X-Idempotency-Replayed", "true")
					for k, v := range dbCompletedRec.ResponseHeaders {
						w.Header()[k] = v
					}
					w.WriteHeader(dbCompletedRec.ResponseStatusCode)
					_, _ = w.Write(dbCompletedRec.ResponseBody)
					return
				}
			}
		}

	executeDownstream:
		// Start bounded lease renewal heartbeat (every 2s) while execution is active
		heartbeatDone := make(chan struct{})
		if s.db != nil {
			go func(tID, k string) {
				ticker := time.NewTicker(2 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-heartbeatDone:
						return
					case <-ticker.C:
						query := `
							UPDATE idempotency_records
							SET updated_at = NOW()
							WHERE tenant_id = $1 AND idempotency_key = $2 AND in_flight = TRUE;
						`
						_, _ = s.db.Exec(context.Background(), query, tID, k)
					}
				}
			}(tenantID, key)
		}

		// Execute downstream handler with response capture
		recWriter := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recWriter, r)

		// Terminate heartbeat immediately when request finishes
		close(heartbeatDone)

		// Cache final response outcome in memory
		rec.ResponseStatusCode = recWriter.statusCode
		rec.ResponseBody = recWriter.bodyBuf.Bytes()
		rec.ResponseHeaders = recWriter.Header().Clone()
		rec.InFlight = false
		rec.mu.Unlock()

		// Persist outcome durably in PostgreSQL if DB enabled
		if s.db != nil {
			_ = s.persistDBCompleted(context.Background(), tenantID, key, rec.ResponseStatusCode, rec.ResponseHeaders, rec.ResponseBody)
		}
	})
}

func (s *IdempotencyStore) fetchDBRecord(ctx context.Context, tenantID, key string) (*dbIdempotencyRecord, bool, error) {
	query := `
		SELECT request_hash, response_status, response_headers, response_body, in_flight, created_at, updated_at, expires_at
		FROM idempotency_records
		WHERE tenant_id = $1 AND idempotency_key = $2
	`
	var reqHash string
	var status *int
	var headersBytes []byte
	var body []byte
	var inFlight bool
	var createdAt, updatedAt, expiresAt time.Time

	err := s.db.QueryRow(ctx, query, tenantID, key).Scan(&reqHash, &status, &headersBytes, &body, &inFlight, &createdAt, &updatedAt, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	headers := make(http.Header)
	if len(headersBytes) > 0 {
		_ = json.Unmarshal(headersBytes, &headers)
	}

	statusCode := http.StatusOK
	if status != nil {
		statusCode = *status
	}

	return &dbIdempotencyRecord{
		RequestHash:        reqHash,
		ResponseStatusCode: statusCode,
		ResponseBody:       body,
		ResponseHeaders:    headers,
		InFlight:           inFlight,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		ExpiresAt:          expiresAt,
	}, true, nil
}

func (s *IdempotencyStore) claimDBInFlight(ctx context.Context, tenantID, key, reqHash string, expiresAt time.Time) (bool, error) {
	query := `
		INSERT INTO idempotency_records (tenant_id, idempotency_key, request_hash, in_flight, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, TRUE, NOW(), NOW(), $4)
		ON CONFLICT (tenant_id, idempotency_key) DO NOTHING
		RETURNING id;
	`
	var id string
	err := s.db.QueryRow(ctx, query, tenantID, key, reqHash, expiresAt).Scan(&id)
	if err != nil {
		// No row returned -> conflict, another pod won the insert
		return false, nil
	}
	return true, nil
}

func (s *IdempotencyStore) reclaimStaleDBInFlight(ctx context.Context, tenantID, key, reqHash string) (bool, error) {
	query := `
		UPDATE idempotency_records
		SET in_flight = TRUE,
		    request_hash = $3,
		    updated_at = NOW()
		WHERE tenant_id = $1
		  AND idempotency_key = $2
		  AND in_flight = TRUE
		  AND updated_at < NOW() - INTERVAL '10 seconds'
		RETURNING id;
	`
	var id string
	err := s.db.QueryRow(ctx, query, tenantID, key, reqHash).Scan(&id)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *IdempotencyStore) persistDBCompleted(ctx context.Context, tenantID, key string, status int, headers http.Header, body []byte) error {
	headersBytes, _ := json.Marshal(headers)
	query := `
		INSERT INTO idempotency_records (
			tenant_id, idempotency_key, request_hash, response_status, response_headers, response_body, in_flight, created_at, updated_at, expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, FALSE, NOW(), NOW(), NOW() + INTERVAL '15 minutes')
		ON CONFLICT (tenant_id, idempotency_key) DO UPDATE
		SET response_status = EXCLUDED.response_status,
		    response_headers = EXCLUDED.response_headers,
		    response_body = EXCLUDED.response_body,
		    in_flight = FALSE,
		    updated_at = NOW();
	`
	_, err := s.db.Exec(ctx, query, tenantID, key, "", status, headersBytes, body)
	if err != nil {
		log.Printf("Warning: Failed to persist idempotency record to DB: %v", err)
	}
	return err
}

func (s *IdempotencyStore) waitForDBCompletion(ctx context.Context, tenantID, key, expectedHash string) (*dbIdempotencyRecord, bool) {
	deadline := time.Now().Add(3 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, false
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, false
			}
			rec, found, err := s.fetchDBRecord(ctx, tenantID, key)
			if err == nil && found {
				if rec.RequestHash != expectedHash {
					return rec, true // Hash conflict detected
				}
				if !rec.InFlight {
					return rec, true
				}
			}
		}
	}
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
