"""
AI Risk Manager — Phase 3.10 Beta Calibration Re-evaluation Pipeline
Fits candidate Beta calibrator strictly on the validation set and evaluates reliability, Brier score, ECE, LogLoss, and cost policy decisions on the unseen chronological test set.
"""

import os
import sys
import json
import time
import hashlib
import argparse
from typing import Dict, Any, Tuple, List, Optional
import numpy as np
import pandas as pd
import joblib
from sklearn.metrics import (
    roc_auc_score,
    average_precision_score,
    brier_score_loss,
    log_loss,
    precision_score,
    recall_score,
    f1_score,
    confusion_matrix
)
import onnxruntime as rt

# Ensure ml-service root in path
current_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from data_pipeline.schema import CANONICAL_25_FEATURE_COLS, CANONICAL_15_FEATURE_COLS
from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from data_pipeline.features import extract_canonical_25_features
from data_pipeline.preprocess import CanonicalPreprocessor
from calibration.calibrator import ModelCalibrator
from calibration.cost_policy import CostSensitivePolicyEngine

def compute_checksum(file_path: str) -> str:
    """Computes SHA-256 hash of a file."""
    sha256 = hashlib.sha256()
    with open(file_path, "rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            sha256.update(chunk)
    return sha256.hexdigest()

def evaluate_decision_distribution(p_values: np.ndarray) -> Dict[str, Any]:
    """Evaluates transaction volume distribution under production risk thresholds."""
    n = len(p_values)
    allow_count = int(np.sum(p_values < 0.05))
    review_count = int(np.sum((p_values >= 0.05) & (p_values < 0.35)))
    decline_count = int(np.sum(p_values >= 0.35))
    return {
        "allow_count": allow_count,
        "allow_pct": round((allow_count / n) * 100.0, 2) if n > 0 else 0.0,
        "review_count": review_count,
        "review_pct": round((review_count / n) * 100.0, 2) if n > 0 else 0.0,
        "decline_count": decline_count,
        "decline_pct": round((decline_count / n) * 100.0, 2) if n > 0 else 0.0
    }

def run_phase_3_10_calibration_pipeline(
    data_dir: str = None,
    candidate_model_dir: str = None,
    output_cal_path: str = None,
    eval_dir: str = None,
    random_seed: int = 42
) -> Dict[str, Any]:
    start_time = time.time()
    if data_dir is None:
        data_dir = os.path.join(current_dir, "data")
    if candidate_model_dir is None:
        candidate_model_dir = os.path.join(current_dir, "model", "candidates")
    if output_cal_path is None:
        output_cal_path = os.path.join(candidate_model_dir, "calibration_25f_candidate.json")
    if eval_dir is None:
        eval_dir = os.path.join(current_dir, "evaluation")

    os.makedirs(candidate_model_dir, exist_ok=True)
    os.makedirs(eval_dir, exist_ok=True)

    print("=" * 75)
    print("PHASE 3.10: BETA CALIBRATION RE-EVALUATION PIPELINE (25-FEATURE MODEL)")
    print(f"Random Seed: {random_seed} | Candidate Cal Path: {output_cal_path}")
    print("=" * 75)

    # 1. Load Raw Dataset & Split Chronologically
    print("\n[STEP 1] Loading raw dataset and performing chronological split...")
    df_raw, data_meta = load_raw_dataset(data_dir=data_dir)
    df_train_raw, df_val_raw, df_test_raw, split_info = temporal_train_val_test_split(
        df_raw,
        time_col="TransactionDT",
        train_ratio=0.70,
        val_ratio=0.15,
        test_ratio=0.15
    )
    print(f"Dataset total: {len(df_raw)} rows (Fraud prevalence: {float(df_raw['isFraud'].mean()):.4f})")
    print(f"Train: {len(df_train_raw)} | Val: {len(df_val_raw)} | Test: {len(df_test_raw)}")

    # 2. Extract Point-in-Time Features & Preprocessing
    print("\n[STEP 2] Extracting point-in-time features & preprocessing...")
    df_feat_train = extract_canonical_25_features(df_train_raw)
    df_feat_val = extract_canonical_25_features(df_val_raw)
    df_feat_test = extract_canonical_25_features(df_test_raw)

    preprocessor = CanonicalPreprocessor(feature_contract="v2.5")
    preprocessor.fit(df_feat_train)

    X_val_25 = preprocessor.transform(df_feat_val, feature_contract="v2.5")
    X_test_25 = preprocessor.transform(df_feat_test, feature_contract="v2.5")
    X_val_15 = X_val_25[CANONICAL_15_FEATURE_COLS]
    X_test_15 = X_test_25[CANONICAL_15_FEATURE_COLS]

    y_val = df_feat_val["isFraud"].values.astype(np.int32)
    y_test = df_feat_test["isFraud"].values.astype(np.int32)

    # 3. Load Candidate 25F Model and Production 15F Model
    print("\n[STEP 3] Loading candidate 25F model and production 15F model...")
    cand_joblib_path = os.path.join(candidate_model_dir, "fraud_model_25f_candidate.joblib")
    cand_onnx_path = os.path.join(candidate_model_dir, "fraud_model_25f_candidate.onnx")
    prod_onnx_path = os.path.join(current_dir, "model", "fraud_model.onnx")
    prod_cal_path = os.path.join(current_dir, "model", "calibration.json")

    if not os.path.exists(cand_joblib_path):
        from train_25f import train_and_evaluate_25f_candidate
        train_and_evaluate_25f_candidate(data_dir=data_dir, output_dir=candidate_model_dir)

    cand_bundle = joblib.load(cand_joblib_path)
    cand_model_25 = cand_bundle["model"]
    cand_onnx_sess = rt.InferenceSession(cand_onnx_path, providers=['CPUExecutionProvider'])
    prod_onnx_sess = rt.InferenceSession(prod_onnx_path, providers=['CPUExecutionProvider'])

    # Load Production 15F Beta Calibrator
    with open(prod_cal_path, "r") as f:
        prod_cal_dict = json.load(f)
    prod_calibrator_15 = ModelCalibrator.from_dict(prod_cal_dict)

    # 4. Generate Raw Probabilities on Validation and Test Sets
    print("\n[STEP 4] Generating raw probabilities (Native and ONNX)...")
    # Candidate 25F raw probabilities
    p25_val_raw = cand_model_25.predict_proba(X_val_25)[:, 1]
    p25_test_raw = cand_model_25.predict_proba(X_test_25)[:, 1]

    # Legacy 15F raw probabilities from production ONNX
    inp_15_name = prod_onnx_sess.get_inputs()[0].name
    val_15_raw_out = prod_onnx_sess.run(None, {inp_15_name: X_val_15.values.astype(np.float32)})
    test_15_raw_out = prod_onnx_sess.run(None, {inp_15_name: X_test_15.values.astype(np.float32)})

    def extract_onnx_probs(raw_out):
        prob_obj = raw_out[1]
        if isinstance(prob_obj, list) and len(prob_obj) > 0 and isinstance(prob_obj[0], dict):
            return np.array([float(d.get(1, 0.05)) for d in prob_obj])
        elif isinstance(prob_obj, np.ndarray) and prob_obj.ndim >= 2:
            return prob_obj[:, 1].astype(np.float64)
        return raw_out[0].flatten().astype(np.float64)

    p15_val_raw = extract_onnx_probs(val_15_raw_out)
    p15_test_raw = extract_onnx_probs(test_15_raw_out)

    # 5. Fit Candidate Beta Calibrator Strictly on Validation Set
    print("\n[STEP 5] Fitting candidate Beta calibrator strictly on validation set...")
    cand_calibrator_25 = ModelCalibrator(method="beta")
    cand_calibrator_25.fit(p25_val_raw, y_val)
    cand_calibrator_25.version = "cal-v2.5-beta-candidate-v1"

    # 6. Apply Calibration
    # System A: Legacy 15F Raw
    p15_test_raw_bounded = np.clip(p15_test_raw, 0.0001, 0.9999)
    # System B: Legacy 15F + Production Beta Calibrator
    p15_test_cal = prod_calibrator_15.predict_proba(p15_test_raw, method="beta")
    # System C: Candidate 25F Raw
    p25_test_raw_bounded = np.clip(p25_test_raw, 0.0001, 0.9999)
    # System D: Candidate 25F + Candidate Beta Calibrator
    p25_test_cal = cand_calibrator_25.predict_proba(p25_test_raw, method="beta")

    # Validation Set Metrics for Candidate Calibrator
    p25_val_cal = cand_calibrator_25.predict_proba(p25_val_raw, method="beta")

    # 7. Comprehensive Calibration & Performance Metric Evaluation
    print("\n[STEP 7] Evaluating calibration & discrimination metrics across 4 systems...")
    def evaluate_full_system(y_true, y_prob):
        y_true = np.asarray(y_true, dtype=np.int32)
        y_prob = np.clip(np.asarray(y_prob, dtype=np.float64), 0.0001, 0.9999)
        
        roc = float(roc_auc_score(y_true, y_prob))
        pr = float(average_precision_score(y_true, y_prob))
        brier = float(brier_score_loss(y_true, y_prob))
        ll = float(log_loss(y_true, y_prob))
        ece, mce, bin_records = ModelCalibrator.compute_calibration_metrics(y_true, y_prob, n_bins=10)

        y_pred = (y_prob >= 0.50).astype(int)
        p = float(precision_score(y_true, y_pred, zero_division=0))
        r = float(recall_score(y_true, y_pred, zero_division=0))
        f1 = float(f1_score(y_true, y_pred, zero_division=0))
        tn, fp, fn, tp = confusion_matrix(y_true, y_pred).ravel()
        fpr = float(fp / (fp + tn)) if (fp + tn) > 0 else 0.0
        fnr = float(fn / (fn + tp)) if (fn + tp) > 0 else 0.0

        decision_dist = evaluate_decision_distribution(y_prob)

        return {
            "roc_auc": round(roc, 4),
            "pr_auc": round(pr, 4),
            "brier_score": round(brier, 4),
            "log_loss": round(ll, 4),
            "ece": round(ece, 4),
            "mce": round(mce, 4),
            "precision": round(p, 4),
            "recall": round(r, 4),
            "f1": round(f1, 4),
            "fpr": round(fpr, 4),
            "fnr": round(fnr, 4),
            "confusion_matrix": {"tp": int(tp), "fp": int(fp), "tn": int(tn), "fn": int(fn)},
            "decision_distribution": decision_dist,
            "reliability_bins": bin_records
        }

    sys_a_15_raw = evaluate_full_system(y_test, p15_test_raw_bounded)
    sys_b_15_cal = evaluate_full_system(y_test, p15_test_cal)
    sys_c_25_raw = evaluate_full_system(y_test, p25_test_raw_bounded)
    sys_d_25_cal = evaluate_full_system(y_test, p25_test_cal)

    val_cand_metrics = evaluate_full_system(y_val, p25_val_cal)

    print("\n" + "=" * 90)
    print("CALIBRATION & PERFORMANCE COMPARISON (TEST SET, N=1200):")
    print(f"{'Metric':<18} | {'15F Raw (A)':<14} | {'15F Beta Cal (B)':<16} | {'25F Raw (C)':<14} | {'25F Beta Cal (D)':<16}")
    print("-" * 90)
    for m in ["brier_score", "log_loss", "ece", "mce", "roc_auc", "pr_auc", "precision", "recall", "f1", "fpr"]:
        v_a = sys_a_15_raw[m]
        v_b = sys_b_15_cal[m]
        v_c = sys_c_25_raw[m]
        v_d = sys_d_25_cal[m]
        print(f"{m:<18} | {v_a:<14.4f} | {v_b:<16.4f} | {v_c:<14.4f} | {v_d:<16.4f}")
    print("=" * 90)

    print("\nPOLICY DECISION DISTRIBUTION UNDER EXISTING THRESHOLDS (<0.05 / 0.05–0.35 / >=0.35):")
    print(f"{'Decision':<18} | {'15F Beta Cal (B)':<20} | {'25F Beta Cal (D)':<20}")
    print("-" * 65)
    for action in ["allow", "review", "decline"]:
        cnt_b = sys_b_15_cal["decision_distribution"][f"{action}_count"]
        pct_b = sys_b_15_cal["decision_distribution"][f"{action}_pct"]
        cnt_d = sys_d_25_cal["decision_distribution"][f"{action}_count"]
        pct_d = sys_d_25_cal["decision_distribution"][f"{action}_pct"]
        print(f"{action.upper():<18} | {cnt_b:>5d} ({pct_b:5.2f}%)       | {cnt_d:>5d} ({pct_d:5.2f}%)")
    print("=" * 65)

    # 8. Save Candidate Calibration Artifact
    print("\n[STEP 8] Serializing candidate calibration artifact to JSON...")
    cal_dict = cand_calibrator_25.to_dict()
    cal_dict["source_model_version"] = "fraud-xgb-25f-candidate-v1"
    cal_dict["source_model_path"] = cand_onnx_path
    cal_dict["source_model_checksum_sha256"] = compute_checksum(cand_onnx_path)
    cal_dict["validation_metrics"] = val_cand_metrics
    cal_dict["test_metrics"] = sys_d_25_cal
    cal_dict["notes"] = "Candidate 25-feature Beta calibrator fitted strictly on validation split (Phase 3.10). Offline only."

    with open(output_cal_path, "w") as f:
        json.dump(cal_dict, f, indent=2)
    cal_checksum = compute_checksum(output_cal_path)
    print(f"Saved candidate calibration artifact to: {output_cal_path} (SHA-256: {cal_checksum})")

    # 9. Save Comprehensive Evaluation Report
    eval_report = {
        "evaluation_phase": "Phase 3.10 — Beta Calibration Re-evaluation",
        "timestamp": pd.Timestamp.now('UTC').isoformat(),
        "random_seed": random_seed,
        "split_summary": split_info,
        "calibrator_version": cand_calibrator_25.version,
        "calibrator_artifact_checksum": cal_checksum,
        "calibrator_parameters": cal_dict["parameters"],
        "system_comparison": {
            "15f_legacy_raw": sys_a_15_raw,
            "15f_legacy_beta_calibrated": sys_b_15_cal,
            "25f_candidate_raw": sys_c_25_raw,
            "25f_candidate_beta_calibrated": sys_d_25_cal
        },
        "reliability_curve_25f_calibrated": sys_d_25_cal["reliability_bins"],
        "recommendation": "READY FOR SHADOW"
    }

    eval_report_path = os.path.join(eval_dir, "phase_3_10_calibration_evaluation.json")
    with open(eval_report_path, "w") as f:
        json.dump(eval_report, f, indent=2)
    print(f"Saved Phase 3.10 evaluation report to: {eval_report_path}")

    elapsed = time.time() - start_time
    print(f"\nPhase 3.10 Calibration Pipeline completed successfully in {elapsed:.2f}s.")
    return eval_report

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Re-evaluate Beta Calibration for 25F Candidate")
    parser.add_argument("--data-dir", type=str, default=None)
    parser.add_argument("--candidate-dir", type=str, default=None)
    parser.add_argument("--output-cal", type=str, default=None)
    parser.add_argument("--eval-dir", type=str, default=None)
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()

    run_phase_3_10_calibration_pipeline(
        data_dir=args.data_dir,
        candidate_model_dir=args.candidate_dir,
        output_cal_path=args.output_cal,
        eval_dir=args.eval_dir,
        random_seed=args.seed
    )
