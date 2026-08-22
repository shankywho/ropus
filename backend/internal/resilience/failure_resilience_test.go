package resilience

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResilience_DegradedModeFallbacks(t *testing.T) {
	hm := NewHealthManager()

	// Register healthy DB, failed Kafka, failed LLM
	hm.RegisterChecker("postgresql", func(ctx context.Context) error { return nil })
	hm.RegisterChecker("kafka_cluster", func(ctx context.Context) error { return errors.New("connection refused: 9092") })
	hm.RegisterChecker("llm_gateway", func(ctx context.Context) error { return errors.New("anthropic api timeout: 504") })

	report := hm.RunHealthChecks(context.Background())
	assert.Equal(t, "DEGRADED", report.OverallStatus)
	assert.True(t, report.Services["postgresql"].IsHealthy)
	assert.False(t, report.Services["kafka_cluster"].IsHealthy)
	assert.False(t, report.Services["llm_gateway"].IsHealthy)

	// Verify core risk decision continues while buffering Kafka event
	hm.BufferFallbackEvent(map[string]interface{}{
		"event":         "risk.decision.created",
		"decision_id":   "dec_fallback_9918",
		"decision":      "BLOCK",
		"risk_score":    0.95,
	})

	buffered := hm.FlushFallbackBuffer()
	require.Equal(t, 1, len(buffered))
	assert.Equal(t, "dec_fallback_9918", buffered[0]["decision_id"])
}
