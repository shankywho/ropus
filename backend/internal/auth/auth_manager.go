package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shankywho/ropus/backend/internal/tenant"
)

// JWTClaims represents customer authentication claims.
type JWTClaims struct {
	UserID    string          `json:"sub"`
	OrgID     string          `json:"org_id"`
	Email     string          `json:"email"`
	Role      tenant.UserRole `json:"role"`
	IssuedAt  int64           `json:"iat"`
	ExpiresAt int64           `json:"exp"`
}

// AuthPlatform provides JWT issuance, session tracking, and RBAC policy enforcement.
type AuthPlatform struct {
	mu          sync.RWMutex
	signingKey  []byte
	sessions    map[string]*JWTClaims
	tokenExpiry time.Duration
}

// NewAuthPlatform initializes the authentication platform.
func NewAuthPlatform(secret string) *AuthPlatform {
	if secret == "" {
		secret = "ropus_enterprise_jwt_secret_key_2026_prod"
	}
	return &AuthPlatform{
		signingKey:  []byte(secret),
		sessions:    make(map[string]*JWTClaims),
		tokenExpiry: 24 * time.Hour,
	}
}

// GenerateToken issues a signed JWT token for a user.
func (a *AuthPlatform) GenerateToken(userID, orgID, email string, role tenant.UserRole) (string, error) {
	now := time.Now().UTC()
	claims := JWTClaims{
		UserID:    userID,
		OrgID:     orgID,
		Email:     email,
		Role:      role,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(a.tokenExpiry).Unix(),
	}

	header := `{"alg":"HS256","typ":"JWT"}`
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(header))

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	data := headerB64 + "." + payloadB64
	sig := a.sign(data)

	token := data + "." + sig

	a.mu.Lock()
	a.sessions[token] = &claims
	a.mu.Unlock()

	return token, nil
}

// ValidateToken verifies token signature, expiration, and returns claims.
func (a *AuthPlatform) ValidateToken(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	data := parts[0] + "." + parts[1]
	expectedSig := a.sign(data)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid token signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	if time.Now().UTC().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// AuthorizeRole ensures caller satisfies required permission level.
func (a *AuthPlatform) AuthorizeRole(callerRole tenant.UserRole, requiredRole tenant.UserRole) bool {
	roleWeights := map[tenant.UserRole]int{
		tenant.RoleOwner:   4,
		tenant.RoleAdmin:   3,
		tenant.RoleAnalyst: 2,
		tenant.RoleViewer:  1,
	}

	return roleWeights[callerRole] >= roleWeights[requiredRole]
}

func (a *AuthPlatform) sign(data string) string {
	mac := hmac.New(sha256.New, a.signingKey)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
