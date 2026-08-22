"""
Point-in-Time Safe Canonical Feature Extraction Module
Strictly prevents future leakage during feature engineering
"""

import numpy as np
import pandas as pd
from typing import Dict, Any, List

def compute_point_in_time_velocities(
    df: pd.DataFrame,
    entity_col: str,
    time_col: str,
    window_seconds: int
) -> np.ndarray:
    """
    Calculates exact historical transaction counts within a rolling time window [T - window, T]
    strictly using past and current records for each entity.
    Guarantees point-in-time correctness: future records (> T) cannot influence transaction at T.
    """
    velocities = np.zeros(len(df), dtype=np.float32)
    
    # Group timestamps by entity in chronological order
    entity_times: Dict[Any, List[int]] = {}
    
    entities = df[entity_col].values
    timestamps = df[time_col].values
    
    for i in range(len(df)):
        entity = entities[i]
        t = timestamps[i]
        
        if pd.isna(entity):
            velocities[i] = 1.0
            continue
            
        if entity not in entity_times:
            entity_times[entity] = [t]
            velocities[i] = 1.0
        else:
            t_list = entity_times[entity]
            # Prune events older than window (sliding window)
            min_t = t - window_seconds
            # Binary search or scan for window boundary
            idx = 0
            while idx < len(t_list) and t_list[idx] < min_t:
                idx += 1
            if idx > 0:
                entity_times[entity] = t_list[idx:]
                t_list = entity_times[entity]
                
            t_list.append(t)
            velocities[i] = float(len(t_list))
            
    return velocities

def compute_point_in_time_device_novelty(
    df: pd.DataFrame,
    account_col: str,
    device_col: str
) -> np.ndarray:
    """
    Computes device_seen_before(T):
    1 if device was associated with account/card prior to time T, 0 if novel/first-time.
    """
    device_seen = np.zeros(len(df), dtype=np.int32)
    known_pairs = set()
    
    accounts = df[account_col].values
    devices = df[device_col].values
    
    for i in range(len(df)):
        acc = accounts[i]
        dev = devices[i]
        
        if pd.isna(acc) or pd.isna(dev):
            device_seen[i] = 0
            continue
            
        pair = (acc, str(dev))
        if pair in known_pairs:
            device_seen[i] = 1
        else:
            device_seen[i] = 0
            known_pairs.add(pair)
            
    return device_seen

def compute_point_in_time_amount_ratio(
    df: pd.DataFrame,
    account_col: str,
    amount_col: str
) -> np.ndarray:
    """
    Computes ratio of transaction amount to cumulative historical mean amount for the account up to time T.
    """
    ratios = np.ones(len(df), dtype=np.float32)
    acc_stats: Dict[Any, List[float]] = {} # [running_sum, running_count]
    
    accounts = df[account_col].values
    amounts = df[amount_col].values
    
    for i in range(len(df)):
        acc = accounts[i]
        amt = float(amounts[i]) if not pd.isna(amounts[i]) else 0.0
        
        if pd.isna(acc):
            ratios[i] = 1.0
            continue
            
        if acc not in acc_stats:
            acc_stats[acc] = [amt, 1.0]
            ratios[i] = 1.0
        else:
            prev_sum, prev_count = acc_stats[acc]
            mean_val = prev_sum / prev_count if prev_count > 0 else amt
            if mean_val > 0:
                ratios[i] = float(np.clip(amt / mean_val, 0.05, 20.0))
            else:
                ratios[i] = 1.0
            acc_stats[acc][0] += amt
            acc_stats[acc][1] += 1.0
            
    return ratios

