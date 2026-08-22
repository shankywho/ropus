# Component 08: Case Review Management & Human Governance

---

## 1. Why It Exists
Autonomous risk scoring and AI investigations cannot replace human accountability in banking operations. Regulators (such as FinCEN, OCC, and the European Banking Authority) mandate that ultimate decision authority and auditability remain under human analyst control.

The **Case Management & Governance Subsystem** (`backend/internal/cases/`, `backend/internal/governance/`) provides:
1. Priority-ranked review queues for human fraud operations analysts.
2. Direct action controls (**Confirm Block**, **Escalate to AML**, **Override False Positive**).
3. Cryptographically immutable, SHA-256 hash-chained audit ledgers.
4. Closed-loop feedback pipelines that route confirmed analyst verdicts to model retraining.

---

## 2. Case Lifecycle State Machine

```text
               [ Elevated Risk Score (Score >= 0.30) ]
                                  │
                                  ▼
                        ┌──────────────────┐
                        │   STATE: OPEN    │ (Priority: P0 Critical to P3 Low)
                        └─────────┬────────┘
                                  │ (Analyst Claims Case)
                                  ▼
                        ┌──────────────────┐
                        │ STATE: IN_REVIEW │
                        └─────────┬────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         ▼                        ▼                        ▼
┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│ CONFIRMED_FRAUD  │    │  ESCALATED_AML   │    │    OVERRIDDEN    │
│ Account frozen;  │    │ SAR package      │    │ False positive   │
│ rule updated     │    │ sent to AML team │    │ whitelist added  │
└────────┬─────────┘    └────────┬─────────┘    └────────┬─────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌──────────────────────────┐
                    │ STATE: CLOSED / RESOLVED │
                    │ Immutable SHA-256 Hash   │
                    │ Chained to Audit Ledger  │
                    └────────────┬─────────────┘
                                 │
                                 ▼
            [ Dispatched to Closed-Loop Retraining Bus ]
```

---

## 3. Immutable SHA-256 Audit Hash Chain

Every analyst action creates an immutable cryptographic block:
$$\text{BlockHash}_n = \text{SHA-256}\left(\text{BlockID}_n \parallel \text{AnalystID} \parallel \text{Action} \parallel \text{Reason} \parallel \text{Timestamp} \parallel \text{BlockHash}_{n-1}\right)$$

Any attempt to tamper with previous audit records breaks the SHA-256 hash chain and triggers an immediate compliance alert.

---

## 4. Key Data Structures (Go)

```go
type Case struct {
    CaseID           string             `json:"case_id" db:"case_id"`
    OrgID            string             `json:"org_id" db:"org_id"`
    TransactionID    string             `json:"transaction_id" db:"transaction_id"`
    Severity         string             `json:"severity"` // "P0_CRITICAL", "P1_HIGH", "P2_MEDIUM"
    Status           string             `json:"status"`   // "OPEN", "IN_REVIEW", "RESOLVED"
    RiskScore        float64            `json:"risk_score" db:"risk_score"`
    AssignedTo       string             `json:"assigned_to,omitempty"`
    Evidence         []string           `json:"evidence"`
    AIInvestigation  *InvestigationDossier `json:"ai_investigation,omitempty"`
    AuditHashChain   string             `json:"audit_hash_chain"`
    CreatedAt        time.Time          `json:"created_at"`
    UpdatedAt        time.Time          `json:"updated_at"`
}
```

---

## 5. Source Code Map
- [`backend/internal/cases/service.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/cases/service.go): Case creation, querying, state transitions, and assignment.
- [`backend/internal/governance/model_risk_management.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/governance/model_risk_management.go): Model risk governance and SR 11-7 validation.

---

## 6. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Automatically opens cases upon high-risk decisions.
- [Component 07: AI Investigators](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/07-ai-investigators.md) — Attaches generated dossiers to the case.
- [Component 04: ML Inference](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/04-ml-inference.md) — Receives resolved case labels for closed-loop model retraining.
