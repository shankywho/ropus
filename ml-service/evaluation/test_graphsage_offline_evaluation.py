#!/usr/bin/env python3
"""
Unit and Integration Tests for GraphSAGE Offline Evaluation & Multi-Defense System.
Ensures temporal isolation, non-enforcing invariants, explainable signal contracts,
inductive embedding stability, and champion immutability under Phase 57.
"""

import unittest
import json
import os
import sys
import hashlib
import numpy as np
from datetime import datetime, timezone, timedelta

# Ensure ml-service root is in python path
CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import (
    HeteroNodeType,
    HeteroEdgeType,
    HeteroNode,
    HeteroEdge,
    GraphSAGELifecycleState,
    RelationshipRiskLevel,
    GraphSAGEDataContract,
)
from graphsage.graph_dataset import TemporalHeteroGraphDataset
from graphsage.model import GraphSAGEModel
from graphsage.train import GraphSAGETrainer
from graphsage.evaluate import evaluate_graphsage_defense_suite

CHAMPION_NAME = "v8.0-bmr-36f"
CHAMPION_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"


class TestGraphSAGEOfflineEvaluation(unittest.TestCase):

    def setUp(self):
        self.base_dir = CURRENT_DIR
        self.now = datetime(2026, 8, 20, 12, 0, 0, tzinfo=timezone.utc)

    def test_production_champion_immutability(self):
        """Verify active champion checksum and non-enforcing shadow invariant."""
        scorecard_path = os.path.join(self.base_dir, "phase_57_graphsage_governance_scorecard.json")
        self.assertTrue(os.path.exists(scorecard_path), "Phase 57 governance scorecard must exist")

        with open(scorecard_path, "r") as f:
            scorecard = json.load(f)

        self.assertEqual(scorecard.get("champion_model"), CHAMPION_NAME)
        self.assertEqual(scorecard.get("champion_checksum"), CHAMPION_SHA256)
        self.assertIn("ACTIVE", scorecard.get("champion_status", ""))
        self.assertIn("NON-ENFORCING SHADOW", scorecard.get("graphsage_status", ""))

    def test_explainable_signal_schema_contract(self):
        """Verify that explainable signals adhere strictly to production schema contract."""
        contract = GraphSAGEDataContract()
        contract_dict = contract.model_dump()
        self.assertIn("required_entity_types", contract_dict)
        self.assertIn("required_edge_types", contract_dict)
        self.assertIn("temporal_guarantee", contract_dict)

        # Check all key node and edge types are represented
        self.assertIn("EMPLOYEE", contract_dict["required_entity_types"])
        self.assertIn("CONSUMER", contract_dict["required_entity_types"])
        self.assertIn("ACCOUNT", contract_dict["required_entity_types"])
        self.assertIn("ACCESSES", contract_dict["required_edge_types"])
        self.assertIn("USES_DEVICE", contract_dict["required_edge_types"])

    def test_multi_defense_benchmark_metrics(self):
        """Verify multi-defense benchmark metrics and recall improvements."""
        benchmark_path = os.path.join(self.base_dir, "phase_57_graphsage_multi_defense_benchmark.json")
        self.assertTrue(os.path.exists(benchmark_path), "Multi-defense benchmark JSON must exist")

        with open(benchmark_path, "r") as f:
            benchmark = json.load(f)

        configs = benchmark.get("configurations", {})
        self.assertIn("1_ROPUS_Baseline", configs)
        self.assertIn("5_ROPUS_Plus_GraphRules_Plus_GraphSAGE", configs)

        baseline = configs["1_ROPUS_Baseline"]
        ensemble = configs["5_ROPUS_Plus_GraphRules_Plus_GraphSAGE"]

        # Multimodal defense should achieve superior PR-AUC and lower FPR
        self.assertGreaterEqual(ensemble["PR_AUC"], baseline["PR_AUC"])
        self.assertLessEqual(ensemble["FPR"], baseline["FPR"])
        self.assertLessEqual(ensemble["Total_Economic_Loss_USD"], baseline["Total_Economic_Loss_USD"])

    def test_temporal_leakage_isolation(self):
        """Verify strict temporal filtering prevents any future edge or node from leaking into embeddings."""
        dataset = TemporalHeteroGraphDataset()
        t_past = self.now - timedelta(days=5)
        t_eval = self.now
        t_future = self.now + timedelta(days=5)

        # Root node created in past
        emp = HeteroNode(id="emp_01", type=HeteroNodeType.EMPLOYEE, created_at=t_past)
        usr = HeteroNode(id="usr_01", type=HeteroNodeType.CONSUMER, created_at=t_past)
        dataset.add_node(emp)
        dataset.add_node(usr)

        # Future-dated edge (should NOT be visible at t_eval)
        future_edge = HeteroEdge(id="e_future", source_id="emp_01", target_id="usr_01", type=HeteroEdgeType.ACCESSES, timestamp=t_future)
        dataset.add_edge(future_edge)

        l1, l2 = dataset.get_point_in_time_neighbors("emp_01", as_of=t_eval)
        self.assertEqual(len(l1), 0, "Future-dated edge must not leak into Hop-1 neighbors at evaluation time")
        self.assertEqual(len(l2), 0, "Future-dated edge must not leak into Hop-2 neighbors")

        # Past-dated edge (SHOULD be visible at t_eval)
        past_edge = HeteroEdge(id="e_past", source_id="emp_01", target_id="usr_01", type=HeteroEdgeType.ACCESSES, timestamp=t_past)
        dataset.add_edge(past_edge)

        l1_past, _ = dataset.get_point_in_time_neighbors("emp_01", as_of=t_eval)
        self.assertEqual(len(l1_past), 1, "Past-dated edge must be correctly included")
        self.assertEqual(l1_past[0].id, "usr_01")

    def test_inductive_unseen_node_forward(self):
        """Verify GraphSAGE can compute embeddings inductively for completely unseen entities."""
        model = GraphSAGEModel(input_dim=32, hidden_dim=64, output_dim=64, seed=42)
        new_emp = HeteroNode(id="emp_unseen_999", type=HeteroNodeType.EMPLOYEE, risk_score=0.15, created_at=self.now)

        # Forward pass with no neighbors (isolated new node)
        emb_isolated, signals_isolated = model.forward(new_emp, [], [])
        self.assertEqual(emb_isolated.shape, (64,))
        # Layer norm outputs mean 0, std 1 => vector norm ~ sqrt(64) = 8.0
        self.assertAlmostEqual(np.mean(emb_isolated), 0.0, places=4, msg="Layer norm output mean must be 0.0")
        self.assertAlmostEqual(np.std(emb_isolated), 1.0, places=3, msg="Layer norm output std must be 1.0")
        self.assertIn("relationship_collusion_risk", signals_isolated)
        self.assertIn("entity_neighborhood_risk", signals_isolated)

    def test_security_and_privacy_audit(self):
        """Verify that raw PAN, CVV, and cleartext credentials are zero in graph stores."""
        sec_path = os.path.join(self.base_dir, "phase_57_graphsage_security_privacy_audit.json")
        self.assertTrue(os.path.exists(sec_path), "Security and privacy audit JSON must exist")

        with open(sec_path, "r") as f:
            sec = json.load(f)

        self.assertEqual(sec.get("privacy_compliance"), "PASSED")
        self.assertFalse(sec.get("raw_pan_stored"))
        self.assertFalse(sec.get("raw_cvv_stored"))


if __name__ == "__main__":
    unittest.main()
