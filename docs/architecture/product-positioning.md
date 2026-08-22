# ROPUS Product Positioning: Traditional Fraud Engines vs. AI Risk Manager

```text
================================================================================
          TRADITIONAL FRAUD PLATFORM vs. ROPUS AI RISK MANAGER
================================================================================
Feature               Legacy (Sift, Feedzai v1)     ROPUS AI Risk Manager
--------------------------------------------------------------------------------
Risk Decision Loop    Transaction -> Score          Transaction -> Context -> Graph
                                                    -> ML -> Investigation -> Action
Explainability        Black-Box Score (0-100)       Exact Additive Factor Breakdown
Fraud Graphs          Slow Batch / Offline ETL      Real-time 3-Hop In-Memory Traversal
Investigations        Manual Analyst Querying       Autonomous AI Agent Dossiers
Governance            Static CSV Export Logs        Cryptographic Hash-Chained Ledgers
Integration           Months of Custom ETL          Standard REST /v1/risk/evaluate
================================================================================
```

---

## 1. Why Point Solutions Fail
Legacy systems treat payments as isolated rows in a database. Modern fraud rings execute multi-hop synthetic identity schemes across distinct merchants, cards, and IPs. Without **real-time graph traversal** and **agentic investigation**, point solutions are blind to coordinated syndicate behavior.

---

## 2. The ROPUS Paradigm: Connected Intelligence
ROPUS fuses **real-time graph neighborhood analysis**, **gradient-boosted ML scoring**, **autonomous AI reasoning**, and **human-in-the-loop governance** into a single sub-10ms evaluation loop.
