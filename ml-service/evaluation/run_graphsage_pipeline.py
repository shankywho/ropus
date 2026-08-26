#!/usr/bin/env python3
"""
Master Execution Pipeline: GraphSAGE Synthetic Training, Point-in-Time Evaluation & Multi-Defense Ablation
"""

import os
import sys
import json
import hashlib
from datetime import datetime, timezone

# Add ml-service to sys.path
CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.graph_dataset import TemporalHeteroGraphDataset
from graphsage.model import GraphSAGEModel
from graphsage.train import GraphSAGETrainer, compute_file_sha256
from graphsage.evaluate import evaluate_graphsage_test_split

PROD_CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
PROD_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"


def main():
    print("=" * 80)
    print("ROPUS GRAPHSAGE OFFLINE TRAINING & MULTI-DEFENSE EVALUATION PIPELINE")
    print("=" * 80)

    # 1. Pre-execution production champion verification
    current_sha256 = compute_file_sha256(PROD_CHAMPION_PATH)
    print(f"[Integrity Check] Production Champion: {PROD_CHAMPION_PATH}")
    print(f"[Integrity Check] Expected SHA-256:    {PROD_EXPECTED_SHA256}")
    print(f"[Integrity Check] Computed SHA-256:    {current_sha256}")
    assert current_sha256 == PROD_EXPECTED_SHA256, (
        f"CRITICAL ERROR: Production champion modified prior to execution!\n"
        f"Expected: {PROD_EXPECTED_SHA256}\nActual: {current_sha256}"
    )

    data_dir = os.path.abspath(os.path.join(ML_SERVICE_DIR, "..", "synthetic_ropus", "data"))
    output_model_dir = os.path.join(ML_SERVICE_DIR, "model", "graphsage")
    output_eval_dir = CURRENT_DIR

    # 2. Load Dataset
    print(f"\n[Pipeline Step 1] Loading Point-in-Time Heterogeneous Dataset from {data_dir}...")
    dataset = TemporalHeteroGraphDataset.load_from_directory(data_dir)

    # 3. Train GraphSAGE on Chronological Train Split
    print(f"\n[Pipeline Step 2] Training 2-Layer GraphSAGE on 70% Chronological Train Split...")
    trainer = GraphSAGETrainer(dataset=dataset, output_dir=output_model_dir)
    train_metadata = trainer.train_on_synthetic_data(
        data_dir=data_dir,
        epochs=3,
        batch_size=64,
        lr=0.005,
        sample_sizes=(10, 5)
    )

    # 4. Evaluate on Untouched 15% Test Split
    print(f"\n[Pipeline Step 3] Evaluating GraphSAGE on Untouched 15% Test Split & Slices...")
    eval_report = evaluate_graphsage_test_split(
        dataset=dataset,
        model=trainer.model,
        data_dir=data_dir
    )

    # 5. Persist Evaluation Results
    eval_output_path = os.path.join(output_eval_dir, "graphsage_offline_evaluation_report.json")
    with open(eval_output_path, "w") as f:
        json.dump(eval_report, f, indent=2)
    print(f"\n[Pipeline Step 4] Saved Evaluation Report: {eval_output_path}")

    # 6. Post-execution production champion verification
    post_sha256 = compute_file_sha256(PROD_CHAMPION_PATH)
    print(f"[Integrity Check] Post-Execution Production SHA-256: {post_sha256}")
    assert post_sha256 == PROD_EXPECTED_SHA256, "CRITICAL: Production model altered during execution!"

    print("\n" + "=" * 80)
    print("PIPELINE EXECUTION COMPLETED SUCCESSFULLY")
    print("=" * 80)


if __name__ == "__main__":
    main()
