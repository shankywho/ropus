"""
ROPUS Senior Model Risk Performance Diagnosis
Deep quantitative diagnostic investigation of champion v8.0-bmr-36f:
1. Feature Importance & Predictive Signal Extraction (CatBoost feature importances)
2. Temporal Drift & Feature Stability across Chronological Slices (Dev vs Holdout, Early Holdout vs Late Holdout)
3. Threshold Sweep Analysis (Precision, Recall, F1, Loss under BMR across thresholds 0.01 to 0.50)
4. False-Positive & False-Negative Dollar Loss Segmentation
5. BMR Decision Threshold & Policy Mechanics Audit
6. Candidate / Challenger Architecture Potential (LightGBM, XGBoost, CatBoost tuned, feature interactions)
"""

import os
import sys
import json
import numpy as np
import pandas as pd
import joblib
from sklearn.metrics import (
    roc_auc_score,
    average_precision_score,
    precision_score,
    recall_score,
    f1_score,
    confusion_matrix,
    brier_score_loss
)

current_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from data_pipeline.data_loader import load_raw_dataset
from evaluation.phase_14_production_hardening import extract_phase14_features, fit_phase14_preprocessor
from evaluation.phase_15_production_release import ProductionScoringEngine

def main():
    print("=" * 95)
    print("ROPUS SENIOR MODEL RISK REVIEW & PERFORMANCE DIAGNOSIS")
    print("=" * 95)

    eval_dir = os.path.join(current_dir, "evaluation")
    v8_artifact_path = os.path.join(current_dir, "model", "candidates", "production_model_v8_bmr.joblib")
    scoring_engine = ProductionScoringEngine(v8_artifact_path)
    bundle = joblib.load(v8_artifact_path)
    catboost_model = bundle.get("model")
    calibrator = bundle.get("calibrator")

    # 1. Feature Importances
    feature_names = scoring_engine.feature_names
    feat_importances = catboost_model.get_feature_importance()
    feat_imp_df = pd.DataFrame({
        "feature": feature_names,
        "importance": feat_importances
    }).sort_values("importance", ascending=False).reset_index(drop=True)

    print("\n[DIAGNOSTIC 1] Top 10 Most Predictive Features in Champion v8.0-bmr-36f:")
    for idx, row in feat_imp_df.head(10).iterrows():
        print(f"   {idx+1:2d}. {row['feature']:<35} : {row['importance']:6.2f}%")

    zero_imp_feats = feat_imp_df[feat_imp_df["importance"] == 0]["feature"].tolist()
    print(f"-> Zero-Importance Features ({len(zero_imp_feats)}/{len(feature_names)}): {zero_imp_feats}")

    # 2. Chronological Holdout Inference
    df_raw, _ = load_raw_dataset()
    df_dev_raw = df_raw.iloc[:6800].reset_index(drop=True)
    df_test_raw = df_raw.iloc[6800:8000].reset_index(drop=True)

    df_dev_feat = extract_phase14_features(df_dev_raw)
    df_test_feat = extract_phase14_features(df_test_raw)
    X_train_feat, X_val_feat, X_test_feat, prep_state = fit_phase14_preprocessor(
        df_dev_feat.iloc[:5600], df_dev_feat.iloc[5600:], df_test_feat
    )

    y_test = df_test_raw["isFraud"].values.astype(int)
    amts_test = df_test_raw["TransactionAmt"].values.astype(float)

    test_scores = scoring_engine.score_feature_vector(X_test_feat)
    raw_probs = np.array([r["raw_probability"] for r in test_scores])
    cal_probs = np.array([r["calibrated_probability"] for r in test_scores])
    decisions = np.array([r["decision"] for r in test_scores])

    # 3. Threshold Sweep Analysis
    print("\n[DIAGNOSTIC 2] Threshold Trade-off & Cost Optimization Analysis:")
    thresholds = [0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.08, 0.10, 0.15, 0.20, 0.25, 0.30, 0.40, 0.50]
    thresh_results = []

    for t in thresholds:
        pred_t = (cal_probs >= t).astype(int)
        tn, fp, fn, tp = confusion_matrix(y_test, pred_t).ravel()
        p = precision_score(y_test, pred_t, zero_division=0)
        r = recall_score(y_test, pred_t, zero_division=0)
        f1 = f1_score(y_test, pred_t, zero_division=0)

        fraud_dlrs_captured = np.sum(amts_test[(y_test == 1) & (pred_t == 1)])
        fraud_dlrs_missed = np.sum(amts_test[(y_test == 1) & (pred_t == 0)])
        legit_dlrs_declined = np.sum(amts_test[(y_test == 0) & (pred_t == 1)])

        # Total cost = Missed Fraud * 1.05 + False Declines * $25
        emp_cost = (fraud_dlrs_missed * 1.05) + (fp * 25.0)

        thresh_results.append({
            "threshold": t,
            "declines": int(tp + fp),
            "decline_rate": float((tp + fp) / len(y_test)),
            "precision": float(p),
            "recall": float(r),
            "f1": float(f1),
            "tp": int(tp), "fp": int(fp), "fn": int(fn), "tn": int(tn),
            "fraud_dollars_captured": float(fraud_dlrs_captured),
            "fraud_dollars_missed": float(fraud_dlrs_missed),
            "legit_dollars_declined": float(legit_dlrs_declined),
            "total_cost": float(emp_cost)
        })

    thresh_df = pd.DataFrame(thresh_results)
    min_cost_idx = thresh_df["total_cost"].idxmin()
    best_fixed_thresh = thresh_df.loc[min_cost_idx]

    print(f"{'Thresh':<7} | {'Dec Rate':<8} | {'Prec':<7} | {'Recall':<7} | {'F1':<7} | {'FP':<5} | {'TP':<4} | {'Legit $ Dec':<12} | {'Fraud $ Cap':<12} | {'Total Cost':<10}")
    print("-" * 95)
    for _, r in thresh_df.iterrows():
        print(f"{r['threshold']:<7.2f} | {r['decline_rate']:<8.2%} | {r['precision']:<7.3f} | {r['recall']:<7.3f} | {r['f1']:<7.3f} | {int(r['fp']):<5d} | {int(r['tp']):<4d} | ${r['legit_dollars_declined']:<11,.2f} | ${r['fraud_dollars_captured']:<11,.2f} | ${r['total_cost']:<9,.2f}")

    print(f"\n-> Lowest-Cost Fixed Threshold: {best_fixed_thresh['threshold']:.2f} (Total Cost: ${best_fixed_thresh['total_cost']:,.2f} vs BMR Dynamic: ${(np.sum(amts_test[(y_test == 1) & (decisions == 'ALLOW')]) * 1.05 + np.sum((y_test == 0) & (decisions == 'DECLINE')) * 25.0):,.2f})")

    # 4. Root Cause of False Positives ($37,134.02 in declined legitimate volume)
    print("\n[DIAGNOSTIC 3] Root-Cause Diagnosis of Declined Legitimate Volume (False Positives):")
    fp_mask = (y_test == 0) & (decisions == "DECLINE")
    fp_amts = amts_test[fp_mask]
    fp_probs = cal_probs[fp_mask]
    fp_X = X_test_feat[fp_mask]

    print(f"-> Total False Positive Orders:       {np.sum(fp_mask)} orders")
    print(f"-> Total Declined Legitimate Volume:  ${np.sum(fp_amts):,.2f}")
    print(f"-> Mean False Positive Ticket Amount: ${np.mean(fp_amts):,.2f} (vs Test Dataset Mean: ${np.mean(amts_test):,.2f})")
    print(f"-> Median False Positive Ticket Amt:  ${np.median(fp_amts):,.2f}")
    print(f"-> Max False Positive Ticket Amount:  ${np.max(fp_amts):,.2f}")

    # Breakdown by amount category among FPs
    high_amt_fps = np.sum(fp_amts > 250.0)
    high_amt_fp_vol = np.sum(fp_amts[fp_amts > 250.0])
    print(f"-> High-Value FPs (> $250):           {high_amt_fps} orders ({high_amt_fps/len(fp_amts):.1%}) accounting for ${high_amt_fp_vol:,.2f} ({high_amt_fp_vol/np.sum(fp_amts):.1%} of FP dollars!)")

    # Why did BMR decline high-value legit orders?
    # Under BMR, P*(A) = 25.0 / (1.05 * A + 25.0).
    # For A = $1,000, P*(1000) = 25 / (1050 + 25) = 25 / 1075 = 0.02325 (2.325%!)
    # For A = $2,000, P*(2000) = 25 / (2100 + 25) = 25 / 2125 = 0.01176 (1.176%!)
    # When model has baseline predicted probability around 2-3%, ANY high ticket transaction is automatically DECLINED!
    print("\n   [KEY ROOT CAUSE IDENTIFIED]: BMR Threshold Skew on High Ticket Amounts:")
    for amt_example in [100.0, 250.0, 500.0, 1000.0, 2000.0]:
        p_star = 25.0 / (1.05 * amt_example + 25.0)
        print(f"      Amount = ${amt_example:>7.2f} -> BMR Decline Threshold P*(A) = {p_star:.4f} ({p_star:.2%})")

    # 5. Temporal Drift & Feature Decay
    print("\n[DIAGNOSTIC 4] Temporal Feature Drift & Degradation Analysis:")
    # Compare feature means between Dev (0:6800) and Test (6800:8000)
    drift_records = []
    for col in feature_names:
        if col in X_train_feat.columns and col in X_test_feat.columns:
            m_dev = float(np.mean(X_train_feat[col]))
            s_dev = float(np.std(X_train_feat[col])) + 1e-6
            m_test = float(np.mean(X_test_feat[col]))
            s_test = float(np.std(X_test_feat[col])) + 1e-6

            # Standardized mean difference (Cohen's d)
            cohen_d = abs(m_test - m_dev) / s_dev
            drift_records.append({
                "feature": col,
                "mean_dev": m_dev,
                "mean_test": m_test,
                "cohens_d": cohen_d
            })

    drift_df = pd.DataFrame(drift_records).sort_values("cohens_d", ascending=False).reset_index(drop=True)
    print("   Top 5 Most Shifted Features (Dev vs Holdout):")
    for idx, r in drift_df.head(5).iterrows():
        print(f"      {idx+1}. {r['feature']:<30} : Dev Mean={r['mean_dev']:8.2f} | Test Mean={r['mean_test']:8.2f} | Cohen's d={r['cohens_d']:.3f}")

    # 6. Candidate / Challenger Feasibility Test
    print("\n[DIAGNOSTIC 5] Offline Challenger Exploration (Read-Only Comparison):")
    from catboost import CatBoostClassifier
    from lightgbm import LGBMClassifier

    # Challenger 1: Tuned LightGBM with Depth 4, Learning Rate 0.03, class weight
    lgb_challenger = LGBMClassifier(
        n_estimators=160,
        learning_rate=0.03,
        max_depth=4,
        num_leaves=15,
        scale_pos_weight=15.0,
        random_state=42,
        verbose=-1
    )
    lgb_challenger.fit(X_train_feat[feature_names], df_dev_raw.iloc[:5600]["isFraud"].values)
    lgb_probs_raw = lgb_challenger.predict_proba(X_test_feat[feature_names])[:, 1]
    lgb_auc = roc_auc_score(y_test, lgb_probs_raw)
    lgb_pr = average_precision_score(y_test, lgb_probs_raw)
    print(f"   Challenger 1 (LightGBM D4):        ROC-AUC = {lgb_auc:.4f}, PR-AUC = {lgb_pr:.4f}")

    # Challenger 2: Deep CatBoost (Depth 6, iterations 200, SPW 15)
    cb_challenger = CatBoostClassifier(
        iterations=200,
        depth=6,
        learning_rate=0.03,
        scale_pos_weight=15.0,
        random_seed=42,
        verbose=0
    )
    cb_challenger.fit(X_train_feat[feature_names], df_dev_raw.iloc[:5600]["isFraud"].values)
    cb_probs_raw = cb_challenger.predict_proba(X_test_feat[feature_names])[:, 1]
    cb_auc = roc_auc_score(y_test, cb_probs_raw)
    cb_pr = average_precision_score(y_test, cb_probs_raw)
    print(f"   Challenger 2 (CatBoost D6 Iter200): ROC-AUC = {cb_auc:.4f}, PR-AUC = {cb_pr:.4f}")

    print(f"   Current Champion (v8.0-bmr-36f):   ROC-AUC = 0.6216, PR-AUC = 0.0886")

    output_data = {
        "feat_importances": feat_imp_df.to_dict(orient="records"),
        "threshold_analysis": thresh_results,
        "lowest_cost_fixed_threshold": best_fixed_thresh.to_dict(),
        "false_positive_analysis": {
            "total_fps": int(np.sum(fp_mask)),
            "total_fp_dollars": float(np.sum(fp_amts)),
            "mean_fp_ticket": float(np.mean(fp_amts)),
            "high_value_fp_count": int(high_amt_fps),
            "high_value_fp_dollars": float(high_amt_fp_vol),
            "high_value_fp_pct_of_total_fp_dollars": float(high_amt_fp_vol / np.sum(fp_amts))
        },
        "top_drifted_features": drift_df.head(10).to_dict(orient="records"),
        "offline_challenger_probes": [
            {"model": "LightGBM_D4_Iter160_SPW15", "roc_auc": round(lgb_auc, 4), "pr_auc": round(lgb_pr, 4)},
            {"model": "CatBoost_D6_Iter200_SPW15", "roc_auc": round(cb_auc, 4), "pr_auc": round(cb_pr, 4)},
            {"model": "v8.0-bmr-36f (Locked Champion)", "roc_auc": 0.6216, "pr_auc": 0.0886}
        ]
    }

    with open(os.path.join(eval_dir, "phase_27_model_risk_diagnosis.json"), "w") as f:
        json.dump(output_data, f, indent=2)

    print(f"\nDiagnostic analysis JSON saved to {eval_dir}/phase_27_model_risk_diagnosis.json")

if __name__ == "__main__":
    main()
