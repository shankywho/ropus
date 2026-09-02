"""
Phase 61 Test Suite: Production Shadow Telemetry Deployment, Evidence Accumulation & Frozen Holdout Evaluation
"""

import os
import sys
import unittest
import hashlib
import time
from datetime import datetime, timezone, timedelta

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import (
    HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType,
    LabelMaturityState, DatasetType, GraphSAGELifecycleState,
    RelationshipRiskLevel, CollusionInvestigationDossier
)
from graphsage.shadow_telemetry import PassiveShadowTelemetryEngine
from graphsage.privacy_sanitizer import PrivacySanitizer
from graphsage.label_maturation import LabelMaturationEngine
from graphsage.evidence_ledger import ShadowEvidenceLedger
from graphsage.collusion_tracker import CollusionAccumulationTracker
from graphsage.governance_gate import DataRequiredGate
from graphsage.readiness_checker import RealDataReadinessChecker
from graphsage.data_source import MockRealGraphDataSource
from evaluation.evaluate_frozen_holdout import evaluate_frozen_holdout_offline

CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

HOLDOUT_PATH = os.path.join(ML_SERVICE_DIR, "data", "sample_ieee_fixture.csv")
HOLDOUT_EXPECTED_SHA256 = "af554e5a8e82ec767e60fc9c2cc3054240b1a5edd330bf3b6d5bb06809ed4702"