def compute_point_in_time_device_velocity_signals(
    df: pd.DataFrame,
    device_col: str,
    time_col: str,
    amount_col: str
) -> Dict[str, np.ndarray]:
    """
    Computes point-in-time multi-window device velocity features (5m, 1h, 24h),
    acceleration, and amount concentration. Strictly uses prior events (< T).
    """
    n = len(df)
    dev_tx_5m = np.zeros(n, dtype=np.float32)
    dev_tx_1h = np.zeros(n, dtype=np.float32)
    dev_amt_24h = np.zeros(n, dtype=np.float32)
    tx_accel_5m_1h = np.zeros(n, dtype=np.float32)
    amt_conc_5m_1h = np.zeros(n, dtype=np.float32)

    device_events: Dict[Any, List[Tuple[int, float]]] = {}  # device -> list of (timestamp, amount)

    devices = df[device_col].values
    timestamps = df[time_col].values
    amounts = df[amount_col].fillna(0.0).values

    for i in range(n):
        dev = devices[i]
        t = int(timestamps[i])
        amt = float(amounts[i])

        if pd.isna(dev) or str(dev) == "" or str(dev) == "nan":
            dev_tx_5m[i] = 0.0
            dev_tx_1h[i] = 0.0
            dev_amt_24h[i] = 0.0
            tx_accel_5m_1h[i] = 0.0
            amt_conc_5m_1h[i] = 0.0
            continue

        if dev not in device_events:
            dev_tx_5m[i] = 0.0
            dev_tx_1h[i] = 0.0
            dev_amt_24h[i] = 0.0
            tx_accel_5m_1h[i] = 0.0
            amt_conc_5m_1h[i] = 0.0
            device_events[dev] = [(t, amt)]
        else:
            ev_list = device_events[dev]
            min_24h = t - 86400
            min_1h = t - 3600
            min_5m = t - 300

            # Prune events older than 24h
            idx = 0
            while idx < len(ev_list) and ev_list[idx][0] < min_24h:
                idx += 1
            if idx > 0:
                ev_list = ev_list[idx:]
                device_events[dev] = ev_list

            c_5m = 0
            c_1h = 0
            sum_24h = 0.0
            sum_5m = 0.0
            sum_1h = 0.0

            for ev_t, ev_amt in ev_list:
                sum_24h += ev_amt
                if ev_t >= min_1h:
                    c_1h += 1
                    sum_1h += ev_amt
                if ev_t >= min_5m:
                    c_5m += 1
                    sum_5m += ev_amt

            dev_tx_5m[i] = float(c_5m)
            dev_tx_1h[i] = float(c_1h)
            dev_amt_24h[i] = float(sum_24h)

            # Acceleration: rate_5m / rate_1h
            rate_5m = float(c_5m) / 300.0
            rate_1h = float(c_1h) / 3600.0
            if rate_1h > 0.0:
                accel = rate_5m / (rate_1h + 0.001)
                tx_accel_5m_1h[i] = float(np.clip(accel, 0.0, 1000.0))
            else:
                tx_accel_5m_1h[i] = 0.0

            # Amount Concentration: sum_5m / sum_1h in [0, 1]
            if sum_1h > 0.0:
                conc = sum_5m / (sum_1h + 0.001)
                amt_conc_5m_1h[i] = float(np.clip(conc, 0.0, 1.0))
            else:
                amt_conc_5m_1h[i] = 0.0

            # Append current event AFTER calculating features
            ev_list.append((t, amt))

    return {
        "device_tx_count_5m": dev_tx_5m,
        "device_tx_count_1h": dev_tx_1h,
        "device_amount_sum_24h": dev_amt_24h,
        "tx_acceleration_5m_1h": tx_accel_5m_1h,
        "device_amount_concentration_5m_1h": amt_conc_5m_1h,
    }

def compute_point_in_time_token_device_linkage(
    df: pd.DataFrame,
    device_col: str,
    token_col: str,
    time_col: str
) -> Dict[str, np.ndarray]:
    """
    Computes distinct tokens per device (1h) and distinct devices per token (1h) point-in-time (< T).
    """
    n = len(df)
    dev_uniq_tok_1h = np.zeros(n, dtype=np.float32)
    tok_uniq_dev_1h = np.zeros(n, dtype=np.float32)

    dev_tokens: Dict[Any, List[Tuple[int, Any]]] = {}  # device -> list of (timestamp, token)
    tok_devices: Dict[Any, List[Tuple[int, Any]]] = {}  # token -> list of (timestamp, device)

    devices = df[device_col].values
    tokens = df[token_col].values
    timestamps = df[time_col].values

    for i in range(n):
        dev = devices[i]
        tok = tokens[i]
        t = int(timestamps[i])

        if pd.isna(dev) or str(dev) == "" or str(dev) == "nan":
            dev_uniq_tok_1h[i] = 0.0
        else:
            if dev not in dev_tokens:
                dev_uniq_tok_1h[i] = 0.0
                dev_tokens[dev] = [(t, tok)]
            else:
                t_list = dev_tokens[dev]
                min_1h = t - 3600
                idx = 0
                while idx < len(t_list) and t_list[idx][0] < min_1h:
                    idx += 1
                if idx > 0:
                    t_list = t_list[idx:]
                    dev_tokens[dev] = t_list
                uniq_tokens = len(set(tok_item[1] for tok_item in t_list))
                dev_uniq_tok_1h[i] = float(uniq_tokens)
                t_list.append((t, tok))

        if pd.isna(tok) or str(tok) == "" or str(tok) == "nan":
            tok_uniq_dev_1h[i] = 0.0
        else:
            if tok not in tok_devices:
                tok_uniq_dev_1h[i] = 0.0
                tok_devices[tok] = [(t, dev)]
            else:
                d_list = tok_devices[tok]
                min_1h = t - 3600
                idx = 0
                while idx < len(d_list) and d_list[idx][0] < min_1h:
                    idx += 1
                if idx > 0:
                    d_list = d_list[idx:]
                    tok_devices[tok] = d_list
                uniq_devs = len(set(dev_item[1] for dev_item in d_list))
                tok_uniq_dev_1h[i] = float(uniq_devs)
                d_list.append((t, dev))

    return {
        "device_unique_tokens_1h": dev_uniq_tok_1h,
        "token_unique_devices_1h": tok_uniq_dev_1h,
    }

