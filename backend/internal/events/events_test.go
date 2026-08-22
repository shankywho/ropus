package events

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvents_KafkaAndDLQ(t *testing.T) {
	ctx := context.Background()
	dlq := NewDeadLetterQueueEngine()
	stream := NewEventStreamingEngine([]string{"localhost:9092"}, dlq)

	receivedCount := 0
	stream.Subscribe("transactions.evaluated", func(ctx context.Context, ev *StreamEvent) error {
		receivedCount++
		return nil
	})

	stream.Subscribe("poison.topic", func(ctx context.Context, ev *StreamEvent) error {
		return fmt.Errorf("simulated consumer parsing failure")
	})

	// 1. Successful message dispatch
	payload := map[string]interface{}{"tx_id": "tx_123", "amount": 500.0}
	ev, err := stream.Publish(ctx, "transactions.evaluated", "tx_123", payload)
	require.NoError(t, err)
	assert.NotEmpty(t, ev.EventID)

	// 2. Poisoned message -> DLQ routing
	_, err = stream.Publish(ctx, "poison.topic", "tx_poison", map[string]interface{}{"corrupt": true})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	// Verify DLQ captured poison event
	dlqEntries := dlq.ListDLQ()
	assert.Equal(t, 1, len(dlqEntries))
	assert.Equal(t, "simulated consumer parsing failure", dlqEntries[0].FailureReason)
}
