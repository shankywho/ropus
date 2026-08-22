# ROPUS Enterprise Production Architecture Whitepaper

## System Architecture

```text
                                [ Inbound API Traffic / Gateway ]
                                                │
                                                ▼
                        ┌───────────────────────────────────────────────┐
                        │     Distributed Token-Bucket Rate Limiter     │
                        │   (Tenant Quotas, Abuse Defense, WAF Rules)   │
                        └───────────────────────┬───────────────────────┘
                                                │
                                                ▼
                        ┌───────────────────────────────────────────────┐
                        │        High Availability Circuit Breaker      │
                        │    (Kafka Buffering, DB Failover Detection)   │
                        └───────┬───────────────┼───────────────┬───────┘
                                │               │               │
                                ▼               ▼               ▼
                       ┌────────────────┐┌─────────────┐┌──────────────┐
                       │ Real ML Engine ││ Fraud Graph ││ Case Manager │
                       │ (Sub-1ms ONNX) ││   Engine    ││  (Postgres)  │
                       └────────┬───────┘└──────┬──────┘└───────┬──────┘
                                │               │               │
                                └───────────────┼───────────────┘
                                                │
                                                ▼
                        ┌───────────────────────────────────────────────┐
                        │       Observability 2.0 & SLO Monitor         │
                        │   (Prometheus Metrics, Distributed Traces)    │
                        └───────────────────────────────────────────────┘
```

---

## 1. High Availability & Subsystem Decoupling
- **Kafka Degraded Mode**: If Apache Kafka brokers experience high latency, the `HealthManager` buffers outbound stream events in a local fallback queue and drains automatically upon broker recovery.
- **Circuit Breaker Fast-Failing**: Protects downstream inference workers from cascading queue exhaustion.
