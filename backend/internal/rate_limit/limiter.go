package rate_limit

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// TokenBucket represents an in-memory token bucket for a tenant.
type TokenBucket struct {
	Capacity       float64
	Tokens         float64
	RefillRate     float64 // tokens per second
	LastRefillTime time.Time
}

// RateLimiterConfig holds default rate settings per tier.
type RateLimiterConfig struct {
	StarterRPS    float64
	GrowthRPS     float64
	EnterpriseRPS float64
}

// DistributedRateLimiter provides token-bucket rate limiting and abuse mitigation.
type DistributedRateLimiter struct {
	mu           sync.Mutex
	buckets      map[string]*TokenBucket
	abuseTracker map[string]int // IP/Tenant -> Violation count
	cfg          RateLimiterConfig
}

// NewDistributedRateLimiter initializes the rate limiter.
func NewDistributedRateLimiter(cfg RateLimiterConfig) *DistributedRateLimiter {
	if cfg.StarterRPS <= 0 {
		cfg.StarterRPS = 100.0
	}
	if cfg.GrowthRPS <= 0 {
		cfg.GrowthRPS = 500.0
	}
	if cfg.EnterpriseRPS <= 0 {
		cfg.EnterpriseRPS = 5000.0
	}

	return &DistributedRateLimiter{
		buckets:      make(map[string]*TokenBucket),
		abuseTracker: make(map[string]int),
		cfg:          cfg,
	}
}

// Allow checks if a request is permitted under the tenant's tier quota.
func (r *DistributedRateLimiter) Allow(tenantID, planTier string, tokensRequested float64) (bool, float64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rps := r.cfg.StarterRPS
	if planTier == "GROWTH" {
		rps = r.cfg.GrowthRPS
	} else if planTier == "ENTERPRISE" {
		rps = r.cfg.EnterpriseRPS
	}

	now := time.Now()
	bucket, exists := r.buckets[tenantID]
	if !exists {
		bucket = &TokenBucket{
			Capacity:       rps * 2.0, // Allow 2x burst
			Tokens:         rps * 2.0,
			RefillRate:     rps,
			LastRefillTime: now,
		}
		r.buckets[tenantID] = bucket
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(bucket.LastRefillTime).Seconds()
	bucket.Tokens = math.Min(bucket.Capacity, bucket.Tokens+(elapsed*bucket.RefillRate))
	bucket.LastRefillTime = now

	if bucket.Tokens >= tokensRequested {
		bucket.Tokens -= tokensRequested
		return true, bucket.Tokens, nil
	}

	// Rate limit exceeded -> Track abuse
	r.abuseTracker[tenantID]++
	return false, bucket.Tokens, fmt.Errorf("rate limit exceeded for tenant '%s' (quota: %.0f req/s)", tenantID, rps)
}

// GetAbuseViolations returns violation count for anomaly monitoring.
func (r *DistributedRateLimiter) GetAbuseViolations(tenantID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.abuseTracker[tenantID]
}
