# Component 11: Storage, Redis Feature Store & Field Encryption

---

## 1. Why It Exists
Financial crime platforms must store two vastly different classes of data:
1. **Ultra-Low Latency Sliding Feature Windows**: Sub-millisecond read/write velocity counters across 10m, 1h, and 24h intervals (handled by **Redis In-Memory Feature Store**).
2. **ACID-Compliant Relational Entity Storage**: Multi-tenant organizations, API keys, rules, cases, and audit logs (handled by **PostgreSQL**).

Furthermore, sensitive user attributes (e.g. device canvas hashes, IP addresses) require **Field-Level AES-256 GCM authenticated encryption** to ensure compliance with strict financial data privacy standards.

---

## 2. Relational Database Schema (PostgreSQL)

```sql
-- Organizations
CREATE TABLE organizations (
    org_id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    plan_tier VARCHAR(32) NOT NULL DEFAULT 'STARTER',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API Keys (Plaintext never stored)
CREATE TABLE api_keys (
    key_id VARCHAR(64) PRIMARY KEY,
    org_id VARCHAR(64) NOT NULL REFERENCES organizations(org_id),
    key_hash VARCHAR(64) NOT NULL UNIQUE,
    key_prefix VARCHAR(16) NOT NULL,
    environment VARCHAR(16) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Cases
CREATE TABLE cases (
    case_id VARCHAR(64) PRIMARY KEY,
    org_id VARCHAR(64) NOT NULL REFERENCES organizations(org_id),
    transaction_id VARCHAR(64) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    risk_score NUMERIC(4,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## 3. Redis Feature Store Key Schemas

| Key Pattern | Data Structure | TTL | Description |
| :--- | :---: | :---: | :--- |
| `feat:{tenant}:{user}:vel_10m` | String (Atomic Counter) | 10 mins | Transaction count in sliding 10m window |
| `feat:{tenant}:{user}:spend_24h` | Float (Atomic Influx) | 24 hours | Cumulative spending amount in 24h window |
| `feat:{tenant}:{device}:seen_users`| Set (HyperLogLog) | 30 days | Count of distinct user accounts using device |

---

## 4. Field-Level AES-256 GCM Encryption

Sensitive fields are encrypted at rest using AES-256 GCM with unique 12-byte initialization vectors (nonces):

$$\text{Ciphertext} \parallel \text{AuthTag} = \text{AES-256-GCM}_{\text{Key}}\left(\text{Plaintext}, \text{Nonce}, \text{AAD}=\text{OrgID}\right)$$

---

## 5. Source Code Map
- [`backend/internal/storage/postgres.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/storage/postgres.go): Connection pooling, query execution, and schema migrations.
- [`backend/internal/features/store/redis_store.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/features/store/redis_store.go): Redis feature store sliding window aggregates.
- [`backend/internal/security/hardening/encryption_manager.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/security/hardening/encryption_manager.go): AES-256 GCM authenticated encryption.

---

## 6. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Reads features and persists decision records.
- [Component 09: Auth & Tenancy](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/09-auth-and-tenancy.md) — Enforces tenant isolation on all database interactions.
