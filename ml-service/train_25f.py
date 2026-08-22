"""
AI Risk Manager — Candidate 25-Feature XGBoost Retraining & Ablation Pipeline (Phase 3.9)
Trains and exports candidate 25-feature model to model/candidates/ with strict leakage prevention
"""

import os
import sys
import json
import time
import hashlib
import argparse
import numpy as np
import pandas as pd
import joblib
import xgboost as xgb
from sklearn.metrics import (
    roc_auc_score,
    average_precision_score,
    precision_score,
    recall_score,
    f1_score,
    confusion_matrix
)
import onnx
from onnxmltools import convert_xgboost
from onnxmltools.convert.common.data_types import FloatTensorType
import onnxruntime as rt

# Ensure ml-service root in path
current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from data_pipeline.schema import CANONICAL_25_FEATURE_COLS, CANONICAL_15_FEATURE_COLS
from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from data_pipeline.features import extract_canonical_25_features
from data_pipeline.preprocess import CanonicalPreprocessor
from data_pipeline.validate import validate_pipeline_integrity

def compute_checksum(file_path: str) -> str:
    """Computes SHA-256 hash of a file."""
    sha256 = hashlib.sha256()
    with open(file_path, "rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            sha256.update(chunk)
    return sha256.hexdigest()

def train_and_evaluate_25f_candidate(
    data_dir: str = None,
    output_dir: str = None,
    eval_dir: str = None,
    random_seed: int = 42
):
    start_time = time.time()
    if data_dir is None:
        data_dir = os.path.join(current_dir, "data")
    if output_dir is None:
        output_dir = os.path.join(current_dir, "model", "candidates")
    if eval_dir is None:
        eval_dir = os.path.join(current_dir, "evaluation")

    os.makedirs(output_dir, exist_ok=True)
    os.makedirs(eval_dir, exist_ok=True)

    print("=" * 70)
    print("PHASE 3.9: 25-FEATURE CANONICAL XGBOOST RETRAINING PIPELINE")
    print(f"Random Seed: {random_seed} | Target Output Dir: {output_dir}")
    print("=" * 70)

    # 1. Load Raw Dataset
    print("\n[STEP 1] Loading raw dataset...")
    df_raw, data_meta = load_raw_dataset(data_dir=data_dir)
    print(f"Loaded {len(df_raw)} records from dataset (Source: {data_meta.get('dataset_source')}).")

    # 2. Strict Chronological Temporal Splitting (70% Train, 15% Val, 15% Test)
    print("\n[STEP 2] Performing strict chronological temporal split...")
    df_train_raw, df_val_raw, df_test_raw, split_info = temporal_train_val_test_split(
        df_raw,
        time_col="TransactionDT",
        train_ratio=0.70,
        val_ratio=0.15,
        test_ratio=0.15
    )
    print(f"Train set: {len(df_train_raw)} rows (Fraud: {split_info['train_fraud_rate']:.4f})")
    print(f"Val set:   {len(df_val_raw)} rows (Fraud: {split_info['val_fraud_rate']:.4f})")
    print(f"Test set:  {len(df_test_raw)} rows (Fraud: {split_info['test_fraud_rate']:.4f})")

    # 3. Point-in-Time Safe 25-Feature Extraction
    print("\n[STEP 3] Extracting point-in-time safe 25-feature representations...")
    df_feat_train = extract_canonical_25_features(df_train_raw)
    df_feat_val = extract_canonical_25_features(df_val_raw)
    df_feat_test = extract_canonical_25_features(df_test_raw)

    # 4. Strict Train-Fitted Preprocessing
    print("\n[STEP 4] Fitting preprocessor strictly on train split...")
    preprocessor = CanonicalPreprocessor(feature_contract="v2.5")
    preprocessor.fit(df_feat_train)

    X_train_25 = preprocessor.transform(df_feat_train, feature_contract="v2.5")
    X_val_25 = preprocessor.transform(df_feat_val, feature_contract="v2.5")
    X_test_25 = preprocessor.transform(df_feat_test, feature_contract="v2.5")

    y_train = df_feat_train["isFraud"].values
    y_val = df_feat_val["isFraud"].values
    y_test = df_feat_test["isFraud"].values

    # 5. Pipeline & Data Integrity Validation
    print("\n[STEP 5] Validating pipeline integrity...")
    is_valid, errors = validate_pipeline_integrity(
        X_train_25, X_val_25, X_test_25,
        train_dts=df_feat_train["TransactionDT"].values,
        val_dts=df_feat_val["TransactionDT"].values,
        test_dts=df_feat_test["TransactionDT"].values,
        expected_cols=CANONICAL_25_FEATURE_COLS
    )
    if not is_valid:
        raise ValueError(f"Pipeline validation failed: {errors}")
    print("Pipeline validation PASSED: 25 features strictly ordered, 0 NaNs, strictly chronological.")

    # 6. Class Imbalance Strategy
    neg_count = int(np.sum(y_train == 0))
    pos_count = int(np.sum(y_train == 1))
    scale_pos_weight = float(neg_count) / max(1.0, float(pos_count))
    print(f"\n[STEP 6] Class imbalance: Neg={neg_count}, Pos={pos_count} => scale_pos_weight={scale_pos_weight:.2f}")

    # 7. Model A: Legacy 15-Feature Baseline (on exact same temporal split)
    print("\n[STEP 7] Training Model A: 15-Feature Baseline...")
    X_train_15 = X_train_25[CANONICAL_15_FEATURE_COLS]
    X_val_15 = X_val_25[CANONICAL_15_FEATURE_COLS]
    X_test_15 = X_test_25[CANONICAL_15_FEATURE_COLS]

    xgb_params = {
        "n_estimators": 100,
        "max_depth": 5,
        "learning_rate": 0.08,
        "subsample": 0.85,
        "colsample_bytree": 0.85,
        "scale_pos_weight": scale_pos_weight,
        "random_state": random_seed,
        "eval_metric": "logloss",
        "tree_method": "hist"
    }

    model_15 = xgb.XGBClassifier(**xgb_params)
    model_15.fit(X_train_15, y_train, eval_set=[(X_val_15, y_val)], verbose=False)
    preds_prob_15 = model_15.predict_proba(X_test_15)[:, 1]

    # 8. Model B: Candidate 25-Feature Model
    print("\n[STEP 8] Training Model B: Candidate 25-Feature Model...")
    model_25 = xgb.XGBClassifier(**xgb_params)
    model_25.fit(X_train_25, y_train, eval_set=[(X_val_25, y_val)], verbose=False)
    preds_prob_25 = model_25.predict_proba(X_test_25)[:, 1]

    # 9. Ablation Study & Metric Evaluation
    print("\n[STEP 9] Evaluating Test Metrics & Ablation Comparison...")
    def evaluate_model(y_true, y_prob):
        roc = roc_auc_score(y_true, y_prob)
        pr = average_precision_score(y_true, y_prob)
        y_pred = (y_prob >= 0.50).astype(int)
        p = precision_score(y_true, y_pred, zero_division=0)
        r = recall_score(y_true, y_pred, zero_division=0)
        f1 = f1_score(y_true, y_pred, zero_division=0)
        tn, fp, fn, tp = confusion_matrix(y_true, y_pred).ravel()
        fpr = fp / (fp + tn) if (fp + tn) > 0 else 0.0
        fnr = fn / (fn + tp) if (fn + tp) > 0 else 0.0
        return {
            "roc_auc": float(roc),
            "pr_auc": float(pr),
            "precision": float(p),
            "recall": float(r),
            "f1": float(f1),
            "fpr": float(fpr),
            "fnr": float(fnr),
            "tp": int(tp),
            "fp": int(fp),
            "tn": int(tn),
            "fn": int(fn)
        }

    m15_metrics = evaluate_model(y_test, preds_prob_15)
    m25_metrics = evaluate_model(y_test, preds_prob_25)

    print("\n" + "=" * 65)
    print("ABLATION STUDY RESULTS (Test Set):")
    print(f"{'Metric':<18} | {'Legacy 15F':<12} | {'Candidate 25F':<14} | {'Delta':<10}")
    print("-" * 65)
    for k in ["roc_auc", "pr_auc", "precision", "recall", "f1", "fpr", "fnr"]:
        v15 = m15_metrics[k]
        v25 = m25_metrics[k]
        delta = v25 - v15
        sign = "+" if delta >= 0 else ""
        print(f"{k:<18} | {v15:<12.4f} | {v25:<14.4f} | {sign}{delta:<10.4f}")
    print("=" * 65)

    # 10. Feature Importance Analysis
    print("\n[STEP 10] Calculating Feature Importance (Gain, Weight, Cover)...")
    booster = model_25.get_booster()
    score_gain = booster.get_score(importance_type="gain")
    score_weight = booster.get_score(importance_type="weight")
    score_cover = booster.get_score(importance_type="cover")

    feature_importances = []
    for i, col in enumerate(CANONICAL_25_FEATURE_COLS):
        # XGBoost assigns f0, f1... or column names
        f_key = col if col in score_gain else f"f{i}"
        gain = score_gain.get(f_key, score_gain.get(col, 0.0))
        weight = score_weight.get(f_key, score_weight.get(col, 0.0))
        cover = score_cover.get(f_key, score_cover.get(col, 0.0))
        feature_importances.append({
            "index": i,
            "feature": col,
            "is_new_feature": i >= 15,
            "gain": float(gain),
            "weight": float(weight),
            "cover": float(cover)
        })

    feature_importances.sort(key=lambda x: x["gain"], reverse=True)
    for rank, fi in enumerate(feature_importances[:10], 1):
        tag = "[NEW]" if fi["is_new_feature"] else "     "
        print(f"{rank:2d}. {tag} {fi['feature']:<35} Gain: {fi['gain']:.4f} | Weight: {fi['weight']:.0f}")

    # 11. Save Joblib Model Bundle
    print("\n[STEP 11] Saving candidate model artifacts...")
    joblib_path = os.path.join(output_dir, "fraud_model_25f_candidate.joblib")
    onnx_path = os.path.join(output_dir, "fraud_model_25f_candidate.onnx")
    meta_path = os.path.join(output_dir, "metadata.json")

    joblib.dump({
        "model": model_25,
        "preprocessor": preprocessor.to_dict(),
        "feature_cols": CANONICAL_25_FEATURE_COLS,
        "metrics": m25_metrics
    }, joblib_path)
    print(f"Saved candidate Joblib model to: {joblib_path}")

    # 12. Export to ONNX
    print("\n[STEP 12] Exporting Candidate Model to ONNX format (opset 15)...")
    booster_25 = model_25.get_booster()
    booster_25.feature_names = [f"f{i}" for i in range(25)]
    initial_type = [('float_input', FloatTensorType([None, 25]))]
    onnx_model = convert_xgboost(model_25, initial_types=initial_type, target_opset=15)
    onnx.save_model(onnx_model, onnx_path)
    print(f"Saved candidate ONNX model to: {onnx_path} (Size: {os.path.getsize(onnx_path)} bytes)")

    # 13. ONNX Runtime Inference Parity Verification
    print("\n[STEP 13] Verifying native XGBoost vs ONNX numerical parity on test set...")
    sess = rt.InferenceSession(onnx_path, providers=['CPUExecutionProvider'])
    input_name = sess.get_inputs()[0].name
    sample_input = X_test_25.values.astype(np.float32)

    onnx_raw = sess.run(None, {input_name: sample_input})
    prob_output = onnx_raw[1]
    if isinstance(prob_output, list) and len(prob_output) > 0 and isinstance(prob_output[0], dict):
        onnx_probs = np.array([float(d.get(1, 0.0)) for d in prob_output])
    elif isinstance(prob_output, np.ndarray) and prob_output.ndim >= 2:
        onnx_probs = onnx_raw[1][:, 1].astype(np.float64)
    else:
        onnx_probs = onnx_raw[0].flatten().astype(np.float64)

    max_diff = float(np.max(np.abs(preds_prob_25 - onnx_probs)))
    print(f"Max absolute probability difference (Native vs ONNX): {max_diff:.8f}")
    assert max_diff < 1e-4, f"Parity violation: max_diff {max_diff} exceeds 1e-4 tolerance"
    assert not np.isnan(onnx_probs).any(), "ONNX outputs contain NaNs"
    print("ONNX Parity Verification PASSED.")

    # 14. Write Metadata JSON
    onnx_checksum = compute_checksum(onnx_path)
    joblib_checksum = compute_checksum(joblib_path)

    metadata = {
        "model_version": "fraud-xgb-25f-candidate-v1",
        "feature_contract": "v2.5",
        "feature_count": 25,
        "algorithm": "XGBoost Classifier",
        "hyperparameters": xgb_params,
        "random_seed": random_seed,
        "trained_at": pd.Timestamp.now('UTC').isoformat(),
        "total_rows": len(df_raw),
        "split": split_info,
        "test_metrics_25f": m25_metrics,
        "test_metrics_15f": m15_metrics,
        "ablation_delta": {k: m25_metrics[k] - m15_metrics[k] for k in m25_metrics if isinstance(m25_metrics[k], float)},
        "feature_importance_ranking": feature_importances,
        "onnx_artifact": {
            "path": onnx_path,
            "checksum_sha256": onnx_checksum,
            "input_dimension": 25,
            "max_parity_diff": max_diff
        },
        "joblib_artifact": {
            "path": joblib_path,
            "checksum_sha256": joblib_checksum
        },
        "is_production_active": False,
        "notes": "Candidate 25-feature model for offline evaluation & shadow scoring (Phase 3.9). Does not serve live production traffic."
    }

    with open(meta_path, "w") as f:
        json.dump(metadata, f, indent=2)
    print(f"Saved candidate metadata to: {meta_path}")

    # Save summary report in evaluation/
    eval_report_path = os.path.join(eval_dir, "phase_3_9_candidate_evaluation.json")
    with open(eval_report_path, "w") as f:
        json.dump(metadata, f, indent=2)

    elapsed = time.time() - start_time
    print(f"\nPhase 3.9 Candidate Retraining Pipeline successfully completed in {elapsed:.2f}s.")
    return metadata

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Retrain 25-Feature Candidate Model")
    parser.add_argument("--data-dir", type=str, default=None)
    parser.add_argument("--output-dir", type=str, default=None)
    parser.add_argument("--eval-dir", type=str, default=None)
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()

    train_and_evaluate_25f_candidate(
        data_dir=args.data_dir,
        output_dir=args.output_dir,
        eval_dir=args.eval_dir,
        random_seed=args.seed
    )
