"""
Automated Test Suite: GraphSAGE Strict Temporal Point-in-Time Isolation & Adversarial Leakage Fixtures
"""

import os
import sys
import unittest
import numpy as np
from datetime import datetime, timezone, timedelta

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType
from graphsage.graph_dataset import TemporalHeteroGraphDataset
from graphsage.model import GraphSAGEModel


class TestGraphSAGETemporalLeakage(unittest.TestCase):

    def setUp(self):
        self.dataset = TemporalHeteroGraphDataset()
        self.model = GraphSAGEModel(input_dim=32, hidden_dim=64, output_dim=64, seed=42)
        self.t_eval = datetime(2025, 6, 15, 12, 0, 0, tzinfo=timezone.utc)

    def test_point_in_time_edge_isolation(self):
        """
        Verifies that only edges with timestamp <= t_eval are visible in neighborhood aggregation.
        """
        t_past = self.t_eval - timedelta(days=1)
        t_future = self.t_eval + timedelta(hours=1)

        self.dataset.add_node(HeteroNode(id="emp_01", type=HeteroNodeType.EMPLOYEE, created_at=t_past))
        self.dataset.add_node(HeteroNode(id="acct_01", type=HeteroNodeType.ACCOUNT, created_at=t_past))
        self.dataset.add_node(HeteroNode(id="acct_02", type=HeteroNodeType.ACCOUNT, created_at=t_past))

        # Past edge (SHOULD be visible)
        self.dataset.add_edge(HeteroEdge(
            id="e_past", source_id="emp_01", target_id="acct_01",
            type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT, timestamp=t_past
        ))
        # Future edge (MUST NOT be visible)
        self.dataset.add_edge(HeteroEdge(
            id="e_future", source_id="emp_01", target_id="acct_02",
            type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT, timestamp=t_future
        ))
        self.dataset.finalize_indices()

        l1, l2 = self.dataset.get_point_in_time_neighbors("emp_01", as_of=self.t_eval)
        l1_ids = [n.id for n in l1]

        self.assertIn("acct_01", l1_ids)
        self.assertNotIn("acct_02", l1_ids, "Future-dated edge MUST NOT leak into point-in-time neighborhood")

    def test_adversarial_temporal_leakage_fixture_fails_without_protection(self):
        """
        Adversarial Fixture: Proves that if an unconstrained neighborhood sampler is used without
        temporal filtering, future high-risk connections cause embedding drift, proving the
        temporal filter is strictly necessary and effective.
        """
        t_past = self.t_eval - timedelta(days=2)
        t_future_attack = self.t_eval + timedelta(days=1)

        ds_protected = TemporalHeteroGraphDataset()

        # Innocent node in past
        emp_node = HeteroNode(id="emp_innocent", type=HeteroNodeType.EMPLOYEE, created_at=t_past)
        mule_node = HeteroNode(id="acct_mule", type=HeteroNodeType.ACCOUNT, risk_score=0.90, is_known_bad=True, created_at=t_past)

        ds_protected.add_node(emp_node)
        ds_protected.add_node(mule_node)

        # Future attack edge
        ds_protected.add_edge(HeteroEdge(
            id="e_future_collusion",
            source_id="emp_innocent",
            target_id="acct_mule",
            type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
            timestamp=t_future_attack
        ))
        ds_protected.finalize_indices()

        # 1. With Point-in-Time Filter at t_eval: Future edge is invisible
        l1_prot, l2_prot = ds_protected.get_point_in_time_neighbors("emp_innocent", as_of=self.t_eval)
        emb_prot, sig_prot = self.model.forward(emp_node, l1_prot, l2_prot)

        self.assertEqual(len(l1_prot), 0)
        self.assertLess(sig_prot["relationship_collusion_risk"], 0.20)

        # 2. Simulated Leakage View (evaluating at future time):
        l1_leaked, l2_leaked = ds_protected.get_point_in_time_neighbors("emp_innocent", as_of=t_future_attack + timedelta(minutes=1))
        emb_leaked, sig_leaked = self.model.forward(emp_node, l1_leaked, l2_leaked)

        self.assertEqual(len(l1_leaked), 1)
        self.assertEqual(l1_leaked[0].id, "acct_mule")
        # Embeddings MUST differ between protected past view and future view
        diff = float(np.linalg.norm(emb_prot - emb_leaked))
        self.assertGreater(diff, 0.01, "Protected point-in-time embedding must isolate future events from leaking")


if __name__ == "__main__":
    unittest.main()
