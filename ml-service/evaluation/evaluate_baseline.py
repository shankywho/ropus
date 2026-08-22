import os
import json
import time
import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.metrics import (
    roc_auc_score,
    average_precision_score,
    precision_score,
    recall_score,
    f1_score,
    confusion_matrix,
    roc_curve,
    precision_recall_curve
)
import onnxruntime as rt
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt

import sys
current_dir = os.path.dirname(os.path.abspath(__file__))
parent_dir = os.path.dirname(current_dir)
if parent_dir not in sys.path:
    sys.path.insert(0, parent_dir)

# Import synthetic generator from train.py
from train import generate_synthetic_data

def run_evaluation(
    onnx_path="model/fraud_model.onnx",
    output_dir="evaluation"
):
    os.makedirs(output_dir, exist_ok=True)
    
    print("==================================================")
    print("PHASE -1: BASELINE ML REPRODUCIBILITY & EVALUATION")
    print("==================================================")
    
    # 1. Dataset Generation (Exact parameters from train.py)
    start_data_time = time.perf_counter()
    df = generate_synthetic_data(n_samples=30000, random_seed=42)
    feature_cols = ['amount', 'ip_velocity_1h', 'token_velocity_24h', 'is_new_device', 'hour_of_day']
    X = df[feature_cols].values
    y = df['is_fraud'].values
    
    total_samples = len(df)
    total_fraud = int(y.sum())
    total_legit = int(total_samples - total_fraud)
    fraud_rate = float(total_fraud / total_samples)
    
    print(f"Dataset Size: {total_samples} samples")
    print(f"Features ({len(feature_cols)}): {feature_cols}")
    print(f"Class Distribution: {total_legit} Legitimate, {total_fraud} Fraud ({fraud_rate*100:.2f}%)")
    
    # Train / Test split (80/20 stratified, seed 42)
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.20, random_state=42, stratify=y
    )
    test_fraud = int(y_test.sum())
    test_legit = int(len(y_test) - test_fraud)
    
    # 2. Load ONNX Runtime Session
    if not os.path.exists(onnx_path):
        raise FileNotFoundError(f"ONNX model not found at {onnx_path}")
        
    model_size_bytes = os.path.getsize(onnx_path)
    session = rt.InferenceSession(onnx_path, providers=['CPUExecutionProvider'])
    input_name = session.get_inputs()[0].name
    
    # 3. Batch & Single-Sample Inference Benchmarking
    latencies = []
    # Warmup
    for _ in range(50):
        _ = session.run(None, {input_name: X_test[:1].astype(np.float32)})
        
    # Measure 1,000 single-item inferences
    for i in range(min(1000, len(X_test))):
        t0 = time.perf_counter()
        _ = session.run(None, {input_name: X_test[i:i+1].astype(np.float32)})
        latencies.append((time.perf_counter() - t0) * 1000.0) # ms
        
    p50_ms = float(np.percentile(latencies, 50))
    p95_ms = float(np.percentile(latencies, 95))
    p99_ms = float(np.percentile(latencies, 99))
    mean_ms = float(np.mean(latencies))
    
    print(f"Inference Latency: p50={p50_ms:.3f}ms, p95={p95_ms:.3f}ms, p99={p99_ms:.3f}ms, mean={mean_ms:.3f}ms")
    
    # Full Test Set Predictions
    raw_preds = session.run(None, {input_name: X_test.astype(np.float32)})
    prob_output = raw_preds[1]
    if isinstance(prob_output, list) and isinstance(prob_output[0], dict):
        y_prob = np.array([p.get(1, 0.0) for p in prob_output])
    elif isinstance(prob_output, np.ndarray) and prob_output.ndim >= 2:
        y_prob = prob_output[:, 1]
    else:
        y_prob = raw_preds[0].flatten()
        
    y_pred = (y_prob >= 0.50).astype(int)
    
    # 4. Metrics Calculation
    roc_auc = float(roc_auc_score(y_test, y_prob))
    pr_auc = float(average_precision_score(y_test, y_prob))
    prec = float(precision_score(y_test, y_pred, zero_division=0))
    rec = float(recall_score(y_test, y_pred, zero_division=0))
    f1 = float(f1_score(y_test, y_pred, zero_division=0))
    
    cm = confusion_matrix(y_test, y_pred)
    tn, fp, fn, tp = [int(v) for v in cm.ravel()]
    
    print(f"ROC-AUC: {roc_auc:.4f}")
    print(f"PR-AUC:  {pr_auc:.4f}")
    print(f"Precision: {prec:.4f}, Recall: {rec:.4f}, F1: {f1:.4f}")
    print(f"Confusion Matrix: TN={tn}, FP={fp}, FN={fn}, TP={tp}")
    
    metrics_data = {
        "evaluation_timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "model_artifact": {
            "path": onnx_path,
            "format": "ONNX (opset 15)",
            "size_bytes": model_size_bytes,
            "runtime": "onnxruntime"
        },
        "dataset": {
            "type": "synthetic",
            "generator_function": "ml-service/train.py:generate_synthetic_data",
            "random_seed": 42,
            "total_samples": total_samples,
            "train_samples": len(X_train),
            "test_samples": len(X_test),
            "features": feature_cols,
            "feature_count": len(feature_cols),
            "fraud_count_total": total_fraud,
            "non_fraud_count_total": total_legit,
            "fraud_ratio_total": round(fraud_rate, 4),
            "test_fraud_count": test_fraud,
            "test_non_fraud_count": test_legit
        },
        "metrics": {
            "decision_threshold": 0.50,
            "roc_auc": round(roc_auc, 4),
            "pr_auc": round(pr_auc, 4),
            "precision": round(prec, 4),
            "recall": round(rec, 4),
            "f1_score": round(f1, 4),
            "true_negatives": tn,
            "false_positives": fp,
            "false_negatives": fn,
            "true_positives": tp,
            "specificity": round(tn / (tn + fp), 4) if (tn + fp) > 0 else 0.0
        },
        "latency_benchmarks_ms": {
            "p50": round(p50_ms, 3),
            "p95": round(p95_ms, 3),
            "p99": round(p99_ms, 3),
            "mean": round(mean_ms, 3),
            "sample_size": len(latencies)
        }
    }
    
    # Save baseline_metrics.json
    metrics_path = os.path.join(output_dir, "baseline_metrics.json")
    with open(metrics_path, "w") as f:
        json.dump(metrics_data, f, indent=2)
    print(f"Saved baseline metrics to: {metrics_path}")
    
    # 5. Visualizations with Matplotlib
    # A. Confusion Matrix Plot
    plt.figure(figsize=(6, 5))
    plt.imshow(cm, interpolation='nearest', cmap=plt.cm.Blues)
    plt.title('Baseline Confusion Matrix (Threshold = 0.50)')
    plt.colorbar()
    tick_marks = np.arange(2)
    plt.xticks(tick_marks, ['Legitimate (0)', 'Fraud (1)'])
    plt.yticks(tick_marks, ['Legitimate (0)', 'Fraud (1)'])
    
    thresh = cm.max() / 2.
    for i in range(cm.shape[0]):
        for j in range(cm.shape[1]):
            plt.text(j, i, format(cm[i, j], 'd'),
                     ha="center", va="center",
                     color="white" if cm[i, j] > thresh else "black")
    plt.ylabel('True Label')
    plt.xlabel('Predicted Label')
    plt.tight_layout()
    cm_path = os.path.join(output_dir, "confusion_matrix.png")
    plt.savefig(cm_path, dpi=150)
    plt.close()
    print(f"Saved confusion matrix plot: {cm_path}")
    
    # B. ROC Curve Plot
    fpr, tpr, _ = roc_curve(y_test, y_prob)
    plt.figure(figsize=(6, 5))
    plt.plot(fpr, tpr, color='darkorange', lw=2, label=f'ROC Curve (AUC = {roc_auc:.4f})')
    plt.plot([0, 1], [0, 1], color='navy', lw=1.5, linestyle='--', label='Random Chance')
    plt.xlim([0.0, 1.0])
    plt.ylim([0.0, 1.05])
    plt.xlabel('False Positive Rate')
    plt.ylabel('True Positive Rate')
    plt.title('Baseline ROC Curve')
    plt.legend(loc="lower right")
    plt.grid(alpha=0.3)
    plt.tight_layout()
    roc_path = os.path.join(output_dir, "roc_curve.png")
    plt.savefig(roc_path, dpi=150)
    plt.close()
    print(f"Saved ROC curve plot: {roc_path}")
    
    # C. Precision-Recall Curve Plot
    precisions, recalls, _ = precision_recall_curve(y_test, y_prob)
    plt.figure(figsize=(6, 5))
    plt.plot(recalls, precisions, color='purple', lw=2, label=f'PR Curve (AP = {pr_auc:.4f})')
    plt.axhline(y=test_fraud/len(y_test), color='gray', linestyle='--', label=f'Baseline ({test_fraud/len(y_test):.3f})')
    plt.xlim([0.0, 1.0])
    plt.ylim([0.0, 1.05])
    plt.xlabel('Recall')
    plt.ylabel('Precision')
    plt.title('Baseline Precision-Recall Curve')
    plt.legend(loc="upper right")
    plt.grid(alpha=0.3)
    plt.tight_layout()
    pr_path = os.path.join(output_dir, "precision_recall_curve.png")
    plt.savefig(pr_path, dpi=150)
    plt.close()
    print(f"Saved Precision-Recall curve plot: {pr_path}")
    
    return metrics_data

if __name__ == "__main__":
    current_dir = os.path.dirname(os.path.abspath(__file__))
    ml_root = os.path.dirname(current_dir)
    onnx_file = os.path.join(ml_root, "model", "fraud_model.onnx")
    out_dir = current_dir
    run_evaluation(onnx_path=onnx_file, output_dir=out_dir)
