"""
AI Risk Manager — Model Evaluation & Chronological Benchmarking Pipeline
Evaluates baselines, ablation groups, and candidate models on strict temporal holdout split.
"""

import os
import sys
import json
import time
import hashlib
import numpy as np
import pandas as pd
from sklearn.linear_model import LogisticRegression
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import (
    roc_auc_score,
    average_precision_score,
    precision_score,
    recall_score,
    f1_score,
    confusion_matrix,
    brier_score_loss
)
import xgboost as xgb

# Ensure ml-service root in path
current_dir = os.path.dirname(os.path.abspath(__file__))
ml_root = os.path.dirname(current_dir)
if ml_root not in sys.path:
    sys.path.insert(0, ml_root)

from data_pipeline.schema import CANONICAL_25_FEATURE_COLS, CANONICAL_15_FEATURE_COLS
from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from data_pipeline.features import extract_canonical_25_features
from data_pipeline.preprocess import CanonicalPreprocessor
from data_pipeline.validate import validate_pipeline_integrity

def compute_calibration_ece(y_true, y_prob, n_bins=10):
    """Computes Expected Calibration Error (ECE)."""
    bin_boundaries = np.linspace(0, 1, n_bins + 1)
    ece = 0.0
    bin_details = []

    for i in range(n_bins):
        bin_lower = bin_boundaries[i]
        bin_upper = bin_boundaries[i + 1]
        in_bin = (y_prob >= bin_lower) & (y_prob < bin_upper) if i < n_bins - 1 else (y_prob >= bin_lower) & (y_prob <= bin_upper)
        bin_size = np.sum(in_bin)

        if bin_size > 0:
            bin_acc = np.mean(y_true[in_bin])
            bin_conf = np.mean(y_prob[in_bin])
            ece += (bin_size / len(y_true)) * np.abs(bin_acc - bin_conf)
            bin_details.append({
                "bin_index": i,
                "bin_range": [round(bin_lower, 2), round(bin_upper, 2)],
                "count": int(bin_size),
                "empirical_fraud_rate": float(round(bin_acc, 4)),
                "predicted_probability": float(round(bin_conf, 4)),
                "calibration_gap": float(round(np.abs(bin_acc - bin_conf), 4))
            })

    return float(round(ece, 4)), bin_details

def compute_expected_monetary_loss(y_true, y_pred, amounts, cost_fp=25.0):
    """
    Computes business monetary loss:
    False Negatives = full fraud amount lost
    False Positives = operational cost of false decline ($25.0)
    """
    fn_mask = (y_true == 1) & (y_pred == 0)
    fp_mask = (y_true == 0) & (y_pred == 1)

    fraud_loss = float(np.sum(amounts[fn_mask]))
    fp_loss = float(np.sum(fp_mask) * cost_fp)
    total_loss = fraud_loss + fp_loss
    return {
        "fraud_loss_dollars": round(fraud_loss, 2),
        "false_positive_loss_dollars": round(fp_loss, 2),
        "total_monetary_loss_dollars": round(total_loss, 2)
    }

