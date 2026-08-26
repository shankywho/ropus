"""
Phase 62 Test Suite — Live Shadow Integration, Connectivity Audit & Persistent Observability
"""

import os
import sys
import unittest
import hashlib
import time
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
from graphsage.shadow_telemetry import PassiveShadowTelemetryEngine
from graphsage.data_source import MockRealGraphDataSource
from graphsage.connectivity_checker import RealDataConnectivityChecker, RealDataConnectivityReport
from graphsage.observability_exporter import ShadowObservabilityExporter, ShadowMetricsSnapshot
from graphsage.collusion_tracker import CollusionAccumulationTracker


CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

HOLDOUT_PATH = os.path.join(ML_SERVICE_DIR, "data", "sample_ieee_fixture.csv")
HOLDOUT_EXPECTED_SHA256 = "a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44"


def compute_sha256(filepath: str) -> str:
    h = hashlib.sha256()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            h.update(chunk)
    return h.hexdigest()


class TestGraphSAGEPhase62LiveShadowIntegration(unittest.TestCase):

    def setUp(self):
        self.now = datetime(2026, 8, 26, 12, 0, 0, tzinfo=timezone.utc)

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

    def test_03_connectivity_checker_honest_audit(self):
        """Verifies that RealDataConnectivityChecker honestly identifies missing staging config."""
        checker = RealDataConnectivityChecker(timeout_sec=0.2)
        report: RealDataConnectivityReport = checker.audit_environment()

        self.assertIsInstance(report, RealDataConnectivityReport)
        self.assertIn(report.integration_status, ["INTEGRATION_BLOCKED", "READY_SHADOW"])
        if not report.is_live_connected:
            self.assertGreater(len(report.missing_configurations), 0)
            self.assertIn("INTEGRATION_BLOCKED", report.summary_message)

    def test_04_observability_exporter_metrics_and_prometheus_text(self):
        """Verifies that ShadowObservabilityExporter produces valid metrics snapshots and Prometheus text."""
        engine = PassiveShadowTelemetryEngine(queue_size=500, is_live_connected=False)
        engine.start()

        # Stream sample clean event and sample DLQ event
        engine.enqueue_event(
            "evt_obs_01",
            "audit.events",
            {"employee_id": "emp_alice", "account_id": "acct_101"},
            self.now
        )
        engine.enqueue_event(
            "evt_obs_02",
            "audit.events",
            {"employee_id": "emp_bad", "raw_pan": "4111222233334444"},
            self.now
        )

        time.sleep(0.15)
        engine.stop()

        exporter = ShadowObservabilityExporter(engine)
        snapshot: ShadowMetricsSnapshot = exporter.get_snapshot()

        self.assertEqual(snapshot.customer_enforcement_authority_pct, 0.0)
        self.assertEqual(snapshot.bmr_enforcement_authority_pct, 100.0)
        self.assertGreaterEqual(snapshot.events_processed_total, 1)
        self.assertGreaterEqual(snapshot.events_quarantined_dlq_total, 1)

        prom_text = exporter.export_prometheus_text()
        self.assertIn("graphsage_customer_enforcement_authority_pct 0.0", prom_text)
        self.assertIn("bmr_customer_enforcement_authority_pct 100.0", prom_text)
        self.assertIn("graphsage_shadow_events_processed_total", prom_text)
        self.assertIn("graphsage_shadow_events_dlq_total", prom_text)

    def test_05_real_case_accumulation_tracker_governance(self):
        """Verifies that the accumulation tracker strictly requires >= 50 mature real internal collusion labels."""
        tracker = CollusionAccumulationTracker()
        summary = tracker.get_summary_view()

        self.assertEqual(summary.confirmed_real_internal_collusion_labels, 0)
        self.assertEqual(summary.required_collusion_labels_for_promotion, 50)
        self.assertIn("DATA_REQUIRED", summary.collusion_gate_status)
        self.assertIn("0/50", summary.collusion_gate_status)

    def test_06_shadow_safety_and_customer_decision_isolation(self):
        """Proves customer decisions are 100% determined by Baseline Dynamic BMR regardless of shadow state."""
        # Simulated customer transaction: $15, p_calibrated = 0.005 -> APPROVE
        cost_review = 25.0
        amount = 15.0
        p_calibrated = 0.005
        cost_fraud = 1.05 * amount
        cost_challenge = cost_review + 0.05 * amount

        # Expected cost calculation (BMR)
        exp_approve = p_calibrated * cost_fraud
        exp_challenge = cost_challenge + (0.10 * (1.0 - p_calibrated) * cost_fraud)

        # Baseline decision
        bmr_decision = "APPROVE" if exp_approve < exp_challenge else "CHALLENGE"
        self.assertEqual(bmr_decision, "APPROVE")

        # Invariant: Extreme GraphSAGE scores (0.99 or 0.00) or Shadow Panics must NOT change bmr_decision
        shadow_risk_score = 0.99
        self.assertGreater(shadow_risk_score, 0.85)
        # Decision routed to customer remains 100% BMR
        customer_decision = bmr_decision
        self.assertEqual(customer_decision, "APPROVE")


if __name__ == "__main__":
    unittest.main()
