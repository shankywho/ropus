# ML Model Evaluation Artifacts

This directory contains offline evaluation plots, metric summaries, and calibration benchmarks for the ROPUS fraud detection models.

## Key Evaluation Artifacts

### 1. Model Performance & Classification Metrics
- **`roc_curve.png` / `precision_recall_curve.png`**: ROC and Precision-Recall curves evaluating the XGBoost classification baseline across temporal train/validation/test splits.
- **`confusion_matrix.png`**: Confusion matrix demonstrating True Positives, False Positives, True Negatives, and False Negatives on the held-out test split.
- **`threshold_analysis.png` / `threshold_analysis.csv`**: Performance metrics (Precision, Recall, F1, FPR) across decision thresholds from 0.05 to 0.95.

### 2. Probability Calibration
- **`calibration_curves.png`**: Reliability diagram comparing raw model probability estimates vs. calibrated posterior probabilities (Beta / Isotonic calibration).
- **`calibration_metrics.json` / `calibration_comparison.csv`**: Expected Calibration Error (ECE) and Brier scores comparing uncalibrated vs. calibrated probability outputs.

### 3. Bayes Minimum Risk (BMR) Cost Analysis
- **`cost_threshold_analysis.png` / `cost_threshold_analysis.csv`**: Financial loss curves evaluating total operating cost (False Positive review cost vs. False Negative fraud loss) across decision thresholds.
- **`review_capacity_analysis.csv`**: Analyst queue volume and expected daily review cost under different operational capacity limits.
- **`bmr_model_comparison.json`**: Comparison of classification models under cost-optimal Bayes Minimum Risk thresholding.

### 4. Offline Evaluators
- **`evaluate_frozen_holdout.py`**: Offline script to compute holdout test split metrics (ROC-AUC, PR-AUC, ECE, Brier score) against the IEEE-CIS fraud dataset fixture without modifying model weights.
- **`shadow_evaluator.py`**: Helper module to evaluate shadow predictions against historical transaction outcomes.
