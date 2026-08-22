package riskengine

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_BurstAndThrottle(t *testing.T) {
	cfg := TenantRateLimiterConfig{
		DefaultRatePerSec: 10.0, // 10 tokens/sec
		DefaultBurstCap:   5.0,  // Burst cap 5
		GlobalRatePerSec:  1000.0,
		GlobalBurstCap:    1000.0,
	}
	limiter := NewTenantRateLimiter(cfg)

	tenantID := "tenant_burst_test"

	// 5 requests should pass burst capacity
	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Allow(tenantID), "Request %d should be allowed", i)
	}

	// 6th request immediately after should be throttled
	assert.False(t, limiter.Allow(tenantID), "6th immediate request should be throttled")

	// Wait 250ms -> should refill ~2.5 tokens
	time.Sleep(250 * time.Millisecond)
	assert.True(t, limiter.Allow(tenantID), "Request after refill should be allowed")
}

func TestRateLimiter_TenantIsolation(t *testing.T) {
	cfg := TenantRateLimiterConfig{
		DefaultRatePerSec: 5.0,
		DefaultBurstCap:   2.0,
		GlobalRatePerSec:  100.0,
		GlobalBurstCap:    100.0,
	}
	limiter := NewTenantRateLimiter(cfg)

	tenantA := "tenant_A"
	tenantB := "tenant_B"

	// Exhaust Tenant A
	assert.True(t, limiter.Allow(tenantA))
	assert.True(t, limiter.Allow(tenantA))
	assert.False(t, limiter.Allow(tenantA), "Tenant A exhausted")

	// Tenant B should still have its full quota
	assert.True(t, limiter.Allow(tenantB), "Tenant B should not be impacted by Tenant A")
	assert.True(t, limiter.Allow(tenantB), "Tenant B second request should succeed")
}

func TestRateLimiter_HighConcurrency(t *testing.T) {
	cfg := TenantRateLimiterConfig{
		DefaultRatePerSec: 10000.0,
		DefaultBurstCap:   10000.0,
		GlobalRatePerSec:  100000.0,
		GlobalBurstCap:    100000.0,
	}
	limiter := NewTenantRateLimiter(cfg)

	var wg sync.WaitGroup
	var allowedCount int64

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				if limiter.Allow("concurrent_tenant") {
					wg.Add(0)
				}
			}
		}(i)
	}
	wg.Wait()
	_ = allowedCount
}