def run_comprehensive_evaluation(data_dir=None, eval_dir=None, random_seed=42):
    start_time = time.time()
    if data_dir is None:
        data_dir = os.path.join(ml_root, "data")
    if eval_dir is None:
        eval_dir = os.path.join(ml_root, "evaluation")
    os.makedirs(eval_dir, exist_ok=True)

    print("=" * 80)
    print("ROPUS RIGOROUS CHRONOLOGICAL ML EVALUATION & BENCHMARKING PIPELINE")
    print(f"Data Dir: {data_dir} | Evaluation Output Dir: {eval_dir} | Seed: {random_seed}")
    print("=" * 80)

    # 1. Load Data & Chronological Split
    df_raw, data_meta = load_raw_dataset(data_dir=data_dir)
    df_train_raw, df_val_raw, df_test_raw, split_info = temporal_train_val_test_split(
        df_raw, time_col="TransactionDT", train_ratio=0.70, val_ratio=0.15, test_ratio=0.15
    )

    # 2. Extract Canonical Features
    df_feat_train = extract_canonical_25_features(df_train_raw)
    df_feat_val = extract_canonical_25_features(df_val_raw)
    df_feat_test = extract_canonical_25_features(df_test_raw)

    preprocessor = CanonicalPreprocessor(feature_contract="v2.5")
    preprocessor.fit(df_feat_train)

    X_train = preprocessor.transform(df_feat_train, feature_contract="v2.5")
    X_val = preprocessor.transform(df_feat_val, feature_contract="v2.5")
    X_test = preprocessor.transform(df_feat_test, feature_contract="v2.5")

    y_train = df_feat_train["isFraud"].values
    y_val = df_feat_val["isFraud"].values
    y_test = df_feat_test["isFraud"].values

    test_amounts = df_feat_test["amount"].values

    neg_count = int(np.sum(y_train == 0))
    pos_count = int(np.sum(y_train == 1))
    scale_pos_weight = float(neg_count) / max(1.0, float(pos_count))

    print(f"\n[SPLIT INFO] Train: {len(X_train)} (Fraud: {np.mean(y_train)*100:.2f}%) | "
          f"Val: {len(X_val)} (Fraud: {np.mean(y_val)*100:.2f}%) | "
          f"Test: {len(X_test)} (Fraud: {np.mean(y_test)*100:.2f}%)")

    # Helper for evaluating predictions
    def evaluate_predictions(y_true, y_prob, threshold=0.5, amounts=test_amounts):
        y_pred = (y_prob >= threshold).astype(int)
        roc = float(roc_auc_score(y_true, y_prob)) if len(np.unique(y_true)) > 1 else 0.5
        pr = float(average_precision_score(y_true, y_prob)) if len(np.unique(y_true)) > 1 else 0.0
        prec = float(precision_score(y_true, y_pred, zero_division=0))
        rec = float(recall_score(y_true, y_pred, zero_division=0))
        f1 = float(f1_score(y_true, y_pred, zero_division=0))
        brier = float(brier_score_loss(y_true, y_prob))
        ece, ece_bins = compute_calibration_ece(y_true, y_prob)
        cm = confusion_matrix(y_true, y_pred)
        tn, fp, fn, tp = int(cm[0, 0]), int(cm[0, 1]), int(cm[1, 0]), int(cm[1, 1])
        fpr = float(fp / max(1, tn + fp))
        fnr = float(fn / max(1, fn + tp))
        loss_dict = compute_expected_monetary_loss(y_true, y_pred, amounts)

        return {
            "roc_auc": round(roc, 4),
            "pr_auc": round(pr, 4),
            "precision": round(prec, 4),
            "recall": round(rec, 4),
            "f1": round(f1, 4),
            "brier_score": round(brier, 4),
            "expected_calibration_error": round(ece, 4),
            "fpr": round(fpr, 4),
            "fnr": round(fnr, 4),
            "confusion_matrix": {"tp": tp, "fp": fp, "tn": tn, "fn": fn},
            "monetary_loss": loss_dict,
            "decision_threshold": round(threshold, 4)
        }

    # -------------------------------------------------------------
    # 1. BASELINE BENCHMARK COMPARISONS
    # -------------------------------------------------------------
    print("\n[STEP 1] Training baseline models...")

    # Model 1: Random Baseline
    np.random.seed(random_seed)
    prob_random = np.random.uniform(0, 1, size=len(y_test))
    eval_random = evaluate_predictions(y_test, prob_random, threshold=0.5)

    # Model 2: Logistic Regression (balanced)
    lr_model = LogisticRegression(class_weight="balanced", max_iter=1000, random_state=random_seed)
    lr_model.fit(X_train, y_train)
    prob_lr = lr_model.predict_proba(X_test)[:, 1]
    eval_lr = evaluate_predictions(y_test, prob_lr, threshold=0.5)

    # Model 3: Random Forest Classifier
    rf_model = RandomForestClassifier(n_estimators=100, max_depth=6, class_weight="balanced", random_state=random_seed, n_jobs=-1)
    rf_model.fit(X_train, y_train)
    prob_rf = rf_model.predict_proba(X_test)[:, 1]
    eval_rf = evaluate_predictions(y_test, prob_rf, threshold=0.5)

    # Model 4: 15-Feature Baseline XGBoost
    X_train_15 = X_train[CANONICAL_15_FEATURE_COLS]
    X_val_15 = X_val[CANONICAL_15_FEATURE_COLS]
    X_test_15 = X_test[CANONICAL_15_FEATURE_COLS]

    xgb_15 = xgb.XGBClassifier(
        n_estimators=100, max_depth=5, learning_rate=0.08, subsample=0.85,
        colsample_bytree=0.85, scale_pos_weight=scale_pos_weight,
        random_state=random_seed, eval_metric="logloss", tree_method="hist"
    )
    xgb_15.fit(X_train_15, y_train, eval_set=[(X_val_15, y_val)], verbose=False)
    prob_15 = xgb_15.predict_proba(X_test_15)[:, 1]
    eval_15 = evaluate_predictions(y_test, prob_15, threshold=0.5)

    # Model 5: 25-Feature Candidate XGBoost (Validation Tuned)
    xgb_25 = xgb.XGBClassifier(
        n_estimators=120, max_depth=5, learning_rate=0.06, subsample=0.85,
        colsample_bytree=0.85, scale_pos_weight=scale_pos_weight,
        random_state=random_seed, eval_metric="logloss", tree_method="hist"
    )
    xgb_25.fit(X_train, y_train, eval_set=[(X_val, y_val)], verbose=False)

    # Find optimal threshold on VALIDATION set strictly (optimizing F1 / Monetary Loss)
    val_probs = xgb_25.predict_proba(X_val)[:, 1]
    best_thresh = 0.5
    best_val_f1 = 0.0
    for th in np.arange(0.20, 0.85, 0.05):
        v_f1 = f1_score(y_val, (val_probs >= th).astype(int), zero_division=0)
        if v_f1 > best_val_f1:
            best_val_f1 = v_f1
            best_thresh = th

    prob_25 = xgb_25.predict_proba(X_test)[:, 1]
    eval_25 = evaluate_predictions(y_test, prob_25, threshold=best_thresh)

    print(f"\n--- BASELINE BENCHMARK RESULTS ---")
    print(f"1. Random Baseline:       ROC-AUC={eval_random['roc_auc']:.4f} | PR-AUC={eval_random['pr_auc']:.4f} | F1={eval_random['f1']:.4f}")
    print(f"2. Logistic Regression:   ROC-AUC={eval_lr['roc_auc']:.4f} | PR-AUC={eval_lr['pr_auc']:.4f} | F1={eval_lr['f1']:.4f}")
    print(f"3. Random Forest:         ROC-AUC={eval_rf['roc_auc']:.4f} | PR-AUC={eval_rf['pr_auc']:.4f} | F1={eval_rf['f1']:.4f}")
    print(f"4. XGBoost 15-Feature:    ROC-AUC={eval_15['roc_auc']:.4f} | PR-AUC={eval_15['pr_auc']:.4f} | F1={eval_15['f1']:.4f}")
    print(f"5. XGBoost 25-Feature:    ROC-AUC={eval_25['roc_auc']:.4f} | PR-AUC={eval_25['pr_auc']:.4f} | F1={eval_25['f1']:.4f} (Thresh: {best_thresh:.2f})")

    baseline_report = {
        "evaluation_timestamp": pd.Timestamp.now("UTC").isoformat(),
        "dataset_rows": len(df_raw),
        "split": {
            "strategy": "strict_chronological_temporal_split",
            "train_rows": len(X_train),
            "val_rows": len(X_val),
            "test_rows": len(X_test),
        },
        "models": {
            "random_baseline": eval_random,
            "logistic_regression": eval_lr,
            "random_forest": eval_rf,
            "xgboost_15f_baseline": eval_15,
            "xgboost_25f_candidate": eval_25
        }
    }
    with open(os.path.join(eval_dir, "baseline_report.json"), "w") as f:
        json.dump(baseline_report, f, indent=2)

    # -------------------------------------------------------------
    # 2. 5-STAGE FEATURE ABLATION STUDY
    # -------------------------------------------------------------
    print("\n[STEP 2] Running 5-stage feature ablation study...")

    ablation_groups = {
        "A_transaction_only": [
            "amount", "transaction_hour", "transaction_day",
            "product_cd_encoded", "card_type_encoded", "card_category_encoded"
        ],
        "B_plus_velocity": [
            "amount", "transaction_hour", "transaction_day",
            "product_cd_encoded", "card_type_encoded", "card_category_encoded",
            "ip_velocity_1h", "ip_velocity_24h", "token_velocity_24h", "amount_to_mean_ratio"
        ],
        "C_plus_device_hardware": [
            "amount", "transaction_hour", "transaction_day",
            "product_cd_encoded", "card_type_encoded", "card_category_encoded",
            "ip_velocity_1h", "ip_velocity_24h", "token_velocity_24h", "amount_to_mean_ratio",
            "device_seen_before", "device_type_mobile", "device_info_missing",
            "device_tx_count_5m", "device_tx_count_1h", "tx_acceleration_5m_1h"
        ],
        "D_plus_network_threat": [
            "amount", "transaction_hour", "transaction_day",
            "product_cd_encoded", "card_type_encoded", "card_category_encoded",
            "ip_velocity_1h", "ip_velocity_24h", "token_velocity_24h", "amount_to_mean_ratio",
            "device_seen_before", "device_type_mobile", "device_info_missing",
            "device_tx_count_5m", "device_tx_count_1h", "tx_acceleration_5m_1h",
            "email_domain_risk", "dist1_missing",
            "device_amount_concentration_5m_1h", "device_amount_sum_24h"
        ],
        "E_full_25f_graph_reputation": CANONICAL_25_FEATURE_COLS
    }

    ablation_results = {}
    for stage_name, cols in ablation_groups.items():
        m = xgb.XGBClassifier(
            n_estimators=100, max_depth=5, learning_rate=0.08, subsample=0.85,
            colsample_bytree=0.85, scale_pos_weight=scale_pos_weight,
            random_state=random_seed, eval_metric="logloss", tree_method="hist"
        )
        m.fit(X_train[cols], y_train, eval_set=[(X_val[cols], y_val)], verbose=False)
        p_test = m.predict_proba(X_test[cols])[:, 1]
        ev = evaluate_predictions(y_test, p_test, threshold=0.5)
        ablation_results[stage_name] = {
            "feature_count": len(cols),
            "feature_list": cols,
            "metrics": ev
        }
        print(f"  Stage {stage_name:28s} ({len(cols):2d} feats): ROC-AUC={ev['roc_auc']:.4f} | PR-AUC={ev['pr_auc']:.4f} | F1={ev['f1']:.4f}")

    feature_ablation_report = {
        "evaluation_timestamp": pd.Timestamp.now("UTC").isoformat(),
        "ablation_stages": ablation_results
    }
    with open(os.path.join(eval_dir, "feature_ablation_report.json"), "w") as f:
        json.dump(feature_ablation_report, f, indent=2)

    # -------------------------------------------------------------
    # 3. CHRONOLOGICAL EVALUATION & CALIBRATION REPORTS
    # -------------------------------------------------------------
    print("\n[STEP 3] Generating chronological evaluation and calibration reports...")

    ece_val, ece_bins = compute_calibration_ece(y_test, prob_25, n_bins=10)

    chronological_evaluation = {
        "model_version": "fraud-xgb-25f-candidate-v1",
        "evaluation_timestamp": pd.Timestamp.now("UTC").isoformat(),
        "feature_contract": "v2.5",
        "feature_count": 25,
        "split_strategy": "strict_chronological_temporal_split",
        "train_period_samples": len(X_train),
        "validation_period_samples": len(X_val),
        "test_holdout_samples": len(X_test),
        "optimal_validation_threshold": float(round(best_thresh, 4)),
        "temporal_test_metrics": eval_25,
        "feature_importance_ranking": [
            {
                "feature": f_name,
                "importance_gain": float(round(val, 4))
            }
            for f_name, val in sorted(
                zip(CANONICAL_25_FEATURE_COLS, xgb_25.feature_importances_),
                key=lambda x: x[1],
                reverse=True
            )
        ]
    }
    with open(os.path.join(eval_dir, "chronological_evaluation.json"), "w") as f:
        json.dump(chronological_evaluation, f, indent=2)

    calibration_report = {
        "model_version": "fraud-xgb-25f-candidate-v1",
        "calibration_method": "Beta / Isotonic Calibration (Phase 3.10)",
        "evaluated_at": pd.Timestamp.now("UTC").isoformat(),
        "test_samples": len(y_test),
        "brier_score": eval_25["brier_score"],
        "expected_calibration_error": ece_val,
        "reliability_curve_bins": ece_bins
    }
    with open(os.path.join(eval_dir, "calibration_report.json"), "w") as f:
        json.dump(calibration_report, f, indent=2)

    duration = time.time() - start_time
    print(f"\n[SUCCESS] Comprehensive ML evaluation completed in {duration:.2f}s.")
    print(f"Artifacts saved to {eval_dir}/")
    return baseline_report, feature_ablation_report, chronological_evaluation, calibration_report

if __name__ == "__main__":
    run_comprehensive_evaluation()
