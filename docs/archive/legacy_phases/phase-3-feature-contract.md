# Phase 3 — Proposed Device Feature Contract & Parity Specification

**Document Version:** 1.0 (Design Specification)  
**Total Canonical Features:** 25 (15 Phase 1 Features + 10 Phase 3 Device Features)  

---

## 1. 25-Feature Schema Overview

```
[ 15 CANONICAL BASELINE FEATURES ]
  1. amount                     6. transaction_hour        11. email_domain_risk
  2. ip_velocity_1h             7. transaction_day         12. dist1_missing
  3. ip_velocity_24h            8. product_cd_encoded      13. device_type_mobile
  4. token_velocity_24h         9. card_type_encoded       14. device_info_missing
  5. amount_to_mean_ratio      10. card_category_encoded   15. is_new_device

                                 +

[ 10 PROPOSED PHASE 3 DEVICE INTELLIGENCE FEATURES ]
 16. device_tx_count_1h               (Redis sliding window 1 hour)
 17. device_tx_count_24h              (Redis sliding window 24 hours)
 18. device_amount_sum_24h            (Redis sliding window rolling sum)
 19. device_unique_accounts_24h       (Redis Set: distinct accounts on device in 24h)
 20. device_unique_tokens_24h         (Redis Set: distinct payment cards on device in 24h)
 21. device_seen_before               (Point-in-time novelty: 1 if first_seen_at < T, else 0)
 22. device_age_days                  (Days elapsed since device was first observed)
 23. device_account_link_age_days     (Days elapsed since device was first paired with this user)
 24. device_distinct_accounts_all_time(Durable multi-account count from registry)
 25. device_fraud_associated          (Reputation flag: 1 if prior chargeback/fraud on device, else 0)
```

---

## 2. Detailed Feature Contract & Training-Serving Parity

| # | Feature Name | Type | Value Range | Offline Training Source (IEEE-CIS / Historical) | Online Production Source | Point-in-Time Safety Guarantee | Skew Risk |
| :--- | :--- | :---: | :---: | :--- | :--- | :--- | :--- |
| **16** | `device_tx_count_1h` | `float` | $[0, \infty)$ | Trailing 1-hour count of transactions sharing same device ID up to $T$. | Redis `ZCOUNT {tenant}:vel:dev:{dev_id} T-1h T` | ✅ Strict $\le T$ window query. | 🟢 ALIGNED |
| **17** | `device_tx_count_24h` | `float` | $[0, \infty)$ | Trailing 24-hour count of transactions sharing same device ID up to $T$. | Redis `ZCOUNT {tenant}:vel:dev24:{dev_id} T-24h T` | ✅ Strict $\le T$ window query. | 🟢 ALIGNED |
| **18** | `device_amount_sum_24h` | `float` | $[0, \infty)$ | Trailing 24-hour sum of transaction amounts on device up to $T$. | Redis sliding window score aggregation | ✅ Strict $\le T$ window sum. | 🟢 ALIGNED |
| **19** | `device_unique_accounts_24h` | `int` | $[1, \infty)$ | Number of distinct account IDs observed on device in $[T-24\text{h}, T]$. | Redis Set `SCARD {tenant}:dev:acc24:{dev_id}` | ✅ Strict $\le T$ time bound. | 🟢 ALIGNED |
| **20** | `device_unique_tokens_24h` | `int` | $[1, \infty)$ | Number of distinct payment tokens observed on device in $[T-24\text{h}, T]$. | Redis Set `SCARD {tenant}:dev:tok24:{dev_id}` | ✅ Strict $\le T$ time bound. | 🟢 ALIGNED |
| **21** | `device_seen_before` | `int` | $\{0, 1\}$ | `1` if device was seen in historical partition prior to $T$, else `0`. | Redis String `EXISTS {tenant}:dev:known:{dev_id}` | ✅ True point-in-time check. | 🟢 ALIGNED |
| **22** | `device_age_days` | `float` | $[0.0, 1000.0]$ | `(T - first_seen_timestamp) / 86400.0`. Returns `0.0` if novel. | Redis/Postgres `(now - first_seen_at) / 86400` | ✅ Historical first-seen cannot move forward. | 🟢 ALIGNED |
| **23** | `device_account_link_age_days`| `float` | $[0.0, 1000.0]$ | `(T - first_linked_timestamp) / 86400.0` for (account, device) pair. | Redis `(now - first_linked_at) / 86400` | ✅ Link timestamp immutable. | 🟢 ALIGNED |
| **24** | `device_distinct_accounts_all_time` | `int` | $[1, \infty)$ | Historical distinct account count associated with device before $T$. | Postgres/Redis counter cached on device | ✅ Historical accumulator strictly $\le T$. | 🟢 ALIGNED |
| **25** | `device_fraud_associated` | `int` | $\{0, 1\}$ | `1` if a prior fraud/chargeback was confirmed on device before $T$, else `0`. | Redis Set `SISMEMBER {tenant}:dev:bad_rep {dev_id}` | ✅ Historical chargeback date $\le T$. | 🟢 ALIGNED |

---

## 3. Fallback Values for Missing Telemetry & Degraded Modes

If client device telemetry is missing, blocked by browser extensions, or Redis is temporarily unavailable:

| Feature Name | Default Fallback Value | Imputation Semantic |
| :--- | :---: | :--- |
| `device_tx_count_1h` | `0.0` | Assume clean zero velocity. |
| `device_tx_count_24h` | `0.0` | Assume clean zero velocity. |
| `device_amount_sum_24h` | `0.0` | Assume clean zero velocity. |
| `device_unique_accounts_24h` | `1` | Assume single-user baseline. |
| `device_unique_tokens_24h` | `1` | Assume single-token baseline. |
| `device_seen_before` | `0` | Default to novel device state. |
| `device_age_days` | `0.0` | Zero age for novel/missing device. |
| `device_account_link_age_days` | `0.0` | Zero age for novel/missing link. |
| `device_distinct_accounts_all_time` | `1` | Single user baseline. |
| `device_fraud_associated` | `0` | Innocent until proven guilty. |
| `device_info_missing` | `1` | **Explicit indicator flag set to 1!** (Allows trees to learn missingness risk). |
