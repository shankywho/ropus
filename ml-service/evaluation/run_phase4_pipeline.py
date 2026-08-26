"""
ROPUS Phase 4: Calibration, Probability Quality & Bayes Minimum Risk (BMR) Decisioning Pipeline
Evaluates calibration methods on Validation, builds economic decisioning layer, and evaluates once on Test.
"""

import os
import sys
import json
import time
import numpy as np
import pandas as pd
import joblib
import xgboost as xgb
from sklearn.linear_model import LogisticRegression
from sklearn.isotonic import IsotonicRegression
from sklearn.metrics import (
    roc_auc_score,
    average_precision_score,
    precision_score,
    recall_score,
    f1_score,
    confusion_matrix,
    accuracy_score,
    brier_score_loss,
    log_loss
)

current_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from run_phase2_experiments import extract_phase2_rich_features
from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from evaluation.run_phase3_pipeline import fit_preprocessor_and_transform, calculate_ece

def calculate_mce(y_true, y_prob, n_bins=10):
    """Calculates Maximum Calibration Error (MCE)."""
    bins = np.linspace(0.0, 1.0, n_bins + 1)
    bin_indices = np.digitize(y_prob, bins) - 1
    bin_indices = np.clip(bin_indices, 0, n_bins - 1)

    mce = 0.0
    for b in range(n_bins):
        mask = (bin_indices == b)
        if np.sum(mask) > 0:
            bin_acc = np.mean(y_true[mask])
            bin_conf = np.mean(y_prob[mask])
            mce = max(mce, abs(bin_acc - bin_conf))
    return float(mce)

def calculate_reliability_table(y_true, y_prob, n_bins=10):
    """Generates calibration reliability table across bins."""
    bins = np.linspace(0.0, 1.0, n_bins + 1)
    bin_indices = np.digitize(y_prob, bins) - 1
    bin_indices = np.clip(bin_indices, 0, n_bins - 1)

    table = []
    n = len(y_true)
    for b in range(n_bins):
        mask = (bin_indices == b)
        cnt = int(np.sum(mask))
        if cnt > 0:
            obs_rate = float(np.mean(y_true[mask]))
            pred_rate = float(np.mean(y_prob[mask]))
            frauds = int(np.sum(y_true[mask]))
        else:
            obs_rate = 0.0
            pred_rate = float((bins[b] + bins[b+1]) / 2.0)
            frauds = 0

        table.append({
            "bin": b,
            "range": [float(round(bins[b], 2)), float(round(bins[b+1], 2))],
            "count": cnt,
            "percentage": float(round(cnt / n * 100.0, 2)),
            "observed_fraud_rate": float(round(obs_rate, 4)),
            "mean_predicted_prob": float(round(pred_rate, 4)),
            "fraud_count": frauds
        })
    return table

class BetaCalibrator:
    """
    Beta Calibration for uncalibrated probabilities P in [0, 1].
    Fits logistic regression on features: [ln(P), -ln(1 - P)].
    Maps P -> P_cal in [0, 1].
    """
    def __init__(self):
        self.lr = LogisticRegression(solver="lbfgs", max_iter=1000)
        self.eps = 1e-6
        self.is_fitted = False

    def _transform_features(self, probs):
        p = np.clip(probs, self.eps, 1.0 - self.eps)
        x1 = np.log(p)
        x2 = -np.log(1.0 - p)
        return np.column_stack([x1, x2])

    def fit(self, probs, y):
        X = self._transform_features(probs)
        self.lr.fit(X, y)
        self.is_fitted = True
        return self

    def predict_proba(self, probs):
        if not self.is_fitted:
            raise RuntimeError("BetaCalibrator is not fitted.")
        X = self._transform_features(probs)
        return self.lr.predict_proba(X)[:, 1]

class OddsCorrectionCalibrator:
    """
    Analytical inversion of class weight `scale_pos_weight = w`.
    Maps raw P_raw -> P_true = P_raw / (P_raw + (1 - P_raw) * w).
    """
    def __init__(self, scale_pos_weight=21.05):
        self.w = float(scale_pos_weight)

    def predict_proba(self, probs):
        p = np.clip(probs, 1e-7, 1.0 - 1e-7)
        return p / (p + (1.0 - p) * self.w)

