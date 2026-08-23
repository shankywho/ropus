# ROPUS Control Plane — 5-Minute Executive Demo Runbook

This runbook defines the exact click path for presenting the ROPUS Risk Control Plane to investors, judges, or enterprise risk committees in approximately 5 minutes.

---

## 1. Demo Narrative Arc

```text
  [0:00 - 0:45]       [0:45 - 1:45]         [1:45 - 2:45]        [2:45 - 3:45]        [3:45 - 4:30]         [4:30 - 5:00]
  COMMAND CENTER  →  RISK DECISIONS  →  FORENSIC GRAPH  →  ANALYST QUEUE  →  CANARY ROUTER  →  CLUSTER HEALTH
    (Overview)          Terminal         & Intelligence       & Resolution       & Retraining           & SLOs
```

---

## 2. Step-by-Step 5-Minute Click Path

### Step 1: Institutional Command Center (`0:00 – 0:45`)
- **Navigate to**: `/` (Overview)
- **Key Talking Points**:
  - Point out the **Live Backend** status and sub-millisecond telemetry.
  - Highlight the 5-metric KPI strip: Decisions/sec, Block Rate (0.42%), Open Review Queue, P99 Latency (1.42ms), and Contractual Availability (99.995%).
  - Highlight the bottom **Infrastructure Mesh Probes** table showing live heartbeats across PostgreSQL, Redis, ML sidecar, Redpanda, and ClickHouse.
- **Action**: Click **"Evaluate Test Txn"** button in top right. Show a live synchronous decision streaming into the table in ~1.4ms.

### Step 2: Synchronous Risk Decision Terminal (`0:45 – 1:45`)
- **Navigate to**: `/transactions` (Risk Decisions)
- **Key Talking Points**:
  - Show the dual-scenario comparison: **Scenario A (Attack Wire: $14,500)** vs **Scenario B (Benign POS: $42.50)**.
  - Click **"Execute"** to score the attack payload against the Go backend and ML sidecar.
  - Point to the **Additive Factor Attribution Table**:
    $$\text{Travel Velocity (+0.22)} + \text{Device Entropy (+0.21)} + \text{IP Threat (+0.20)} + \text{Volume (+0.18)} + \text{Drain (+0.17)} = 0.96 \implies \text{BLOCK}$$
  - Expand **"TECHNICAL SPECIFICATIONS & 25-FEATURE VECTOR CONTRACT"** to prove real ONNX beta-calibrated model serving.

### Step 3: Forensic Graph & Threat Intelligence (`1:45 – 2:45`)
- **Navigate to**: `/graph` (Forensic Graph)
- **Key Talking Points**:
  - Point out the **Analytical View** provenance badge.
  - Filter between **1-Hop**, **2-Hop**, and **3-Hop** BFS topology to reveal the coordinated syndicate mule account (`payout_offshore_882`).
  - Click on the node `dev_mule_cluster_99` to show shared canvas entropy across 14 synthetic accounts in the Entity Inspector.
- **Navigate to**: `/investigations` (Investigation Workspace)
  - Highlight the strict **Tripartite Explainability** structure: *1. Observed Facts*, *2. Inferred Patterns*, *3. Recommended Actions*.

### Step 4: Fraud Analyst Queue & Human-In-The-Loop (`2:45 – 3:45`)
- **Navigate to**: `/cases` (Cases)
- **Key Talking Points**:
  - Point out that this is backed by real **PostgreSQL persistence**.
  - Select an active case (`CASE-88419`).
  - Review the evidentiary dossier.
  - Enter analyst justification notes and click **"Confirm Decline (Block)"**.
  - Show the instant database update and audit trail entry.

### Step 5: Model Governance & Dynamic Canary Router (`3:45 – 4:30`)
- **Navigate to**: `/models` (Model Registry)
- **Key Talking Points**:
  - Show the 3-tier architecture: Active Production (`fraud-xgb-25f-v3.0`), Canary Candidate, and Zero-Dependency Standby Fallback (`fraud-xgb-15f-v1.5`).
  - Drag the **Canary Traffic Split Slider** to **25%** and click **"Apply Canary Split"**.
  - Explain that the Go router dynamically adjusts traffic weights with zero-downtime hot reloading.

### Step 6: Operations & Emergency Safety Controls (`4:30 – 5:00`)
- **Navigate to**: `/operations` (Operations)
- **Key Talking Points**:
  - Show contractual **99.99% SLO error budgets** and sub-millisecond component probes.
  - Point out the isolated **MUTATE** safety controls (Maintenance Mode, Model Freeze Lock, Disaster Recovery state sync).
- **Conclude**: *"ROPUS is a complete, production-ready, sub-millisecond risk decisioning and safety control plane built on developer-first infrastructure."*

---

## 3. Alternative Fast Scenario: Automated 7-Stage Walkthrough
- If presenting to an audience that prefers a hands-off cinematic simulation, navigate directly to **`/demo`**.
- Click **"START DEMO"** to automatically advance through the 7-stage attack scenario with the operational execution timeline.
