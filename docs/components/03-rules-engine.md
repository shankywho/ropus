# Component 03: Rules Engine & Velocity Evaluator

---

## 1. Why It Exists
While machine learning excels at detecting complex non-linear patterns, human risk operations teams require the ability to **instantly deploy deterministic controls** without waiting for model retraining.

For example, when a financial institution experiences an active zero-day card testing attack or a sudden regulatory sanction, fraud analysts must be able to deploy a rule like:
`IF amount > 5000 AND merchant_type == "CRYPTO" AND country != "US" THEN BLOCK`
and have it take effect globally within milliseconds.

The **Rules Engine** (`backend/internal/rules/`) provides this dynamic, declarative policy evaluation capability with sub-millisecond evaluation times and zero server restarts.

---

## 2. Architecture & Step-by-Step Workflow

```text
[ Incoming Transaction Features ]
                │
                ▼
┌──────────────────────────────────────────────┐
│ 1. Load Active Rules for Tenant (OrgID)      │
│    • In-memory rule cache ordered by priority │
│    • Filter: IsActive == true                │
└───────────────────────┬──────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────┐
│ 2. Evaluate Rule Conditions (Sequential/AND) │
│    • Extract feature value by field name     │
│    • Execute typed Operator comparison       │
│    • Short-circuit evaluation on first FALSE │
└───────────────────────┬──────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────┐
│ 3. Match Evaluation & Action Aggregation     │
│    • If all conditions TRUE -> Rule MATCHED  │
│    • Record matched rule ID & action payload │
│    • Apply highest priority action           │
└───────────────────────┬──────────────────────┘
                        │
                        ▼
[ Return Matched Rules & Actions to Risk Engine ]
```

---

## 3. Supported Declarative Operators

| Operator | Supported Types | Evaluation Semantics |
| :--- | :--- | :--- |
| `EQUALS` | String, Float, Bool | Exact equality match ($A == B$) |
| `NOT_EQUALS` | String, Float, Bool | Negated equality ($A \ne B$) |
| `GREATER_THAN` | Float, Int | Numerical strictly greater than ($A > B$) |
| `LESS_THAN` | Float, Int | Numerical strictly less than ($A < B$) |
| `GREATER_THAN_OR_EQUAL` | Float, Int | Numerical threshold ($A \ge B$) |
| `LESS_THAN_OR_EQUAL` | Float, Int | Numerical threshold ($A \le B$) |
| `IN` | String Array | Element membership ($A \in \{B_1, B_2, ...\}$) |
| `NOT_IN` | String Array | Element exclusion ($A \notin \{B_1, B_2, ...\}$) |
| `CONTAINS` | String | Substring match (`strings.Contains(A, B)`) |
| `REGEX_MATCH` | String | Compiled POSIX regular expression match |

---

## 4. Key Data Structures (Go)

```go
type Rule struct {
    RuleID      string          `json:"rule_id" db:"rule_id"`
    OrgID       string          `json:"org_id" db:"org_id"`
    Name        string          `json:"name" db:"name"`
    Description string          `json:"description" db:"description"`
    Priority    int             `json:"priority" db:"priority"` // Higher number = higher precedence
    Action      string          `json:"action" db:"action"`     // "APPROVE", "REVIEW", "CHALLENGE", "BLOCK"
    Conditions  []RuleCondition `json:"conditions" db:"conditions"`
    IsActive    bool            `json:"is_active" db:"is_active"`
    CreatedAt   time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type RuleCondition struct {
    Field    string      `json:"field"`    // e.g. "amount", "country", "device_seen_days"
    Operator string      `json:"operator"` // e.g. "GREATER_THAN", "EQUALS", "IN"
    Value    interface{} `json:"value"`    // Target comparison value
}

type RuleMatchRecord struct {
    RuleID   string `json:"rule_id"`
    RuleName string `json:"rule_name"`
    Action   string `json:"action"`
    Priority int    `json:"priority"`
}
```

---

## 5. Failure Modes & Edge Cases
- **Type Mismatch in Condition**: If a rule specifies `GREATER_THAN` on a string field, the condition evaluator safely returns `false` and increments a type mismatch metric rather than causing a runtime panic.
- **Malformed Regular Expression**: Regex strings are validated and compiled upon rule creation/update; invalid regex patterns are rejected with `400 Bad Request` before persisting.

---

## 6. Performance & Concurrency
- **Rule Set Traversal**: Active rules per tenant are cached in a read-heavy `sync.RWMutex` in-memory structure.
- **Evaluation Speed**: Evaluating 50 rules with 4 conditions each takes $< 0.15\text{ms}$ on a single core.

---

## 7. Source Code Map
- [`backend/internal/rules/service.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/rules/service.go): Rule store, in-memory caching, and condition evaluation engine.
- [`backend/internal/rules/rules_test.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/rules/rules_test.go): Comprehensive operator and priority ordering unit tests.

---

## 8. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Upstream ingestion.
- [Component 02: Risk Evaluation Engine](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/02-risk-engine.md) — Consumes rule match results for precedence resolution.
