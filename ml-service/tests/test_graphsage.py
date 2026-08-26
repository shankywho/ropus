import os
import sys
import unittest
import numpy as np
from datetime import datetime, timezone, timedelta

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))

from graphsage.schema import HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType, GraphSAGELifecycleState
from graphsage.graph_dataset import TemporalHeteroGraphDataset
from graphsage.model import GraphSAGEModel
from graphsage.train import GraphSAGETrainer
from graphsage.evaluate import evaluate_graphsage_defense_suite


class TestGraphSAGEPythonSuite(unittest.TestCase):

    def setUp(self):
        self.dataset = TemporalHeteroGraphDataset()
        self.now = datetime.now(timezone.utc)

    def test_heterogeneous_graph_insertion(self):
        node = HeteroNode(
            id="emp_001",
            type=HeteroNodeType.EMPLOYEE,
            risk_score=0.05,
            created_at=self.now
        )
        self.dataset.add_node(node)
        self.assertIn("emp_001", self.dataset.nodes)
        self.assertEqual(self.dataset.nodes["emp_001"].type, HeteroNodeType.EMPLOYEE)

    def test_temporal_leakage_prevention(self):
        t0 = datetime(2026, 8, 1, 12, 0, 0, tzinfo=timezone.utc)
        t_past = t0 - timedelta(hours=1)
        t_future = t0 + timedelta(hours=1)

        self.dataset.add_node(HeteroNode(id="emp_root", type=HeteroNodeType.EMPLOYEE, created_at=t_past))
        self.dataset.add_node(HeteroNode(id="usr_past", type=HeteroNodeType.CONSUMER, created_at=t_past))
        self.dataset.add_node(HeteroNode(id="usr_future", type=HeteroNodeType.CONSUMER, created_at=t_past))

        self.dataset.add_edge(HeteroEdge(
            id="e_past", source_id="emp_root", target_id="usr_past",
            type=HeteroEdgeType.ACCESSES, timestamp=t_past
        ))
        self.dataset.add_edge(HeteroEdge(
            id="e_future", source_id="emp_root", target_id="usr_future",
            type=HeteroEdgeType.ACCESSES, timestamp=t_future
        ))

        l1, l2 = self.dataset.get_point_in_time_neighbors("emp_root", as_of=t0)
        l1_ids = [n.id for n in l1]

        self.assertIn("usr_past", l1_ids)
        self.assertNotIn("usr_future", l1_ids, "Future node MUST NOT leak into past evaluation")

    def test_graphsage_forward_pass(self):
        model = GraphSAGEModel(input_dim=32, hidden_dim=64, output_dim=64, seed=42)
        root = HeteroNode(id="usr_alice", type=HeteroNodeType.CONSUMER, risk_score=0.10, created_at=self.now)
        l1 = [HeteroNode(id="acc_1", type=HeteroNodeType.ACCOUNT, risk_score=0.05, created_at=self.now)]
        l2 = [HeteroNode(id="dev_1", type=HeteroNodeType.DEVICE, risk_score=0.05, created_at=self.now)]

        emb, signals = model.forward(root, l1, l2)
        self.assertEqual(emb.shape, (64,))
        self.assertIn("relationship_collusion_risk", signals)
        self.assertIn("entity_neighborhood_risk", signals)
        self.assertLess(signals["relationship_collusion_risk"], 0.50)

    def test_training_pipeline_data_required_state(self):
        trainer = GraphSAGETrainer(self.dataset)
        report = trainer.run_training_pipeline()

        self.assertEqual(report["lifecycle_state"], GraphSAGELifecycleState.DATA_REQUIRED.value)
        self.assertEqual(report["status"], "DATA_REQUIRED")
        self.assertIn("minimum_required_collusion_cases", report)

    def test_defense_suite_evaluation(self):
        np.random.seed(42)
        n = 500
        labels = np.zeros(n, dtype=int)
        labels[:25] = 1 # 5% fraud rate

        base_scores = np.random.uniform(0.01, 0.08, size=n)
        base_scores[:25] = np.random.uniform(0.20, 0.80, size=25)

        rule_scores = np.random.uniform(0.01, 0.05, size=n)
        rule_scores[:25] = np.random.uniform(0.15, 0.70, size=25)

        graphsage_scores = np.random.uniform(0.01, 0.05, size=n)
        graphsage_scores[:25] = np.random.uniform(0.25, 0.85, size=25)

        amounts = np.random.uniform(10.0, 500.0, size=n)

        results = evaluate_graphsage_defense_suite(
            holdout_labels=labels,
            baseline_scores=base_scores,
            graph_rule_scores=rule_scores,
            graphsage_scores=graphsage_scores,
            amounts=amounts
        )

        self.assertIn("configurations", results)
        self.assertEqual(len(results["configurations"]), 5)
        self.assertEqual(results["champion_status"], "v8.0-bmr-36f UNCHANGED (AUTHORITATIVE)")


if __name__ == "__main__":
    unittest.main()
