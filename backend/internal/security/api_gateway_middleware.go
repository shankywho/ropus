package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// SecurityContext contains validated tenant and caller attributes.
type SecurityContext struct {
	TenantID      string
	CallerRole    string
	IsRateLimited bool
	IsIPBlocked   bool
	Authenticated bool
}

// EnterpriseSecurityGateway enforces Zero-Trust API protection, rate limiting, and request signing.
type EnterpriseSecurityGateway struct {
	mu            sync.RWMutex
	blockedIPs    map[string]bool
	tenantLimits  map[string]int // tenantID -> max requests per minute
	requestCounts map[string]int // tenantID:minute -> current count
}

// NewEnterpriseSecurityGateway initializes the enterprise security gateway.
func NewEnterpriseSecurityGateway() *EnterpriseSecurityGateway {
	return &EnterpriseSecurityGateway{
		blockedIPs: map[string]bool{
			"198.51.100.44": true, // Bulletproof proxy
			"203.0.113.88":  true, // Known malicious tor exit node
		},
		tenantLimits:  make(map[string]int),
		requestCounts: make(map[string]int),
	}
}

// VerifyRequestSignature verifies HMAC-SHA256 request signatures for bank integration.
func (g *EnterpriseSecurityGateway) VerifyRequestSignature(payload []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ValidateRequest checks IP blocklists and enforces tenant rate quotas.
func (g *EnterpriseSecurityGateway) ValidateRequest(tenantID, clientIP string) (*SecurityContext, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 1. IP Blocklist Check
	if g.blockedIPs[clientIP] {
		return &SecurityContext{
			TenantID:      tenantID,
			IsIPBlocked:   true,
			Authenticated: false,
		}, fmt.Errorf("security violation: client IP '%s' is on global blocklist", clientIP)
	}

	// 2. Rate Limiting Check
	limit, hasLimit := g.tenantLimits[tenantID]
	if !hasLimit {
		limit = 10000 // Default 10k req/min per tenant
	}

	nowMin := time.Now().UTC().Format("2006-01-02-15-04")
	bucketKey := fmt.Sprintf("%s:%s", tenantID, nowMin)
	g.requestCounts[bucketKey]++

	if g.requestCounts[bucketKey] > limit {
		return &SecurityContext{
			TenantID:      tenantID,
			IsRateLimited: true,
			Authenticated: true,
		}, fmt.Errorf("rate limit exceeded for tenant '%s' (%d > %d req/min)", tenantID, g.requestCounts[bucketKey], limit)
	}

	return &SecurityContext{
		TenantID:      tenantID,
		CallerRole:    "TENANT_SERVICE",
		IsRateLimited: false,
		IsIPBlocked:   false,
		Authenticated: true,
	}, nil
}
