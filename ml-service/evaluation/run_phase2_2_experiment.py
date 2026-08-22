"""
Phase 2.2 — Calibration Selection & Score Resolution Experiment
Evaluates RAW, PLATT, ISOTONIC, and BETA calibrators across statistical, resolution, and operational metrics.
"""

import os
import sys
import json
from typing import Tuple, Dict, Any, List
import numpy as np
import pandas as pd
from scipy.stats import entropy
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import (
    brier_score_loss,
    log_loss,
    roc_auc_score,
    average_precision_score
)

# Ensure ml-service root in path
current_dir = os.path.dirname(os.path.abspath(__file__))
ml_root = os.path.dirname(current_dir)
if ml_root not in sys.path:
    sys.path.insert(0, ml_root)

from data_pipeline.schema import CANONICAL_FEATURE_COLS
from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from data_pipeline.features import extract_canonical_features
from data_pipeline.preprocess import CanonicalPreprocessor
from calibration.calibrator import ModelCalibrator
import xgboost as xgb

def compute_calibration_slope_intercept(y_true: np.ndarray, y_prob: np.ndarray) -> Tuple[float, float]:
    """
    Fits logistic regression of y_true on logit(y_prob) to find calibration slope and intercept.
    Ideal calibration: slope = 1.0, intercept = 0.0.
    """
    p_clip = np.clip(y_prob, 1e-6, 1.0 - 1e-6)
    logit_p = np.log(p_clip / (1.0 - p_clip)).reshape(-1, 1)
    
    lr = LogisticRegression(solver="lbfgs", max_iter=1000, random_state=42)
    lr.fit(logit_p, y_true)
    slope = float(lr.coef_[0][0])
    intercept = float(lr.intercept_[0])
    return slope, intercept

def compute_resolution_and_entropy(y_prob: np.ndarray) -> Dict[str, Any]:
    """
    Computes score resolution, modal share, and empirical Shannon entropy.
    """
    n = len(y_prob)
    rounded = np.round(y_prob, 4)
    unique_vals, counts = np.unique(rounded, return_counts=True)
    n_unique = len(unique_vals)
    
    max_idx = np.argmax(counts)
    modal_val = float(unique_vals[max_idx])
    modal_count = int(counts[max_idx])
    modal_share = float(modal_count / n)
    resolution = float(n_unique / n)
    
    probs_dist = counts / n
    ent = float(entropy(probs_dist, base=2))
    
    return {
        "n_unique": n_unique,
        "resolution": round(resolution, 4),
        "modal_val": round(modal_val, 4),
        "modal_count": modal_count,
        "modal_share": round(modal_share, 4),
        "entropy": round(ent, 4)
    }

