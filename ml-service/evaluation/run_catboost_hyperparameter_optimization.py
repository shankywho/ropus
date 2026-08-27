"""
AI Risk Manager — CatBoost Hyperparameter Optimization & Validation Tuning Engine (P3)
Performs rigorous grid search, class-weight calibration, and regularization strictly
using Train and Validation partitions, evaluating final generalization on the frozen Test holdout.
"""

import os
import sys
import json
import numpy as np
import pandas as pd
from catboost import CatBoostClassifier
from xgboost import XGBClassifier
from sklearn.metrics import (
    roc_auc_score,
    average_precision_score,
    precision_score,
    recall_score,
    f1_score,
    brier_score_loss
)

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from data_pipeline.features import extract_canonical_25_features
from data_pipeline.preprocess import CanonicalPreprocessor
from evaluation.run_catboost_calibration_and_bmr import extract_extended_raw_signals
from evaluation.shadow_evaluator import compute_ece

def run_optimization():
    print("=" * 80)
    print("CATBOOST HYPERPARAMETER OPTIMIZATION & VALIDATION TUNING (P3)")
    print("=" * 80)

    data_dir = os.path.join(ML_SERVICE_DIR, "data")
    df_raw, data_meta = load_raw_dataset(data_dir=data_dir)
    df_train_raw, df_val_raw, df_test_raw, split_info = temporal_train_val_test_split(
        df_raw, time_col="TransactionDT", train_ratio=0.70, val_ratio=0.15, test_ratio=0.15
    )

    y_train = df_train_raw["isFraud"].values
    y_val = df_val_raw["isFraud"].values
    y_test = df_test_raw["isFraud"].values

    # Base 25-Feature Extraction
    df_f_tr = extract_canonical_25_features(df_train_raw)
    df_f_va = extract_canonical_25_features(df_val_raw)
    df_f_te = extract_canonical_25_features(df_test_raw)
    prep = CanonicalPreprocessor(feature_contract="v2.5")
    prep.fit(df_f_tr)
    X_tr_25 = prep.transform(df_f_tr, feature_contract="v2.5")
    X_va_25 = prep.transform(df_f_va, feature_contract="v2.5")
    X_te_25 = prep.transform(df_f_te, feature_contract="v2.5")

    # Extended 33-Signal Extraction
    df_ext_tr, tr_med, tr_freq = extract_extended_raw_signals(df_train_raw)
    df_ext_va, _, _ = extract_extended_raw_signals(df_val_raw, train_medians=tr_med, train_freq_card=tr_freq)
    df_ext_te, _, _ = extract_extended_raw_signals(df_test_raw, train_medians=tr_med, train_freq_card=tr_freq)

    X_train_58 = np.hstack([X_tr_25, df_ext_tr.values])
    X_val_58 = np.hstack([X_va_25, df_ext_va.values])
    X_test_58 = np.hstack([X_te_25, df_ext_te.values])

    print(f"Data Partition: Train={len(y_train)} (Fraud={sum(y_train)}), Val={len(y_val)} (Fraud={sum(y_val)}), Test={len(y_test)} (Fraud={sum(y_test)})")
    print(f"Feature Dimension: {X_train_58.shape[1]} features")

    # 1. Hyperparameter Grid on Validation Split
    param_grid = [
        {"depth": 3, "l2_leaf_reg": 3.0, "scale_pos_weight": 2.0, "learning_rate": 0.05, "iterations": 150},
        {"depth": 4, "l2_leaf_reg": 3.0, "scale_pos_weight": 2.0, "learning_rate": 0.05, "iterations": 150}, # Baseline
        {"depth": 4, "l2_leaf_reg": 8.0, "scale_pos_weight": 2.0, "learning_rate": 0.05, "iterations": 150}, # Regularized
        {"depth": 4, "l2_leaf_reg": 3.0, "scale_pos_weight": 3.5, "learning_rate": 0.05, "iterations": 150}, # High Class Weight
        {"depth": 5, "l2_leaf_reg": 5.0, "scale_pos_weight": 2.0, "learning_rate": 0.03, "iterations": 180},
    ]

    val_results = []
    trained_models = []

    print("\nEvaluating Hyperparameter Grid on Validation Partition:")
    for idx, params in enumerate(param_grid):
        cb = CatBoostClassifier(
            iterations=params["iterations"],
            depth=params["depth"],
            learning_rate=params["learning_rate"],
            l2_leaf_reg=params["l2_leaf_reg"],
            scale_pos_weight=params["scale_pos_weight"],
            random_seed=42,
            verbose=0
        )
        cb.fit(X_train_58, y_train, eval_set=(X_val_58, y_val), early_stopping_rounds=25, verbose=False)
        p_val = cb.predict_proba(X_val_58)[:, 1]

        pr_val = float(average_precision_score(y_val, p_val))
        roc_val = float(roc_auc_score(y_val, p_val))

        val_results.append({
            "config_id": idx,
            "params": params,
            "val_pr_auc": round(pr_val, 4),
            "val_roc_auc": round(roc_val, 4)
        })
        trained_models.append(cb)
        print(f"  Config {idx}: depth={params['depth']}, reg={params['l2_leaf_reg']}, spw={params['scale_pos_weight']} -> Val PR-AUC={pr_val:.4f}, ROC-AUC={roc_val:.4f}")

    # Select Best Model based strictly on Validation PR-AUC
    best_config_idx = int(np.argmax([r["val_pr_auc"] for r in val_results]))
    best_model = trained_models[best_config_idx]
    best_params = param_grid[best_config_idx]
    print(f"\nOptimal Config Selected on Validation: Config {best_config_idx} (Val PR-AUC={val_results[best_config_idx]['val_pr_auc']})")

    # 2. Optimal Threshold Selection on Validation Split
    p_val_best = best_model.predict_proba(X_val_58)[:, 1]
    best_th = 0.170
    best_f1_val = 0.0
    for th in np.linspace(0.02, 0.40, 39):
        y_pred = (p_val_best >= th).astype(int)
        f1 = f1_score(y_val, y_pred, zero_division=0)
        if f1 > best_f1_val:
            best_f1_val = f1
            best_th = float(th)
    print(f"Validation-Optimal Threshold: tau* = {best_th:.3f} (Val F1 = {best_f1_val:.4f})")

    # 3. Final Generalization Assessment on Frozen Holdout Test Split
    p_test = best_model.predict_proba(X_test_58)[:, 1]
    roc_test = float(roc_auc_score(y_test, p_test))
    pr_test = float(average_precision_score(y_test, p_test))

    y_pred_test = (p_test >= best_th).astype(int)
    prec_test = float(precision_score(y_test, y_pred_test, zero_division=0))
    rec_test = float(recall_score(y_test, y_pred_test, zero_division=0))
    f1_test = float(f1_score(y_test, y_pred_test, zero_division=0))
    brier_test = float(brier_score_loss(y_test, p_test))
    ece_test = compute_ece(y_test, p_test)

    # 4. Bootstrap Confidence Intervals on Test Split (1000 resamples)
    np.random.seed(42)
    boot_pr, boot_roc = [], []
    for _ in range(1000):
        idx = np.random.choice(len(y_test), size=len(y_test), replace=True)
        if len(np.unique(y_test[idx])) > 1:
            boot_pr.append(average_precision_score(y_test[idx], p_test[idx]))
            boot_roc.append(roc_auc_score(y_test[idx], p_test[idx]))

    ci_pr = [float(np.percentile(boot_pr, 2.5)), float(np.percentile(boot_pr, 97.5))]
    ci_roc = [float(np.percentile(boot_roc, 2.5)), float(np.percentile(boot_roc, 97.5))]

    print("\nTest Holdout Generalization Metrics (Selected Model):")
    print(f"  ROC-AUC: {roc_test:.4f} (95% CI: [{ci_roc[0]:.4f}, {ci_roc[1]:.4f}])")
    print(f"  PR-AUC:  {pr_test:.4f} (95% CI: [{ci_pr[0]:.4f}, {ci_pr[1]:.4f}])")
    print(f"  Precision: {prec_test:.4f} | Recall: {rec_test:.4f} | F1: {f1_test:.4f} (at tau*={best_th:.3f})")
    print(f"  Brier: {brier_test:.4f} | ECE: {ece_test:.4f}")

    report = {
        "validation_grid_results": val_results,
        "selected_configuration": {
            "config_id": best_config_idx,
            "params": best_params,
            "validation_pr_auc": val_results[best_config_idx]["val_pr_auc"],
            "validation_optimal_threshold": round(best_th, 3),
            "validation_f1": round(best_f1_val, 4)
        },
        "frozen_test_holdout_evaluation": {
            "roc_auc": round(roc_test, 4),
            "roc_auc_95_ci": [round(ci_roc[0], 4), round(ci_roc[1], 4)],
            "pr_auc": round(pr_test, 4),
            "pr_auc_95_ci": [round(ci_pr[0], 4), round(ci_pr[1], 4)],
            "precision_at_tau_star": round(prec_test, 4),
            "recall_at_tau_star": round(rec_test, 4),
            "f1_at_tau_star": round(f1_test, 4),
            "brier_score": round(brier_test, 4),
            "ece": round(ece_test, 4)
        }
    }

    out_path = os.path.join(CURRENT_DIR, "catboost_optimization_report.json")
    with open(out_path, "w") as f:
        json.dump(report, f, indent=2)
    print(f"\n[SUCCESS] Optimization report saved to {out_path}")
    return report

if __name__ == "__main__":
    run_optimization()
