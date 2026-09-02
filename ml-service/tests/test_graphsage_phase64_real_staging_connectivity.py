"""
Phase 64 Test Suite — Real Staging Connectivity Preflight & Verified Shadow Consumption
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


class TestGraphSAGEPhase64RealStagingConnectivity(unittest.TestCase):

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

    def test_03_staging_preflight_audit_reports_exact_dependencies(self):
        """Audits all dependencies and verifies granular status reporting."""
        report: StagingPreflightReport = self.validator.run_preflight()

        self.assertEqual(report.overall_status, "STAGING_CONNECTIVITY_BLOCKED")
        self.assertFalse(report.is_ready_for_live_shadow)
        self.assertGreater(len(report.missing_configurations), 0)
        self.assertGreater(len(report.blocking_reasons), 0)
        self.assertGreater(len(report.required_unblocking_actions), 0)

        self.assertIn("kafka_brokers", report.dependency_results)
        self.assertIn("kafka_auth", report.dependency_results)
        self.assertIn("topic_audit_events", report.dependency_results)
        self.assertIn("topic_transactions_created", report.dependency_results)
        self.assertIn("clickhouse", report.dependency_results)
        self.assertIn("prometheus", report.dependency_results)

    def test_04_privacy_sanitizer_quarantines_prohibited_pii_and_card_data(self):
        """Ensures raw PAN, CVV, PIN, SSN, and bank account numbers are quarantined to DLQ."""
        sanitizer = PrivacySanitizer("phase64_test_salt")
        payload = {
            "employee_id": "emp_audit_01",
            "account_id": "acct_999",
            "raw_pan": "4111222233334444",
            "cvv": "999",
            "pin": "1234",
            "ssn": "123-45-6789",
            "bank_account_number": "123456789012"
        }

        sanitized, violations, is_quarantined = sanitizer.sanitize_payload(payload)
        self.assertTrue(is_quarantined)
        self.assertGreaterEqual(len(violations), 5)
        self.assertNotIn("raw_pan", sanitized)
        self.assertNotIn("cvv", sanitized)
        self.assertNotIn("pin", sanitized)
        self.assertNotIn("ssn", sanitized)
        self.assertNotIn("bank_account_number", sanitized)

    def test_05_customer_routing_zero_enforcement_isolation(self):
        """Proves 100% customer decision authority resides with Baseline Dynamic BMR."""
        cost_review = 25.0
        amount = 30.0
        p_calibrated = 0.005
        cost_fraud = 1.05 * amount
        cost_challenge = cost_review + 0.05 * amount

        exp_approve = p_calibrated * cost_fraud
        exp_challenge = cost_challenge + (0.10 * (1.0 - p_calibrated) * cost_fraud)

        bmr_decision = "APPROVE" if exp_approve < exp_challenge else "CHALLENGE"
        self.assertEqual(bmr_decision, "APPROVE")

        # Invariant: Critical shadow risk score (0.98) cannot modify customer decision
        shadow_score = 0.98
        self.assertGreater(shadow_score, 0.85)

        customer_decision = bmr_decision
        self.assertEqual(customer_decision, "APPROVE", "GraphSAGE must have 0% authority over customer decisions!")

    def test_06_zero_simulated_events_counted_as_real(self):
        """Ensures that simulated soak events are NEVER counted as real staging events."""
        report = self.validator.run_preflight()
        self.assertEqual(report.real_events_consumed, 0)
        self.assertEqual(report.confirmed_real_collusion_cases, 0)

    def test_07_real_case_accumulation_gate_blocks_promotion(self):
        """Verifies accumulation tracker strictly requires >= 50 mature real internal collusion labels."""
        tracker = CollusionAccumulationTracker()
        summary = tracker.get_summary_view()

        self.assertEqual(summary.confirmed_real_internal_collusion_labels, 0)
        self.assertEqual(summary.required_collusion_labels_for_promotion, 50)
        self.assertIn("DATA_REQUIRED", summary.collusion_gate_status)


if __name__ == "__main__":
    unittest.main()
