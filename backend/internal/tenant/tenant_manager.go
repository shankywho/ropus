package tenant

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// UserRole defines permission tiers.
type UserRole string

const (
	RoleOwner   UserRole = "OWNER"
	RoleAdmin   UserRole = "ADMIN"
	RoleAnalyst UserRole = "ANALYST"
	RoleViewer  UserRole = "VIEWER"
)

// Organization represents a customer enterprise tenant.
type Organization struct {
	OrgID         string    `json:"org_id"`
	Name          string    `json:"name"`
	PlanTier      string    `json:"plan_tier"` // "ENTERPRISE", "GROWTH", "STARTER"
	MonthlyQuota  int64     `json:"monthly_quota"`
	UsedThisMonth int64     `json:"used_this_month"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserAccount represents an authenticated member of an organization.
type UserAccount struct {
	UserID    string    `json:"user_id"`
	OrgID     string    `json:"org_id"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// APIKeyRecord represents an API key linked to a tenant organization.
type APIKeyRecord struct {
	KeyID        string    `json:"key_id"`
	OrgID        string    `json:"org_id"`
	Name         string    `json:"name"`
	HashedKey    string    `json:"-"`
	Prefix       string    `json:"prefix"` // e.g. "ropus_live_abc..."
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsedAt   time.Time `json:"last_used_at"`
}

// TenantManager handles multi-tenant lifecycle, API keys, and rate limits.
type TenantManager struct {
	mu       sync.RWMutex
	orgs     map[string]*Organization
	users    map[string]*UserAccount
	apiKeys  map[string]*APIKeyRecord // hashedKey -> APIKeyRecord
	keyIndex map[string]*APIKeyRecord // keyID -> APIKeyRecord
}

// NewTenantManager initializes the tenant management platform.
func NewTenantManager() *TenantManager {
	tm := &TenantManager{
		orgs:     make(map[string]*Organization),
		users:    make(map[string]*UserAccount),
		apiKeys:  make(map[string]*APIKeyRecord),
		keyIndex: make(map[string]*APIKeyRecord),
	}
	// Seed default enterprise organization
	tm.CreateOrganization("org_default", "Acme Global Financial", "ENTERPRISE", 10000000)
	return tm
}

// CreateOrganization provisions a new tenant organization.
func (m *TenantManager) CreateOrganization(orgID, name, plan string, quota int64) *Organization {
	m.mu.Lock()
	defer m.mu.Unlock()

	org := &Organization{
		OrgID:         orgID,
		Name:          name,
		PlanTier:      plan,
		MonthlyQuota:  quota,
		UsedThisMonth: 0,
		CreatedAt:     time.Now().UTC(),
	}
	m.orgs[orgID] = org
	return org
}

// GenerateAPIKey creates a cryptographically secure API key for an organization.
func (m *TenantManager) GenerateAPIKey(orgID, keyName string) (string, *APIKeyRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.orgs[orgID]; !exists {
		return "", nil, fmt.Errorf("organization '%s' does not exist", orgID)
	}

	rawBytes := make([]byte, 24)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", nil, err
	}
	secretHex := hex.EncodeToString(rawBytes)
	rawKey := fmt.Sprintf("ropus_live_%s", secretHex)

	sum := sha256.Sum256([]byte(rawKey))
	hashedKey := hex.EncodeToString(sum[:])
	keyID := fmt.Sprintf("key_%d", time.Now().UnixNano())

	rec := &APIKeyRecord{
		KeyID:      keyID,
		OrgID:      orgID,
		Name:       keyName,
		HashedKey:  hashedKey,
		Prefix:     rawKey[:15] + "...",
		IsActive:   true,
		CreatedAt:  time.Now().UTC(),
		LastUsedAt: time.Now().UTC(),
	}

	m.apiKeys[hashedKey] = rec
	m.keyIndex[keyID] = rec
	return rawKey, rec, nil
}

// AuthenticateAPIKey validates a raw API key and increments usage.
func (m *TenantManager) AuthenticateAPIKey(rawKey string) (*APIKeyRecord, error) {
	sum := sha256.Sum256([]byte(rawKey))
	hashedKey := hex.EncodeToString(sum[:])

	m.mu.Lock()
	defer m.mu.Unlock()

	rec, exists := m.apiKeys[hashedKey]
	if !exists || !rec.IsActive {
		return nil, fmt.Errorf("invalid or inactive API key")
	}

	org, orgExists := m.orgs[rec.OrgID]
	if !orgExists {
		return nil, fmt.Errorf("organization not found")
	}

	if org.UsedThisMonth >= org.MonthlyQuota {
		return nil, fmt.Errorf("organization monthly quota exceeded (%d/%d)", org.UsedThisMonth, org.MonthlyQuota)
	}

	org.UsedThisMonth++
	rec.LastUsedAt = time.Now().UTC()
	return rec, nil
}

// RotateAPIKey deactivates an old key and issues a replacement.
func (m *TenantManager) RotateAPIKey(oldKeyID string) (string, *APIKeyRecord, error) {
	m.mu.Lock()
	oldRec, exists := m.keyIndex[oldKeyID]
	if !exists {
		m.mu.Unlock()
		return "", nil, fmt.Errorf("key ID '%s' not found", oldKeyID)
	}
	oldRec.IsActive = false
	orgID := oldRec.OrgID
	name := oldRec.Name + " (Rotated)"
	m.mu.Unlock()

	return m.GenerateAPIKey(orgID, name)
}
