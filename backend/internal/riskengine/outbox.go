package riskengine

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// OutboxEvent represents a structured domain event queued for reliable asynchronous dispatch.
type OutboxEvent struct {
	EventID        string            `json:"event_id"`
	EventType      string            `json:"event_type"`
	Timestamp      time.Time         `json:"timestamp"`
	CorrelationID  string            `json:"correlation_id"`
	Payload        []byte            `json:"payload"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	RetryCount     int               `json:"retry_count"`
	MaxRetries     int               `json:"max_retries"`
	InitialBackoff time.Duration     `json:"initial_backoff"`
}

// EventDispatcher is the sink handler for dispatching outbox events (e.g. Kafka, ClickHouse, Webhooks).
type EventDispatcher interface {
	Dispatch(ctx context.Context, event OutboxEvent) error
}

// DurableOutbox delivers domain events asynchronously with bounded queueing, retries, and dead-letter safety.
type DurableOutbox struct {
	queue          chan OutboxEvent
	dispatcher     EventDispatcher
	workers        int
	maxQueueSize   int
	isClosed       int32
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	deadLetterLog  []OutboxEvent
	deadLetterMu   sync.Mutex
	queuedCount    int64
	deliveredCount int64
	failedCount    int64
	droppedCount   int64
}

// NewDurableOutbox initializes a thread-safe, non-blocking outbox.
func NewDurableOutbox(dispatcher EventDispatcher, workers, maxQueueSize int) *DurableOutbox {
	if workers <= 0 {
		workers = 4
	}
	if maxQueueSize <= 0 {
		maxQueueSize = 10000
	}

	ctx, cancel := context.WithCancel(context.Background())
	o := &DurableOutbox{
		queue:        make(chan OutboxEvent, maxQueueSize),
		dispatcher:   dispatcher,
		workers:      workers,
		maxQueueSize: maxQueueSize,
		ctx:          ctx,
		cancel:       cancel,
	}

	o.startWorkers()
	return o
}

func (o *DurableOutbox) startWorkers() {
	for i := 0; i < o.workers; i++ {
		o.wg.Add(1)
		go o.workerLoop()
	}
}

func (o *DurableOutbox) workerLoop() {
	defer o.wg.Done()
	for {
		select {
		case <-o.ctx.Done():
			return
		case event, ok := <-o.queue:
			if !ok {
				return
			}
			o.processWithRetries(event)
		}
	}
}

// processWithRetries executes delivery with exponential backoff.
func (o *DurableOutbox) processWithRetries(event OutboxEvent) {
	if event.MaxRetries <= 0 {
		event.MaxRetries = 3
	}
	if event.InitialBackoff <= 0 {
		event.InitialBackoff = 20 * time.Millisecond
	}

	var lastErr error
	for attempt := 0; attempt <= event.MaxRetries; attempt++ {
		if o.ctx.Err() != nil {
			return
		}

		ctx, cancel := context.WithTimeout(o.ctx, 5*time.Second)
		err := o.dispatcher.Dispatch(ctx, event)
		cancel()

		if err == nil {
			atomic.AddInt64(&o.deliveredCount, 1)
			return
		}

		lastErr = err
		if attempt < event.MaxRetries {
			backoff := time.Duration(float64(event.InitialBackoff) * math.Pow(2, float64(attempt)))
			select {
			case <-time.After(backoff):
			case <-o.ctx.Done():
				return
			}
		}
	}

	// Delivery exhausted retries: push to dead-letter storage
	atomic.AddInt64(&o.failedCount, 1)
	o.deadLetterMu.Lock()
	if len(o.deadLetterLog) < 1000 {
		o.deadLetterLog = append(o.deadLetterLog, event)
	}
	o.deadLetterMu.Unlock()
	log.Printf("[OUTBOX_DEAD_LETTER] Event %s (%s) dropped after %d retries: %v", event.EventID, event.EventType, event.MaxRetries, lastErr)
}

// Enqueue adds an event to the outbox queue. Returns false if the queue is full (strictly non-blocking).
func (o *DurableOutbox) Enqueue(event OutboxEvent) bool {
	if atomic.LoadInt32(&o.isClosed) == 1 {
		return false
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	select {
	case o.queue <- event:
		atomic.AddInt64(&o.queuedCount, 1)
		return true
	default:
		// Bounded queue full: drop rather than block synchronous inference path
		atomic.AddInt64(&o.droppedCount, 1)
		return false
	}
}

// QueueDepth returns the current number of pending items in the channel.
func (o *DurableOutbox) QueueDepth() int {
	return len(o.queue)
}

// Stats returns operational counters.
func (o *DurableOutbox) Stats() (queued, delivered, failed, dropped int64, depth int) {
	return atomic.LoadInt64(&o.queuedCount),
		atomic.LoadInt64(&o.deliveredCount),
		atomic.LoadInt64(&o.failedCount),
		atomic.LoadInt64(&o.droppedCount),
		len(o.queue)
}

// FlushAndClose drains remaining in-flight events up to the given timeout.
func (o *DurableOutbox) FlushAndClose(timeout time.Duration) error {
	if !atomic.CompareAndSwapInt32(&o.isClosed, 0, 1) {
		return nil // Already closed
	}

	close(o.queue)

	doneCh := make(chan struct{})
	go func() {
		o.wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		o.cancel()
		return nil
	case <-time.After(timeout):
		o.cancel()
		return fmt.Errorf("outbox flush timed out after %v", timeout)
	}
}
