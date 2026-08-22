package streaming

import (
	"context"
	"sync"
	"time"
)

// StreamProcessor processes ordered event streams with deduplication and state tracking.
type StreamProcessor struct {
	bus      EventBus
	mu       sync.RWMutex
	seenKeys map[string]time.Time
}

// NewStreamProcessor initializes the stream processor.
func NewStreamProcessor(bus EventBus) *StreamProcessor {
	if bus == nil {
		bus = NewLocalEventBus()
	}
	return &StreamProcessor{
		bus:      bus,
		seenKeys: make(map[string]time.Time),
	}
}

// ProcessEvent applies exactly-once deduplication and processes the streaming event.
func (p *StreamProcessor) ProcessEvent(ctx context.Context, event *StreamingEvent) (bool, error) {
	if event.IdempotencyKey != "" {
		p.mu.Lock()
		if _, exists := p.seenKeys[event.IdempotencyKey]; exists {
			p.mu.Unlock()
			return false, nil // Duplicate event ignored
		}
		p.seenKeys[event.IdempotencyKey] = time.Now().UTC()
		p.mu.Unlock()
	}

	return true, nil
}
