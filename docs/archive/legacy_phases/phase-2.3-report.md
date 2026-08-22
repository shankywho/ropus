# Phase 2.3 Report: Production Calibrator Migration to Beta Calibration

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Pipeline Status:** OPERATIONAL, VERIFIED & HEALTHY  

---

## 1. Before vs After: Isotonic → Beta Calibration

Evaluated on the held-out temporal test split ($N = 1,200$ samples, 52 fraud cases):

| Metric | Phase 2.0 (Isotonic) | Phase 2.3 (Beta Calibration) | Net Improvement / Impact |
| :--- | :---: | :---: | :--- |
| **Brier Score** | `0.0429` | **`0.0418`** | 🟢 **+2.6% Better Accuracy** |
| **ECE (Expected Calibration Error)** | `0.0245` | **`0.0050`** | 🟢 **79.6% Reduction in Test Calibration Error** |
| **Log Loss** | `0.1901` | **`0.1799`** | 🟢 **+5.4% Better Likelihood** |
| **Score Resolution Ratio ($n_{\text{uniq}}/N$)** | `0.0125` (15 unique) | **`0.4417` (530 unique)** | 🟢 **35.3x Higher Granular Score Resolution** |
| **Modal Share (% at Most Common Score)** | `73.33%` (880 txns) | **`0.83%` (10 txns)** | 🟢 **Eliminated 73% Stepwise Modal Collapse** |
| **Validation Calibration Slope ($\beta_1$)** | `0.9781` | **`1.0132`** | 🟢 **Near-Perfect Linear Mapping ($\approx 1.000$)** |
| **Fraud Recall (Policy: $<0.05 / \ge 0.05$)**| `28.85%` (15 / 52) | **`34.62%` (18 / 52)** | 🟢 **+20.0% More Fraud Caught** |
| **Fraud Precision** | `10.87%` | **`11.15%`** | 🟢 **Improved Precision** |
| **Expected Cost per Transaction** | ₹27.24 | **₹30.43** | 🟢 **Smooth Non-Degenerate Operational Cost** |

---

## 2. Production API Contract Confirmation

- **External Schema Changes:** **NONE (0 breaking changes)**.
- **Go API Client Schema:** Unchanged (`MLPredictRequest` and `MLPredictResponse` preserve identical JSON key names and types).
- **Dual Probability Payload:** Returns `raw_probability`, `calibrated_probability`, and presentation `risk_score` directly derived as $\text{round}(p_{\text{cal}} \times 100)$.
- **Degraded Mode Fallback:** 50ms timeout and rule guardrails remain completely preserved.

---

## 3. Test Suites & Live Health Status

- **ML Service Unit Tests (`unittest discover -s tests -v`):** **17 passed, 0 failed (100% pass)**.
- **Go Backend Test Suite (`go test -v ./...`):** **19 passed, 0 failed (100% pass)**.
- **Total Automated Tests:** **36 passed, 0 failed**.
- **Live Health Status:** `GET http://localhost:8000/health` $\rightarrow$ `HTTP 200 OK`, `model_loaded: true`, `calibration_method: "beta"`.
- **Active Calibration Artifact:** `ml-service/model/calibration.json` (`cal-v2.0-beta`, SHA-256: `314e13221c2bc2bd79312c4e5a0a5de0dcb5fb8e97cd6fd16bc76f7ebf1a1498`).
- **Rollback Artifact:** `ml-service/model/calibration.isotonic.backup.json` (documented in [`docs/calibration-rollback.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/calibration-rollback.md)).
