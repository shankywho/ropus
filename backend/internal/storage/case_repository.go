package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StoredCase represents a fraud investigation case entity in PostgreSQL.
type StoredCase struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transaction_id"`
	Priority      string    `json:"priority"` // "CRITICAL", "HIGH", "MEDIUM", "LOW"
	Status        string    `json:"status"`   // "OPEN", "INVESTIGATING", "RESOLVED", "CLOSED"
	AssignedAgent string    `json:"assigned_agent"`
	Evidence      string    `json:"evidence"` // JSON payload of evidence signals
	Resolution    string    `json:"resolution"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CaseRepository provides data access for fraud cases.
type CaseRepository struct {
	db       *DatabaseManager
	mu       sync.RWMutex
	memStore map[string]*StoredCase
}

// NewCaseRepository initializes the case repository.
func NewCaseRepository(db *DatabaseManager) *CaseRepository {
	return &CaseRepository{
		db:       db,
		memStore: make(map[string]*StoredCase),
	}
}

// Save persists or updates a case.
func (r *CaseRepository) Save(ctx context.Context, c *StoredCase) error {
	if c == nil || c.ID == "" {
		return fmt.Errorf("invalid case")
	}

	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now().UTC()
	}
	c.UpdatedAt = time.Now().UTC()

	if r.db != nil && r.db.DB() != nil {
		query := `
			INSERT INTO fraud_cases (
				id, transaction_id, priority, status, assigned_agent, evidence, resolution, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE SET
				priority = EXCLUDED.priority,
				status = EXCLUDED.status,
				assigned_agent = EXCLUDED.assigned_agent,
				evidence = EXCLUDED.evidence,
				resolution = EXCLUDED.resolution,
				updated_at = EXCLUDED.updated_at;
		`
		_, err := r.db.DB().ExecContext(
			ctx, query,
			c.ID, c.TransactionID, c.Priority, c.Status, c.AssignedAgent, c.Evidence, c.Resolution, c.CreatedAt, c.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert case into postgres: %w", err)
		}
	}

	r.mu.Lock()
	r.memStore[c.ID] = c
	r.mu.Unlock()

	return nil
}

// FindByID retrieves a case by ID.
func (r *CaseRepository) FindByID(ctx context.Context, id string) (*StoredCase, error) {
	if r.db != nil && r.db.DB() != nil {
		query := `
			SELECT id, transaction_id, priority, status, assigned_agent, evidence, resolution, created_at, updated_at
			FROM fraud_cases WHERE id = $1;
		`
		row := r.db.DB().QueryRowContext(ctx, query, id)
		var c StoredCase
		err := row.Scan(
			&c.ID, &c.TransactionID, &c.Priority, &c.Status, &c.AssignedAgent, &c.Evidence, &c.Resolution, &c.CreatedAt, &c.UpdatedAt,
		)
		if err == nil {
			return &c, nil
		}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	c, exists := r.memStore[id]
	if !exists {
		return nil, fmt.Errorf("case '%s' not found", id)
	}
	return c, nil
}
