package riskengine

import (
	"sync"
	"time"
)

// TenantBucket implements a thread-safe token bucket rate limiter for an individual tenant.
type TenantBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func newTenantBucket(rate, burst float64) *TenantBucket {
	return &TenantBucket{
		tokens:     burst,
		capacity:   burst,
		refillRate: rate,
		lastRefill: time.Now().UTC(),
	}
}

func (tb *TenantBucket) allow(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now().UTC()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// Refill tokens
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	reqTokens := float64(n)
	if tb.tokens >= reqTokens {
		tb.tokens -= reqTokens
		return true
	}
	return false
}

// TenantRateLimiterConfig sets rate and burst limits for global and per-tenant traffic.
type TenantRateLimiterConfig struct {
	DefaultRatePerSec float64 // Default tokens/sec per tenant
	DefaultBurstCap   float64 // Default burst capacity per tenant
	GlobalRatePerSec  float64 // Global platform capacity cap
	GlobalBurstCap    float64 // Global platform burst cap
	CleanupInterval   time.Duration
}

// DefaultRateLimiterConfig returns standard production defaults.
func DefaultRateLimiterConfig() TenantRateLimiterConfig {
	return TenantRateLimiterConfig{
		DefaultRatePerSec: 5000.0, // 5,000 req/sec per tenant
		DefaultBurstCap:   10000.0, // 10,000 burst per tenant
		GlobalRatePerSec:  50000.0, // 50,000 req/sec global
		GlobalBurstCap:    100000.0, // 100,000 burst global
		CleanupInterval:   10 * time.Minute,
	}
}

// TenantRateLimiter enforces token-bucket rate limits per tenant and globally.
type TenantRateLimiter struct {
	config  TenantRateLimiterConfig
	global  *TenantBucket
	mu      sync.RWMutex
	tenants map[string]*TenantBucket
	enabled bool
}

// NewTenantRateLimiter initializes a TenantRateLimiter.
func NewTenantRateLimiter(cfg TenantRateLimiterConfig) *TenantRateLimiter {
	return &TenantRateLimiter{
		config:  cfg,
		global:  newTenantBucket(cfg.GlobalRatePerSec, cfg.GlobalBurstCap),
		tenants: make(map[string]*TenantBucket),
		enabled: true,
	}
}

// SetEnabled toggles rate limiting.
func (rl *TenantRateLimiter) SetEnabled(enabled bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.enabled = enabled
}

// Allow returns true if the tenant request conforms to rate limits.
func (rl *TenantRateLimiter) Allow(tenantID string) bool {
	return rl.AllowN(tenantID, 1)
}

// AllowN checks both global and tenant-specific token buckets.
func (rl *TenantRateLimiter) AllowN(tenantID string, n int) bool {
	rl.mu.RLock()
	if !rl.enabled {
		rl.mu.RUnlock()
		return true
	}
	rl.mu.RUnlock()

	// 1. Check Global Platform Rate
	if rl.global != nil && !rl.global.allow(n) {
		return false
	}

	// 2. Resolve/Create Tenant Bucket
	if tenantID == "" {
		tenantID = "default"
	}

	rl.mu.RLock()
	bucket, exists := rl.tenants[tenantID]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		bucket, exists = rl.tenants[tenantID]
		if !exists {
			bucket = newTenantBucket(rl.config.DefaultRatePerSec, rl.config.DefaultBurstCap)
			rl.tenants[tenantID] = bucket
		}
		rl.mu.Unlock()
	}

	return bucket.allow(n)
}
