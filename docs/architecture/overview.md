# ROPUS — System Architecture

```text
                    CUSTOMER
                       │
                       ▼
                 API GATEWAY
                       │
                       ▼
              TENANT + AUTH
                       │
                       ▼
             FEATURE PIPELINE
                       │
        ┌──────────────┼──────────────┐
        │              │              │
      RULES           ML            GRAPH
        │              │              │
        └──────────────┼──────────────┘
                       │
                       ▼
                 RISK ENGINE
                       │
                       ▼
              EXPLAINABILITY
                       │
             ┌─────────┴─────────┐
             │                   │
           CASE              WEBHOOK
             │
             ▼
       AI INVESTIGATOR
             │
             ▼
      HUMAN DECISION
```

---

## Component Descriptions

1. **Customer**: Bank, payment processor, or e-commerce merchant sending real-time transaction requests.
2. **API Gateway**: Edge routing layer enforcing TLS 1.3, rate limits, and IP reputation filtering.
3. **Tenant + Auth**: Resolves organization credentials against one-way SHA-256 API key hashes and enforces 4-tier RBAC boundaries.
4. **Feature Pipeline**: Ingests transaction telemetry, extracts velocity metrics, and checks impossible travel anomalies.
5. **Rules Engine**: Evaluates declarative policy conditions and velocity ceilings.
6. **ML Engine**: Real gradient-boosted decision tree (XGBoost/LightGBM) computing continuous mathematical fraud probabilities.
7. **Graph Engine**: Traverses in-memory entity graph to detect multi-hop links to synthetic identities, mules, and emulator clusters.
8. **Risk Engine**: Fuses rule scores, ML probabilities, graph centrality, and threat intelligence into a unified normalized verdict.
9. **Explainability**: Mathematically decomposes composite risk into exact additive factor weights.
10. **Case Management**: Automatically opens review cases for escalated risk decisions.
11. **Webhook Service**: Dispatches HMAC-SHA256 signed event notifications to customer endpoints.
12. **AI Investigator**: Synthesizes structured dossiers distinguishing observed facts from inferred attack patterns.
13. **Human Decision**: Analysts maintain final authority to approve, block, or override decisions, feeding closed-loop retraining.
