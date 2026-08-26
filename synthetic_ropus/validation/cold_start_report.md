# Inductive Cold-Start Evaluation Report

**Dataset:** `synthetic_ropus/data/`
**Purpose:** Quantify unseen entity coverage in the test split to validate GraphSAGE inductive graph learning capabilities.

---

## 1. Executive Summary

GraphSAGE was specifically designed for **inductive node and edge representation learning**, allowing embeddings to be generated for unseen nodes at inference time without retraining the entire graph.

To ensure the test split genuinely exercises this capability, the temporal split (`splits.csv`) was audited to quantify entities appearing in the test period (`2025-11-07` to `2026-01-01`) that never appeared during the training period (`2025-01-01` to `2025-09-14`).

---

## 2. Unseen Entity Statistics in Test Split

| Entity Type | Total in Test Split | Unseen (Never in Train) | Cold-Start Ratio (%) |
| :--- | :--- | :--- | :--- |
| **Accounts** | 9,055 | **801** | **8.85%** |
| **Payment Tokens** | 9,541 | **929** | **9.74%** |
| **Devices** | 7,329 | **73** | **1.00%** |
| **IP Addresses** | 6,888 | **20** | **0.29%** |
| **Active Employees** | 259 | **9** | **3.47%** |
| **Merchants** | 1,200 | **0** | **0.00%** |

---

## 3. Dedicated Cold-Start Evaluation Subset

We construct a dedicated evaluation subset:
- **Definition:** Transactions in the test split where at least one entity (`account_id`, `payment_token_id`, or `device_id`) was completely unobserved during training.
- **Transactions in Cold-Start Subset:** **2,287 transactions** (**18.28% of the 12,509 test transactions**).
- **Fraud Representation in Cold-Start:** Includes new accounts created under fraud ring, ATO, and card testing scenarios.

---

## 4. Recommended GraphSAGE Evaluation Protocol

1. **Training Phase:** Train GraphSAGE using nodes and edges up to the train cutoff timestamp ($T_{\text{train}} = \text{2025-09-14 21:53:00}$).
2. **Inductive Inference Phase:** For any test transaction arriving at $T_{\text{eval}} > T_{\text{train}}$, sample local 2-hop computation graphs using available historical and instantaneous edges without recomputing or fine-tuning node embedding matrices from scratch.
3. **Metric Reporting:** Report ROC AUC, PR AUC, and Top-1% Recall separately on:
   - Full Test Set (12,509 transactions)
   - Cold-Start Subset (2,287 transactions)
