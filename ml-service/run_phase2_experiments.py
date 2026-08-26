"""
ROPUS Phase 2 Controlled Experiments Suite
Evaluates feature additions and ranking improvements strictly on Train + Validation splits.
"""

import os
import sys
import json
import numpy as np
import pandas as pd
import xgboost as xgb
from sklearn.metrics import (
    roc_auc_score,
    average_precision_score,
    precision_score,
    recall_score,
    f1_score,
    confusion_matrix,
    accuracy_score
)

# Add ml-service to path
current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from data_pipeline.features import (
    compute_point_in_time_velocities,
    compute_point_in_time_device_novelty,
    compute_point_in_time_amount_ratio,
    compute_point_in_time_device_velocity_signals,
    compute_point_in_time_token_device_linkage,
    compute_point_in_time_reputation_signals
)
from data_pipeline.schema import CANONICAL_25_FEATURE_COLS

def extract_phase2_rich_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Extracts canonical features plus point-in-time safe Phase 2 candidate features.
    All calculations are strictly causal and use information prior to or at time T.
    """
    df_sorted = df.sort_values(by="TransactionDT", ascending=True).reset_index(drop=True)
    n = len(df_sorted)

    # Base canonical features with robust fallback for missing fields
    amount = df_sorted["TransactionAmt"].fillna(0.0).values.astype(np.float32) if "TransactionAmt" in df_sorted.columns else np.zeros(n, dtype=np.float32)
    dev_col = "DeviceInfo" if "DeviceInfo" in df_sorted.columns else ("DeviceType" if "DeviceType" in df_sorted.columns else "device_id")
    if dev_col not in df_sorted.columns:
        df_sorted[dev_col] = "unknown"
    if "addr1" not in df_sorted.columns:
        df_sorted["addr1"] = 0.0
    if "card1" not in df_sorted.columns:
        df_sorted["card1"] = 0
    if "TransactionDT" not in df_sorted.columns:
        df_sorted["TransactionDT"] = 0
    if "TransactionAmt" not in df_sorted.columns:
        df_sorted["TransactionAmt"] = 0.0

    ip_vel_1h = compute_point_in_time_velocities(df_sorted, "addr1", "TransactionDT", window_seconds=3600)
    ip_vel_24h = compute_point_in_time_velocities(df_sorted, "addr1", "TransactionDT", window_seconds=86400)
    token_vel_24h = compute_point_in_time_velocities(df_sorted, "card1", "TransactionDT", window_seconds=86400)

    device_seen = compute_point_in_time_device_novelty(df_sorted, "card1", dev_col)

    dts = df_sorted["TransactionDT"].values
    tx_hour = ((dts % 86400) // 3600).astype(np.int32)
    tx_day = ((dts // 86400) % 7).astype(np.int32)

    prod_cd = df_sorted["ProductCD"].fillna("W").values if "ProductCD" in df_sorted.columns else np.full(n, "W")
    card_type = df_sorted["card4"].fillna("visa").values if "card4" in df_sorted.columns else np.full(n, "visa")
    card_cat = df_sorted["card6"].fillna("debit").values if "card6" in df_sorted.columns else np.full(n, "debit")
    p_email = df_sorted["P_emaildomain"].fillna("missing").values if "P_emaildomain" in df_sorted.columns else np.full(n, "missing")

    dist1_missing = df_sorted["dist1"].isna().astype(np.int32).values if "dist1" in df_sorted.columns else np.ones(n, dtype=np.int32)
    device_mobile = (df_sorted["DeviceType"] == "mobile").astype(np.int32).values if "DeviceType" in df_sorted.columns else np.zeros(n, dtype=np.int32)
    dev_missing = df_sorted[dev_col].isna().astype(np.int32).values if dev_col in df_sorted.columns else np.ones(n, dtype=np.int32)

    amt_ratio = compute_point_in_time_amount_ratio(df_sorted, "card1", "TransactionAmt")
    dev_vel_signals = compute_point_in_time_device_velocity_signals(df_sorted, dev_col, "TransactionDT", "TransactionAmt")
    tok_dev_signals = compute_point_in_time_token_device_linkage(df_sorted, dev_col, "card1", "TransactionDT")
    rep_signals = compute_point_in_time_reputation_signals(df_sorted, dev_col, "TransactionDT", "isFraud")

    # ----------------- NEW PHASE 2 CANDIDATE FEATURES (Point-in-time safe) -----------------
    # 1. Log Amount (Stabilizes heavy tail)
    log_amount = np.log1p(np.maximum(0.0, amount)).astype(np.float32)

    # 2. Cyclic Hour Encodings (Captures midnight continuity)
    sin_hour = np.sin(2.0 * np.pi * tx_hour / 24.0).astype(np.float32)
    cos_hour = np.cos(2.0 * np.pi * tx_hour / 24.0).astype(np.float32)

    # 3. Velocity Burst Ratios (Ratio of recent 1h surge to 24h baseline)
    ip_burst_ratio = ((ip_vel_1h + 1.0) / (ip_vel_24h + 1.0)).astype(np.float32)
    dev_burst_ratio = ((dev_vel_signals["device_tx_count_1h"] + 1.0) / (dev_vel_signals["device_amount_sum_24h"] / (amount + 1.0) + 1.0)).astype(np.float32)

    # 4. Compound Interactions
    # High value on new device
    amt_novelty_risk = (log_amount * (1.0 - device_seen)).astype(np.float32)
    # High value with high IP velocity
    amt_ip_vel_risk = (log_amount * np.log1p(ip_vel_1h)).astype(np.float32)

    # 5. Causal IP-to-Device Fan-out (Tracking distinct devices on same IP in past 1h)
    ip_dev_fanout_1h = np.zeros(n, dtype=np.float32)
    ip_devices = {}
    ips = df_sorted["addr1"].values
    devices = df_sorted[dev_col].values

    for i in range(n):
        ip = ips[i]
        dev = devices[i]
        t = int(dts[i])

        if pd.isna(ip) or str(ip) == "" or pd.isna(dev):
            ip_dev_fanout_1h[i] = 1.0
            continue

        if ip not in ip_devices:
            ip_devices[ip] = [(t, dev)]
            ip_dev_fanout_1h[i] = 1.0
        else:
            d_list = ip_devices[ip]
            min_1h = t - 3600
            idx = 0
            while idx < len(d_list) and d_list[idx][0] < min_1h:
                idx += 1
            if idx > 0:
                d_list = d_list[idx:]
                ip_devices[ip] = d_list
            uniq_d = len(set(item[1] for item in d_list))
            ip_dev_fanout_1h[i] = float(uniq_d)
            d_list.append((t, dev))

    # 6. Causal Device Recency Delta (Seconds since previous transaction on same device)
    recency_device_seconds = np.full(n, 86400.0, dtype=np.float32)
    last_seen_device = {}
    for i in range(n):
        dev = devices[i]
        t = int(dts[i])
        if not pd.isna(dev) and str(dev) != "":
            if dev in last_seen_device:
                delta = t - last_seen_device[dev]
                recency_device_seconds[i] = float(min(86400, delta))
            last_seen_device[dev] = t
    log_device_recency = np.log1p(recency_device_seconds).astype(np.float32)

    feat_df = pd.DataFrame({
        "TransactionID": df_sorted["TransactionID"],
        "TransactionDT": df_sorted["TransactionDT"],
        # Core 15
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
        # Advanced 10
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
        # Phase 2 Candidates
        "log_amount": log_amount,
        "sin_tx_hour": sin_hour,
        "cos_tx_hour": cos_hour,
        "ip_burst_ratio": ip_burst_ratio,
        "dev_burst_ratio": dev_burst_ratio,
        "amt_novelty_risk": amt_novelty_risk,
        "amt_ip_vel_risk": amt_ip_vel_risk,
        "ip_dev_fanout_1h": ip_dev_fanout_1h,
        "log_device_recency": log_device_recency,
    })

    if "isFraud" in df_sorted.columns:
        feat_df["isFraud"] = df_sorted["isFraud"].values

    return feat_df

def run_experiments():
    print("=" * 80)
    print("PHASE 2: FEATURE SIGNAL & RANKING IMPROVEMENT EXPERIMENTS")
    print("=" * 80)

    # 1. Load and Split
    df_raw, _ = load_raw_dataset()
    df_train_raw, df_val_raw, df_test_raw, split_info = temporal_train_val_test_split(df_raw)

    print(f"Train: {len(df_train_raw)} (Frauds: {split_info['train_fraud_rate']:.4%})")
    print(f"Val:   {len(df_val_raw)} (Frauds: {split_info['val_fraud_rate']:.4%})")
    print(f"Test:  {len(df_test_raw)} (Frauds: {split_info['test_fraud_rate']:.4%}) [KEPT UNTOUCHED]")

    # 2. Extract Phase 2 Features
    df_feat_train = extract_phase2_rich_features(df_train_raw)
    df_feat_val = extract_phase2_rich_features(df_val_raw)

    # Categorical and numerical mappings fitted strictly on train
    unique_prods = sorted(df_feat_train["raw_product_cd"].dropna().unique().tolist())
    product_map = {prod: idx for idx, prod in enumerate(unique_prods)}

    unique_cards = sorted(df_feat_train["raw_card_type"].dropna().unique().tolist())
    card_type_map = {card: idx for idx, card in enumerate(unique_cards)}

    unique_cats = sorted(df_feat_train["raw_card_category"].dropna().unique().tolist())
    card_cat_map = {cat: idx for idx, cat in enumerate(unique_cats)}

    global_fraud_prior = float(df_feat_train["isFraud"].mean())
    domain_stats = df_feat_train.groupby("raw_email_domain")["isFraud"].agg(["count", "sum"])
    email_domain_risk = {}
    for domain, row in domain_stats.iterrows():
        cnt = row["count"]
        fraud_sum = row["sum"]
        smoothed_risk = (fraud_sum + (20.0 * global_fraud_prior)) / (cnt + 20.0)
        email_domain_risk[str(domain)] = float(round(smoothed_risk, 4))

    def transform_split(df_f):
        out = pd.DataFrame(index=df_f.index)
        for col in [
            "amount", "ip_velocity_1h", "ip_velocity_24h", "token_velocity_24h", "amount_to_mean_ratio",
            "device_tx_count_5m", "device_tx_count_1h", "device_amount_sum_24h", "tx_acceleration_5m_1h",
            "device_amount_concentration_5m_1h", "device_unique_tokens_1h", "token_unique_devices_1h",
            "device_reputation_score", "device_fraud_rate", "device_dispute_rate",
            "log_amount", "sin_tx_hour", "cos_tx_hour", "ip_burst_ratio", "dev_burst_ratio",
            "amt_novelty_risk", "amt_ip_vel_risk", "ip_dev_fanout_1h", "log_device_recency"
        ]:
            out[col] = df_f[col].fillna(0.0).astype(np.float32)

        out["device_seen_before"] = df_f["device_seen_before"].fillna(0).astype(np.int32)
        out["transaction_hour"] = df_f["transaction_hour"].fillna(12).astype(np.int32)
        out["transaction_day"] = df_f["transaction_day"].fillna(0).astype(np.int32)
        out["dist1_missing"] = df_f["dist1_missing"].fillna(1).astype(np.int32)
        out["device_type_mobile"] = df_f["device_type_mobile"].fillna(0).astype(np.int32)
        out["device_info_missing"] = df_f["device_info_missing"].fillna(0).astype(np.int32)

        out["product_cd_encoded"] = df_f["raw_product_cd"].map(lambda x: product_map.get(str(x), -1)).astype(np.int32)
        out["card_type_encoded"] = df_f["raw_card_type"].map(lambda x: card_type_map.get(str(x), -1)).astype(np.int32)
        out["card_category_encoded"] = df_f["raw_card_category"].map(lambda x: card_cat_map.get(str(x), -1)).astype(np.int32)
        out["email_domain_risk"] = df_f["raw_email_domain"].map(lambda x: email_domain_risk.get(str(x), global_fraud_prior)).astype(np.float32)
        return out

    X_train_all = transform_split(df_feat_train)
    X_val_all = transform_split(df_feat_val)

    y_train = df_feat_train["isFraud"].values
    y_val = df_feat_val["isFraud"].values

    scale_pos_weight = float(np.sum(y_train == 0)) / max(1.0, float(np.sum(y_train == 1)))

    # Define Controlled Experiments
    experiments = {
        "Exp A: Baseline 25F": {
            "features": CANONICAL_25_FEATURE_COLS,
            "params": {"max_depth": 5, "learning_rate": 0.08, "n_estimators": 100, "subsample": 0.85, "colsample_bytree": 0.85}
        },
        "Exp B: Prune Low-Signal (23F)": {
            "features": [c for c in CANONICAL_25_FEATURE_COLS if c not in ["token_unique_devices_1h", "device_dispute_rate"]],
            "params": {"max_depth": 5, "learning_rate": 0.08, "n_estimators": 100, "subsample": 0.85, "colsample_bytree": 0.85}
        },
        "Exp C: Amount & Time Cyclic (+log_amt, sin/cos_hr)": {
            "features": [c for c in CANONICAL_25_FEATURE_COLS if c not in ["token_unique_devices_1h", "device_dispute_rate"]] + [
                "log_amount", "sin_tx_hour", "cos_tx_hour"
            ],
            "params": {"max_depth": 5, "learning_rate": 0.08, "n_estimators": 100, "subsample": 0.85, "colsample_bytree": 0.85}
        },
        "Exp D: Add Burst Ratios & Recency (+burst, +recency)": {
            "features": [c for c in CANONICAL_25_FEATURE_COLS if c not in ["token_unique_devices_1h", "device_dispute_rate"]] + [
                "log_amount", "sin_tx_hour", "cos_tx_hour", "ip_burst_ratio", "log_device_recency"
            ],
            "params": {"max_depth": 5, "learning_rate": 0.08, "n_estimators": 100, "subsample": 0.85, "colsample_bytree": 0.85}
        },
        "Exp E: Add Interactions (+amt_novelty, +amt_ip_vel, +fanout)": {
            "features": [c for c in CANONICAL_25_FEATURE_COLS if c not in ["token_unique_devices_1h", "device_dispute_rate"]] + [
                "log_amount", "sin_tx_hour", "cos_tx_hour", "ip_burst_ratio", "log_device_recency",
                "amt_novelty_risk", "amt_ip_vel_risk", "ip_dev_fanout_1h"
            ],
            "params": {"max_depth": 5, "learning_rate": 0.08, "n_estimators": 100, "subsample": 0.85, "colsample_bytree": 0.85}
        },
        "Exp F: Tuned Combined Candidate (Regularized & Depth 4)": {
            "features": [c for c in CANONICAL_25_FEATURE_COLS if c not in ["token_unique_devices_1h", "device_dispute_rate"]] + [
                "log_amount", "sin_tx_hour", "cos_tx_hour", "ip_burst_ratio", "log_device_recency",
                "amt_novelty_risk", "amt_ip_vel_risk", "ip_dev_fanout_1h"
            ],
            "params": {"max_depth": 4, "learning_rate": 0.05, "n_estimators": 120, "subsample": 0.80, "colsample_bytree": 0.80, "reg_lambda": 2.0, "min_child_weight": 3}
        }
    }

    results = {}
    print("\n" + "=" * 90)
    print(f"{'Experiment':<42} | {'Val ROC':<9} | {'Val PR-AUC':<11} | {'Val F1':<8} | {'Val Prec':<9} | {'Val Rec':<8}")
    print("=" * 90)

    for exp_name, cfg in experiments.items():
        feat_cols = cfg["features"]
        X_tr = X_train_all[feat_cols]
        X_va = X_val_all[feat_cols]

        p = cfg["params"]
        model = xgb.XGBClassifier(
            max_depth=p.get("max_depth", 5),
            learning_rate=p.get("learning_rate", 0.08),
            n_estimators=p.get("n_estimators", 100),
            subsample=p.get("subsample", 0.85),
            colsample_bytree=p.get("colsample_bytree", 0.85),
            reg_lambda=p.get("reg_lambda", 1.0),
            min_child_weight=p.get("min_child_weight", 1),
            scale_pos_weight=scale_pos_weight,
            random_state=42,
            eval_metric="logloss",
            tree_method="hist"
        )

        model.fit(X_tr, y_train, eval_set=[(X_va, y_val)], verbose=False)
        probs_val = model.predict_proba(X_va)[:, 1]
        preds_val = (probs_val >= 0.50).astype(int)

        roc_v = float(roc_auc_score(y_val, probs_val))
        pr_v = float(average_precision_score(y_val, probs_val))
        acc_v = float(accuracy_score(y_val, preds_val))
        prec_v = float(precision_score(y_val, preds_val, zero_division=0))
        rec_v = float(recall_score(y_val, preds_val, zero_division=0))
        f1_v = float(f1_score(y_val, preds_val, zero_division=0))
        tn_v, fp_v, fn_v, tp_v = confusion_matrix(y_val, preds_val).ravel()
        fpr_v = float(fp_v / (fp_v + tn_v)) if (fp_v + tn_v) > 0 else 0.0

        results[exp_name] = {
            "feature_count": len(feat_cols),
            "features": feat_cols,
            "hyperparameters": p,
            "val_metrics": {
                "roc_auc": roc_v,
                "pr_auc": pr_v,
                "f1": f1_v,
                "precision": prec_v,
                "recall": rec_v,
                "fpr": fpr_v,
                "accuracy": acc_v,
                "tp": int(tp_v),
                "fp": int(fp_v),
                "tn": int(tn_v),
                "fn": int(fn_v)
            }
        }

        print(f"{exp_name:<42} | {roc_v:<9.4f} | {pr_v:<11.4f} | {f1_v:<8.4f} | {prec_v:<9.4f} | {rec_v:<8.4f}")

    print("=" * 90)

    # Save validation experiment results
    eval_dir = os.path.join(current_dir, "evaluation")
    os.makedirs(eval_dir, exist_ok=True)
    exp_out_path = os.path.join(eval_dir, "phase_2_experiment_results.json")
    with open(exp_out_path, "w") as f:
        json.dump(results, f, indent=2)
    print(f"\nSaved experiment results to: {exp_out_path}")

    return results

if __name__ == "__main__":
    run_experiments()