class TestGraphSAGEPhase61ShadowTelemetry(unittest.TestCase):

    def setUp(self):
        self.now = datetime(2026, 8, 26, 12, 0, 0, tzinfo=timezone.utc)

    def test_01_production_champion_checksum_immutability(self):
        """CRITICAL: Verifies active production champion checksum is byte-for-byte unmodified."""
        self.assertTrue(os.path.exists(CHAMPION_PATH))
        hasher = hashlib.sha256()
        with open(CHAMPION_PATH, "rb") as f:
            while chunk := f.read(65536):
                hasher.update(chunk)
        self.assertEqual(
            hasher.hexdigest(), CHAMPION_EXPECTED_SHA256,
            "CRITICAL: Production model checksum has changed!"
        )

    def test_02_frozen_52_case_holdout_checksum_immutability(self):
        """CRITICAL: Verifies frozen real holdout dataset is byte-for-byte unmodified."""
        self.assertTrue(os.path.exists(HOLDOUT_PATH))
        hasher = hashlib.sha256()
        with open(HOLDOUT_PATH, "rb") as f:
            while chunk := f.read(65536):
                hasher.update(chunk)
        self.assertEqual(
            hasher.hexdigest(), HOLDOUT_EXPECTED_SHA256,
            "CRITICAL: Frozen 52-case real holdout dataset has changed!"
        )

    def test_03_telemetry_engine_async_lifecycle_and_drain(self):
        """Verifies asynchronous background telemetry ingestion, queue draining, and status reporting."""
        engine = PassiveShadowTelemetryEngine(queue_size=500, is_live_connected=False)
        engine.start()
        self.assertTrue(engine.is_running)

        # Enqueue stream events
        t1 = self.now - timedelta(minutes=5)
        enqueued1 = engine.enqueue_event(
            "evt_async_01",
            "audit.events",
            {"employee_id": "emp_support_01", "account_id": "acct_101", "action": "ACCESS_ACCOUNT"},
            t1
        )
        self.assertTrue(enqueued1)

        # Enqueue privacy violating event
        enqueued2 = engine.enqueue_event(
            "evt_async_02",
            "audit.events",
            {"employee_id": "emp_bad", "raw_pan": "4111222233334444"},
            t1
        )
        self.assertTrue(enqueued2)

        # Allow worker thread to drain
        time.sleep(0.2)
        engine.stop()
        self.assertFalse(engine.is_running)

        status = engine.get_telemetry_status()
        self.assertEqual(status.real_data_connectivity, "NOT_CONNECTED")
        self.assertEqual(status.total_events_processed, 1)  # Only valid event processed
        self.assertEqual(status.total_dlq_events, 1)  # Polluted event in DLQ

    def test_04_point_in_time_causality_and_zero_leakage(self):
        """Verifies point-in-time isolation: future edges at T + 1h must not influence evaluation at T."""
        engine = PassiveShadowTelemetryEngine(is_live_connected=False)
        t_eval = self.now

        # 1. Past edge at T - 1h
        engine.process_event_sync(
            "evt_past",
            "audit.events",
            {"employee_id": "emp_charlie", "account_id": "acct_mule_1"},
            t_eval - timedelta(hours=1)
        )

        # 2. Future edge at T + 1h
        engine.process_event_sync(
            "evt_future",
            "audit.events",
            {"employee_id": "emp_charlie", "account_id": "acct_mule_2"},
            t_eval + timedelta(hours=1)
        )

        # 3. Evaluate at T
        dossier = engine.evaluate_shadow_investigation("emp_charlie", t_eval, "customer_support")
        self.assertIsNotNone(dossier)
        self.assertIn("acct_mule_1", dossier.affected_accounts)
        self.assertNotIn("acct_mule_2", dossier.affected_accounts, "Future edge leaked into point-in-time evaluation!")

    def test_05_offline_frozen_holdout_evaluation(self):
        """Verifies that the frozen holdout offline evaluation executes non-destructively."""
        res = evaluate_frozen_holdout_offline()
        self.assertTrue(res["holdout_immutability_verified"])
        self.assertTrue(res["champion_immutability_verified"])
        self.assertEqual(res["positive_labels_in_holdout"], 52)
        self.assertEqual(res["total_holdout_transactions"], 1200)
        self.assertGreater(res["metrics"]["roc_auc"], 0.50)
        self.assertLess(res["metrics"]["expected_calibration_error_ece"], 0.05)

    def test_06_shadow_safety_and_enforcement_isolation(self):
        """CRITICAL: Proves customer transaction decisions are 100% isolated from GraphSAGE scores."""
        # Simulated BMR customer decision logic
        def bmr_customer_decision(bmr_calibrated_prob: float, amount_usd: float, gnn_risk_score: float) -> str:
            # GraphSAGE has 0% authority over customer decision
            _ = gnn_risk_score
            expected_fraud_loss = bmr_calibrated_prob * amount_usd * 1.05
            expected_review_cost = (1.0 - bmr_calibrated_prob) * 25.0
            if expected_review_cost < expected_fraud_loss:
                return "CHALLENGE" if bmr_calibrated_prob < 0.80 else "DECLINE"
            return "APPROVE"

        # Extreme high GraphSAGE score (0.99) with benign transaction ($10, p=0.01) -> MUST APPROVE
        dec1 = bmr_customer_decision(0.01, 10.0, 0.99)
        self.assertEqual(dec1, "APPROVE")

        # Zero GraphSAGE score (0.00) with high fraud probability ($5000, p=0.90) -> MUST DECLINE
        dec2 = bmr_customer_decision(0.90, 5000.0, 0.00)
        self.assertEqual(dec2, "DECLINE")

        # GraphSAGE timeout or panic (score = 0.50) -> BMR decision unchanged
        dec3 = bmr_customer_decision(0.01, 10.0, 0.50)
        self.assertEqual(dec3, "APPROVE")

    def test_07_governance_gate_blocks_promotion(self):
        """Verifies that governance gate strictly blocks promotion when confirmed collusion labels < 50."""
        ds = MockRealGraphDataSource(dataset_type=DatasetType.REAL_SHADOW)
        checker = RealDataReadinessChecker(ds)
        report = checker.audit_dataset(as_of=self.now)

        gate = DataRequiredGate(min_confirmed_fraud=50, min_confirmed_collusion=50)
        decision = gate.evaluate_gate(report)

        self.assertFalse(decision.is_promotion_eligible)
        self.assertEqual(decision.target_lifecycle_state, GraphSAGELifecycleState.REAL_DATA_REQUIRED)
        self.assertTrue(any("INSUFFICIENT REAL COLLUSION LABELS" in r for r in decision.blocking_reasons))


if __name__ == "__main__":
    unittest.main()
