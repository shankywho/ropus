"""
AI Risk Manager — Canonical ML Training, Calibration & Cost-Sensitive Policy Pipeline
Integrates Platt / Isotonic Calibration, ECE Evaluation, and Expected Loss Optimization
"""

import os
import sys
import json
import time
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
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt

# Ensure ml-service root in path
current_dir = os.path.dirname(os.path.abspath(__file__))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from data_pipeline.schema import CANONICAL_FEATURE_COLS, LEGACY_FEATURE_COLS
from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from data_pipeline.features import extract_canonical_features
from data_pipeline.preprocess import CanonicalPreprocessor
from data_pipeline.validate import validate_pipeline_integrity
from calibration.calibrator import ModelCalibrator, evaluate_calibration_methods
from calibration.cost_policy import (
    CostSensitivePolicyEngine,
    run_cost_threshold_analysis,
    run_review_capacity_analysis,
    run_cost_sensitivity_scenarios
)

def train_and_evaluate_pipeline(
    data_dir: str = None,
    output_model_dir: str = None,
    eval_dir: str = None,
    config_dir: str = None
):
    if data_dir is None:
        data_dir = os.path.join(current_dir, "data")
    if output_model_dir is None:
        output_model_dir = os.path.join(current_dir, "model")
    if eval_dir is None:
        eval_dir = os.path.join(current_dir, "evaluation")
    if config_dir is None:
        config_dir = os.path.join(os.path.dirname(current_dir), "config")
        
    os.makedirs(output_model_dir, exist_ok=True)
    os.makedirs(eval_dir, exist_ok=True)
    
    print("==================================================")
    print("PHASE 2: CALIBRATION & COST-SENSITIVE POLICY PIPELINE")
    print("==================================================")
    
    # 1. Load Dataset
    raw_df, load_meta = load_raw_dataset(data_dir=data_dir)
    print(f"Dataset Loaded: {load_meta['dataset_source']} ({len(raw_df)} rows, {load_meta.get('fraud_count')} fraud, rate: {load_meta.get('fraud_ratio', 0)*100:.2f}%)")
    
    # 2. Extract Canonical Features
    print("Extracting canonical point-in-time features...")
    feat_df = extract_canonical_features(raw_df)
    
    # 3. Temporal Chronological Split (70% Train, 15% Val, 15% Test)
    print("Executing chronological split (70% Train, 15% Val, 15% Test)...")
    df_train, df_val, df_test, split_info = temporal_train_val_test_split(
        feat_df, time_col="TransactionDT", train_ratio=0.70, val_ratio=0.15, test_ratio=0.15
    )
    print(f"Split sizes: Train={len(df_train)}, Val={len(df_val)}, Test={len(df_test)}")
    
    # 4. Strict Train-Fitted Preprocessing
    print("Fitting canonical preprocessor strictly on train set...")
    preprocessor = CanonicalPreprocessor()
    preprocessor.fit(df_train)
    
    X_train = preprocessor.transform(df_train)
    y_train = df_train["isFraud"].values
    
    X_val = preprocessor.transform(df_val)
    y_val = df_val["isFraud"].values
    
    X_test = preprocessor.transform(df_test)
    y_test = df_test["isFraud"].values
    test_amounts = df_test["amount"].values
    
    # Pipeline integrity validation
    is_valid, errors = validate_pipeline_integrity(
        X_train, X_val, X_test,
        df_train["TransactionDT"].values,
        df_val["TransactionDT"].values,
        df_test["TransactionDT"].values
    )
    if not is_valid:
        raise RuntimeError(f"Pipeline integrity validation failed: {errors}")
    print("✓ Pipeline integrity & temporal point-in-time safety verified.")
    
    # Calculate Class Imbalance Weighting
    n_neg = int(np.sum(y_train == 0))
    n_pos = int(np.sum(y_train == 1))
    scale_pos_weight = float(n_neg / n_pos) if n_pos > 0 else 1.0
    print(f"Class imbalance handling: scale_pos_weight={scale_pos_weight:.2f}")
    
    # 5. Train Base XGBoost Model (Model B on Train Set 70%)
    print("\n--- Training Base XGBoost Model (Model B) ---")
    model_b = xgb.XGBClassifier(
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
    model_b.fit(
        X_train[CANONICAL_FEATURE_COLS], y_train,
        eval_set=[(X_val[CANONICAL_FEATURE_COLS], y_val)],
        verbose=False
    )
    
    # Generate Raw Probabilities
    y_prob_raw_val = model_b.predict_proba(X_val[CANONICAL_FEATURE_COLS])[:, 1]
    y_prob_raw_test = model_b.predict_proba(X_test[CANONICAL_FEATURE_COLS])[:, 1]
    
    # 6. Fit & Evaluate Probability Calibration (Validation Split Selection)
    print("\n--- Fitting & Evaluating Probability Calibration ---")
    calibrator, cal_metrics = evaluate_calibration_methods(
        y_true_val=y_val,
        y_prob_raw_val=y_prob_raw_val,
        y_true_test=y_test,
        y_prob_raw_test=y_prob_raw_test,
        output_dir=eval_dir
    )
    # Production calibrator migration: Beta calibration selected for high continuous resolution
    calibrator.method = "beta"
    print(f"Selected Production Calibrator: {calibrator.method} ({cal_metrics['selection_rationale']})")
    
    # Calibrate Test Set Probabilities using selected calibrator
    y_prob_cal_test = calibrator.predict_proba(y_prob_raw_test, method="beta")
    
    # 7. Evaluate Cost-Sensitive Decision Engine
    print("\n--- Running Cost-Sensitive Decision Engine & Scenario Analyses ---")
    policy_cfg_path = os.path.join(config_dir, "risk-policy.json")
    if not os.path.exists(policy_cfg_path):
        policy_cfg_path = os.path.join(current_dir, "risk-policy.json")
        
    policy = CostSensitivePolicyEngine(
        false_positive_cost=500.0,
        manual_review_cost=100.0,
        fraud_multiplier=1.0,
        residual_review_rate=0.05,
        review_capacity_pct=0.10
    )
    
    # A. Cost Threshold Sweep
    cost_thresh_csv = os.path.join(eval_dir, "cost_threshold_analysis.csv")
    cost_thresh_png = os.path.join(eval_dir, "cost_threshold_analysis.png")
    df_cost_sweep = run_cost_threshold_analysis(
        y_true=y_test,
        p_calibrated=y_prob_cal_test,
        amounts=test_amounts,
        policy=policy,
        output_csv=cost_thresh_csv,
        output_png=cost_thresh_png
    )
    
    # B. Review Capacity Analysis (1%, 5%, 10%, 20%)
    review_cap_csv = os.path.join(eval_dir, "review_capacity_analysis.csv")
    df_review_cap = run_review_capacity_analysis(
        y_true=y_test,
        p_calibrated=y_prob_cal_test,
        amounts=test_amounts,
        policy=policy,
        output_csv=review_cap_csv,
        capacities=[0.01, 0.05, 0.10, 0.20]
    )
    
    # C. Cost Sensitivity Multi-Scenario Analysis (Scenarios A, B, C, D)
    scenario_results = run_cost_sensitivity_scenarios(
        y_true=y_test,
        p_calibrated=y_prob_cal_test,
        amounts=test_amounts,
        config_path=policy_cfg_path
    )
    
    # 8. Export Versioned Calibration Artifact
    cal_json_path = os.path.join(output_model_dir, "calibration.json")
    with open(cal_json_path, "w") as f:
        json.dump(calibrator.to_dict(), f, indent=2)
    print(f"Saved calibration artifact to: {cal_json_path}")
    
    # 9. Feature Importances
    booster = model_b.get_booster()
    score_dict = booster.get_score(importance_type="gain")
    total_gain = sum(score_dict.values()) if score_dict else 1.0
    global_importances = {k: round(float(v / total_gain), 4) for k, v in score_dict.items()}
    
    # 10. Model Metadata Export
    metadata_path = os.path.join(output_model_dir, "model_metadata.json")
    model_metadata = {
        "model_version": "xgb-ieee-canonical-v2-calibrated",
        "feature_schema_version": "1.0.0",
        "calibration_version": calibrator.version,
        "calibration_method": calibrator.method,
        "training_timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "dataset": load_meta["dataset_source"],
        "is_full_raw_dataset": load_meta["is_full_raw_dataset"],
        "train_rows": len(df_train),
        "validation_rows": len(df_val),
        "test_rows": len(df_test),
        "fraud_rate_train": round(float(np.mean(y_train)), 4),
        "fraud_rate_test": round(float(np.mean(y_test)), 4),
        "scale_pos_weight": round(scale_pos_weight, 2),
        "random_seed": 42,
        "feature_count": len(CANONICAL_FEATURE_COLS),
        "features": CANONICAL_FEATURE_COLS,
        "feature_importances": global_importances,
        "calibration_evaluation": cal_metrics,
        "cost_policy_scenarios": scenario_results,
        "preprocessor_state": preprocessor.to_dict(),
        "python_version": sys.version.split()[0],
        "xgboost_version": xgb.__version__
    }
    with open(metadata_path, "w") as f:
        json.dump(model_metadata, f, indent=2)
    print(f"Saved model metadata to: {metadata_path}")
    
    # 11. Save Joblib Model Bundle
    joblib_path = os.path.join(output_model_dir, "fraud_model.joblib")
    joblib.dump({
        "model": model_b,
        "preprocessor": preprocessor,
        "calibrator": calibrator,
        "metadata": model_metadata
    }, joblib_path)
    print(f"Saved Joblib model bundle to: {joblib_path}")
    
    # 12. Export ONNX Model
    onnx_path = os.path.join(output_model_dir, "fraud_model.onnx")
    from export_onnx import export_canonical_model_to_onnx
    export_canonical_model_to_onnx(model_b, preprocessor, onnx_path=onnx_path, n_features=len(CANONICAL_FEATURE_COLS))
    
    print("\n✓ Phase 2 Training, Calibration & Cost-Sensitive Policy Pipeline Complete.")
    return model_metadata

if __name__ == "__main__":
    train_and_evaluate_pipeline()
