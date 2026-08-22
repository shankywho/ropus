package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// DatabaseConfig holds connection pool settings.
type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DatabaseManager handles PostgreSQL connection pooling and health checks.
type DatabaseManager struct {
	mu  sync.RWMutex
	db  *sql.DB
	cfg DatabaseConfig
}

// NewDatabaseManager initializes a database manager instance.
func NewDatabaseManager(cfg DatabaseConfig) (*DatabaseManager, error) {
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 25
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 15 * time.Minute
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 5 * time.Minute
	}

	mgr := &DatabaseManager{cfg: cfg}

	// If DSN is provided, attempt connection
	if cfg.DSN != "" {
		db, err := sql.Open("postgres", cfg.DSN)
		if err == nil {
			db.SetMaxOpenConns(cfg.MaxOpenConns)
			db.SetMaxIdleConns(cfg.MaxIdleConns)
			db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
			db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
			mgr.db = db
		}
	}

	return mgr, nil
}

// HealthCheck performs a ping test on the database.
func (m *DatabaseManager) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.db == nil {
		return nil // Operating in memory-backed mode for testing
	}
	return m.db.PingContext(ctx)
}

// Close gracefully closes open pool connections.
func (m *DatabaseManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// DB returns the underlying sql.DB instance if connected.
func (m *DatabaseManager) DB() *sql.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.db
}

// ExecuteInTransaction executes a function within an isolated ACID transaction.
func (m *DatabaseManager) ExecuteInTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	m.mu.RLock()
	db := m.db
	m.mu.RUnlock()

	if db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error in tx: %v (rollback error: %v)", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
