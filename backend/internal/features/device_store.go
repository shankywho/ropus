package features

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDeviceNotFound     = errors.New("device record not found")
	ErrReputationNotFound = errors.New("device reputation not found")
)

// DeviceRecord models the primary PostgreSQL device registry entity.
type DeviceRecord struct {
	DeviceID             string    `json:"device_id"`   // UUID primary key in PostgreSQL
	TenantID             string    `json:"tenant_id"`   // UUID tenant identifier
	DeviceHash           string    `json:"device_hash"` // 64-character SHA-256 canonical hash
	EncryptedFingerprint string    `json:"-"`           // Encrypted ciphertext (AES-256-GCM)
	FirstSeenAt          time.Time `json:"first_seen_at"`
	LastSeenAt           time.Time `json:"last_seen_at"`
	TotalTxCount         int64     `json:"total_tx_count"`
	UniqueAccountCount   int       `json:"unique_account_count"`
	TrustScore           int       `json:"trust_score"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// DeviceAccountRecord models a device ↔ account linkage.
type DeviceAccountRecord struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	DeviceID         string    `json:"device_id"`
	AccountID        string    `json:"account_id"`
	FirstSeenAt      time.Time `json:"first_seen_at"`
	LastSeenAt       time.Time `json:"last_seen_at"`
	TransactionCount int64     `json:"transaction_count"`
	CreatedAt        time.Time `json:"created_at"`
}

// DevicePaymentInstrumentRecord models a device ↔ payment token linkage.
type DevicePaymentInstrumentRecord struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	DeviceID         string    `json:"device_id"`
	PaymentToken     string    `json:"payment_token"`
	FirstSeenAt      time.Time `json:"first_seen_at"`
	LastSeenAt       time.Time `json:"last_seen_at"`
	TransactionCount int64     `json:"transaction_count"`
	CreatedAt        time.Time `json:"created_at"`
}

// DeviceEventRecord models an immutable event in the historical device ledger.
type DeviceEventRecord struct {
	ID             string                 `json:"id"`
	TenantID       string                 `json:"tenant_id"`
	DeviceID       string                 `json:"device_id"`
	EventType      string                 `json:"event_type"` // TRANSACTION, LOGIN, CHARGEBACK, FRAUD_CONFIRMATION
	AccountID      string                 `json:"account_id,omitempty"`
	PaymentToken   string                 `json:"payment_token,omitempty"`
	EventTime      time.Time              `json:"event_time"`
	Amount         int64                  `json:"amount,omitempty"`
	Currency       string                 `json:"currency,omitempty"`
	RiskDecisionID *string                `json:"risk_decision_id,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// DeviceReputationRecord models the persistent device reputation and dispute history.
type DeviceReputationRecord struct {
	ID                         string     `json:"id"`
	TenantID                   string     `json:"tenant_id"`
	DeviceID                   string     `json:"device_id"`
	FraudCount                 int        `json:"fraud_count"`
	ChargebackCount            int        `json:"chargeback_count"`
	ConfirmedLegitimateCount   int        `json:"confirmed_legitimate_count"`
	ReputationScore            int        `json:"reputation_score"`
	RiskBand                   string     `json:"risk_band"` // TRUSTED, NEUTRAL, SUSPICIOUS, BLACKLISTED
	LastFraudAt                *time.Time `json:"last_fraud_at,omitempty"`
	LastChargebackAt           *time.Time `json:"last_chargeback_at,omitempty"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
}

// DeviceStore provides PostgreSQL relational operations for durable device intelligence.
type DeviceStore struct {
	db *pgxpool.Pool
}

// NewDeviceStore constructs a new PostgreSQL DeviceStore.
func NewDeviceStore(db *pgxpool.Pool) *DeviceStore {
	return &DeviceStore{db: db}
}

// GetDeviceByHash retrieves a device record by tenant ID and canonical SHA-256 hash.
func (s *DeviceStore) GetDeviceByHash(ctx context.Context, tenantID, deviceHash string) (*DeviceRecord, error) {
	if s == nil || s.db == nil {
		return nil, ErrDeviceNotFound
	}

	query := `
		SELECT device_id, tenant_id, device_hash, encrypted_fingerprint,
		       first_seen_at, last_seen_at, total_tx_count, unique_account_count,
		       trust_score, created_at, updated_at
		FROM devices
		WHERE tenant_id = $1 AND device_hash = $2
	`

	var d DeviceRecord
	var encFP *string
	err := s.db.QueryRow(ctx, query, tenantID, deviceHash).Scan(
		&d.DeviceID,
		&d.TenantID,
		&d.DeviceHash,
		&encFP,
		&d.FirstSeenAt,
		&d.LastSeenAt,
		&d.TotalTxCount,
		&d.UniqueAccountCount,
		&d.TrustScore,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("failed to query device by hash: %w", err)
	}
	if encFP != nil {
		d.EncryptedFingerprint = *encFP
	}
	return &d, nil
}

// UpsertDeviceSeen inserts a new device or updates last_seen_at and increments total_tx_count atomically.
func (s *DeviceStore) UpsertDeviceSeen(ctx context.Context, tenantID, deviceHash, encryptedFP string) (*DeviceRecord, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}

	query := `
		INSERT INTO devices (
			tenant_id, device_hash, encrypted_fingerprint, first_seen_at, last_seen_at,
			total_tx_count, unique_account_count, trust_score, created_at, updated_at
		)
		VALUES ($1, $2, $3, NOW(), NOW(), 1, 1, 50, NOW(), NOW())
		ON CONFLICT (tenant_id, device_hash) DO UPDATE
		SET last_seen_at = NOW(),
		    total_tx_count = devices.total_tx_count + 1,
		    encrypted_fingerprint = COALESCE(EXCLUDED.encrypted_fingerprint, devices.encrypted_fingerprint),
		    updated_at = NOW()
		RETURNING device_id, tenant_id, device_hash, encrypted_fingerprint,
		          first_seen_at, last_seen_at, total_tx_count, unique_account_count,
		          trust_score, created_at, updated_at
	`

	var d DeviceRecord
	var enc *string
	err := s.db.QueryRow(ctx, query, tenantID, deviceHash, encryptedFP).Scan(
		&d.DeviceID,
		&d.TenantID,
		&d.DeviceHash,
		&enc,
		&d.FirstSeenAt,
		&d.LastSeenAt,
		&d.TotalTxCount,
		&d.UniqueAccountCount,
		&d.TrustScore,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert device: %w", err)
	}
	if enc != nil {
		d.EncryptedFingerprint = *enc
	}
	return &d, nil
}

// LinkAccount links a device with a customer account, updating timestamps and counts atomically.
func (s *DeviceStore) LinkAccount(ctx context.Context, tenantID, deviceUUID, accountID string) error {
	if s == nil || s.db == nil || accountID == "" {
		return nil
	}

	query := `
		INSERT INTO device_accounts (
			tenant_id, device_id, account_id, first_seen_at, last_seen_at, transaction_count, created_at, updated_at
		)
		VALUES ($1, $2, $3, NOW(), NOW(), 1, NOW(), NOW())
		ON CONFLICT (tenant_id, device_id, account_id) DO UPDATE
		SET last_seen_at = NOW(),
		    transaction_count = device_accounts.transaction_count + 1,
		    updated_at = NOW()
	`

	_, err := s.db.Exec(ctx, query, tenantID, deviceUUID, accountID)
	if err != nil {
		return fmt.Errorf("failed to link device account: %w", err)
	}
	return nil
}

// LinkPaymentToken links a device with a tokenized payment instrument atomically.
func (s *DeviceStore) LinkPaymentToken(ctx context.Context, tenantID, deviceUUID, paymentToken string) error {
	if s == nil || s.db == nil || paymentToken == "" {
		return nil
	}

	query := `
		INSERT INTO device_payment_instruments (
			tenant_id, device_id, payment_token, first_seen_at, last_seen_at, transaction_count, created_at, updated_at
		)
		VALUES ($1, $2, $3, NOW(), NOW(), 1, NOW(), NOW())
		ON CONFLICT (tenant_id, device_id, payment_token) DO UPDATE
		SET last_seen_at = NOW(),
		    transaction_count = device_payment_instruments.transaction_count + 1,
		    updated_at = NOW()
	`

	_, err := s.db.Exec(ctx, query, tenantID, deviceUUID, paymentToken)
	if err != nil {
		return fmt.Errorf("failed to link device payment token: %w", err)
	}
	return nil
}

// RecordDeviceEvent writes an audit ledger entry to the device_events table.
func (s *DeviceStore) RecordDeviceEvent(ctx context.Context, event *DeviceEventRecord) error {
	if s == nil || s.db == nil || event == nil {
		return nil
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO device_events (
			tenant_id, device_id, event_type, account_id, payment_token,
			event_time, amount, currency, risk_decision_id, metadata, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
	`

	_, err = s.db.Exec(ctx, query,
		event.TenantID,
		event.DeviceID,
		event.EventType,
		event.AccountID,
		event.PaymentToken,
		event.EventTime,
		event.Amount,
		event.Currency,
		event.RiskDecisionID,
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert device event: %w", err)
	}
	return nil
}

// GetDeviceReputation retrieves the current reputation score and risk band for a device.
func (s *DeviceStore) GetDeviceReputation(ctx context.Context, tenantID, deviceUUID string) (*DeviceReputationRecord, error) {
	if s == nil || s.db == nil {
		return nil, ErrReputationNotFound
	}

	query := `
		SELECT id, tenant_id, device_id, fraud_count, chargeback_count,
		       confirmed_legitimate_count, reputation_score, risk_band,
		       last_fraud_at, last_chargeback_at, created_at, updated_at
		FROM device_reputation
		WHERE tenant_id = $1 AND device_id = $2
	`

	var r DeviceReputationRecord
	err := s.db.QueryRow(ctx, query, tenantID, deviceUUID).Scan(
		&r.ID,
		&r.TenantID,
		&r.DeviceID,
		&r.FraudCount,
		&r.ChargebackCount,
		&r.ConfirmedLegitimateCount,
		&r.ReputationScore,
		&r.RiskBand,
		&r.LastFraudAt,
		&r.LastChargebackAt,
		&r.CreatedAt,
		&r.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrReputationNotFound
		}
		return nil, fmt.Errorf("failed to query device reputation: %w", err)
	}
	return &r, nil
}

