package api_keys

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// APIKeyMetadata contains public properties of an API key (secret is never stored in plaintext).
type APIKeyMetadata struct {
	KeyID       string    `json:"key_id"`
	OrgID       string    `json:"org_id"`
	Name        string    `json:"name"`
	KeyPrefix   string    `json:"key_prefix"` // e.g. "rop_live_8a19"
	KeyHash     string    `json:"-"`
	Environment string    `json:"environment"` // "live" or "test"
	Permissions []string  `json:"permissions"` // e.g. ["risk:evaluate", "cases:write", "models:read"]
	IsRevoked   bool      `json:"is_revoked"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
}

// KeyGenerationResult returns the plaintext secret once upon creation.
type KeyGenerationResult struct {
	PlaintextKey string         `json:"plaintext_key"`
	Metadata     APIKeyMetadata `json:"metadata"`
}

// APIKeyService manages API key lifecycle, verification, and rotation.
type APIKeyService struct {
	mu   sync.RWMutex
	keys map[string]*APIKeyMetadata // keyHash -> Metadata
}

// NewAPIKeyService initializes the API key service.
func NewAPIKeyService() *APIKeyService {
	return &APIKeyService{
		keys: make(map[string]*APIKeyMetadata),
	}
}

// GenerateKey provisions a new cryptographically random API key.
func (s *APIKeyService) GenerateKey(orgID, name, env string, permissions []string, validDays int) (*KeyGenerationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if env != "live" && env != "test" {
		env = "live"
	}

	// 32 random bytes -> 64 hex chars
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate random entropy: %w", err)
	}

	secret := hex.EncodeToString(b)
	plaintext := fmt.Sprintf("rop_%s_%s", env, secret)
	prefix := plaintext[:12] + "..."

	sum := sha256.Sum256([]byte(plaintext))
	hash := hex.EncodeToString(sum[:])

	now := time.Now().UTC()
	keyID := fmt.Sprintf("key_%s", hex.EncodeToString(b[:6]))
	expires := now.Add(time.Duration(validDays) * 24 * time.Hour)
	if validDays == 0 {
		expires = now.Add(365 * 24 * time.Hour)
	}

	meta := APIKeyMetadata{
		KeyID:       keyID,
		OrgID:       orgID,
		Name:        name,
		KeyPrefix:   prefix,
		KeyHash:     hash,
		Environment: env,
		Permissions: permissions,
		IsRevoked:   false,
		CreatedAt:   now,
		ExpiresAt:   expires,
	}

	s.keys[hash] = &meta

	return &KeyGenerationResult{
		PlaintextKey: plaintext,
		Metadata:     meta,
	}, nil
}

// VerifyKey authenticates an incoming request Bearer token and returns metadata.
func (s *APIKeyService) VerifyKey(token string) (*APIKeyMetadata, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !strings.HasPrefix(token, "rop_") {
		return nil, fmt.Errorf("invalid API key format: must start with 'rop_'")
	}

	sum := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(sum[:])

	meta, exists := s.keys[hash]
	if !exists {
		return nil, fmt.Errorf("unauthorized: API key not recognized")
	}

	if meta.IsRevoked {
		return nil, fmt.Errorf("unauthorized: API key has been revoked")
	}

	if time.Now().UTC().After(meta.ExpiresAt) {
		return nil, fmt.Errorf("unauthorized: API key has expired")
	}

	meta.LastUsedAt = time.Now().UTC()
	return meta, nil
}

// RevokeKey immediately invalidates an API key.
func (s *APIKeyService) RevokeKey(orgID, keyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, meta := range s.keys {
		if meta.OrgID == orgID && meta.KeyID == keyID {
			meta.IsRevoked = true
			return nil
		}
	}
	return fmt.Errorf("key '%s' not found for organization '%s'", keyID, orgID)
}

// ListKeys returns all active and revoked keys for an organization.
func (s *APIKeyService) ListKeys(orgID string) []APIKeyMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []APIKeyMetadata
	for _, meta := range s.keys {
		if meta.OrgID == orgID {
			result = append(result, *meta)
		}
	}
	return result
}
