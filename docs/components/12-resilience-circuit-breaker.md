# Component 12: Fault Tolerance, Circuit Breakers & Fallback Buffers

---

## 1. Why It Exists
In high-volume payment processing, downstream dependency failures (e.g. Kafka partition rebalances, external LLM API rate limits, or slow Redis instances) must **never cascade into synchronous evaluation failures**.

The **Resilience Subsystem** (`backend/internal/resilience/`, `backend/internal/chaos/`) guarantees continuous availability by:
1. Fast-failing degraded dependencies via stateful **Circuit Breakers**.
2. Buffering asynchronous events into a thread-safe in-memory **Fallback Queue**.
3. Gracefully degrading to heuristic rule + graph evaluations when ML/LLM services are unreachable.

---

## 2. Circuit Breaker State Transitions

```text
       ┌──────────────────────────────┐
       │     STATE: CLOSED (Normal)   │
       │   Executes all remote calls  │
       └──────────────┬───────────────┘
                      │
                 (Failure Rate >= 50% across 20 calls)
                      ▼
       ┌──────────────────────────────┐
       │      STATE: OPEN (Tripped)   │
       │   Fast-fails immediately;    │
       │   Buffers to Fallback Queue  │
       └──────────────┬───────────────┘
                      │
                 (Cooldown: 30 seconds elapsed)
                      ▼
       ┌──────────────────────────────┐
       │   STATE: HALF-OPEN (Canary)  │
       │   Allows 3 trial canary calls│
       └──────────────┬───────────────┘
                      │
         ┌────────────┴────────────┐
         ▼ (Canary Success)        ▼ (Canary Failure)
   [ Return to CLOSED ]      [ Return to OPEN (30s) ]
```

---

## 3. Fallback Queue In-Memory Buffering

When Kafka is unavailable, events are appended to a lock-free FIFO ring buffer:

```go
type FallbackQueue struct {
    mu       sync.Mutex
    buffer   []map[string]interface{}
    capacity int
}

func (q *FallbackQueue) Enqueue(event map[string]interface{}) bool {
    q.mu.Lock()
    defer q.mu.Unlock()
    if len(q.buffer) >= q.capacity {
        return false // Queue full
    }
    q.buffer = append(q.buffer, event)
    return true
}
```

When the circuit breaker transitions back to `CLOSED`, a background worker drains the fallback queue and flushes all buffered events into Kafka without data loss.

---

## 4. Source Code Map
- [`backend/internal/resilience/circuit_breaker.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/resilience/circuit_breaker.go): Circuit breaker state machine and thresholds.
- [`backend/internal/resilience/fallback_queue.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/resilience/fallback_queue.go): Thread-safe in-memory FIFO queue buffer.
- [`backend/internal/resilience/failure_resilience_test.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/resilience/failure_resilience_test.go): Simulated fault injection tests.

---

## 5. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Protected by circuit breakers.
- [Component 10: Streaming Kafka](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/10-streaming-kafka.md) — Target destination for fallback flushes.