def compute_point_in_time_reputation_signals(
    df: pd.DataFrame,
    device_col: str,
    time_col: str,
    label_col: str = "isFraud"
) -> Dict[str, np.ndarray]:
    """
    Computes deterministic reputation score, fraud rate, and dispute rate strictly prior to T (< T).
    """
    n = len(df)
    rep_scores = np.full(n, 0.50, dtype=np.float32)
    fraud_rates = np.zeros(n, dtype=np.float32)
    dispute_rates = np.zeros(n, dtype=np.float32)

    # device -> [total_tx, success_tx, fraud_tx, dispute_tx, first_seen_t]
    dev_rep_stats: Dict[Any, List[Any]] = {}

    devices = df[device_col].values
    timestamps = df[time_col].values
    labels = df[label_col].values if label_col in df.columns else np.zeros(n)

    for i in range(n):
        dev = devices[i]
        t = int(timestamps[i])
        is_fraud = int(labels[i]) if (not pd.isna(labels[i]) and labels[i] == 1) else 0

        if pd.isna(dev) or str(dev) == "" or str(dev) == "nan":
            rep_scores[i] = 0.50
            fraud_rates[i] = 0.0
            dispute_rates[i] = 0.0
            continue

        if dev not in dev_rep_stats:
            rep_scores[i] = 0.50
            fraud_rates[i] = 0.0
            dispute_rates[i] = 0.0
            # Initialize: total=1, success=(1 - is_fraud), fraud=is_fraud, dispute=is_fraud, first_seen=t
            dev_rep_stats[dev] = [1, 1 - is_fraud, is_fraud, is_fraud, t]
        else:
            tot, succ, fraud_cnt, disp_cnt, first_seen = dev_rep_stats[dev]

            # Calculate rates on past transactions (< T)
            eff_total = max(1.0, float(tot))
            f_rate = float(fraud_cnt) / eff_total
            d_rate = float(disp_cnt) / eff_total
            fraud_rates[i] = float(np.clip(f_rate, 0.0, 1.0))
            dispute_rates[i] = float(np.clip(d_rate, 0.0, 1.0))

            # Calculate deterministic score matching Phase 3.7
            score = 0.50
            if succ > 0:
                score -= min(0.30, float(succ) * 0.03)
            days_seen = max(0.0, float(t - first_seen) / 86400.0)
            if days_seen > 0:
                score -= min(0.15, days_seen * 0.005)

            if fraud_cnt > 0:
                score += 0.50
            if disp_cnt > 0:
                score += min(0.50, float(disp_cnt) * 0.25)
            if tot >= 3 and d_rate >= 0.10:
                score += 0.20

            rep_scores[i] = float(np.clip(score, 0.0, 1.0))

            # Update stats AFTER scoring
            dev_rep_stats[dev][0] += 1
            dev_rep_stats[dev][1] += (1 - is_fraud)
            dev_rep_stats[dev][2] += is_fraud
            dev_rep_stats[dev][3] += is_fraud

    return {
        "device_reputation_score": rep_scores,
        "device_fraud_rate": fraud_rates,
        "device_dispute_rate": dispute_rates,
    }

