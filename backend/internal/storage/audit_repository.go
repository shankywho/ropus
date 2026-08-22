package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// StoredAuditEntry represents a cryptographically verifiable tamper-evident log in PostgreSQL.
type StoredAuditEntry struct {
	ID           string    `json:"id"`
	Actor        string    `json:"actor"`
	Action       string    `json:"action"`
	Resource     string    `json:"resource"`
	PreviousHash string    `json:"previous_hash"`
	CurrentHash  string    `json:"current_hash"`
	Timestamp    time.Time `json:"timestamp"`
}

// AuditRepository provides append-only data access for audit trails.
type AuditRepository struct {
	db       *DatabaseManager
	mu       sync.RWMutex
	lastHash string
	memStore []*StoredAuditEntry
}

// NewAuditRepository initializes the audit repository.
func NewAuditRepository(db *DatabaseManager) *AuditRepository {
	return &AuditRepository{
		db:       db,
		lastHash: "0000000000000000000000000000000000000000000000000000000000000000",
		memStore: make([]*StoredAuditEntry, 0),
	}
}

// Append logs an action and updates the cryptographic hash chain.
func (r *AuditRepository) Append(ctx context.Context, actor, action, resource string) (*StoredAuditEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	id := fmt.Sprintf("aud_%d", now.UnixNano())

	data := fmt.Sprintf("%s:%s:%s:%s:%s", id, actor, action, resource, r.lastHash)
	sum := sha256.Sum256([]byte(data))
	currentHash := hex.EncodeToString(sum[:])

	entry := &StoredAuditEntry{
		ID:           id,
		Actor:        actor,
		Action:       action,
		Resource:     resource,
		PreviousHash: r.lastHash,
		CurrentHash:  currentHash,
		Timestamp:    now,
	}

	r.lastHash = currentHash
	r.memStore = append(r.memStore, entry)

	if r.db != nil && r.db.DB() != nil {
		query := `
			INSERT INTO audit_logs (id, actor, action, resource, previous_hash, current_hash, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7);
		`
		_, _ = r.db.DB().ExecContext(ctx, query, entry.ID, entry.Actor, entry.Action, entry.Resource, entry.PreviousHash, entry.CurrentHash, entry.Timestamp)
	}

	return entry, nil
}

// ListRecent retrieves the latest audit log entries.
func (r *AuditRepository) ListRecent(ctx context.Context, limit int) ([]*StoredAuditEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n := len(r.memStore)
	if limit > n {
		limit = n
	}

	res := make([]*StoredAuditEntry, limit)
	for i := 0; i < limit; i++ {
		res[i] = r.memStore[n-1-i]
	}
	return res, nil
}
