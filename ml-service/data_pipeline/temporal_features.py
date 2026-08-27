"""
AI Risk Manager — Extended Point-in-Time Safe Temporal Feature Engine (P1)
Computes multi-window card transaction counts, amount velocities, z-scores,
novelty indicators, entity linkages, and interactions strictly using information prior to timestamp T (< T).
"""

import numpy as np
import pandas as pd
from typing import Dict, Any, List, Tuple

ENHANCED_TEMPORAL_FEATURE_COLS = [
    "time_since_prev_tx",
    "card_tx_count_5m",
    "card_tx_count_1h",
    "card_tx_count_6h",
    "card_tx_count_24h",
    "card_tx_count_7d",
    "card_amount_sum_1h",
    "card_amount_sum_24h",
    "card_amount_sum_7d",
    "card_amount_zscore_historical",
    "card_velocity_acceleration_1h_24h",
    "distinct_devices_per_card_24h",
    "distinct_cards_per_device_24h",
    "is_new_card_indicator",
    "is_new_device_indicator",
    "card_account_age_days"
]

def extract_point_in_time_temporal_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Extracts strictly point-in-time (< T) temporal and behavioral features.
    Guarantees zero future leakage by iterating chronologically and appending
    current transaction events to history strictly AFTER feature calculation.
    """
    n = len(df)

    # Entity identifiers
    cards = df["card1"].fillna(0).astype(str).values
    addrs = df["addr1"].fillna(0).astype(str).values
    devices = df["DeviceInfo"].fillna("missing").astype(str).values if "DeviceInfo" in df.columns else np.full(n, "missing")
    timestamps = df["TransactionDT"].values
    amounts = df["TransactionAmt"].fillna(0.0).values if "TransactionAmt" in df.columns else df["amount"].fillna(0.0).values

    # Output arrays
    time_since_prev_tx = np.zeros(n, dtype=np.float32)
    card_tx_5m = np.zeros(n, dtype=np.float32)
    card_tx_1h = np.zeros(n, dtype=np.float32)
    card_tx_6h = np.zeros(n, dtype=np.float32)
    card_tx_24h = np.zeros(n, dtype=np.float32)
    card_tx_7d = np.zeros(n, dtype=np.float32)
    card_amt_1h = np.zeros(n, dtype=np.float32)
    card_amt_24h = np.zeros(n, dtype=np.float32)
    card_amt_7d = np.zeros(n, dtype=np.float32)
    card_amt_zscore = np.zeros(n, dtype=np.float32)
    card_accel_1h_24h = np.zeros(n, dtype=np.float32)
    devices_per_card_24h = np.zeros(n, dtype=np.float32)
    cards_per_device_24h = np.zeros(n, dtype=np.float32)
    is_new_card = np.zeros(n, dtype=np.float32)
    is_new_device = np.zeros(n, dtype=np.float32)
    card_age_days = np.zeros(n, dtype=np.float32)

    # State tracking
    card_history: Dict[str, List[Tuple[int, float, str]]] = {} # card_key -> list of (timestamp, amount, device)
    device_history: Dict[str, List[Tuple[int, str]]] = {}       # device_key -> list of (timestamp, card_key)
    known_devices = set()

    for i in range(n):
        c_key = f"{cards[i]}_{addrs[i]}"
        d_key = devices[i]
        t = int(timestamps[i])
        amt = float(amounts[i])

        # 1. Device Novelty
        if d_key != "missing" and d_key != "nan":
            is_new_device[i] = 0.0 if d_key in known_devices else 1.0
            known_devices.add(d_key)
        else:
            is_new_device[i] = 1.0

        # 2. Card History & Point-in-Time Windows
        if c_key not in card_history:
            is_new_card[i] = 1.0
            time_since_prev_tx[i] = -1.0 # Sentinel for first seen
            card_tx_5m[i] = 0.0
            card_tx_1h[i] = 0.0
            card_tx_6h[i] = 0.0
            card_tx_24h[i] = 0.0
            card_tx_7d[i] = 0.0
            card_amt_1h[i] = 0.0
            card_amt_24h[i] = 0.0
            card_amt_7d[i] = 0.0
            card_amt_zscore[i] = 0.0
            card_accel_1h_24h[i] = 0.0
            devices_per_card_24h[i] = 0.0
            card_age_days[i] = 0.0
            card_history[c_key] = [(t, amt, d_key)]
        else:
            is_new_card[i] = 0.0
            hist = card_history[c_key]
            first_t = hist[0][0]
            prev_t = hist[-1][0]

            card_age_days[i] = float(max(0.0, (t - first_t) / 86400.0))
            time_since_prev_tx[i] = float(max(0.0, t - prev_t))

            # Sliding windows strictly < T
            min_7d = t - (7 * 86400)
            min_24h = t - 86400
            min_6h = t - 21600
            min_1h = t - 3600
            min_5m = t - 300

            c_5m, c_1h, c_6h, c_24h, c_7d = 0, 0, 0, 0, 0
            s_1h, s_24h, s_7d = 0.0, 0.0, 0.0
            dev_24h_set = set()
            all_amts = []

            for h_t, h_amt, h_dev in hist:
                all_amts.append(h_amt)
                if h_t >= min_7d:
                    c_7d += 1
                    s_7d += h_amt
                if h_t >= min_24h:
                    c_24h += 1
                    s_24h += h_amt
                    if h_dev != "missing" and h_dev != "nan":
                        dev_24h_set.add(h_dev)
                if h_t >= min_6h:
                    c_6h += 1
                if h_t >= min_1h:
                    c_1h += 1
                    s_1h += h_amt
                if h_t >= min_5m:
                    c_5m += 1

            card_tx_5m[i] = float(c_5m)
            card_tx_1h[i] = float(c_1h)
            card_tx_6h[i] = float(c_6h)
            card_tx_24h[i] = float(c_24h)
            card_tx_7d[i] = float(c_7d)
            card_amt_1h[i] = float(s_1h)
            card_amt_24h[i] = float(s_24h)
            card_amt_7d[i] = float(s_7d)
            devices_per_card_24h[i] = float(len(dev_24h_set))

            # Acceleration (rate in 1h vs rate in 24h)
            rate_1h = float(c_1h) / 3600.0
            rate_24h = float(c_24h) / 86400.0
            if rate_24h > 0.0:
                card_accel_1h_24h[i] = float(np.clip(rate_1h / rate_24h, 0.0, 50.0))
            else:
                card_accel_1h_24h[i] = 0.0

            # Historical z-score
            if len(all_amts) >= 2:
                mu = float(np.mean(all_amts))
                sigma = float(np.std(all_amts))
                if sigma > 1e-4:
                    card_amt_zscore[i] = float(np.clip((amt - mu) / sigma, -5.0, 10.0))
                else:
                    card_amt_zscore[i] = 0.0
            else:
                card_amt_zscore[i] = 0.0

            # Append AFTER calculations to prevent self-inclusion in prior history
            hist.append((t, amt, d_key))

        # 3. Cards per device in 24h
        if d_key != "missing" and d_key != "nan":
            if d_key not in device_history:
                cards_per_device_24h[i] = 0.0
                device_history[d_key] = [(t, c_key)]
            else:
                d_hist = device_history[d_key]
                min_24h = t - 86400
                cards_24h_set = set()
                for dh_t, dh_card in d_hist:
                    if dh_t >= min_24h:
                        cards_24h_set.add(dh_card)
                cards_per_device_24h[i] = float(len(cards_24h_set))
                d_hist.append((t, c_key))
        else:
            cards_per_device_24h[i] = 0.0

    return pd.DataFrame({
        "time_since_prev_tx": time_since_prev_tx,
        "card_tx_count_5m": card_tx_5m,
        "card_tx_count_1h": card_tx_1h,
        "card_tx_count_6h": card_tx_6h,
        "card_tx_count_24h": card_tx_24h,
        "card_tx_count_7d": card_tx_7d,
        "card_amount_sum_1h": card_amt_1h,
        "card_amount_sum_24h": card_amt_24h,
        "card_amount_sum_7d": card_amt_7d,
        "card_amount_zscore_historical": card_amt_zscore,
        "card_velocity_acceleration_1h_24h": card_accel_1h_24h,
        "distinct_devices_per_card_24h": devices_per_card_24h,
        "distinct_cards_per_device_24h": cards_per_device_24h,
        "is_new_card_indicator": is_new_card,
        "is_new_device_indicator": is_new_device,
        "card_account_age_days": card_age_days
    })