def run_phase2_2_experiment():
    print("==================================================")
    print("PHASE 2.2: CALIBRATION SELECTION & SCORE RESOLUTION")
    print("==================================================")
    
    # 1. Load Data & Preprocess
    raw_df, meta = load_raw_dataset()
    feat_df = extract_canonical_features(raw_df)
    df_train, df_val, df_test, split_info = temporal_train_val_test_split(feat_df)
    
    prep = CanonicalPreprocessor()
    prep.fit(df_train)
    
    X_train = prep.transform(df_train)
    y_train = df_train["isFraud"].values
    
    X_val = prep.transform(df_val)
    y_val = df_val["isFraud"].values
    val_amounts = df_val["amount"].values
    
    X_test = prep.transform(df_test)
    y_test = df_test["isFraud"].values
    test_amounts = df_test["amount"].values
    
    # 2. Fit Base Model on Train (70%)
    n_neg = int(np.sum(y_train == 0))
    n_pos = int(np.sum(y_train == 1))
    scale_pos_weight = float(n_neg / n_pos) if n_pos > 0 else 1.0
    
    model = xgb.XGBClassifier(
        n_estimators=150,
        max_depth=5,
        learning_rate=0.05,
        subsample=0.85,
        colsample_bytree=0.85,
        scale_pos_weight=scale_pos_weight,
        eval_metric="logloss",
        tree_method="hist",
        random_state=42
    )
    model.fit(X_train[CANONICAL_FEATURE_COLS], y_train)
    
    y_prob_raw_val = model.predict_proba(X_val[CANONICAL_FEATURE_COLS])[:, 1]
    y_prob_raw_test = model.predict_proba(X_test[CANONICAL_FEATURE_COLS])[:, 1]
    
    # 3. Fit Calibrator on Validation (15%)
    calibrator = ModelCalibrator(method="isotonic")
    calibrator.fit(y_prob_raw_val, y_val)
    
    methods = ["raw", "platt", "isotonic", "beta"]
    
    val_probs = {
        "raw": y_prob_raw_val,
        "platt": calibrator.predict_proba(y_prob_raw_val, method="platt"),
        "isotonic": calibrator.predict_proba(y_prob_raw_val, method="isotonic"),
        "beta": calibrator.predict_proba(y_prob_raw_val, method="beta")
    }
    
    test_probs = {
        "raw": y_prob_raw_test,
        "platt": calibrator.predict_proba(y_prob_raw_test, method="platt"),
        "isotonic": calibrator.predict_proba(y_prob_raw_test, method="isotonic"),
        "beta": calibrator.predict_proba(y_prob_raw_test, method="beta")
    }
    
    # -------------------------------------------------------------
    # 1. Calibration Selection CSV (Validation Metrics)
    # -------------------------------------------------------------
    selection_rows = []
    
    for m in methods:
        p_v = val_probs[m]
        brier = brier_score_loss(y_val, p_v)
        ece, _ = ModelCalibrator.compute_expected_calibration_error(y_val, p_v)
        ll = log_loss(y_val, np.clip(p_v, 1e-6, 1.0 - 1e-6))
        roc = roc_auc_score(y_val, p_v)
        pr = average_precision_score(y_val, p_v)
        slope, intercept = compute_calibration_slope_intercept(y_val, p_v)
        res_info = compute_resolution_and_entropy(p_v)
        
        selection_rows.append({
            "method": m,
            "brier_score": round(float(brier), 4),
            "ece": round(float(ece), 4),
            "log_loss": round(float(ll), 4),
            "pr_auc": round(float(pr), 4),
            "roc_auc": round(float(roc), 4),
            "calibration_slope": round(slope, 4),
            "calibration_intercept": round(intercept, 4),
            "unique_probabilities": res_info["n_unique"],
            "modal_probability": res_info["modal_val"],
            "modal_share": res_info["modal_share"],
            "resolution": res_info["resolution"],
            "probability_entropy": res_info["entropy"],
            "p01": round(float(np.percentile(p_v, 1)), 4),
            "p10": round(float(np.percentile(p_v, 10)), 4),
            "p25": round(float(np.percentile(p_v, 25)), 4),
            "p50": round(float(np.percentile(p_v, 50)), 4),
            "p75": round(float(np.percentile(p_v, 75)), 4),
            "p90": round(float(np.percentile(p_v, 90)), 4),
            "p99": round(float(np.percentile(p_v, 99)), 4),
            "fraction_below_0_05": round(float(np.mean(p_v < 0.05)), 4),
            "fraction_0_05_to_0_35": round(float(np.mean((p_v >= 0.05) & (p_v < 0.35))), 4),
            "fraction_above_0_35": round(float(np.mean(p_v >= 0.35)), 4)
        })
        
    df_selection = pd.DataFrame(selection_rows)
    selection_csv_path = os.path.join(current_dir, "calibration_selection.csv")
    df_selection.to_csv(selection_csv_path, index=False)
    print(f"Saved calibration selection CSV: {selection_csv_path}")
    
    # -------------------------------------------------------------
    # 2. Calibration Resolution CSV
    # -------------------------------------------------------------
    resolution_rows = []
    for m in methods:
        p_v = val_probs[m]
        res_info = compute_resolution_and_entropy(p_v)
        res_flag = res_info["resolution"] < 0.10
        collapse_flag = res_info["modal_share"] > 0.50
        
        verdict = "HEALTHY"
        if collapse_flag and res_flag:
            verdict = "SEVERE_STEPWISE_COLLAPSE"
        elif collapse_flag:
            verdict = "HIGH_MODAL_CONCENTRATION"
        elif res_flag:
            verdict = "LOW_RESOLUTION"
            
        resolution_rows.append({
            "method": m,
            "total_samples": len(p_v),
            "unique_probabilities": res_info["n_unique"],
            "resolution_ratio": res_info["resolution"],
            "modal_probability": res_info["modal_val"],
            "modal_count": res_info["modal_count"],
            "modal_share_pct": round(res_info["modal_share"] * 100, 2),
            "resolution_flag": res_flag,
            "modal_collapse_flag": collapse_flag,
            "probability_entropy": res_info["entropy"],
            "verdict": verdict
        })
    df_resolution = pd.DataFrame(resolution_rows)
    resolution_csv_path = os.path.join(current_dir, "calibration_resolution.csv")
    df_resolution.to_csv(resolution_csv_path, index=False)
    print(f"Saved calibration resolution CSV: {resolution_csv_path}")
    
    # -------------------------------------------------------------
    # 3. Calibration Operational Comparison CSV (Phase 2 Policy Simulation)
    # Policy: ALLOW (<0.05), MANUAL_REVIEW (0.05-0.35), DECLINE (>=0.35)
    # -------------------------------------------------------------
    op_rows = []
    n_val = len(y_val)
    total_val_fraud = int(np.sum(y_val))
    
    # Load cost model from config/risk-policy.json
    fp_cost = 500.0
    review_cost = 100.0
    fraud_mult = 1.0
    residual_review_rate = 0.05
    
    for m in methods:
        p_v = val_probs[m]
        
        is_allow = (p_v < 0.05)
        is_review = (p_v >= 0.05) & (p_v < 0.35)
        is_decline = (p_v >= 0.35)
        
        pct_allow = float(np.mean(is_allow) * 100)
        pct_review = float(np.mean(is_review) * 100)
        pct_decline = float(np.mean(is_decline) * 100)
        
        # Fraud outcomes
        fraud_in_allow = int(np.sum((is_allow) & (y_val == 1))) # Missed (FN)
        fraud_in_review = int(np.sum((is_review) & (y_val == 1)))
        fraud_in_decline = int(np.sum((is_decline) & (y_val == 1)))
        fraud_caught = fraud_in_review + fraud_in_decline
        
        legit_in_decline = int(np.sum((is_decline) & (y_val == 0))) # Hard False Positives
        
        recall = float(fraud_caught / total_val_fraud) if total_val_fraud > 0 else 0.0
        total_flagged = int(np.sum(is_review | is_decline))
        precision = float(fraud_caught / total_flagged) if total_flagged > 0 else 0.0
        
        # Financial Cost Calculation
        fn_loss = float(np.sum(val_amounts[is_allow & (y_val == 1)] * fraud_mult))
        fp_loss = float(legit_in_decline * fp_cost)
        rev_ops_cost = float(np.sum(is_review) * review_cost)
        residual_rev_loss = float(np.sum(val_amounts[is_review & (y_val == 1)] * residual_review_rate * fraud_mult))
        
        total_cost = fn_loss + fp_loss + rev_ops_cost + residual_rev_loss
        cost_per_txn = total_cost / n_val
        
        op_rows.append({
            "method": m,
            "pct_allow": round(pct_allow, 2),
            "pct_manual_review": round(pct_review, 2),
            "pct_decline": round(pct_decline, 2),
            "fraud_recall": round(recall, 4),
            "fraud_precision": round(precision, 4),
            "fraud_caught": fraud_caught,
            "fraud_missed": fraud_in_allow,
            "false_positives": legit_in_decline,
            "false_negatives": fraud_in_allow,
            "total_realized_cost": round(total_cost, 2),
            "expected_cost_per_txn": round(cost_per_txn, 2)
        })
        
    df_op = pd.DataFrame(op_rows)
    op_csv_path = os.path.join(current_dir, "calibration_operational_comparison.csv")
    df_op.to_csv(op_csv_path, index=False)
    print(f"Saved calibration operational comparison CSV: {op_csv_path}")
    
    # -------------------------------------------------------------
    # 4. Untouched Test Set Evaluation
    # -------------------------------------------------------------
    test_eval_rows = []
    n_test = len(y_test)
    total_test_fraud = int(np.sum(y_test))
    
    for m in methods:
        p_t = test_probs[m]
        brier = brier_score_loss(y_test, p_t)
        ece, _ = ModelCalibrator.compute_expected_calibration_error(y_test, p_t)
        ll = log_loss(y_test, np.clip(p_t, 1e-6, 1.0 - 1e-6))
        roc = roc_auc_score(y_test, p_t)
        pr = average_precision_score(y_test, p_t)
        res_info = compute_resolution_and_entropy(p_t)
        
        # Policy simulation on test
        is_allow = (p_t < 0.05)
        is_review = (p_t >= 0.05) & (p_t < 0.35)
        is_decline = (p_t >= 0.35)
        
        fraud_in_allow = int(np.sum((is_allow) & (y_test == 1)))
        fraud_in_review = int(np.sum((is_review) & (y_test == 1)))
        fraud_in_decline = int(np.sum((is_decline) & (y_test == 1)))
        fraud_caught = fraud_in_review + fraud_in_decline
        legit_in_decline = int(np.sum((is_decline) & (y_test == 0)))
        
        fn_loss = float(np.sum(test_amounts[is_allow & (y_test == 1)] * fraud_mult))
        fp_loss = float(legit_in_decline * fp_cost)
        rev_ops_cost = float(np.sum(is_review) * review_cost)
        residual_rev_loss = float(np.sum(test_amounts[is_review & (y_test == 1)] * residual_review_rate * fraud_mult))
        tot_test_cost = fn_loss + fp_loss + rev_ops_cost + residual_rev_loss
        
        test_eval_rows.append({
            "method": m,
            "brier_score": round(float(brier), 4),
            "ece": round(float(ece), 4),
            "log_loss": round(float(ll), 4),
            "pr_auc": round(float(pr), 4),
            "roc_auc": round(float(roc), 4),
            "resolution": res_info["resolution"],
            "modal_share": res_info["modal_share"],
            "fraud_recall": round(float(fraud_caught / total_test_fraud), 4),
            "fraud_caught": fraud_caught,
            "total_realized_cost": round(tot_test_cost, 2),
            "expected_cost_per_txn": round(tot_test_cost / n_test, 2)
        })
        
    df_test_eval = pd.DataFrame(test_eval_rows)
    print("\n--- Test Set Evaluation Summary ---")
    print(df_test_eval.to_string(index=False))
    
    return {
        "selection": df_selection.to_dict(orient="records"),
        "resolution": df_resolution.to_dict(orient="records"),
        "operational": df_op.to_dict(orient="records"),
        "test_eval": df_test_eval.to_dict(orient="records")
    }

if __name__ == "__main__":
    run_phase2_2_experiment()
