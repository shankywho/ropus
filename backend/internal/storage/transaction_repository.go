package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StoredTransaction represents a persistent transaction entity in PostgreSQL.
type StoredTransaction struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	CustomerHash string    `json:"customer_hash"`
	MerchantID   string    `json:"merchant_id"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	LocationHash string    `json:"location_hash"`
	DeviceHash   string    `json:"device_hash"`
	RiskScore    float64   `json:"risk_score"`
	Decision     string    `json:"decision"`
	ModelVersion string    `json:"model_version"`
	CreatedAt    time.Time `json:"created_at"`
}

// TransactionRepository provides data access for financial transactions.
type TransactionRepository struct {
	db       *DatabaseManager
	mu       sync.RWMutex
	memStore map[string]*StoredTransaction
}

// NewTransactionRepository initializes the transaction repository.
func NewTransactionRepository(db *DatabaseManager) *TransactionRepository {
	return &TransactionRepository{
		db:       db,
		memStore: make(map[string]*StoredTransaction),
	}
}

// Save persists a transaction record.
func (r *TransactionRepository) Save(ctx context.Context, tx *StoredTransaction) error {
	if tx == nil || tx.ID == "" {
		return fmt.Errorf("invalid transaction")
	}

	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now().UTC()
	}

	if r.db != nil && r.db.DB() != nil {
		query := `
			INSERT INTO transactions (
				id, tenant_id, customer_hash, merchant_id, amount, currency,
				location_hash, device_hash, risk_score, decision, model_version, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO UPDATE SET
				risk_score = EXCLUDED.risk_score,
				decision = EXCLUDED.decision,
				model_version = EXCLUDED.model_version;
		`
		_, err := r.db.DB().ExecContext(
			ctx, query,
			tx.ID, tx.TenantID, tx.CustomerHash, tx.MerchantID, tx.Amount, tx.Currency,
			tx.LocationHash, tx.DeviceHash, tx.RiskScore, tx.Decision, tx.ModelVersion, tx.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert transaction into postgres: %w", err)
		}
	}

	r.mu.Lock()
	r.memStore[tx.ID] = tx
	r.mu.Unlock()

	return nil
}

// FindByID retrieves a transaction by ID.
func (r *TransactionRepository) FindByID(ctx context.Context, id string) (*StoredTransaction, error) {
	if r.db != nil && r.db.DB() != nil {
		query := `
			SELECT id, tenant_id, customer_hash, merchant_id, amount, currency,
			       location_hash, device_hash, risk_score, decision, model_version, created_at
			FROM transactions WHERE id = $1;
		`
		row := r.db.DB().QueryRowContext(ctx, query, id)
		var tx StoredTransaction
		err := row.Scan(
			&tx.ID, &tx.TenantID, &tx.CustomerHash, &tx.MerchantID, &tx.Amount, &tx.Currency,
			&tx.LocationHash, &tx.DeviceHash, &tx.RiskScore, &tx.Decision, &tx.ModelVersion, &tx.CreatedAt,
		)
		if err == nil {
			return &tx, nil
		}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, exists := r.memStore[id]
	if !exists {
		return nil, fmt.Errorf("transaction '%s' not found", id)
	}
	return tx, nil
}