// GetDeviceAccounts retrieves all account linkages associated with a device.
func (s *DeviceStore) GetDeviceAccounts(ctx context.Context, tenantID, deviceUUID string) ([]DeviceAccountRecord, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}

	query := `
		SELECT id, tenant_id, device_id, account_id, first_seen_at, last_seen_at, transaction_count, created_at
		FROM device_accounts
		WHERE tenant_id = $1 AND device_id = $2
		ORDER BY last_seen_at DESC
	`

	rows, err := s.db.Query(ctx, query, tenantID, deviceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query device accounts: %w", err)
	}
	defer rows.Close()

	var records []DeviceAccountRecord
	for rows.Next() {
		var r DeviceAccountRecord
		if err := rows.Scan(&r.ID, &r.TenantID, &r.DeviceID, &r.AccountID, &r.FirstSeenAt, &r.LastSeenAt, &r.TransactionCount, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// GetAccountDevices retrieves all devices associated with a specific customer account.
func (s *DeviceStore) GetAccountDevices(ctx context.Context, tenantID, accountID string) ([]DeviceAccountRecord, error) {
	if s == nil || s.db == nil || accountID == "" {
		return nil, nil
	}

	query := `
		SELECT id, tenant_id, device_id, account_id, first_seen_at, last_seen_at, transaction_count, created_at
		FROM device_accounts
		WHERE tenant_id = $1 AND account_id = $2
		ORDER BY last_seen_at DESC
	`

	rows, err := s.db.Query(ctx, query, tenantID, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query account devices: %w", err)
	}
	defer rows.Close()

	var records []DeviceAccountRecord
	for rows.Next() {
		var r DeviceAccountRecord
		if err := rows.Scan(&r.ID, &r.TenantID, &r.DeviceID, &r.AccountID, &r.FirstSeenAt, &r.LastSeenAt, &r.TransactionCount, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// GetAccountDeviceRelationship retrieves the specific linkage between a device and account.
func (s *DeviceStore) GetAccountDeviceRelationship(ctx context.Context, tenantID, deviceUUID, accountID string) (*DeviceAccountRecord, error) {
	if s == nil || s.db == nil || accountID == "" {
		return nil, errors.New("relationship not found")
	}

	query := `
		SELECT id, tenant_id, device_id, account_id, first_seen_at, last_seen_at, transaction_count, created_at
		FROM device_accounts
		WHERE tenant_id = $1 AND device_id = $2 AND account_id = $3
	`

	var r DeviceAccountRecord
	err := s.db.QueryRow(ctx, query, tenantID, deviceUUID, accountID).Scan(
		&r.ID, &r.TenantID, &r.DeviceID, &r.AccountID, &r.FirstSeenAt, &r.LastSeenAt, &r.TransactionCount, &r.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("relationship not found")
		}
		return nil, fmt.Errorf("failed to query account device relationship: %w", err)
	}
	return &r, nil
}

// GetDevicePaymentInstruments retrieves all payment instrument linkages for a device.
func (s *DeviceStore) GetDevicePaymentInstruments(ctx context.Context, tenantID, deviceUUID string) ([]DevicePaymentInstrumentRecord, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}

	query := `
		SELECT id, tenant_id, device_id, payment_token, first_seen_at, last_seen_at, transaction_count, created_at
		FROM device_payment_instruments
		WHERE tenant_id = $1 AND device_id = $2
		ORDER BY last_seen_at DESC
	`

	rows, err := s.db.Query(ctx, query, tenantID, deviceUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query device payment instruments: %w", err)
	}
	defer rows.Close()

	var records []DevicePaymentInstrumentRecord
	for rows.Next() {
		var r DevicePaymentInstrumentRecord
		if err := rows.Scan(&r.ID, &r.TenantID, &r.DeviceID, &r.PaymentToken, &r.FirstSeenAt, &r.LastSeenAt, &r.TransactionCount, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// GetPaymentInstrumentDevices retrieves all devices associated with a payment token.
func (s *DeviceStore) GetPaymentInstrumentDevices(ctx context.Context, tenantID, paymentToken string) ([]DevicePaymentInstrumentRecord, error) {
	if s == nil || s.db == nil || paymentToken == "" {
		return nil, nil
	}

	query := `
		SELECT id, tenant_id, device_id, payment_token, first_seen_at, last_seen_at, transaction_count, created_at
		FROM device_payment_instruments
		WHERE tenant_id = $1 AND payment_token = $2
		ORDER BY last_seen_at DESC
	`

	rows, err := s.db.Query(ctx, query, tenantID, paymentToken)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment instrument devices: %w", err)
	}
	defer rows.Close()

	var records []DevicePaymentInstrumentRecord
	for rows.Next() {
		var r DevicePaymentInstrumentRecord
		if err := rows.Scan(&r.ID, &r.TenantID, &r.DeviceID, &r.PaymentToken, &r.FirstSeenAt, &r.LastSeenAt, &r.TransactionCount, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// GetDevicePaymentInstrumentRelationship retrieves the specific linkage between a device and payment token.
func (s *DeviceStore) GetDevicePaymentInstrumentRelationship(ctx context.Context, tenantID, deviceUUID, paymentToken string) (*DevicePaymentInstrumentRecord, error) {
	if s == nil || s.db == nil || paymentToken == "" {
		return nil, errors.New("relationship not found")
	}

	query := `
		SELECT id, tenant_id, device_id, payment_token, first_seen_at, last_seen_at, transaction_count, created_at
		FROM device_payment_instruments
		WHERE tenant_id = $1 AND device_id = $2 AND payment_token = $3
	`

	var r DevicePaymentInstrumentRecord
	err := s.db.QueryRow(ctx, query, tenantID, deviceUUID, paymentToken).Scan(
		&r.ID, &r.TenantID, &r.DeviceID, &r.PaymentToken, &r.FirstSeenAt, &r.LastSeenAt, &r.TransactionCount, &r.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("relationship not found")
		}
		return nil, fmt.Errorf("failed to query payment instrument relationship: %w", err)
	}
	return &r, nil
}


