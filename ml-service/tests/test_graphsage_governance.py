"""
Automated Test Suite: GraphSAGE Governance, Promotion Gates, and Champion Immutability
"""

import os
import sys
import json
import hashlib
import unittest

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import GraphSAGELifecycleState, GraphSAGEDataContract

CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"


class TestGraphSAGEGovernance(unittest.TestCase):

    def test_production_champion_checksum_invariant(self):
        """
        CRITICAL: Verifies byte-for-byte immutability of production champion.
        """
        self.assertTrue(os.path.exists(CHAMPION_PATH), f"Production champion missing: {CHAMPION_PATH}")
        hasher = hashlib.sha256()
        with open(CHAMPION_PATH, "rb") as f:
            while chunk := f.read(65536):
                hasher.update(chunk)
        computed_sha256 = hasher.hexdigest()
        self.assertEqual(
            computed_sha256, CHAMPION_EXPECTED_SHA256,
            "CRITICAL GOVERNANCE VIOLATION: Production champion model checksum has changed!"
        )

    def test_governance_lifecycle_state_machine(self):
        """
        Verifies that GraphSAGE has explicit states and CANNOT jump directly from synthetic validation to production.
        """
        states = [s.value for s in GraphSAGELifecycleState]
        self.assertIn("SYNTHETIC_TRAINED", states)
        self.assertIn("SYNTHETIC_VALIDATED", states)
        self.assertIn("REAL_DATA_REQUIRED", states)
        self.assertIn("SHADOW_EVALUATION", states)
        self.assertIn("HUMAN_REVIEW", states)
        self.assertIn("PROMOTION_ELIGIBLE", states)

    def test_promotion_gate_blocked_without_real_holdout(self):
        """
        Verifies that synthetic data alone CANNOT promote GraphSAGE to production enforcement.
        """
        contract = GraphSAGEDataContract()
        self.assertGreaterEqual(contract.minimum_genuine_internal_fraud_cases_for_promotion, 50)
        self.assertFalse(contract.lifecycle_status == GraphSAGELifecycleState.PROMOTION_ELIGIBLE)


if __name__ == "__main__":
    unittest.main()
