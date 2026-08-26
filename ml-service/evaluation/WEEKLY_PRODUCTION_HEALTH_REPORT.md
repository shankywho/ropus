# ROPUS — Weekly Production Health & Steady-State Observability Report

## 1. Executive Operations Summary

- **Reporting Period**: `2026-08-25T16:02:00.733987+00:00`
- **Active Production Model**: `v8.0-bmr-36f`
- **SHA-256 Checksum**: `d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7`
- **Model Checksum Match**: **`PASS`**
- **Current Operational State**: **`NORMAL`**
- **Evidence Classification**: **`LOCAL_OPERATIONAL_TEST`**

---

## 2. Operational Key Performance Indicators

| Metric Category | Measured Value | Operational SLA / Target | Status |
| :--- | :---: | :---: | :---: |
| **Inference Latency (p50)** | `1.49 ms` | $< 5.0\text{ ms}$ | **PASS** |
| **Inference Latency (p95)** | `1.73 ms` | $< 15.0\text{ ms}$ | **PASS** |
| **Inference Latency (p99)** | `1.73 ms` | $< 25.0\text{ ms}$ | **PASS** |
| **HTTP 5xx Error Rate** | `0.000%` | $< 0.100\%$ | **PASS** |
| **HTTP 4xx Error Rate** | `0.000%` | $< 2.000\%$ | **PASS** |
| **Calibrated P(Fraud) PSI** | `0.0144` | $< 0.1000$ (Normal) | **PASS** |
| **Transaction Amount PSI** | `0.0096` | $< 0.1000$ (Normal) | **PASS** |

---

## 3. Financial & Business Risk Proxies

- **Total Gross Processed Volume (GPV)**: `$133,833.57`
- **Gross Declined Volume**: `$35,974.36`
- **Dollar Decline Rate**: `26.88%`
- **Mean Declined Ticket Size**: `$599.57`
- **Transaction Allow Rate**: `94.00%` (`940` orders)
- **Transaction Decline Rate**: `6.00%` (`60` orders)

---

## 4. Delayed Dispute / Ground-Truth Ingestion

- **Matched Labels in Buffer**: `5`
- **Label Ingestion Status**: `ACCUMULATING_LABELS`

---

## 5. Escalation & Maintenance Determination

- **Operational Determination**: **`NORMAL`**
- **Action Triggers**: All metrics within normal operational bounds
- **Model Change Governance**: Model `v8.0-bmr-36f` remains locked in `STEADY_STATE_PRODUCTION`. No change request is authorized.
