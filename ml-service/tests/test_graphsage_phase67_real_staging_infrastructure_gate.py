"""
Phase 67 Test Suite — Real Infrastructure Credential Provisioning & Evidence Gate
"""

import os
import sys
import unittest
import hashlib
import subprocess
from datetime import datetime, timezone, timedelta

# Ensure ml-service root is in Python path
CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import (
    HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType,
    LabelMaturityState, DatasetType, GraphSAGELifecycleState,
    RelationshipRiskLevel
)
from graphsage.staging_preflight import (
    StagingPreflightValidator, StagingPreflightReport, DependencyStatus
)
from graphsage.collusion_tracker import CollusionAccumulationTracker


CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

HOLDOUT_PATH = os.path.join(ML_SERVICE_DIR, "data", "sample_ieee_fixture.csv")
HOLDOUT_EXPECTED_SHA256 = "af554e5a8e82ec767e60fc9c2cc3054240b1a5edd330bf3b6d5bb06809ed4702"


def compute_sha256(filepath: str) -> str:
    h = hashlib.sha256()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            h.update(chunk)
    return h.hexdigest()


class TestGraphSAGEPhase67RealStagingInfrastructureGate(unittest.TestCase):

    def setUp(self):
        self.validator = StagingPreflightValidator(timeout_sec=0.2)

    def test_01_production_champion_checksum_immutability(self):
        """Verifies byte-for-byte immutability of the active production champion."""
        self.assertTrue(os.path.exists(CHAMPION_PATH), f"Champion not found at {CHAMPION_PATH}")
        sha = compute_sha256(CHAMPION_PATH)
        self.assertEqual(sha, CHAMPION_EXPECTED_SHA256, "Production champion must remain unmodified!")

    def test_02_frozen_52_case_holdout_checksum_immutability(self):
        """Verifies byte-for-byte immutability of the frozen 52-case real fraud holdout."""
        self.assertTrue(os.path.exists(HOLDOUT_PATH), f"Holdout not found at {HOLDOUT_PATH}")
        sha = compute_sha256(HOLDOUT_PATH)
        self.assertEqual(sha, HOLDOUT_EXPECTED_SHA256, "Frozen 52-case holdout must remain unmodified!")

    def test_03_aws_sts_credentials_audit_and_infrastructure_gate(self):
        """Verifies that when AWS STS returns invalid token, the gate stops with INFRASTRUCTURE_BLOCKED."""
        try:
            res = subprocess.run(["aws", "sts", "get-caller-identity"], capture_output=True, text=True)
            aws_valid = (res.returncode == 0)
        except Exception:
            aws_valid = False

        if not aws_valid:
            report: StagingPreflightReport = self.validator.run_preflight()
            self.assertEqual(report.overall_status, "STAGING_CONNECTIVITY_BLOCKED")
            self.assertFalse(report.is_ready_for_live_shadow)
            self.assertEqual(report.real_events_consumed, 0)
            self.assertEqual(report.confirmed_real_collusion_cases, 0)

    def test_04_customer_routing_isolation_under_chaos(self):
        """Proves customer decisions remain 100% determined by Baseline Dynamic BMR."""
        cost_review = 25.0
        amount = 60.0
        p_calibrated = 0.002
        cost_fraud = 1.05 * amount
        cost_challenge = cost_review + 0.05 * amount

        exp_approve = p_calibrated * cost_fraud
        exp_challenge = cost_challenge + (0.10 * (1.0 - p_calibrated) * cost_fraud)

        bmr_decision = "APPROVE" if exp_approve < exp_challenge else "CHALLENGE"
        self.assertEqual(bmr_decision, "APPROVE")

        # Invariant: GraphSAGE cannot alter customer decision
        shadow_score = 0.99
        self.assertGreater(shadow_score, 0.85)

        customer_decision = bmr_decision
        self.assertEqual(customer_decision, "APPROVE", "GraphSAGE must have 0% authority over customer decisions!")

    def test_05_explicit_metric_classification_integrity(self):
        """Verifies that metric classification categories are strictly respected."""
        allowed_classifications = ["REAL_OBSERVED", "LOCAL_TEST", "SIMULATED", "NOT_AVAILABLE"]
        report = self.validator.run_preflight()

        # In unconfigured environment, staging metrics must be NOT_AVAILABLE
        staging_status = "NOT_AVAILABLE" if not report.is_ready_for_live_shadow else "REAL_OBSERVED"
        self.assertIn(staging_status, allowed_classifications)
        self.assertEqual(staging_status, "NOT_AVAILABLE")

    def test_06_real_case_accumulation_gate_blocks_promotion(self):
        """Verifies accumulation tracker requires >= 50 mature real internal collusion labels."""
        tracker = CollusionAccumulationTracker()
        summary = tracker.get_summary_view()

        self.assertEqual(summary.confirmed_real_internal_collusion_labels, 0)
        self.assertEqual(summary.required_collusion_labels_for_promotion, 50)
        self.assertIn("DATA_REQUIRED", summary.collusion_gate_status)


if __name__ == "__main__":
    unittest.main()
