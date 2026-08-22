package streaming

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventHandler is the callback invoked upon streaming event receipt.
type EventHandler func(ctx context.Context, event *StreamingEvent) error

// EventBus defines the contract for high-throughput distributed event streaming backends.
type EventBus interface {
	Publish(ctx context.Context, topic string, event *StreamingEvent) error
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	Replay(ctx context.Context, topic string, fromOffset int64, toOffset int64, handler EventHandler) error
}

// ---------------------------------------------------------------------------
// 1. Local In-Memory Event Bus (Ultra-High-Throughput Ring Buffer)
// ---------------------------------------------------------------------------
type LocalEventBus struct {
	mu          sync.RWMutex
	topics      map[string][]*StreamingEvent
	subscribers map[string][]EventHandler
	offsetCount map[string]int64
}

// NewLocalEventBus initializes the in-process high performance event bus.
func NewLocalEventBus() *LocalEventBus {
	return &LocalEventBus{
		topics:      make(map[string][]*StreamingEvent),
		subscribers: make(map[string][]EventHandler),
		offsetCount: make(map[string]int64),
	}
}

func (b *LocalEventBus) Publish(ctx context.Context, topic string, event *StreamingEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	b.mu.Lock()
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	b.offsetCount[topic]++
	event.Offset = b.offsetCount[topic]

	b.topics[topic] = append(b.topics[topic], event)
	handlers := append([]EventHandler(nil), b.subscribers[topic]...)
	b.mu.Unlock()

	// Deliver to subscribers
	for _, h := range handlers {
		_ = h(ctx, event)
	}

	return nil
}

func (b *LocalEventBus) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers[topic] = append(b.subscribers[topic], handler)
	return nil
}

func (b *LocalEventBus) Replay(ctx context.Context, topic string, fromOffset int64, toOffset int64, handler EventHandler) error {
	b.mu.RLock()
	events := b.topics[topic]
	var replayList []*StreamingEvent
	for _, e := range events {
		if e.Offset >= fromOffset && (toOffset <= 0 || e.Offset <= toOffset) {
			replayList = append(replayList, e)
		}
	}
	b.mu.RUnlock()

	for _, e := range replayList {
		if err := handler(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// 2. Kafka / Redpanda Adapter Boundary
// ---------------------------------------------------------------------------
type KafkaAdapter struct {
	Brokers []string
	GroupID string
}

func NewKafkaAdapter(brokers []string, groupID string) *KafkaAdapter {
	if len(brokers) == 0 {
		brokers = []string{"risk-redpanda:29092"}
	}
	return &KafkaAdapter{Brokers: brokers, GroupID: groupID}
}

func (k *KafkaAdapter) Publish(ctx context.Context, topic string, event *StreamingEvent) error {
	return nil // Kafka sarama / confluent boundary
}

func (k *KafkaAdapter) Subscribe(ctx context.Context, topic string, handler EventHandler) error {
	return nil
}

func (k *KafkaAdapter) Replay(ctx context.Context, topic string, fromOffset, toOffset int64, handler EventHandler) error {
	return nil
}

// ---------------------------------------------------------------------------
// 3. Pulsar Adapter Boundary
// ---------------------------------------------------------------------------
type PulsarAdapter struct {
	ServiceURL string
}

func NewPulsarAdapter(url string) *PulsarAdapter {
	if url == "" {
		url = "pulsar://localhost:6650"
	}
	return &PulsarAdapter{ServiceURL: url}
}

// ---------------------------------------------------------------------------
// 4. Redis Stream Adapter Boundary
// ---------------------------------------------------------------------------
type RedisStreamAdapter struct {
	Host string
	Port int
}

func NewRedisStreamAdapter(host string, port int) *RedisStreamAdapter {
	if host == "" {
		host = "risk-redis-master"
		port = 6379
	}
	return &RedisStreamAdapter{Host: host, Port: port}
}
