"""
Phase 60 Test Suite: Real-Data Shadow Ingestion, Evidence Ledger, Label Maturation & Holdout Protection
"""

import os
import sys
import unittest
import hashlib
from datetime import datetime, timezone, timedelta

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import (
    HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType,
    LabelMaturityState, DatasetType, GraphSAGELifecycleState,
    ShadowInvestigationOutput, RelationshipRiskLevel, CollusionInvestigationDossier
)
from graphsage.privacy_sanitizer import PrivacySanitizer
from graphsage.label_maturation import LabelMaturationEngine
from graphsage.evidence_ledger import ShadowEvidenceLedger
from graphsage.collusion_tracker import CollusionAccumulationTracker
from graphsage.passive_ingestion import PassiveShadowIngestionAdapter
from graphsage.governance_gate import DataRequiredGate
from graphsage.readiness_checker import RealDataReadinessChecker
from graphsage.data_source import MockRealGraphDataSource

CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

HOLDOUT_PATH = os.path.join(ML_SERVICE_DIR, "data", "sample_ieee_fixture.csv")
HOLDOUT_EXPECTED_SHA256 = "a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44"


class TestGraphSAGEPhase60ShadowIngestion(unittest.TestCase):

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

    def test_03_privacy_boundary_rejects_and_quarantines_prohibited_fields(self):
        """Verifies that raw PAN, CVV, PIN, and SSN are quarantined and stripped before GraphSAGE processing."""
        sanitizer = PrivacySanitizer()

        # Intentionally polluted payload with multiple prohibited fields
        polluted_payload = {
            "employee_id": "emp_alice_01",
            "account_id": "acct_999",
            "raw_pan": "4111222233334444",
            "cvv": "999",
            "ssn": "123-45-6789",
            "memo": "Customer support payment assistance"
        }

        sanitized, violations, is_quarantined = sanitizer.sanitize_payload(polluted_payload)

        self.assertTrue(is_quarantined, "Polluted payload must be flagged for quarantine")
        self.assertGreaterEqual(len(violations), 3, "Expected at least 3 privacy violations")
        self.assertNotIn("raw_pan", sanitized)
        self.assertNotIn("cvv", sanitized)
        self.assertNotIn("ssn", sanitized)
        self.assertIn("memo", sanitized)

    def test_04_passive_ingestion_dedup_and_out_of_order_handling(self):
        """Verifies that the passive shadow adapter deduplicates events and handles late-arriving timestamps."""
        ds = MockRealGraphDataSource()
        adapter = PassiveShadowIngestionAdapter(target_data_source=ds, is_live_connected=False)

        t1 = self.now - timedelta(minutes=10)
        t2 = self.now - timedelta(minutes=5)
        t_late = self.now - timedelta(minutes=20)  # Out-of-order

        # 1. Normal event
        ok1, msg1 = adapter.consume_event("evt_001", "audit.events", {"employee_id": "emp_01", "account_id": "acct_01"}, t1)
        self.assertTrue(ok1)
        self.assertEqual(msg1, "INGESTED_TO_SHADOW_GRAPH")

        # 2. Duplicate event
        ok2, msg2 = adapter.consume_event("evt_001", "audit.events", {"employee_id": "emp_01", "account_id": "acct_01"}, t2)
        self.assertFalse(ok2)
        self.assertEqual(msg2, "DUPLICATE_EVENT_DROPPED")

        # 3. Out-of-order late arrival
        ok3, msg3 = adapter.consume_event("evt_002", "audit.events", {"employee_id": "emp_02", "account_id": "acct_02"}, t_late)
        self.assertTrue(ok3)

        # 4. Polluted event -> DLQ
        ok4, msg4 = adapter.consume_event("evt_003", "audit.events", {"employee_id": "emp_03", "raw_pan": "5500000000000004"}, self.now)
        self.assertFalse(ok4)
        self.assertIn("QUARANTINED_PRIVACY_VIOLATION", msg4)

        metrics = adapter.get_metrics()
        self.assertEqual(metrics.events_consumed, 4)
        self.assertEqual(metrics.events_duplicated, 1)
        self.assertEqual(metrics.events_out_of_order, 1)
        self.assertEqual(metrics.events_quarantined, 1)
        self.assertEqual(metrics.dlq_size, 1)
        self.assertEqual(metrics.real_data_connectivity, "NOT_CONNECTED")

    def test_05_label_maturation_lifecycle_and_threshold_tracking(self):
        """Verifies label state transitions and progress tracking toward the >= 50 real collusion requirement."""
        engine = LabelMaturationEngine(dispute_window_days=90)

        # Fresh transaction
        t_event = self.now - timedelta(days=100)
        rec = engine.register_transaction("txn_mat_01", t_event)
        self.assertEqual(rec.state, LabelMaturityState.UNLABELED)

        # Flag suspected
        engine.flag_suspected("txn_mat_01", t_event + timedelta(days=1), "GNN threshold alert")
        self.assertEqual(engine.records["txn_mat_01"].state, LabelMaturityState.SUSPECTED)

        # Open case
        engine.open_investigation("txn_mat_01", "CASE-9001", t_event + timedelta(days=2))
        self.assertEqual(engine.records["txn_mat_01"].state, LabelMaturityState.UNDER_INVESTIGATION)

        # Confirm internal collusion
        engine.confirm_label(
            "txn_mat_01",
            LabelMaturityState.CONFIRMED_FRAUD,
            "FORMAL_CASE_DISPOSITION",
            t_event + timedelta(days=15),
            is_collusion=True
        )
        self.assertEqual(engine.records["txn_mat_01"].state, LabelMaturityState.CONFIRMED_FRAUD)
        self.assertTrue(engine.records["txn_mat_01"].is_collusion)

        # Auto-mature clean transaction
        engine.register_transaction("txn_clean_02", self.now - timedelta(days=95))
        matured = engine.auto_mature_clean_transactions(as_of=self.now)
        self.assertEqual(matured, 1)
        self.assertEqual(engine.records["txn_clean_02"].state, LabelMaturityState.CONFIRMED_LEGITIMATE)

        report = engine.generate_report(as_of=self.now)
        self.assertEqual(report.confirmed_internal_collusion_count, 1)
        self.assertEqual(report.progress_to_collusion_gate_pct, 2.0)  # 1/50 = 2%
        self.assertIn("DATA_REQUIRED", report.governance_status)

    def test_06_shadow_evidence_ledger_append_only_integrity(self):
        """Verifies that the evidence ledger records dossiers with SHA-256 provenance hashes."""
        ledger = ShadowEvidenceLedger()

        dossier = CollusionInvestigationDossier(
            dossier_id="dos_test_001",
            generated_at=self.now,
            overall_risk_score=0.88,
            risk_level=RelationshipRiskLevel.HIGH,
            employee_id="emp_alice_01",
            employee_role="customer_support",
            affected_accounts=["acct_1", "acct_2"],
            affected_consumers=["usr_1"],
            relationship_paths=["(emp_alice_01) -> (acct_1) -> (usr_1)"],
            temporal_evidence={"supporting_evidence": ["Tight concentration"], "counter_evidence": ["Support role"]},
            transaction_summary={"total_amount_usd": 5000.0},
            shared_infrastructure={"device_overlap_count": 1},
            confidence_score=0.90,
            human_explanation="Investigation dossier test payload.",
            governance_notice="INVESTIGATION INTELLIGENCE ONLY - STRICTLY NON-ENFORCING"
        )

        entry = ledger.record_dossier(dossier)
        self.assertEqual(entry.dossier_id, "dos_test_001")
        self.assertTrue(entry.entry_id.startswith("ev_ledg_"))
        self.assertEqual(len(entry.provenance_hash), 64)  # Valid SHA-256

        stats = ledger.get_summary_statistics()
        self.assertEqual(stats["total_dossiers_recorded"], 1)
        self.assertEqual(stats["high_risk_dossiers_count"], 1)
        self.assertEqual(stats["storage_mode"], "APPEND_ONLY_AUDIT")

    def test_07_collusion_accumulation_tracker_summary(self):
        """Verifies that the collusion accumulation tracker monitors employees, concentrations, and self-approvals."""
        tracker = CollusionAccumulationTracker()

        # Add benign employee edges
        for i in range(10):
            edge_benign = HeteroEdge(
                id=f"e_b_{i}",
                source_id="emp_support_1",
                target_id=f"acct_benign_{i}",
                type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
                timestamp=self.now.replace(hour=10)
            )
            tracker.process_edge(edge_benign)

        # Add rogue employee concentrated edges + self-approval
        for i in range(4):
            edge_rogue = HeteroEdge(
                id=f"e_r_{i}",
                source_id="emp_rogue_1",
                target_id="acct_mule_1",
                type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
                timestamp=self.now.replace(hour=22)  # Off-hours
            )
            tracker.process_edge(edge_rogue)

        edge_appr = HeteroEdge(
            id="e_appr_1",
            source_id="emp_rogue_1",
            target_id="txn_1",
            type=HeteroEdgeType.EMPLOYEE_APPROVES_TRANSACTION,
            timestamp=self.now
        )
        tracker.process_edge(edge_appr)

        view = tracker.get_summary_view()
        self.assertEqual(view.total_employees_observed, 2)
        self.assertEqual(view.employees_with_account_access, 2)
        self.assertEqual(view.employees_with_concentrated_access, 1)  # emp_rogue_1 has HHI = 1.0 > 0.25
        self.assertEqual(view.total_self_approval_events, 1)
        self.assertEqual(view.total_off_hours_access_events, 4)
        self.assertIn("DATA_REQUIRED", view.collusion_gate_status)

    def test_08_governance_gate_blocks_promotion_until_real_50_cases_exist(self):
        """Verifies that the governance gate remains BLOCKED with 0 real labels."""
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