def extract_canonical_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Extracts canonical point-in-time safe features (15 core + 10 advanced) from raw IEEE-CIS dataframe.
    """
    df_sorted = df.sort_values(by="TransactionDT", ascending=True).reset_index(drop=True)
    n = len(df_sorted)
    
    # 1. Amount
    amount = df_sorted["TransactionAmt"].fillna(0.0).values.astype(np.float32)
    
    # 2. IP & Card/Token Proxies
    dev_col = "DeviceInfo" if "DeviceInfo" in df_sorted.columns else "DeviceType"
    card_proxy = df_sorted["card1"].astype(str) + "_" + df_sorted["card2"].fillna(0).astype(str)
    
    # 3. Rolling Velocities
    ip_vel_1h = compute_point_in_time_velocities(df_sorted, "addr1", "TransactionDT", window_seconds=3600)
    ip_vel_24h = compute_point_in_time_velocities(df_sorted, "addr1", "TransactionDT", window_seconds=86400)
    token_vel_24h = compute_point_in_time_velocities(df_sorted, "card1", "TransactionDT", window_seconds=86400)
    
    # 4. Device Novelty
    device_seen = compute_point_in_time_device_novelty(df_sorted, "card1", dev_col)
    
    # 5. Temporal Features
    dts = df_sorted["TransactionDT"].values
    tx_hour = ((dts % 86400) // 3600).astype(np.int32)
    tx_day = ((dts // 86400) % 7).astype(np.int32)
    
    # 6. Categoricals Raw
    prod_cd = df_sorted["ProductCD"].fillna("W").values
    card_type = df_sorted["card4"].fillna("visa").values
    card_cat = df_sorted["card6"].fillna("debit").values
    p_email = df_sorted["P_emaildomain"].fillna("missing").values
    
    # 7. Missing Value Indicators
    dist1_missing = df_sorted["dist1"].isna().astype(np.int32).values
    device_mobile = (df_sorted["DeviceType"] == "mobile").astype(np.int32).values if "DeviceType" in df_sorted.columns else np.zeros(n, dtype=np.int32)
    dev_missing = df_sorted[dev_col].isna().astype(np.int32).values if dev_col in df_sorted.columns else np.zeros(n, dtype=np.int32)
    
    # 8. Amount to Mean Ratio
    amt_ratio = compute_point_in_time_amount_ratio(df_sorted, "card1", "TransactionAmt")

    # 9. 10 New Advanced Behavioral, Velocity, Token, and Reputation Features (Phase 3.5 - 3.7)
    dev_vel_signals = compute_point_in_time_device_velocity_signals(df_sorted, dev_col, "TransactionDT", "TransactionAmt")
    tok_dev_signals = compute_point_in_time_token_device_linkage(df_sorted, dev_col, "card1", "TransactionDT")
    rep_signals = compute_point_in_time_reputation_signals(df_sorted, dev_col, "TransactionDT", "isFraud")
    
    feat_df = pd.DataFrame({
        "TransactionID": df_sorted["TransactionID"],
        "TransactionDT": df_sorted["TransactionDT"],
        # Core 15 Features
        "amount": amount,
        "ip_velocity_1h": ip_vel_1h,
        "ip_velocity_24h": ip_vel_24h,
        "token_velocity_24h": token_vel_24h,
        "device_seen_before": device_seen,
        "transaction_hour": tx_hour,
        "transaction_day": tx_day,
        "raw_product_cd": prod_cd,
        "raw_card_type": card_type,
        "raw_card_category": card_cat,
        "raw_email_domain": p_email,
        "dist1_missing": dist1_missing,
        "device_type_mobile": device_mobile,
        "device_info_missing": dev_missing,
        "amount_to_mean_ratio": amt_ratio,
        # 10 New Advanced Features
        "device_tx_count_5m": dev_vel_signals["device_tx_count_5m"],
        "device_tx_count_1h": dev_vel_signals["device_tx_count_1h"],
        "device_amount_sum_24h": dev_vel_signals["device_amount_sum_24h"],
        "tx_acceleration_5m_1h": dev_vel_signals["tx_acceleration_5m_1h"],
        "device_amount_concentration_5m_1h": dev_vel_signals["device_amount_concentration_5m_1h"],
        "device_unique_tokens_1h": tok_dev_signals["device_unique_tokens_1h"],
        "token_unique_devices_1h": tok_dev_signals["token_unique_devices_1h"],
        "device_reputation_score": rep_signals["device_reputation_score"],
        "device_fraud_rate": rep_signals["device_fraud_rate"],
        "device_dispute_rate": rep_signals["device_dispute_rate"],
    })
    
    if "isFraud" in df_sorted.columns:
        feat_df["isFraud"] = df_sorted["isFraud"].values
        
    return feat_df

extract_canonical_25_features = extract_canonical_features

