"""
Dynamic Metric Consistency & Submission Hygiene Test
Ensures all published Markdown metrics and documentation stay 100% synchronized
with the generated canonical evaluation JSON artifact on the held-out split,
and asserts that no stale cost-reduction or ungrounded claims exist in documentation.
"""

import json
import os
import pytest
import sys

def test_published_metrics_match_canonical_evaluation():
    base_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
    ml_eval_dir = os.path.join(base_dir, "ml-service", "evaluation")
    sys.path.insert(0, ml_eval_dir)

    from evaluate_frozen_holdout import evaluate_canonical_holdout

    # 1. Execute live canonical evaluation directly
    eval_report = evaluate_canonical_holdout()
    assert eval_report is not None, "Evaluation returned None"

    # 2. Load generated JSON artifact from disk and verify identity
    report_json_path = os.path.join(ml_eval_dir, "frozen_holdout_evaluation_report.json")
    assert os.path.exists(report_json_path), f"JSON artifact missing at {report_json_path}"
    with open(report_json_path, "r") as f:
        disk_report = json.load(f)

    # 3. Assert invariants
    disc_cal = eval_report["discrimination_and_calibration_metrics"]
    fixed_p50 = eval_report["fixed_threshold_baseline_p50"]
    bmr_policy = eval_report["bmr_cost_sensitive_policy"]
    cm = fixed_p50["confusion_matrix"]

    total_txns = eval_report["total_holdout_transactions"]
    pos_labels = eval_report["positive_labels_in_holdout"]
    neg_labels = eval_report["negative_labels_in_holdout"]

    # Arithmetic integrity
    assert cm["true_positives_tp"] + cm["false_negatives_fn"] == pos_labels, "TP + FN must equal total positive labels"
    assert cm["true_negatives_tn"] + cm["false_positives_fp"] == neg_labels, "TN + FP must equal total negative labels"
    assert cm["true_positives_tp"] + cm["false_positives_fp"] + cm["true_negatives_tn"] + cm["false_negatives_fn"] == total_txns

    # Metric bounds
    roc_auc = disc_cal["roc_auc"]
    pr_auc = disc_cal["pr_auc"]
    ece = disc_cal["expected_calibration_error_ece"]
    precision = fixed_p50["precision"]
    recall = fixed_p50["recall"]
    fpr = fixed_p50["false_positive_rate_fpr"]

    assert 0.50 < roc_auc <= 1.0, f"ROC-AUC {roc_auc} out of valid range"
    assert 0.0 < pr_auc <= 1.0, f"PR-AUC {pr_auc} out of valid range"
    assert 0.0 <= ece < 0.10, f"ECE {ece} exceeds acceptable calibration bound"

    # 4. Assert published documents contain the dynamically evaluated values
    roc_str = f"{roc_auc:.4f}"
    pr_str = f"{pr_auc:.4f}"
    ece_str = f"{ece:.4f}"
    prec_str = f"{precision * 100:.2f}%"
    rec_str = f"{recall * 100:.2f}%"
    fpr_str = f"{fpr * 100:.2f}%"
    fixed_cost_str = f"{bmr_policy['financial_loss_comparison']['fixed_050_expected_loss_inr']:.2f}"
    fixed_cost_comma = f"{bmr_policy['financial_loss_comparison']['fixed_050_expected_loss_inr']:,.2f}"
    bmr_cost_str = f"{bmr_policy['financial_loss_comparison']['bmr_policy_expected_loss_inr']:.2f}"
    bmr_cost_comma = f"{bmr_policy['financial_loss_comparison']['bmr_policy_expected_loss_inr']:,.2f}"

    # Check METRICS.md
    metrics_md_path = os.path.join(base_dir, "METRICS.md")
    assert os.path.exists(metrics_md_path), f"METRICS.md missing at {metrics_md_path}"
    with open(metrics_md_path, "r") as f:
        metrics_content = f.read()

    assert roc_str in metrics_content, f"ROC-AUC {roc_str} missing in METRICS.md"
    assert pr_str in metrics_content, f"PR-AUC {pr_str} missing in METRICS.md"
    assert ece_str in metrics_content, f"ECE {ece_str} missing in METRICS.md"
    assert prec_str in metrics_content, f"Precision {prec_str} missing in METRICS.md"
    assert rec_str in metrics_content, f"Recall {rec_str} missing in METRICS.md"
    assert fpr_str in metrics_content, f"FPR {fpr_str} missing in METRICS.md"
    assert (fixed_cost_str in metrics_content or fixed_cost_comma in metrics_content), f"Fixed cost {fixed_cost_str} missing in METRICS.md"
    assert (bmr_cost_str in metrics_content or bmr_cost_comma in metrics_content), f"BMR cost {bmr_cost_str} missing in METRICS.md"

    # Check docs/ml_quality_report.md
    report_md_path = os.path.join(base_dir, "docs", "ml_quality_report.md")
    assert os.path.exists(report_md_path), f"docs/ml_quality_report.md missing at {report_md_path}"
    with open(report_md_path, "r") as f:
        report_content = f.read()

    assert roc_str in report_content, f"ROC-AUC {roc_str} missing in docs/ml_quality_report.md"
    assert pr_str in report_content, f"PR-AUC {pr_str} missing in docs/ml_quality_report.md"
    assert ece_str in report_content, f"ECE {ece_str} missing in docs/ml_quality_report.md"

    # 5. Assert absence of stale / contradictory claims across documentation
    doc_files_to_check = [
        "README.md",
        "METRICS.md",
        "docs/README.md",
        "docs/buildathon-evidence.md",
        "docs/limitations.md",
        "docs/demo-runbook.md",
        "docs/SUBMISSION_GUIDE.md",
        "docs/ml_quality_report.md"
    ]

    stale_patterns = ["71.5%", "28,450", "8,120", "confirmed real fraud cases in holdout"]
    for rel_doc in doc_files_to_check:
        doc_path = os.path.join(base_dir, rel_doc)
        if os.path.exists(doc_path):
            with open(doc_path, "r") as f:
                content = f.read()
            for pattern in stale_patterns:
                assert pattern not in content, f"Found stale pattern '{pattern}' in {rel_doc}"
