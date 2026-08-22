package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StreamEvent represents an atomic streaming message payload.
type StreamEvent struct {
	EventID   string                 `json:"event_id"`
	Topic     string                 `json:"topic"`
	Key       string                 `json:"key"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventHandler defines the consumer callback function signature.
type EventHandler func(ctx context.Context, event *StreamEvent) error

// EventStreamingEngine abstracts Kafka, Redis Streams, and memory-backed streaming backbones.
type EventStreamingEngine struct {
	mu           sync.RWMutex
	brokers      []string
	subscribers  map[string][]EventHandler
	eventBuffer  []*StreamEvent
	dlqEngine    *DeadLetterQueueEngine
	isKafkaLive  bool
}

// NewEventStreamingEngine initializes the streaming backbone.
func NewEventStreamingEngine(brokers []string, dlq *DeadLetterQueueEngine) *EventStreamingEngine {
	if dlq == nil {
		dlq = NewDeadLetterQueueEngine()
	}
	return &EventStreamingEngine{
		brokers:     brokers,
		subscribers: make(map[string][]EventHandler),
		eventBuffer: make([]*StreamEvent, 0),
		dlqEngine:   dlq,
		isKafkaLive: len(brokers) > 0,
	}
}

// Publish dispatches an event into the target topic.
func (e *EventStreamingEngine) Publish(ctx context.Context, topic, key string, payload map[string]interface{}) (*StreamEvent, error) {
	now := time.Now().UTC()
	eventID := fmt.Sprintf("evt_%d_%s", now.UnixNano(), key)

	event := &StreamEvent{
		EventID:   eventID,
		Topic:     topic,
		Key:       key,
		Payload:   payload,
		Timestamp: now,
	}

	e.mu.Lock()
	e.eventBuffer = append(e.eventBuffer, event)
	handlers := append([]EventHandler(nil), e.subscribers[topic]...)
	e.mu.Unlock()

	// Dispatch to subscribers asynchronously with error handling & DLQ routing
	for _, handler := range handlers {
		go func(h EventHandler, ev *StreamEvent) {
			if err := h(context.Background(), ev); err != nil {
				e.dlqEngine.RouteToDeadLetter(ev, err.Error())
			}
		}(handler, event)
	}

	return event, nil
}

// Subscribe registers a consumer callback to a specific topic.
func (e *EventStreamingEngine) Subscribe(topic string, handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.subscribers[topic] = append(e.subscribers[topic], handler)
}

// GetBufferedEvents retrieves recent streaming events.
func (e *EventStreamingEngine) GetBufferedEvents() []*StreamEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res := make([]*StreamEvent, len(e.eventBuffer))
	copy(res, e.eventBuffer)
	return res
}
