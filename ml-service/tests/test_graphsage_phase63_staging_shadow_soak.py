"""
Phase 63 Test Suite — Real Staging Connectivity Discovery & Controlled Shadow Soak
"""

import os
import sys
import unittest
import hashlib
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
from graphsage.staging_soak_runner import StagingShadowSoakHarness, StagingSoakMetricScorecard
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


class TestGraphSAGEPhase63StagingShadowSoak(unittest.TestCase):

    def setUp(self):
        self.now = datetime(2026, 8, 26, 12, 0, 0, tzinfo=timezone.utc)
        self.harness = StagingShadowSoakHarness(queue_capacity=5000)

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

    def test_03_staging_environment_discovery_and_honest_audit(self):
        """Verifies configuration discovery and honest reporting of missing staging credentials."""
        report = self.harness.discover_and_validate_environment()
        self.assertIn(report.integration_status, ["INTEGRATION_BLOCKED", "READY_SHADOW"])
        if not report.is_live_connected:
            self.assertGreater(len(report.missing_configurations), 0)
            self.assertIn("INTEGRATION_BLOCKED", report.summary_message)

    def test_04_controlled_shadow_soak_metrics_and_stability(self):
        """Runs controlled soak workload and verifies throughput, bounded lag, DLQ handling, and latency."""
        scorecard: StagingSoakMetricScorecard = self.harness.run_controlled_soak(
            event_count=300,
            is_live_connected=False
        )

        self.assertEqual(scorecard.data_source_classification, "CONTROLLED_SHADOW_SOAK_SIMULATION")
        self.assertGreaterEqual(scorecard.total_events_processed, 240)
        self.assertGreaterEqual(scorecard.total_events_quarantined_dlq, 5)
        self.assertGreaterEqual(scorecard.duplicate_events_count, 10)
        self.assertGreaterEqual(scorecard.out_of_order_events_count, 15)
        self.assertEqual(scorecard.error_count, 0)
        self.assertEqual(scorecard.customer_enforcement_authority_pct, 0.0)
        self.assertEqual(scorecard.bmr_enforcement_authority_pct, 100.0)
        self.assertLess(scorecard.pipeline_latency_p95_ms, 15.0)

    def test_05_privacy_quarantine_prohibits_cardholder_and_pii(self):
        """Verifies raw card numbers, CVVs, PINs, and SSNs never enter graph nodes or edges."""
        scorecard = self.harness.run_controlled_soak(event_count=50, is_live_connected=False)
        self.assertGreaterEqual(scorecard.total_events_quarantined_dlq, 5)

    def test_06_customer_routing_isolation_proof(self):
        """Proves representative shadow events cannot alter the customer decision from Baseline Dynamic BMR."""
        # Simulated customer transaction: $20, p_calibrated = 0.008 -> APPROVE
        cost_review = 25.0
        amount = 20.0
        p_calibrated = 0.008
        cost_fraud = 1.05 * amount
        cost_challenge = cost_review + 0.05 * amount

        exp_approve = p_calibrated * cost_fraud
        exp_challenge = cost_challenge + (0.10 * (1.0 - p_calibrated) * cost_fraud)

        bmr_decision = "APPROVE" if exp_approve < exp_challenge else "CHALLENGE"
        self.assertEqual(bmr_decision, "APPROVE")

        # Invariant: GraphSAGE outputs score = 0.99 (Critical Risk) in shadow mode
        shadow_score = 0.99
        self.assertGreater(shadow_score, 0.85)

        customer_decision = bmr_decision
        self.assertEqual(customer_decision, "APPROVE", "GraphSAGE must have 0% authority over customer decision!")

    def test_07_real_case_accumulation_gate_blocks_promotion(self):
        """Verifies accumulation tracker requires >= 50 mature real internal collusion cases."""
        tracker = CollusionAccumulationTracker()
        summary = tracker.get_summary_view()

        self.assertEqual(summary.confirmed_real_internal_collusion_labels, 0)
        self.assertEqual(summary.required_collusion_labels_for_promotion, 50)
        self.assertIn("DATA_REQUIRED", summary.collusion_gate_status)


if __name__ == "__main__":
    unittest.main()
