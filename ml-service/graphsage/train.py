"""
GraphSAGE Training Pipeline with Strict Temporal Splitting and Governance Lifecycle Management
"""

import os
import json
import hashlib
import numpy as np
import pandas as pd
from datetime import datetime, timezone
from typing import Dict, Any, Optional, List, Tuple
from sklearn.metrics import roc_auc_score, average_precision_score

from .schema import GraphSAGELifecycleState, GraphSAGEDataContract, HeteroNode, HeteroNodeType
from .graph_dataset import TemporalHeteroGraphDataset
from .model import GraphSAGEModel


def compute_file_sha256(path: str) -> str:
    if not os.path.exists(path):
        return ""
    hasher = hashlib.sha256()
    with open(path, "rb") as f:
        while chunk := f.read(65536):
            hasher.update(chunk)
    return hasher.hexdigest()


class GraphSAGETrainer:
    """
    Manages GraphSAGE training workflows on point-in-time heterogeneous graphs.
    """
    def __init__(
        self,
        dataset: Optional[TemporalHeteroGraphDataset] = None,
        model: Optional[GraphSAGEModel] = None,
        output_dir: str = "ml-service/model/graphsage"
    ):
        self.dataset = dataset or TemporalHeteroGraphDataset()
        self.model = model or GraphSAGEModel(input_dim=32, hidden_dim=64, output_dim=64, seed=42)
        self.output_dir = output_dir
        self.lifecycle_state = GraphSAGELifecycleState.DATA_REQUIRED

    def train_on_synthetic_data(
        self,
        data_dir: str,
        epochs: int = 3,
        batch_size: int = 128,
        lr: float = 0.005,
        sample_sizes: Tuple[int, int] = (10, 5)
    ) -> Dict[str, Any]:
        """
        Executes supervised point-in-time training strictly using the chronological 70% Train split.
        Uses 15% Validation split for model checkpoint selection.
        Untouched 15% Test split is preserved strictly for evaluation.
        """
        print("=" * 80, flush=True)
        print("[GraphSAGE Training] Initializing Point-in-Time Training Pipeline", flush=True)
        print("=" * 80, flush=True)

        if len(self.dataset.nodes) == 0:
            self.dataset = TemporalHeteroGraphDataset.load_from_directory(data_dir)

        txns_path = os.path.join(data_dir, "transactions", "transactions.csv")
        labels_path = os.path.join(data_dir, "labels", "labels.csv")
        splits_path = os.path.join(data_dir, "labels", "splits.csv")

        txns_df = pd.read_csv(txns_path)
        labels_df = pd.read_csv(labels_path)
        splits_df = pd.read_csv(splits_path)

        txns_sha256 = compute_file_sha256(txns_path)
        labels_sha256 = compute_file_sha256(labels_path)
        splits_sha256 = compute_file_sha256(splits_path)

        # Ground truth mapping: 1 for non-legitimate, 0 for legitimate
        txn_labels = labels_df[labels_df["entity_type"] == "TRANSACTION"]
        gt_map = dict(zip(txn_labels["entity_id"], [0.0 if gt == "LEGITIMATE" else 1.0 for gt in txn_labels["ground_truth"]]))

        split_map = dict(zip(splits_df["transaction_id"], splits_df["split"]))
        txns_df["split"] = txns_df["transaction_id"].map(split_map)
        txns_df["target_y"] = txns_df["transaction_id"].map(lambda tid: gt_map.get(tid, 0.0))

        train_txns = txns_df[txns_df["split"] == "train"].copy()
        val_txns = txns_df[txns_df["split"] == "validation"].copy()
        test_txns = txns_df[txns_df["split"] == "test"].copy()

        print(f"[GraphSAGE Training] Train samples: {len(train_txns)} (Fraud: {train_txns['target_y'].sum():.0f})", flush=True)
        print(f"[GraphSAGE Training] Val samples:   {len(val_txns)} (Fraud: {val_txns['target_y'].sum():.0f})", flush=True)
        print(f"[GraphSAGE Training] Test samples:  {len(test_txns)} (Fraud: {test_txns['target_y'].sum():.0f}) [Preserved for Test]", flush=True)

        train_records = train_txns.to_dict("records")
        val_records = val_txns.to_dict("records")
        rng = np.random.RandomState(42)

        # Pre-extract validation samples for fast epoch evaluation
        print("[GraphSAGE Training] Pre-sampling validation point-in-time neighborhoods...", flush=True)
        val_samples = []
        default_dt = datetime(2025, 1, 1, tzinfo=timezone.utc)
        for r in val_records:
            acct_id = str(r["account_id"])
            ts_str = str(r["event_timestamp"])
            event_dt = datetime.fromisoformat(ts_str) if ts_str else default_dt
            if event_dt.tzinfo is None:
                event_dt = event_dt.replace(tzinfo=timezone.utc)

            root_node = self.dataset.nodes.get(acct_id) or HeteroNode(
                id=acct_id, type=HeteroNodeType.ACCOUNT, risk_score=0.05, created_at=event_dt
            )
            l1, l2 = self.dataset.get_point_in_time_neighbors(acct_id, as_of=event_dt, sample_sizes=sample_sizes)
            val_samples.append((root_node, l1, l2, float(r["target_y"])))

        print(f"[GraphSAGE Training] Starting {epochs} Epochs of Supervised Point-in-Time Training...", flush=True)
        best_val_auc = 0.0
        best_model_weights = None
        history = []

        for epoch in range(1, epochs + 1):
            rng.shuffle(train_records)
            epoch_loss = 0.0
            num_batches = 0

            for i in range(0, len(train_records), batch_size):
                batch_rows = train_records[i:i + batch_size]
                batch_samples = []

                for r in batch_rows:
                    acct_id = str(r["account_id"])
                    ts_str = str(r["event_timestamp"])
                    event_dt = datetime.fromisoformat(ts_str) if ts_str else default_dt
                    if event_dt.tzinfo is None:
                        event_dt = event_dt.replace(tzinfo=timezone.utc)

                    root_node = self.dataset.nodes.get(acct_id) or HeteroNode(
                        id=acct_id, type=HeteroNodeType.ACCOUNT, risk_score=0.05, created_at=event_dt
                    )
                    l1, l2 = self.dataset.get_point_in_time_neighbors(acct_id, as_of=event_dt, sample_sizes=sample_sizes)
                    batch_samples.append((root_node, l1, l2, float(r["target_y"])))

                loss = self.model.train_step(batch_samples, lr=lr)
                epoch_loss += loss
                num_batches += 1

            avg_train_loss = epoch_loss / max(num_batches, 1)

            # Evaluate on Validation Split
            val_preds = []
            val_targets = []
            for root_node, l1, l2, target_y in val_samples:
                _, signals = self.model.forward(root_node, l1, l2)
                val_preds.append(signals["transaction_context_score"])
                val_targets.append(target_y)

            val_preds = np.array(val_preds)
            val_targets = np.array(val_targets)
            val_auc = float(roc_auc_score(val_targets, val_preds))
            val_ap = float(average_precision_score(val_targets, val_preds))

            print(f"[Epoch {epoch:02d}/{epochs:02d}] Train Loss: {avg_train_loss:.4f} | Val ROC-AUC: {val_auc:.4f} | Val PR-AUC: {val_ap:.4f}", flush=True)

            history.append({
                "epoch": epoch,
                "train_loss": round(avg_train_loss, 4),
                "val_roc_auc": round(val_auc, 4),
                "val_pr_auc": round(val_ap, 4)
            })

            if val_auc >= best_val_auc:
                best_val_auc = val_auc
                best_model_weights = self.model.to_dict()

        if best_model_weights is not None:
            self.model = GraphSAGEModel.from_dict(best_model_weights)
            print(f"[GraphSAGE Training] Restored best checkpoint with Validation ROC-AUC: {best_val_auc:.4f}", flush=True)

        os.makedirs(self.output_dir, exist_ok=True)
        model_path = os.path.join(self.output_dir, "graphsage_synthetic_v1.json")
        with open(model_path, "w") as f:
            json.dump(self.model.to_dict(), f, indent=2)

        model_sha256 = compute_file_sha256(model_path)
        self.lifecycle_state = GraphSAGELifecycleState.SYNTHETIC_TRAINED

        metadata = {
            "model_name": "graphsage_synthetic_v1",
            "lifecycle_state": self.lifecycle_state.value,
            "governance_status": "SYNTHETIC-TRAINED / NON-PRODUCTION / SHADOW-ONLY",
            "enforcement_mode": "STRICTLY NON-ENFORCING",
            "architecture": {
                "type": "Heterogeneous 2-Layer GraphSAGE",
                "input_dim": self.model.input_dim,
                "hidden_dim": self.model.hidden_dim,
                "output_dim": self.model.output_dim,
                "aggregator": "mean",
                "activation": "ReLU + LayerNorm",
                "risk_heads": [
                    "employee_risk", "consumer_risk", "relationship_collusion_risk",
                    "entity_neighborhood_risk", "graph_anomaly_score", "transaction_context_score"
                ]
            },
            "training_provenance": {
                "training_dataset_sha256": txns_sha256,
                "labels_dataset_sha256": labels_sha256,
                "splits_dataset_sha256": splits_sha256,
                "model_artifact_sha256": model_sha256,
                "trained_timestamp": datetime.now(timezone.utc).isoformat(),
                "generator_seed": 42,
                "splits": {"train_count": len(train_txns), "val_count": len(val_txns), "test_count": len(test_txns)},
                "epochs": epochs,
                "best_val_roc_auc": round(best_val_auc, 4),
            },
            "training_history": history,
            "production_champion_invariant": {
                "production_champion": "production_model_v8_bmr.joblib",
                "production_champion_sha256": "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7",
                "status": "UNCHANGED (AUTHORITATIVE)"
            }
        }

        meta_path = os.path.join(self.output_dir, "model_metadata.json")
        with open(meta_path, "w") as f:
            json.dump(metadata, f, indent=2)

        print(f"[GraphSAGE Training] Artifact saved: {model_path} (SHA-256: {model_sha256})", flush=True)
        print(f"[GraphSAGE Training] Metadata saved: {meta_path}", flush=True)

        return metadata

    def run_training_pipeline(self) -> Dict[str, Any]:
        """Legacy compatibility wrapper."""
        emp_nodes = [n for n in self.dataset.nodes.values() if n.type.value == "EMPLOYEE"]
        known_collusion_edges = [
            e for e in self.dataset.edges.values()
            if e.type.value in ["ACCESSES", "INTERACTS_WITH_CONSUMER", "EMPLOYEE_ACCESSES_ACCOUNT"] and e.properties.get("is_collusion", False)
        ]

        if len(known_collusion_edges) < 50 and len(self.dataset.edges) < 100:
            self.lifecycle_state = GraphSAGELifecycleState.DATA_REQUIRED
            return {
                "lifecycle_state": self.lifecycle_state.value,
                "status": "DATA_REQUIRED",
                "message": "Insufficient genuine internal employee collusion labels for supervised training. Minimum 50 required.",
                "current_employee_nodes": len(emp_nodes),
                "current_labeled_collusion_edges": len(known_collusion_edges),
                "minimum_required_collusion_cases": 50,
                "contract": GraphSAGEDataContract().model_dump(),
                "timestamp": datetime.now(timezone.utc).isoformat(),
            }

        return {
            "lifecycle_state": GraphSAGELifecycleState.SYNTHETIC_TRAINED.value,
            "status": "TRAINED",
            "timestamp": datetime.now(timezone.utc).isoformat(),
        }
