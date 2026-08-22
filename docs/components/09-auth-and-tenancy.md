# Component 09: Authentication, API Keys & Multi-Tenant Isolation

---

## 1. Why It Exists
ROPUS operates as a multi-tenant enterprise SaaS platform. A single physical cluster serves multiple independent banks, payment processors, and marketplaces.

Under no circumstances may Tenant A access, inspect, or mutate Tenant B's transactions, rules, cases, or machine learning models. 

The **Auth & SaaS Subsystem** (`backend/internal/auth/api_keys/`, `backend/internal/saas/`) enforces:
1. One-way cryptographic SHA-256 API key hashing.
2. 4-tier Role-Based Access Control (**OWNER**, **ADMIN**, **ANALYST**, **VIEWER**).
3. Mandatory tenant boundary scoping at the repository data layer.

---

## 2. API Key Cryptographic Lifecycle

```text
[ 1. Key Generation ]
  Plaintext Secret: "rop_live_8a19bc7f2e41c6d39a04..."
  SHA-256 Hash:     "4f8a1e9c7a2b5d8e3f1c...e91c"
  Database Write:   INSERT INTO api_keys (key_hash, org_id, environment)
                    (Plaintext token is returned ONCE to client, NEVER stored)

[ 2. Inbound Request Verification ]
  Header:           Authorization: Bearer rop_live_8a19bc7f2e41c6d39a04...
  Compute Hash:     hash = SHA-256("rop_live_8a19bc7f2e41c6d39a04...")
  Lookup:           SELECT org_id, environment FROM api_keys WHERE key_hash = hash
  Result:           org_id = "org_bank_acme", environment = "live"
```

---

## 3. 4-Tier RBAC Permission Matrix

| Capability | OWNER | ADMIN | ANALYST | VIEWER |
| :--- | :---: | :---: | :---: | :---: |
| **Evaluate Transaction Risk** | Yes | Yes | Yes | No |
| **View Cases & Investigation Dossiers** | Yes | Yes | Yes | Read-Only |
| **Execute Analyst Decision (Confirm/Override)** | Yes | Yes | Yes | No |
| **Create / Modify Declarative Rules** | Yes | Yes | Read-Only | No |
| **Manage API Keys & Webhook Secrets** | Yes | Yes | No | No |
| **Manage SaaS Plan & Invoices** | Yes | No | No | No |

---

## 4. Key Data Structures (Go)

```go
type APIKeyMetadata struct {
    KeyID       string    `json:"key_id" db:"key_id"`
    OrgID       string    `json:"org_id" db:"org_id"`
    KeyPrefix   string    `json:"key_prefix"` // e.g. "rop_live_8a19..."
    KeyHash     string    `json:"-" db:"key_hash"`
    Environment string    `json:"environment"` // "live" or "test"
    CreatedAt   time.Time `json:"created_at"`
    ExpiresAt   time.Time `json:"expires_at"`
    IsRevoked   bool      `json:"is_revoked"`
}
```

---

## 5. Source Code Map
- [`backend/internal/auth/api_keys/key_service.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/auth/api_keys/key_service.go): API key generation, hashing, and cache-backed verification.
- [`backend/internal/saas/organization.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/saas/organization.go): Multi-tenant organization and user membership.
- [`backend/internal/saas/tenant_isolation_test.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/saas/tenant_isolation_test.go): Adversarial cross-tenant isolation test suite.

---

## 6. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — First middleware step in every request.
- [Component 11: Storage Persistence](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/11-storage-persistence.md) — Enforces `WHERE org_id = $1` on all SQL queries.
