package rate_limit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimit_TokenBucketAndAbuse(t *testing.T) {
	limiter := NewDistributedRateLimiter(RateLimiterConfig{
		StarterRPS:    10.0,
		GrowthRPS:     50.0,
		EnterpriseRPS: 200.0,
	})

	// 1. Normal Requests
	allowed, remaining, err := limiter.Allow("tenant_starter", "STARTER", 1.0)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Greater(t, remaining, 0.0)

	// 2. Consume burst capacity to trigger limit
	allowed, _, _ = limiter.Allow("tenant_starter", "STARTER", 50.0)
	assert.False(t, allowed)

	violations := limiter.GetAbuseViolations("tenant_starter")
	assert.Greater(t, violations, 0)
}
