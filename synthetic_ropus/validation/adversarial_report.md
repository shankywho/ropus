# Adversarial and Stress-Test Evaluation Report

**Dataset:** `synthetic_ropus/data/`
**Scope:** Graph poisoning, flood nodes, injected relationships, and velocity evasion evaluation.
**Classification:** **Synthetic Stress-Test Evidence Only** (Not Real-World Security Certification).

---

## 1. Adversarial Scenarios Overview

The dataset incorporates 51 explicit adversarial tags across three distinct threat vectors modeled in `scenarios/graph_poisoning.py`, alongside dedicated bot and Sybil attacks.

```
                  +-------------------------------------------------+
                  |          Adversarial Threat Scenarios           |
                  +-------------------------------------------------+
                                      |
         +----------------------------+----------------------------+
         |                            |                            |
         v                            v                            v
+------------------+         +------------------+         +------------------+
| High-Degree Node |         | Injected Relat.  |         |   Low-and-Slow   |
| (Degree Flooding)|         | (Spoofed Access) |         | (Velocity Evasion|
| device_000724    |         | 10 emp->acct     |         | 40 transactions  |
+------------------+         +------------------+         +------------------+
```

---

## 2. Threat Vector Breakdown

### 2.1 High-Degree Attack Node (Degree Flooding)
- **Target Entity:** `device_000724`
- **Graph Degree:** **204** (Mean graph device degree is 6.61; P99 is 14.0).
- **Transaction Footprint:** Only **4 transactions** despite connecting to 200 distinct accounts.
- **Objective:** Tests whether degree-based or unweighted GNN neighborhood aggregation is distorted when an attacker creates a dense cluster of empty device connections.

### 2.2 Injected Misleading Relationships (Spoofed Access)
- **Entities:** 10 `employee_accesses_account` edges.
- **Characteristics:** Injected with plausible timestamps during business hours, but completely detached from any legitimate case review, transaction approval, or customer-support session.
- **Objective:** Tests whether multi-hop GNNs overfit to the mere existence of an employee-account edge without verifying surrounding session, case, or transaction context.

### 2.3 Low-and-Slow Attack (Velocity Evasion)
- **Transactions:** 40 transactions across poisoned accounts.
- **Pattern:** Distributed across rotating devices, IP addresses, and payment tokens over several months with small amounts ($5–$60).
- **Objective:** Evaluates whether temporal graph models can correlate disparate low-frequency events that completely evade traditional short-window velocity rules (e.g. 5 transactions in 10 minutes).

### 2.4 Bot & Sybil Swarms
- **Bot Attack (`bot_attack.py`):** 918 transactions characterized by deterministic cadence (2s, 3s, 5s inter-arrival intervals) and rotating tokens.
- **Sybil Clusters (`sybil.py`):** 328 transactions across multi-account clusters funneled through small shared device/IP footprints.

---

## 3. Dedicated Evaluation Subsets

The generator exports dedicated evaluation subsets in `data/labels/`:
- `bot_attack_test.csv` (918 tags)
- `fraud_ring_test.csv` (1,758 tags, including Sybil clusters)
- `graph_poisoning_test.csv` (51 tags: flood, injected edges, low-and-slow)
- `employee_collusion_test.csv` (2,930 tags)
- `hard_negative_test.csv` (12,280 tags)

---

## 4. Methodological Note

> [!IMPORTANT]
> **Synthetic Evidence Boundary:** These adversarial scenarios evaluate GraphSAGE architectural behavior, embedding robustness, and aggregation stability against known synthetic attacks. They do not constitute proof of real-world adversarial robustness against adaptive fraud actors.
