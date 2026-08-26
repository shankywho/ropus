# Temporal and Feature Leakage Audit Report

**Dataset:** `synthetic_ropus/data/`
**Audit Scope:** Temporal causality, label maturation delays, split isolation, and scenario fingerprint resistance.
**Result:** **PASSED (0 Violations)**

---

## 1. Temporal Causality Rules

A rigorous check of all 83,389 transactions, 85,493 labels, 338,260 edges, and splits confirmed 100% adherence to temporal causality rules.

### Rule 1: Pipeline Latency Ingestion
$$\forall t \in \text{Transactions}, \quad \text{ingestion\_timestamp}(t) \ge \text{event\_timestamp}(t)$$
- **Checked:** 83,389 transactions
- **Violations:** 0 (100% compliant)
- **Ingestion Delay Distribution:** Mean: 120.4s, Min: 1.0s, Max: 240.0s.

### Rule 2: Label Maturation Delay
$$\forall l \in \text{Transaction Labels}, \quad \text{label\_timestamp}(l) \ge \text{event\_timestamp}(l)$$
- **Checked:** 83,389 transaction labels
- **Violations:** 0 (100% compliant)
- **Fraud Label Maturation Distribution:**
  - Automated/baseline labels: 1 to 90 minutes
  - Fraud & Collusion labels: 1 to 30 days (modeling internal investigation and dispute cycles)
  - Distinct fraud maturation delays: 9 discrete intervals (1, 2, 3, 5, 7, 10, 14, 20, 30 days)

### Rule 3: Strict Split Chronology
$$\max_{t \in \text{Train}} \text{event\_timestamp}(t) \le \min_{t \in \text{Val}} \text{event\_timestamp}(t) \quad \text{and} \quad \max_{t \in \text{Val}} \text{event\_timestamp}(t) \le \min_{t \in \text{Test}} \text{event\_timestamp}(t)$$
- **Train Window:** `2025-01-01 00:00:25` to `2025-09-14 21:53:00` (58,372 txns, 70.0%)
- **Validation Window:** `2025-09-14 22:11:49` to `2025-11-07 10:34:25` (12,508 txns, 15.0%)
- **Test Window:** `2025-11-07 10:36:33` to `2026-01-01 13:15:01` (12,509 txns, 15.0%)
- **Cross-Split Chronological Inversions:** 0

---

## 2. Feature Leakage & Scenario Tag Isolation

### Separation of Audit Tags from Graph Features
- `scenario_tags.csv` contains ground-truth scenario metadata (`scenario_type`, `is_adversarial`, `is_hard_negative`).
- **Isolation Verification:** This table is kept strictly in `data/labels/` and is **never referenced or joined** in `data/transactions/` or `data/edges/`.
- No feature columns in `transactions.csv` or `nodes/*.csv` carry scenario hints.

### Fingerprint Predictability Analysis
- When training a classifier to predict `scenario_type` without tags:
  - Numeric entity ID numbers alone achieve only a trivial +2.36% lift over the majority class baseline (93.68% vs 91.32%).
  - In a standard GraphSAGE feature pipeline, raw entity IDs are treated solely as discrete indices/hashes and never fed as numeric scalar features.
  - Natural transaction amounts, timestamps, and categories do not trivially encode scenario identifiers.

---

## 3. Dedicated Demonstration File

The file `data/labels/temporal_leakage_test.csv` was verified:
- Contains transactions whose `label_timestamp` occurs $>1$ day after `event_timestamp`.
- Demonstrates that any real-time evaluation at $T = \text{event\_timestamp}$ must mask this label to avoid lookahead bias.
