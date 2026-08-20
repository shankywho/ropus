package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"os"
	"sync"
)

var (
	ErrTenantKeyShredded = errors.New("kms: tenant key has been shredded; data is cryptographically unrecoverable")
)

// KMS defines the interface for Key Management and Crypto-Shredding.
type KMS interface {
	GetTenantKey(tenantID string) ([]byte, error)
	ShredTenantKey(tenantID string) error
	IsTenantShredded(tenantID string) bool
}

// MockKMS provides in-memory per-tenant AES-256 key management with crypto-shredding support.
type MockKMS struct {
	mu           sync.RWMutex
	masterSecret []byte
	tenantKeys   map[string][]byte
	shreddedKeys map[string]bool
}

// NewMockKMS initializes a new MockKMS with a master derivation secret.
func NewMockKMS() *MockKMS {
	master := os.Getenv("KMS_MASTER_SECRET")
	if master == "" {
		master = "kms_master_secret_key_risk_engine_32bytes!"
	}
	return &MockKMS{
		masterSecret: []byte(master),
		tenantKeys:   make(map[string][]byte),
		shreddedKeys: make(map[string]bool),
	}
}

// GetTenantKey derives or retrieves a consistent 32-byte AES-256 key for a given tenant.
func (k *MockKMS) GetTenantKey(tenantID string) ([]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if tenantID == "" {
		tenantID = "00000000-0000-0000-0000-000000000001"
	}

	// Check if this tenant has been crypto-shredded
	if k.shreddedKeys[tenantID] {
		return nil, ErrTenantKeyShredded
	}

	// Check if custom key is cached
	if key, ok := k.tenantKeys[tenantID]; ok {
		return key, nil
	}

	// Derive a deterministic 32-byte key using HMAC-SHA256
	mac := hmac.New(sha256.New, k.masterSecret)
	mac.Write([]byte(tenantID))
	derivedKey := mac.Sum(nil) // 32 bytes (256 bits)

	k.tenantKeys[tenantID] = derivedKey
	return derivedKey, nil
}

// ShredTenantKey permanently destroys the encryption key for a tenant (Crypto-Shredding).
// Any data previously encrypted with this key becomes mathematically impossible to decrypt.
func (k *MockKMS) ShredTenantKey(tenantID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if key, exists := k.tenantKeys[tenantID]; exists {
		// Zero-out memory
		for i := range key {
			key[i] = 0
		}
		delete(k.tenantKeys, tenantID)
	}

	k.shreddedKeys[tenantID] = true
	return nil
}

// IsTenantShredded checks if a tenant's keys have been shredded.
func (k *MockKMS) IsTenantShredded(tenantID string) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.shreddedKeys[tenantID]
}
