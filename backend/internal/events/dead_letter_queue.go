package events

import (
	"fmt"
	"sync"
	"time"
)

// DeadLetterEntry represents a failed message trapped in the DLQ.
type DeadLetterEntry struct {
	DLQID         string       `json:"dlq_id"`
	OriginalEvent *StreamEvent `json:"original_event"`
	FailureReason string       `json:"failure_reason"`
	AttemptCount  int          `json:"attempt_count"`
	TrappedAt     time.Time    `json:"trapped_at"`
}

// DeadLetterQueueEngine manages failed message isolation and replay.
type DeadLetterQueueEngine struct {
	mu      sync.RWMutex
	entries []*DeadLetterEntry
}

// NewDeadLetterQueueEngine initializes the DLQ engine.
func NewDeadLetterQueueEngine() *DeadLetterQueueEngine {
	return &DeadLetterQueueEngine{
		entries: make([]*DeadLetterEntry, 0),
	}
}

// RouteToDeadLetter captures a failed stream message into the DLQ.
func (d *DeadLetterQueueEngine) RouteToDeadLetter(ev *StreamEvent, reason string) *DeadLetterEntry {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	entry := &DeadLetterEntry{
		DLQID:         fmt.Sprintf("dlq_%d", now.UnixNano()),
		OriginalEvent: ev,
		FailureReason: reason,
		AttemptCount:  3,
		TrappedAt:     now,
	}

	d.entries = append(d.entries, entry)
	return entry
}

// ListDLQ retrieves all trapped messages for inspection or replay.
func (d *DeadLetterQueueEngine) ListDLQ() []*DeadLetterEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()

	res := make([]*DeadLetterEntry, len(d.entries))
	copy(res, d.entries)
	return res
}
