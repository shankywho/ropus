# Component 10: Event Streaming & Apache Kafka Architecture

---

## 1. Why It Exists
Synchronous payment evaluation must complete in $< 10\text{ms}$. However, auxiliary downstream workloads—such as updating 24-hour sliding feature stores, archiving compliance audit logs, training offline ML models, and delivering customer webhooks—require substantial processing time.

The **Event Streaming Subsystem** (`backend/internal/streaming/`, `backend/internal/events/`) leverages **Apache Kafka** to decouple high-speed synchronous risk decisioning from asynchronous downstream event processing.

---

## 2. Topic Topology & Partitioning

```text
[ Inbound Transaction ] ──> [ Evaluated in < 2ms ]
                                    │
                                    ▼ (Publish Event)
                     ┌──────────────────────────────┐
                     │ Topic: transactions.evaluated│ (Key: tenant_id, 16 Partitions)
                     └──────────────┬───────────────┘
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         ▼ (Consumer Group A)       ▼ (Consumer Group B)       ▼ (Consumer Group C)
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│ Feature Store    │       │ Compliance Audit │       │ Webhook Egress   │
│ Aggregator       │       │ Archiver         │       │ Dispatcher       │
│ Updates Redis    │       │ Writes to S3 WAL │       │ Signs & Dispatches│
│ velocity counts  │       │ cold storage     │       │ HMAC HTTP hooks  │
└──────────────────┘       └──────────────────┘       └──────────────────┘
```

---

## 3. Dead-Letter Queue (DLQ) & Fallback Architecture

If a downstream consumer fails to process an event after 3 retries (e.g. malformed metadata or database deadlocks), the event is routed to `dlq.failed_events` for manual inspection:

```go
type DeadLetterEvent struct {
    EventID       string                 `json:"event_id"`
    OriginalTopic string                 `json:"original_topic"`
    Payload       map[string]interface{} `json:"payload"`
    ErrorMessage  string                 `json:"error_message"`
    FailedAt      time.Time              `json:"failed_at"`
    RetryCount    int                    `json:"retry_count"`
}
```

---

## 4. Source Code Map
- [`backend/internal/streaming/kafka_producer.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/streaming/kafka_producer.go): Partitioned Kafka producer with batching and backpressure.
- [`backend/internal/events/event_bus.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/events/event_bus.go): In-memory pub/sub event bus with consumer group fanout.

---

## 5. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Publishes decision events.
- [Component 12: Resilience & Fallbacks](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/12-resilience-circuit-breaker.md) — Buffers events locally if Kafka brokers are unreachable.
- [Component 13: Webhooks](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/13-webhooks.md) — Consumes events to dispatch customer webhooks.
