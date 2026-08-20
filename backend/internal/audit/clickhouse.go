package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// AuditRecord represents a flattened audit log entry in ClickHouse.
type AuditRecord struct {
	TransactionID   string    `json:"transaction_id"`
	RiskScore       int32     `json:"risk_score"`
	RuleTriggered   string    `json:"rule_triggered"`
	FeatureSnapshot string    `json:"feature_snapshot"`
	TenantID        string    `json:"tenant_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// ClickHouseClient manages connection to ClickHouse server.
type ClickHouseClient struct {
	conn driver.Conn
}

// NewClickHouseClient initializes a native ClickHouse connection.
func NewClickHouseClient(addr, database, username, password string) (*ClickHouseClient, error) {
	if addr == "" {
		addr = "localhost:9000"
	}
	if database == "" {
		database = "default"
	}
	if username == "" {
		username = "default"
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	return &ClickHouseClient{conn: conn}, nil
}

// Ping verifies connectivity to ClickHouse.
func (c *ClickHouseClient) Ping(ctx context.Context) error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("clickhouse connection is uninitialized")
	}
	return c.conn.Ping(ctx)
}

// InsertAuditRecord inserts a single audit record into risk_audit_log.
func (c *ClickHouseClient) InsertAuditRecord(ctx context.Context, record AuditRecord) error {
	if c == nil || c.conn == nil {
		return nil
	}

	query := `
		INSERT INTO risk_audit_log (
			transaction_id, risk_score, rule_triggered, feature_snapshot, tenant_id, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?
		)
	`
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}

	err := c.conn.Exec(ctx, query,
		record.TransactionID,
		record.RiskScore,
		record.RuleTriggered,
		record.FeatureSnapshot,
		record.TenantID,
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert audit record into clickhouse: %w", err)
	}

	return nil
}

// Close terminates the ClickHouse connection.
func (c *ClickHouseClient) Close() error {
	if c != nil && c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