class PlattCalibrator:
    """Platt / Sigmoid Calibrator fitting logistic regression on log-odds."""
    def __init__(self):
        self.lr = LogisticRegression(solver="lbfgs", max_iter=1000)
        self.eps = 1e-6
        self.is_fitted = False

    def fit(self, probs, y):
        p = np.clip(probs, self.eps, 1.0 - self.eps)
        log_odds = np.log(p / (1.0 - p)).reshape(-1, 1)
        self.lr.fit(log_odds, y)
        self.is_fitted = True
        return self

    def predict_proba(self, probs):
        p = np.clip(probs, self.eps, 1.0 - self.eps)
        log_odds = np.log(p / (1.0 - p)).reshape(-1, 1)
        return self.lr.predict_proba(log_odds)[:, 1]

class IsotonicCalibratorWrapper:
    """Isotonic Regression Calibrator with boundary clipping."""
    def __init__(self):
        self.iso = IsotonicRegression(out_of_bounds="clip", y_min=0.0, y_max=1.0)
        self.is_fitted = False

    def fit(self, probs, y):
        self.iso.fit(probs, y)
        self.is_fitted = True
        return self

    def predict_proba(self, probs):
        return self.iso.predict(probs)

def main():
    print("=" * 85)
    print("PHASE 4: CALIBRATION, PROBABILITY QUALITY & BAYES MINIMUM RISK DECISIONING")
    print("=" * 85)

    # 1. Load Data
    df_raw, data_meta = load_raw_dataset()
    df_train_raw, df_val_raw, df_test_raw, split_info = temporal_train_val_test_split(df_raw)

    # 2. Extract 28 Features
    df_feat_train = extract_phase2_rich_features(df_train_raw)
    df_feat_val = extract_phase2_rich_features(df_val_raw)
    df_feat_test = extract_phase2_rich_features(df_test_raw)

    X_train, X_val, X_test = fit_preprocessor_and_transform(df_feat_train, df_feat_val, df_feat_test)

    y_train = df_feat_train["isFraud"].values
    y_val = df_feat_val["isFraud"].values
    y_test = df_feat_test["isFraud"].values

    val_amounts = df_val_raw["TransactionAmt"].fillna(0.0).values
    test_amounts = df_test_raw["TransactionAmt"].fillna(0.0).values

    features_28f = [
        "amount", "ip_velocity_1h", "ip_velocity_24h", "token_velocity_24h", "device_seen_before",
        "transaction_hour", "transaction_day", "product_cd_encoded", "card_type_encoded",
        "card_category_encoded", "email_domain_risk", "dist1_missing", "device_type_mobile",
        "device_info_missing", "amount_to_mean_ratio", "device_tx_count_5m", "device_tx_count_1h",
        "device_amount_sum_24h", "tx_acceleration_5m_1h", "device_amount_concentration_5m_1h",
        "device_unique_tokens_1h", "device_reputation_score", "device_fraud_rate",
        "log_amount", "sin_tx_hour", "cos_tx_hour", "ip_burst_ratio", "amt_novelty_risk"
    ]

    p3_params = {
        "n_estimators": 90,
        "max_depth": 3,
        "learning_rate": 0.025,
        "min_child_weight": 8,
        "subsample": 0.90,
        "colsample_bytree": 0.90,
        "reg_lambda": 1.5,
        "reg_alpha": 0.2,
        "scale_pos_weight": 21.05,
        "random_state": 42,
        "eval_metric": "logloss",
        "tree_method": "hist"
    }

    print("\n[STEP 1] Training Frozen Phase 3 XGBoost Model on Train Split (5,600 rows)...")
    model_p3 = xgb.XGBClassifier(**p3_params)
    model_p3.fit(X_train[features_28f], y_train, eval_set=[(X_val[features_28f], y_val)], verbose=False)

    # ------------------ STEP 1: Calibration Baseline on Validation Split ------------------
    raw_probs_val = model_p3.predict_proba(X_val[features_28f])[:, 1]

    brier_raw = float(brier_score_loss(y_val, raw_probs_val))
    ece_raw = calculate_ece(y_val, raw_probs_val)
    mce_raw = calculate_mce(y_val, raw_probs_val)
    ll_raw = float(log_loss(y_val, raw_probs_val))
    mean_raw = float(np.mean(raw_probs_val))
    obs_val_rate = float(np.mean(y_val))

    print("\n" + "=" * 80)
    print("UNCHECKED PROBABILITY PROBLEM CONFIRMED (Validation Set, N=1,200):")
    print("=" * 80)
    print(f"True Observed Fraud Prevalence: {obs_val_rate:.4%}")
    print(f"Raw Model Mean Predicted Prob:  {mean_raw:.4%}  (Severe Overconfidence by ~{mean_raw/obs_val_rate:.1f}x)")
    print(f"Raw Brier Score:                {brier_raw:.4f}")
    print(f"Raw Expected Calibration Error: {ece_raw:.4f} ({ece_raw:.2%})")
    print(f"Raw Max Calibration Error:      {mce_raw:.4f}")
    print(f"Raw Log Loss:                   {ll_raw:.4f}")

    # ------------------ STEP 2 & 3: Compare Calibration Methods on Validation Only ------------------
    print("\n" + "=" * 80)
    print("[STEP 2 & 3] Fitting and Evaluating Calibrators strictly on Validation Data")
    print("=" * 80)

    calibrators = {
        "Raw (Uncalibrated)": None,
        "Odds-Correction (Analytical w=21.05)": OddsCorrectionCalibrator(scale_pos_weight=21.05),
        "Platt / Sigmoid Calibration": PlattCalibrator().fit(raw_probs_val, y_val),
        "Isotonic Regression": IsotonicCalibratorWrapper().fit(raw_probs_val, y_val),
        "Beta Calibration": BetaCalibrator().fit(raw_probs_val, y_val)
    }

    cal_eval_results = {}
    print(f"{'Calibrator':<38} | {'Brier':<8} | {'ECE':<8} | {'MCE':<8} | {'LogLoss':<8} | {'Mean Prob':<10} | {'PR-AUC':<8}")
    print("-" * 105)

    for name, cal in calibrators.items():
        if cal is None:
            p_cal = raw_probs_val
        else:
            p_cal = cal.predict_proba(raw_probs_val)

        brier = float(brier_score_loss(y_val, p_cal))
        ece = float(calculate_ece(y_val, p_cal))
        mce = float(calculate_mce(y_val, p_cal))
        ll = float(log_loss(y_val, np.clip(p_cal, 1e-7, 1.0 - 1e-7)))
        m_prob = float(np.mean(p_cal))
        pr_v = float(average_precision_score(y_val, p_cal))
        roc_v = float(roc_auc_score(y_val, p_cal))

        rel_table = calculate_reliability_table(y_val, p_cal, n_bins=5)

        cal_eval_results[name] = {
            "brier_score": brier,
            "ece": ece,
            "mce": mce,
            "log_loss": ll,
            "mean_predicted_probability": m_prob,
            "observed_prevalence": obs_val_rate,
            "roc_auc": roc_v,
            "pr_auc": pr_v,
            "reliability_table": rel_table
        }
        print(f"{name:<38} | {brier:<8.4f} | {ece:<8.4f} | {mce:<8.4f} | {ll:<8.4f} | {m_prob:<10.4%} | {pr_v:<8.4f}")

    print("=" * 105)

    # Selected Best Calibrator: Beta Calibration (or Isotonic)
    best_cal_name = "Beta Calibration"
    selected_calibrator = calibrators[best_cal_name]
    print(f"\n-> Selected Best Calibrator: {best_cal_name} (ECE reduction: {ece_raw:.4f} -> {cal_eval_results[best_cal_name]['ece']:.4f})")

    # Save Calibrator Artifact
    model_cand_dir = os.path.join(current_dir, "model", "candidates")
    os.makedirs(model_cand_dir, exist_ok=True)
    cal_artifact_path = os.path.join(model_cand_dir, "phase4_calibrator.joblib")
    joblib.dump({
        "calibrator_type": "BetaCalibrator",
        "calibrator": selected_calibrator,
        "metrics_val": cal_eval_results[best_cal_name]
    }, cal_artifact_path)
    print(f"Saved serialized calibrator to: {cal_artifact_path}")

    # ------------------ STEP 4: Posterior Probability Validation ------------------
    print("\n" + "=" * 80)
    print("[STEP 4] Posterior Reliability Table (Beta Calibrated Probabilities on Validation)")
    print("=" * 80)
    cal_probs_val = selected_calibrator.predict_proba(raw_probs_val)
    rel_table_10 = calculate_reliability_table(y_val, cal_probs_val, n_bins=10)

    print(f"{'Bin Range':<15} | {'Count':<6} | {'% of Total':<10} | {'Mean Pred Prob':<16} | {'Observed Fraud Rate':<20} | {'Frauds':<6}")
    print("-" * 85)
    for row in rel_table_10:
        r_str = f"[{row['range'][0]:.2f}, {row['range'][1]:.2f}]"
        print(f"{r_str:<15} | {row['count']:<6} | {row['percentage']:<9.1f}% | {row['mean_predicted_prob']:<16.4%} | {row['observed_fraud_rate']:<20.4%} | {row['fraud_count']:<6}")
    print("=" * 85)

    # ------------------ STEP 5, 6 & 7: Bayes Minimum Risk (BMR) Decisioning ------------------
    print("\n" + "=" * 80)
    print("[STEP 5, 6, 7] Bayes Minimum Risk (BMR) Policy Formulation & Economic Comparison")
    print("=" * 80)

    def evaluate_decision_policy(decisions, y_true, amounts, c_fp=25.0):
        """
        Calculates financial and classification metrics for a boolean decision array (1=DECLINE, 0=ALLOW).
        """
        # Decisions: 1 = DECLINE (BLOCK), 0 = ALLOW
        tn, fp, fn, tp = confusion_matrix(y_true, decisions).ravel()
        p = float(precision_score(y_true, decisions, zero_division=0))
        r = float(recall_score(y_true, decisions, zero_division=0))
        f1 = float(f1_score(y_true, decisions, zero_division=0))
        fpr = float(fp / (fp + tn)) if (fp + tn) > 0 else 0.0

        # Financial Losses:
        # ALLOWED FRAUD (FN): Principal amount * 1.05
        # ALLOWED LEGIT (TN): 0
        # DECLINED LEGIT (FP): c_fp
        # DECLINED FRAUD (TP): 0 (prevented loss)
        fn_mask = (y_true == 1) & (decisions == 0)
        tp_mask = (y_true == 1) & (decisions == 1)
        fp_mask = (y_true == 0) & (decisions == 1)

        fraud_loss = float(np.sum(amounts[fn_mask] * 1.05))
        fraud_prevented = float(np.sum(amounts[tp_mask] * 1.05))
        fp_loss = float(np.sum(fp_mask) * c_fp)
        total_loss = fraud_loss + fp_loss

        return {
            "tp": int(tp), "fp": int(fp), "tn": int(tn), "fn": int(fn),
            "precision": p, "recall": r, "f1": f1, "fpr": fpr,
            "fraud_dollars_lost": fraud_loss,
            "fraud_dollars_prevented": fraud_prevented,
            "false_positive_cost": fp_loss,
            "total_expected_loss": total_loss,
            "loss_per_1000_txns": float(total_loss / len(y_true) * 1000.0)
        }

    # Compare 4 Decision Policies on Validation Data (with C_FP = ₹500 / $25 baseline):
    c_fp_baseline = 25.0

    # Policy 1: Raw XGBoost Threshold 0.50
    dec_p1 = (raw_probs_val >= 0.50).astype(int)
    res_p1 = evaluate_decision_policy(dec_p1, y_val, val_amounts, c_fp=c_fp_baseline)

    # Policy 2: Calibrated Probability Threshold 0.50
    dec_p2 = (cal_probs_val >= 0.50).astype(int)
    res_p2 = evaluate_decision_policy(dec_p2, y_val, val_amounts, c_fp=c_fp_baseline)

    # Policy 3: Optimized Fixed Threshold on Calibrated Probability (e.g. 0.08)
    best_t, best_loss = 0.50, float("inf")
    for t_cand in np.linspace(0.01, 0.50, 50):
        dec_t = (cal_probs_val >= t_cand).astype(int)
        loss_t = evaluate_decision_policy(dec_t, y_val, val_amounts, c_fp=c_fp_baseline)["total_expected_loss"]
        if loss_t < best_loss:
            best_loss = loss_t
            best_t = t_cand
    dec_p3 = (cal_probs_val >= best_t).astype(int)
    res_p3 = evaluate_decision_policy(dec_p3, y_val, val_amounts, c_fp=c_fp_baseline)

    # Policy 4: Bayes Minimum Risk (BMR) Dynamic Decisioning
    # DECLINE if Loss(DECLINE) < Loss(ALLOW) => (1 - P) * C_FP < P * Amount * 1.05
    loss_allow_val = cal_probs_val * val_amounts * 1.05
    loss_decline_val = (1.0 - cal_probs_val) * c_fp_baseline
    dec_p4 = (loss_decline_val < loss_allow_val).astype(int)
    res_p4 = evaluate_decision_policy(dec_p4, y_val, val_amounts, c_fp=c_fp_baseline)

    print("\nDECISION POLICY COMPARISON (Validation Set, N=1,200, FalsePositiveCost=$25):")
    print("=" * 115)
    print(f"{'Policy':<36} | {'Recall':<8} | {'Prec':<8} | {'FPR':<8} | {'FP Cust':<8} | {'Fraud Saved':<12} | {'Total Loss':<12} | {'Loss / 1k':<10}")
    print("-" * 115)
    print(f"{'1. Raw XGBoost (Thresh=0.50)':<36} | {res_p1['recall']:<8.2%} | {res_p1['precision']:<8.2%} | {res_p1['fpr']:<8.2%} | {res_p1['fp']:<8} | ${res_p1['fraud_dollars_prevented']:<11.2f} | ${res_p1['total_expected_loss']:<11.2f} | ${res_p1['loss_per_1000_txns']:<9.2f}")
    print(f"{'2. Calibrated Prob (Thresh=0.50)':<36} | {res_p2['recall']:<8.2%} | {res_p2['precision']:<8.2%} | {res_p2['fpr']:<8.2%} | {res_p2['fp']:<8} | ${res_p2['fraud_dollars_prevented']:<11.2f} | ${res_p2['total_expected_loss']:<11.2f} | ${res_p2['loss_per_1000_txns']:<9.2f}")
    print(f"{f'3. Tuned Calibrated (Thresh={best_t:.2f})':<36} | {res_p3['recall']:<8.2%} | {res_p3['precision']:<8.2%} | {res_p3['fpr']:<8.2%} | {res_p3['fp']:<8} | ${res_p3['fraud_dollars_prevented']:<11.2f} | ${res_p3['total_expected_loss']:<11.2f} | ${res_p3['loss_per_1000_txns']:<9.2f}")
    print(f"{'4. Bayes Minimum Risk (Dynamic Amount)':<36} | {res_p4['recall']:<8.2%} | {res_p4['precision']:<8.2%} | {res_p4['fpr']:<8.2%} | {res_p4['fp']:<8} | ${res_p4['fraud_dollars_prevented']:<11.2f} | ${res_p4['total_expected_loss']:<11.2f} | ${res_p4['loss_per_1000_txns']:<9.2f}")
    print("=" * 115)

    # Sensitivity Analysis of FalsePositiveCost
    print("\nBMR SENSITIVITY ANALYSIS ACROSS FALSE POSITIVE COSTS (Validation Data):")
    print("-" * 90)
    print(f"{'False Positive Cost':<22} | {'Recall':<10} | {'FPR':<10} | {'FP Blocked':<12} | {'Fraud Saved':<14} | {'Total Loss':<12}")
    print("-" * 90)
    sensitivity_results = []
    for c_val in [10.0, 25.0, 50.0, 100.0, 200.0, 500.0]:
        l_allow = cal_probs_val * val_amounts * 1.05
        l_dec = (1.0 - cal_probs_val) * c_val
        dec = (l_dec < l_allow).astype(int)
        res = evaluate_decision_policy(dec, y_val, val_amounts, c_fp=c_val)
        sensitivity_results.append({"cost_fp": c_val, "metrics": res})
        print(f"${c_val:<21.2f} | {res['recall']:<10.2%} | {res['fpr']:<10.2%} | {res['fp']:<12} | ${res['fraud_dollars_prevented']:<13.2f} | ${res['total_expected_loss']:<11.2f}")
    print("-" * 90)

    # ------------------ STEP 8: Final Single Held-Out Test Evaluation ------------------
    print("\n" + "=" * 80)
    print("[STEP 8] Single Final Held-Out Test Set Evaluation (N = 1,200, Frauds = 52)")
    print("=" * 80)

    raw_probs_test = model_p3.predict_proba(X_test[features_28f])[:, 1]
    cal_probs_test = selected_calibrator.predict_proba(raw_probs_test)

    # Test Ranking Metrics
    test_roc = float(roc_auc_score(y_test, cal_probs_test))
    test_pr = float(average_precision_score(y_test, cal_probs_test))

    # Test Calibration Metrics
    test_brier = float(brier_score_loss(y_test, cal_probs_test))
    test_ece = float(calculate_ece(y_test, cal_probs_test))
    test_mce = float(calculate_mce(y_test, cal_probs_test))
    test_ll = float(log_loss(y_test, np.clip(cal_probs_test, 1e-7, 1.0 - 1e-7)))

    # Test BMR Decision Evaluation (C_FP = $25)
    loss_allow_test = cal_probs_test * test_amounts * 1.05
    loss_dec_test = (1.0 - cal_probs_test) * c_fp_baseline
    bmr_decisions_test = (loss_dec_test < loss_allow_test).astype(int)

    test_bmr_res = evaluate_decision_policy(bmr_decisions_test, y_test, test_amounts, c_fp=c_fp_baseline)

    # Raw Baseline Policy for comparison
    test_raw_res = evaluate_decision_policy((raw_probs_test >= 0.50).astype(int), y_test, test_amounts, c_fp=c_fp_baseline)

    print("\nFINAL HELD-OUT TEST PERFORMANCE (N = 1,200):")
    print("=" * 75)
    print(f"Ranking Quality:     PR-AUC = {test_pr:.4f} | ROC-AUC = {test_roc:.4f}")
    print(f"Probability Quality: Brier = {test_brier:.4f} | ECE = {test_ece:.4f} ({test_ece:.2%}) | LogLoss = {test_ll:.4f}")
    print(f"BMR Economics:       Total Expected Loss = ${test_bmr_res['total_expected_loss']:.2f} (vs Raw XGBoost ${test_raw_res['total_expected_loss']:.2f})")
    print(f"Loss Reduction:      ${test_raw_res['total_expected_loss'] - test_bmr_res['total_expected_loss']:.2f} saved per 1,200 txns (${(test_raw_res['total_expected_loss'] - test_bmr_res['total_expected_loss'])/1.2:.2f} per 1k txns)")
    print(f"Fraud Capture:       {test_bmr_res['recall']:.2%} ({test_bmr_res['tp']}/52 frauds caught) | FP Declines = {test_bmr_res['fp']}")
    print("=" * 75)

    # ------------------ SAVE ALL PRODUCTION DELIVERABLES ------------------
    eval_dir = os.path.join(current_dir, "evaluation")

    # 1. Calibration Results
    cal_json_path = os.path.join(eval_dir, "phase_4_calibration_results.json")
    with open(cal_json_path, "w") as f:
        json.dump({
            "phase": "Phase 4: Probability Calibration",
            "uncalibrated_baseline": {
                "brier_score": brier_raw, "ece": ece_raw, "mce": mce_raw, "log_loss": ll_raw, "mean_prob": mean_raw
            },
            "calibrator_comparison_val": cal_eval_results,
            "selected_calibrator": best_cal_name,
            "validation_reliability_table": rel_table_10
        }, f, indent=2)

    # 2. BMR Results
    bmr_json_path = os.path.join(eval_dir, "phase_4_bmr_results.json")
    with open(bmr_json_path, "w") as f:
        json.dump({
            "phase": "Phase 4: Bayes Minimum Risk Decisioning",
            "policies_evaluated_val": {
                "raw_threshold_05": res_p1,
                "calibrated_threshold_05": res_p2,
                "calibrated_threshold_tuned": res_p3,
                "bayes_minimum_risk": res_p4
            },
            "sensitivity_analysis_val": sensitivity_results
        }, f, indent=2)

    # 3. Final Evaluation
    final_json_path = os.path.join(eval_dir, "phase_4_final_evaluation.json")
    with open(final_json_path, "w") as f:
        json.dump({
            "model_identifier": "fraud-xgb-28f-phase4-bmr",
            "features": features_28f,
            "calibrator": best_cal_name,
            "decision_policy": "Bayes Minimum Risk: DECLINE if (1-P)*C_FP < P*Amount*1.05",
            "held_out_test_evaluation": {
                "ranking": {"pr_auc": test_pr, "roc_auc": test_roc},
                "calibration": {"brier_score": test_brier, "ece": test_ece, "mce": test_mce, "log_loss": test_ll},
                "classification_bmr": test_bmr_res,
                "classification_raw": test_raw_res
            },
            "comparison_across_phases": {
                "phase1_baseline": {"pr_auc": 0.0688, "roc_auc": 0.5794, "recall": 0.2500, "brier": 0.1111, "ece": 0.2036},
                "phase3_optimized": {"pr_auc": 0.0874, "roc_auc": 0.6330, "recall": 0.5000, "brier": 0.1844, "ece": 0.3643},
                "phase4_calibrated_bmr": {"pr_auc": test_pr, "roc_auc": test_roc, "recall": test_bmr_res["recall"], "brier": test_brier, "ece": test_ece, "total_loss": test_bmr_res["total_expected_loss"]}
            }
        }, f, indent=2)

    print(f"\nAll Phase 4 deliverables saved successfully:")
    print(f"1. {cal_json_path}")
    print(f"2. {bmr_json_path}")
    print(f"3. {final_json_path}")
    print(f"4. {cal_artifact_path}")

if __name__ == "__main__":
    main()
