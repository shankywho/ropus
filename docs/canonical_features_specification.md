# Canonical 25-Feature Specification & Parity Standard

## Overview
This document defines the authoritative, mathematical specification for all 25 features utilized in the ROPUS risk decisioning pipeline.
Every feature is guaranteed to be **strictly point-in-time consistent** ($< T$), with zero temporal future data leakage, and exact mathematical equivalence between the Python training pipeline (`ml-service/data_pipeline/features.py`) and the Go runtime inference engine (`backend/internal/riskengine/ml_features.go`).

---

## Canonical Feature Catalog

| # | Feature Key | Data Type | Source / Entity | Sliding Window / Computation | Missing Default | Go / Python Equivalence |
|---|---|---|---|---|---|---|
| 1 | `amount` | Float32 | Transaction Payload | Transaction monetary amount ($) | `100.0` | Direct extract from payload `Amount` / 100.0 |
| 2 | `ip_velocity_1h` | Float32 | Redis Feature Store / IP | Count of prior txns from IP in last 3600s ($< T$) | `0.0` | `VelocityStore.GetIPVelocity(ip, 1h)` |
| 3 | `ip_velocity_24h` | Float32 | Redis Feature Store / IP | Count of prior txns from IP in last 86400s ($< T$) | `0.0` | `VelocityStore.GetIPVelocity(ip, 24h)` |
| 4 | `token_velocity_24h` | Float32 | Redis Feature Store / Token | Count of prior txns from payment token in last 86400s | `0.0` | `PaymentTokenStore.GetVelocity(token, 24h)` |
| 5 | `amount_to_mean_ratio` | Float32 | Redis Feature Store / Account | $\text{Amount} / \max(1.0, \bar{A}_{\text{account}})$ | `1.0` | Computed via historical rolling mean in Redis |
| 6 | `device_seen_before` | Int32 | Redis Feature Store / Device | 1 if device previously associated with account, else 0 | `0` | `AccountDeviceGraphStore.HasEdge(account, device)` |
| 7 | `transaction_hour` | Int32 | Temporal Context | $(T \pmod{86400}) / 3600 \in [0, 23]$ | `12` | UTC hour of transaction timestamp |
| 8 | `transaction_day` | Int32 | Temporal Context | $(T / 86400) \pmod 7 \in [0, 6]$ | `0` | UTC day of week |
| 9 | `product_cd_encoded` | Int32 | Transaction Context | Ordinal encoding of ProductCD: `W:4, C:0, H:1, R:2, S:3` | `-1` | Constant mapping table matching training |
| 10 | `card_type_encoded` | Int32 | Payment Method | Ordinal encoding of card brand: `visa:3, mastercard:2, ...` | `-1` | Constant mapping table matching training |
| 11 | `card_category_encoded` | Int32 | Payment Method | Ordinal encoding of card type: `debit:1, credit:0` | `-1` | Constant mapping table matching training |
| 12 | `email_domain_risk` | Float32 | PII / Email Domain | Laplace-smoothed historical fraud risk of domain | `0.035` | Train-fitted smoothed risk lookup table |
| 13 | `dist1_missing` | Int32 | Transaction Context | 1 if billing/shipping distance missing, else 0 | `1` | Check for null/empty distance context |
| 14 | `device_type_mobile` | Int32 | Client Telemetry | 1 if client user agent indicates mobile/tablet, else 0 | `0` | User agent parser classification |
| 15 | `device_info_missing` | Int32 | Client Telemetry | 1 if device fingerprint is missing/empty, else 0 | `0` | Check `device_fingerprint == ""` |
| 16 | `device_tx_count_5m` | Float32 | Redis Feature Store / Device | Count of prior txns from device fingerprint in last 300s | `0.0` | `DeviceVelocityStore.GetCount(device, 5m)` |
| 17 | `device_tx_count_1h` | Float32 | Redis Feature Store / Device | Count of prior txns from device fingerprint in last 3600s | `0.0` | `DeviceVelocityStore.GetCount(device, 1h)` |
| 18 | `device_amount_sum_24h`| Float32 | Redis Feature Store / Device | Sum of amounts ($) processed on device in last 86400s | `0.0` | `DeviceVelocityStore.GetAmountSum(device, 24h)` |
| 19 | `tx_acceleration_5m_1h` | Float32 | Derived Velocity | $\text{count}_{5m} / \max(1.0, \text{count}_{1h} / 12.0)$ | `0.0` | Exact formula verified in both Go & Python |
| 20 | `device_amount_concentration_5m_1h` | Float32 | Derived Velocity | $\text{amt}_{5m} / \max(1.0, \text{amt}_{1h})$ | `0.0` | Exact formula verified in both Go & Python |
| 21 | `device_unique_tokens_1h` | Float32 | Redis Feature Store / Device | Count of distinct payment tokens used on device in last 1h | `0.0` | `DeviceVelocityStore.GetUniqueTokens(device, 1h)` |
| 22 | `token_unique_devices_1h` | Float32 | Redis Feature Store / Token | Count of distinct devices using payment token in last 1h | `0.0` | `PaymentTokenStore.GetUniqueDevices(token, 1h)` |
| 23 | `device_reputation_score` | Float32 | Redis Feature Store / Device | Multi-factor device trust score $[0.0, 1.0]$ | `0.50` | `DeviceReputationStore.GetScore(device)` |
| 24 | `device_fraud_rate` | Float32 | Redis Feature Store / Device | Historical confirmed fraud rate on device $< T$ | `0.0` | `DeviceReputationStore.GetFraudRate(device)` |
| 25 | `device_dispute_rate` | Float32 | Redis Feature Store / Device | Historical dispute/chargeback rate on device $< T$ | `0.0` | `DeviceReputationStore.GetDisputeRate(device)` |

---

## Leakage Prevention Guarantees
1. **Sliding Windows are Strictly $< T$**: All features are computed strictly on transactions occurring prior to the current transaction timestamp.
2. **Train-Only Preprocessor Statistics**: Imputation medians and target encoding dictionaries are computed strictly on the training partition and frozen before inference.
3. **No Target Leakage**: Feature computation never accesses the current transaction's `isFraud` outcome or future dispute updates.
