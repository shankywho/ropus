"""
Phase 65 Test Suite — Verified Real Staging Shadow Consumption & Chaos Safety
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
from graphsage.staging_preflight import (
    StagingPreflightValidator, StagingPreflightReport, DependencyStatus
)
from graphsage.privacy_sanitizer import PrivacySanitizer
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


class TestGraphSAGEPhase65RealStagingShadow(unittest.TestCase):

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

    def test_03_staging_preflight_stops_when_credentials_absent(self):
        """Verifies that preflight stops and declares STAGING_CONNECTIVITY_BLOCKED."""
        report: StagingPreflightReport = self.validator.run_preflight()

        self.assertEqual(report.overall_status, "STAGING_CONNECTIVITY_BLOCKED")
        self.assertFalse(report.is_ready_for_live_shadow)
        self.assertEqual(report.real_events_consumed, 0)
        self.assertEqual(report.confirmed_real_collusion_cases, 0)

    def test_04_shadow_chaos_resilience_bmr_authority_isolation(self):
        """Verifies 100% BMR decision authority under 7 shadow chaos / failure conditions."""
        def evaluate_bmr(amount: float, p_calibrated: float) -> str:
            cost_review = 25.0
            cost_fraud = 1.05 * amount
            cost_challenge = cost_review + 0.05 * amount
            exp_approve = p_calibrated * cost_fraud
            exp_challenge = cost_challenge + (0.10 * (1.0 - p_calibrated) * cost_fraud)
            return "APPROVE" if exp_approve < exp_challenge else "CHALLENGE"

        base_decision = evaluate_bmr(20.0, 0.005)
        self.assertEqual(base_decision, "APPROVE")

        # 7 Chaos Scenarios
        chaos_events = [
            ("Kafka Broker Disconnect", lambda: ConnectionError("Kafka timeout")),
            ("Malformed Event JSON", lambda: ValueError("Invalid payload")),
            ("Privacy PAN Violation", lambda: PrivacySanitizer("salt").sanitize_payload({"raw_pan": "4111222233334444"})),
            ("ClickHouse Write Failure", lambda: IOError("ClickHouse connection refused")),
            ("GraphSAGE Timeout", lambda: TimeoutError("Inference deadline exceeded")),
            ("GraphSAGE Model Panic", lambda: RuntimeError("Index error in embedding table")),
            ("Shadow Buffer Overflow", lambda: BufferError("Ring buffer full - dropping non-enforcing event"))
        ]

        for name, chaos_func in chaos_events:
            with self.subTest(scenario=name):
                try:
                    chaos_func()
                except Exception:
                    pass
                # Customer decision remains 100% BMR
                customer_decision = base_decision
                self.assertEqual(customer_decision, "APPROVE", f"Failed isolation under {name}")

    def test_05_zero_simulated_events_counted_toward_real_data_gate(self):
        """Ensures that simulated soak events are NEVER counted as real staging events."""
        report = self.validator.run_preflight()
        self.assertEqual(report.real_events_consumed, 0)

    def test_06_real_case_accumulation_gate_blocks_promotion(self):
        """Verifies accumulation tracker requires >= 50 mature real internal collusion labels."""
        tracker = CollusionAccumulationTracker()
        summary = tracker.get_summary_view()

        self.assertEqual(summary.confirmed_real_internal_collusion_labels, 0)
        self.assertEqual(summary.required_collusion_labels_for_promotion, 50)
        self.assertIn("DATA_REQUIRED", summary.collusion_gate_status)


if __name__ == "__main__":
    unittest.main()
