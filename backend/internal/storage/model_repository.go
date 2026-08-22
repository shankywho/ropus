package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StoredModel represents a registered ML model artifact in PostgreSQL.
type StoredModel struct {
	ID               string    `json:"id"`
	Version          string    `json:"version"`
	Algorithm        string    `json:"algorithm"`
	Metrics          string    `json:"metrics"` // JSON string of metrics
	ApprovalStatus   string    `json:"approval_status"`
	ArtifactLocation string    `json:"artifact_location"`
	DeployedAt       time.Time `json:"deployed_at"`
}

// ModelRepository provides data access for model registry.
type ModelRepository struct {
	db       *DatabaseManager
	mu       sync.RWMutex
	memStore map[string]*StoredModel
}

// NewModelRepository initializes the model repository.
func NewModelRepository(db *DatabaseManager) *ModelRepository {
	return &ModelRepository{
		db:       db,
		memStore: make(map[string]*StoredModel),
	}
}

// Save persists a model entry.
func (r *ModelRepository) Save(ctx context.Context, m *StoredModel) error {
	if m == nil || m.ID == "" {
		return fmt.Errorf("invalid model")
	}

	if m.DeployedAt.IsZero() {
		m.DeployedAt = time.Now().UTC()
	}

	if r.db != nil && r.db.DB() != nil {
		query := `
			INSERT INTO model_registry (
				id, version, algorithm, metrics, approval_status, artifact_location, deployed_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				metrics = EXCLUDED.metrics,
				approval_status = EXCLUDED.approval_status,
				artifact_location = EXCLUDED.artifact_location,
				deployed_at = EXCLUDED.deployed_at;
		`
		_, err := r.db.DB().ExecContext(
			ctx, query,
			m.ID, m.Version, m.Algorithm, m.Metrics, m.ApprovalStatus, m.ArtifactLocation, m.DeployedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert model into postgres: %w", err)
		}
	}

	r.mu.Lock()
	r.memStore[m.ID] = m
	r.mu.Unlock()

	return nil
}

// FindByID retrieves a model by ID.
func (r *ModelRepository) FindByID(ctx context.Context, id string) (*StoredModel, error) {
	if r.db != nil && r.db.DB() != nil {
		query := `
			SELECT id, version, algorithm, metrics, approval_status, artifact_location, deployed_at
			FROM model_registry WHERE id = $1;
		`
		row := r.db.DB().QueryRowContext(ctx, query, id)
		var m StoredModel
		err := row.Scan(
			&m.ID, &m.Version, &m.Algorithm, &m.Metrics, &m.ApprovalStatus, &m.ArtifactLocation, &m.DeployedAt,
		)
		if err == nil {
			return &m, nil
		}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	m, exists := r.memStore[id]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found", id)
	}
	return m, nil
}
