"""
Automated Test Suite: GraphSAGE Feature Leakage Prevention & Anti-Shortcut Verification
"""

import os
import sys
import unittest
import numpy as np
from datetime import datetime, timezone

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import HeteroNode, HeteroNodeType
from graphsage.model import GraphSAGEModel


class TestGraphSAGELeakagePrevention(unittest.TestCase):

    def setUp(self):
        self.model = GraphSAGEModel(input_dim=32, hidden_dim=64, output_dim=64, seed=42)
        self.now = datetime(2025, 6, 1, tzinfo=timezone.utc)

    def test_zero_raw_numeric_id_in_node_features(self):
        """
        Tests that node ID numbers (e.g. acct_000001 vs acct_099999) produce ZERO scalar magnitude in features.
        Prevents model from learning shortcut scenario partitions based on sequential allocation.
        """
        node_low_id = HeteroNode(id="acct_000001", type=HeteroNodeType.ACCOUNT, risk_score=0.05, created_at=self.now)
        node_high_id = HeteroNode(id="acct_099999", type=HeteroNodeType.ACCOUNT, risk_score=0.05, created_at=self.now)

        feat_low = self.model.compute_node_features(node_low_id)
        feat_high = self.model.compute_node_features(node_high_id)

        # Identical node types and properties MUST produce identical feature vectors regardless of ID
        np.testing.assert_array_almost_equal(
            feat_low, feat_high,
            err_msg="Node features MUST NOT depend on entity ID numbers or strings"
        )

    def test_forbidden_columns_absent_from_feature_schema(self):
        """
        Verifies that scenario tags, label provenance, and ground truth labels cannot enter feature vectors.
        """
        forbidden_keys = {
            "scenario_type", "is_adversarial", "is_hard_negative",
            "ground_truth", "label_source", "label_confidence", "label_timestamp"
        }

        # Even if a malicious dictionary passes forbidden properties, compute_node_features must ignore them
        polluted_node = HeteroNode(
            id="acct_mule_001",
            type=HeteroNodeType.ACCOUNT,
            risk_score=0.05,
            created_at=self.now,
            properties={
                "scenario_type": "employee_collusion",
                "is_adversarial": True,
                "label_confidence": 0.99,
                "ground_truth": "INTERNAL_COLLUSION"
            }
        )
        clean_node = HeteroNode(id="acct_mule_001", type=HeteroNodeType.ACCOUNT, risk_score=0.05, created_at=self.now)

        feat_polluted = self.model.compute_node_features(polluted_node)
        feat_clean = self.model.compute_node_features(clean_node)

        np.testing.assert_array_almost_equal(
            feat_polluted, feat_clean,
            err_msg="Forbidden scenario tags or label fields MUST NOT leak into feature vectors"
        )


if __name__ == "__main__":
    unittest.main()
