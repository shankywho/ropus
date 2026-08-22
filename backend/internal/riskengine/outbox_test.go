package riskengine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDispatcher struct {
	mu           sync.Mutex
	delivered    []OutboxEvent
	failAttempts int
	callCount    int64
}

func (m *mockDispatcher) Dispatch(ctx context.Context, event OutboxEvent) error {
	count := atomic.AddInt64(&m.callCount, 1)
	m.mu.Lock()
	defer m.mu.Unlock()

	if int(count) <= m.failAttempts {
		return errors.New("transient dispatch failure")
	}

	m.delivered = append(m.delivered, event)
	return nil
}

func TestOutbox_SuccessfulAsyncDelivery(t *testing.T) {
	dispatcher := &mockDispatcher{}
	outbox := NewDurableOutbox(dispatcher, 2, 100)
	defer func() { _ = outbox.FlushAndClose(1 * time.Second) }()

	for i := 0; i < 50; i++ {
		ok := outbox.Enqueue(OutboxEvent{
			EventID:       fmt.Sprintf("evt_%d", i),
			EventType:     "RISK_EVALUATION_RECORDED",
			CorrelationID: "corr_123",
			Payload:       []byte(`{"score":45}`),
		})
		assert.True(t, ok)
	}

	err := outbox.FlushAndClose(2 * time.Second)
	require.NoError(t, err)

	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	assert.Equal(t, 50, len(dispatcher.delivered))
}

func TestOutbox_RetryAndDeadLetterHandling(t *testing.T) {
	// Dispatcher fails first 2 attempts, succeeds on 3rd
	dispatcher := &mockDispatcher{failAttempts: 2}
	outbox := NewDurableOutbox(dispatcher, 1, 100)

	ok := outbox.Enqueue(OutboxEvent{
		EventID:        "evt_retry_01",
		EventType:      "MODEL_PROMOTED",
		MaxRetries:     3,
		InitialBackoff: 5 * time.Millisecond,
	})
	assert.True(t, ok)

	err := outbox.FlushAndClose(1 * time.Second)
	require.NoError(t, err)

	dispatcher.mu.Lock()
	assert.Equal(t, 1, len(dispatcher.delivered), "Event should eventually succeed after retries")
	dispatcher.mu.Unlock()
}

func TestOutbox_BufferFullNonBlocking(t *testing.T) {
	// Dummy blocking dispatcher
	blockCh := make(chan struct{})
	blockingDispatcher := &blockingMockDispatcher{blockCh: blockCh}
	outbox := NewDurableOutbox(blockingDispatcher, 1, 5) // tiny buffer of 5

	// Fill the buffer
	for i := 0; i < 15; i++ {
		_ = outbox.Enqueue(OutboxEvent{EventID: fmt.Sprintf("evt_overflow_%d", i)})
	}

	// Stats should reflect dropped events without blocking or deadlocking
	queued, _, _, dropped, _ := outbox.Stats()
	assert.Greater(t, dropped, int64(0), "Excess events must be dropped non-blockingly")
	assert.LessOrEqual(t, queued, int64(6))

	close(blockCh)
	_ = outbox.FlushAndClose(500 * time.Millisecond)
}

type blockingMockDispatcher struct {
	blockCh chan struct{}
}

func (b *blockingMockDispatcher) Dispatch(ctx context.Context, event OutboxEvent) error {
	<-b.blockCh
	return nil
}
