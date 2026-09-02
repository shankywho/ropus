# 📖 ROPUS Documentation Index

Welcome to the documentation suite for **ROPUS (Track 02: AI Risk Manager)**.

This index is organized by objective so evaluators, engineers, and judges can quickly locate verifiable evidence, operational runbooks, and architectural specifications.

---

## 🗺️ Documentation Map

### 1. 🚀 Run It
- [Local Development & Service Startup](local-development.md) — How to spin up PostgreSQL, Redis, Redpanda, ClickHouse, Go API, FastAPI ML sidecar, and React frontend natively or via Docker Compose.
- [Docker Compose Quickstart](../README.md#⚡-60-second-quickstart-docker-compose) — Single-command startup (`make up` / `docker compose up -d`).

### 2. 📊 Evaluate It
- [METRICS.md](../METRICS.md) — The single canonical metrics document: ROC-AUC, PR-AUC, ECE, Brier score, confusion matrix, and BMR cost reduction.
- [ML Quality & Calibration Report](ml_quality_report.md) — Comprehensive data distributions, calibration curve comparisons (Raw vs Platt vs Isotonic vs Beta), and ablation studies.
- [Reproducibility Test](../ml-service/tests/test_metrics_consistency.py) — Programmatic test asserting that published metrics match the generated evaluation artifact `phase_3_10_calibration_evaluation.json`.

### 3. 🎬 Demo It
- [Five-Minute Demo Video Runbook](demo-runbook.md) — Time-budgeted 5-minute evaluation walkthrough covering signed webhook ingestion, BMR decisions, failure recovery, and audit ledger verification.
- [Buildathon Demonstration Script](../scripts/demo_buildathon.sh) — Automated shell script executing low-risk allow, high-risk decline, forged signature rejection, idempotent replay, and SHA-256 audit verification.
- [Judge Quick Reference & Cheatsheet](demo/judge-cheatsheet.md) — Concise reference for Buildathon track scoring criteria.

### 4. 🏗️ Architecture & Requirements
- [Buildathon Submission Evidence](buildathon-evidence.md) — Authoritative requirement-to-evidence matrix mapping every Track 02 prompt item to concrete source files, tests, and CLI outputs.
- [System Architecture Overview](architecture/overview.md) — Component boundaries, request flow, and synchronous risk decision pipeline.
- [API Reference](api/api-reference.md) — Endpoints for webhooks, evaluations, decisions, cases, and rules.

### 5. ⚠️ Scope & Known Limitations
- [Technical Scope & Limitations](limitations.md) — Transparent disclosure of what is locally implemented, synthetic fixture boundaries, and non-production disclaimers.

---

## 🛠️ Verification Commands

```bash
# Verify entire test suite (Go, Python, Frontend)
make test

# Reproduce holdout evaluation metrics
make evaluate

# Run the live signed webhook demonstration
make demo-webhook

# Verify documentation structure
make docs-check
```
