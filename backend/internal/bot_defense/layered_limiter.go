package bot_defense

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// LayeredLimiterConfig defines token bucket limits per operational layer.
type LayeredLimiterConfig struct {
	// Tenant Limits (Req / sec)
	TenantStarterRPS    float64
	TenantGrowthRPS     float64
	TenantEnterpriseRPS float64

	// Entity Granular Limits (Req / min)
	IPMaxRPM      float64 // Default: 60 req/min per IP
	IPBurst       float64 // Default: 120 burst capacity
	DeviceMaxRPM  float64 // Default: 30 req/min per Device
	DeviceBurst   float64 // Default: 60 burst capacity
	AccountMaxRPM float64 // Default: 20 req/min per Account
	AccountBurst  float64 // Default: 40 burst capacity
	CardMaxRPM    float64 // Default: 15 req/min per Card Hash
	CardBurst     float64 // Default: 30 burst capacity
}

// DefaultLayeredLimiterConfig returns hardened production defaults.
func DefaultLayeredLimiterConfig() LayeredLimiterConfig {
	return LayeredLimiterConfig{
		TenantStarterRPS:    100.0,
		TenantGrowthRPS:     500.0,
		TenantEnterpriseRPS: 5000.0,
		IPMaxRPM:            60.0,
		IPBurst:             120.0,
		DeviceMaxRPM:        30.0,
		DeviceBurst:         60.0,
		AccountMaxRPM:       20.0,
		AccountBurst:        40.0,
		CardMaxRPM:          15.0,
		CardBurst:           30.0,
	}
}

// TokenBucket holds state for an individual rate-limited entity key.
type TokenBucket struct {
	Capacity       float64
	Tokens         float64
	RefillRate     float64 // tokens per second
	LastRefillTime time.Time
}

// LayeredLimiter provides multi-dimensional rate limiting across Tenant, IP, Device, Account, and Card.
type LayeredLimiter struct {
	mu      sync.Mutex
	buckets map[string]*TokenBucket
	cfg     LayeredLimiterConfig
}

// NewLayeredLimiter initializes the multi-dimensional token bucket limiter.
func NewLayeredLimiter(cfg LayeredLimiterConfig) *LayeredLimiter {
	return &LayeredLimiter{
		buckets: make(map[string]*TokenBucket),
		cfg:     cfg,
	}
}

// CheckLimit checks if a token bucket allows a request for a specific key.
func (l *LayeredLimiter) checkBucket(key string, capacity, refillRate float64, tokensReq float64, now time.Time) bool {
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			Capacity:       capacity,
			Tokens:         capacity - tokensReq,
			RefillRate:     refillRate,
			LastRefillTime: now,
		}
		l.buckets[key] = bucket
		return true
	}

	// Refill tokens
	elapsed := now.Sub(bucket.LastRefillTime).Seconds()
	bucket.Tokens = math.Min(bucket.Capacity, bucket.Tokens+(elapsed*bucket.RefillRate))
	bucket.LastRefillTime = now

	if bucket.Tokens >= tokensReq {
		bucket.Tokens -= tokensReq
		return true
	}

	return false
}

// EvaluateLimits checks all 5 tiers and returns breach flags and violation details.
func (l *LayeredLimiter) EvaluateLimits(ctx *BotDefenseContext) (signals BotSignals, breachedRules []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()

	// 1. Tenant Plan Tier Limiting
	tenantRPS := l.cfg.TenantStarterRPS
	if ctx.PlanTier == "GROWTH" {
		tenantRPS = l.cfg.TenantGrowthRPS
	} else if ctx.PlanTier == "ENTERPRISE" {
		tenantRPS = l.cfg.TenantEnterpriseRPS
	}
	tenantKey := fmt.Sprintf("tenant:%s", ctx.TenantID)
	if !l.checkBucket(tenantKey, tenantRPS*2.0, tenantRPS, 1.0, now) {
		signals.TenantLimited = true
		breachedRules = append(breachedRules, fmt.Sprintf("Tenant quota exceeded for %s (limit: %.0f req/s)", ctx.TenantID, tenantRPS))
	}

	// 2. IP Address Limiting
	if ctx.IPAddress != "" {
		ipKey := fmt.Sprintf("ip:%s", ctx.IPAddress)
		ipRPS := l.cfg.IPMaxRPM / 60.0
		if !l.checkBucket(ipKey, l.cfg.IPBurst, ipRPS, 1.0, now) {
			signals.IPLimited = true
			breachedRules = append(breachedRules, fmt.Sprintf("IP rate limit exceeded for %s (limit: %.0f req/min)", ctx.IPAddress, l.cfg.IPMaxRPM))
		}
	}

	// 3. Device Identifier Limiting
	if ctx.DeviceFingerprint != "" {
		devKey := fmt.Sprintf("dev:%s", ctx.DeviceFingerprint)
		devRPS := l.cfg.DeviceMaxRPM / 60.0
		if !l.checkBucket(devKey, l.cfg.DeviceBurst, devRPS, 1.0, now) {
			signals.DeviceLimited = true
			breachedRules = append(breachedRules, fmt.Sprintf("Device rate limit exceeded for %s (limit: %.0f req/min)", ctx.DeviceFingerprint, l.cfg.DeviceMaxRPM))
		}
	}

	// 4. Customer Account Limiting
	if ctx.AccountID != "" {
		accKey := fmt.Sprintf("acc:%s", ctx.AccountID)
		accRPS := l.cfg.AccountMaxRPM / 60.0
		if !l.checkBucket(accKey, l.cfg.AccountBurst, accRPS, 1.0, now) {
			signals.AccountLimited = true
			breachedRules = append(breachedRules, fmt.Sprintf("Account rate limit exceeded for %s (limit: %.0f req/min)", ctx.AccountID, l.cfg.AccountMaxRPM))
		}
	}

	// 5. Payment Card Hash Limiting
	if ctx.CardHash != "" {
		cardKey := fmt.Sprintf("card:%s", ctx.CardHash)
		cardRPS := l.cfg.CardMaxRPM / 60.0
		if !l.checkBucket(cardKey, l.cfg.CardBurst, cardRPS, 1.0, now) {
			signals.CardLimited = true
			breachedRules = append(breachedRules, fmt.Sprintf("Payment instrument velocity exceeded for card hash %s (limit: %.0f req/min)", ctx.CardHash[:min(8, len(ctx.CardHash))], l.cfg.CardMaxRPM))
		}
	}

	return signals, breachedRules
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
