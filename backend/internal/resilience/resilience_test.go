package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResilience_CircuitBreakerAndHealth(t *testing.T) {
	// 1. Circuit Breaker Test
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		Timeout:          50 * time.Millisecond,
	})

	assert.Equal(t, StateClosed, cb.GetState())

	// Fail twice -> Trip to OPEN
	_ = cb.Execute(func() error { return errors.New("error 1") })
	_ = cb.Execute(func() error { return errors.New("error 2") })
	assert.Equal(t, StateOpen, cb.GetState())

	// Open -> Fast fail
	err := cb.Execute(func() error { return nil })
	assert.Equal(t, ErrCircuitOpen, err)

	// Wait timeout -> Half Open -> Closed on success
	time.Sleep(60 * time.Millisecond)
	err = cb.Execute(func() error { return nil })
	require.NoError(t, err)

	// 2. Health Manager & Fallback Queue Test
	hm := NewHealthManager()
	hm.RegisterChecker("database", func(ctx context.Context) error { return nil })
	hm.RegisterChecker("kafka", func(ctx context.Context) error { return errors.New("broker timeout") })

	report := hm.RunHealthChecks(context.Background())
	assert.Equal(t, "DEGRADED", report.OverallStatus)
	assert.True(t, report.Services["database"].IsHealthy)
	assert.False(t, report.Services["kafka"].IsHealthy)

	// Buffer fallback
	hm.BufferFallbackEvent(map[string]interface{}{"tx_id": "tx_fb_01"})
	flushed := hm.FlushFallbackBuffer()
	assert.Equal(t, 1, len(flushed))
}
