"""
Phase 58 Test Suite: Real-Data Shadow Readiness, Governance Gates, Label Maturity & Dossier Verification
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
    ShadowInvestigationOutput
)
from graphsage.data_source import MockRealGraphDataSource, PointInTimeReplayHarness
from graphsage.readiness_checker import RealDataReadinessChecker
from graphsage.dossier_generator import CollusionDossierGenerator
from graphsage.governance_gate import DataRequiredGate
from graphsage.model import GraphSAGEModel

CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"


class TestGraphSAGEPhase58Readiness(unittest.TestCase):

    def setUp(self):
        self.now = datetime(2026, 8, 26, 12, 0, 0, tzinfo=timezone.utc)
        self.model = GraphSAGEModel(input_dim=32, hidden_dim=64, output_dim=64, seed=42)

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

    def test_02_label_maturity_state_machine(self):
        """Verifies explicit label maturity states and prevents suspected flags from acting as ground truth."""
        states = [s.value for s in LabelMaturityState]
        self.assertEqual(states, [
            "UNLABELED", "SUSPECTED", "UNDER_INVESTIGATION",
            "CONFIRMED_FRAUD", "CONFIRMED_LEGITIMATE", "REJECTED"
        ])

        source = MockRealGraphDataSource(dataset_type=DatasetType.REAL_SHADOW)
        source.AddLabel({
            "entity_id": "txn_001",
            "maturity_state": LabelMaturityState.SUSPECTED.value,
            "label_timestamp": self.now.isoformat()
        })
        source.AddLabel({
            "entity_id": "txn_002",
            "maturity_state": LabelMaturityState.CONFIRMED_FRAUD.value,
            "label_timestamp": self.now.isoformat()
        })

        mature_labels = source.get_labels(as_of=self.now, mature_only=True)
        self.assertEqual(len(mature_labels), 1)
        self.assertEqual(mature_labels[0]["entity_id"], "txn_002")

    def test_03_point_in_time_replay_harness_causality(self):
        """Verifies that the replay harness strictly enforces edge_timestamp <= T."""
        t_past = self.now - timedelta(days=5)
        t_eval = self.now
        t_future = self.now + timedelta(days=2)

        emp = HeteroNode(id="emp_01", type=HeteroNodeType.EMPLOYEE, created_at=t_past)
        acct = HeteroNode(id="acct_01", type=HeteroNodeType.ACCOUNT, created_at=t_past)
        edge_past = HeteroEdge(id="e_past", source_id="emp_01", target_id="acct_01", type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT, timestamp=t_past)
        edge_future = HeteroEdge(id="e_future", source_id="emp_01", target_id="acct_01", type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT, timestamp=t_future)

        source = MockRealGraphDataSource(
            nodes=[emp, acct],
            edges=[edge_past, edge_future],
            transactions=[{
                "transaction_id": "txn_eval_01",
                "account_id": "acct_01",
                "amount": 500.0,
                "event_timestamp": t_eval.isoformat()
            }]
        )
        harness = PointInTimeReplayHarness(source)
        stream = list(harness.replay_stream(start_time=t_past, end_time=t_future))

        self.assertEqual(len(stream), 1)
        txn, nodes_as_of, edges_as_of = stream[0]
        edge_ids = [e.id for e in edges_as_of]
        self.assertIn("e_past", edge_ids)
        self.assertNotIn("e_future", edge_ids, "Future-dated edge leaked into replay stream!")

    def test_04_readiness_checker_privacy_and_pan_detection(self):
        """Verifies that raw PAN and CVV exposure are caught and block readiness."""
        emp = HeteroNode(id="emp_01", type=HeteroNodeType.EMPLOYEE, created_at=self.now)
        polluted_acct = HeteroNode(
            id="acct_polluted",
            type=HeteroNodeType.ACCOUNT,
            created_at=self.now,
            properties={"raw_pan": "4111222233334444", "cvv": "123"}
        )
        source = MockRealGraphDataSource(
            nodes=[emp, polluted_acct],
            dataset_type=DatasetType.REAL_SHADOW
        )
        checker = RealDataReadinessChecker(source)
        report = checker.audit_dataset(as_of=self.now)

        self.assertFalse(report.is_ready_for_real_validation)
        self.assertEqual(report.privacy_and_security["tokenization_compliance"], "FAILED")
        self.assertGreater(report.privacy_and_security["raw_pan_violations"], 0)

    def test_05_governance_gate_blocks_synthetic_and_insufficient_labels(self):
        """Verifies that synthetic data and datasets with < 50 real confirmed labels are blocked."""
        # 1. Synthetic dataset check
        source_synth = MockRealGraphDataSource(dataset_type=DatasetType.SYNTHETIC)
        checker_synth = RealDataReadinessChecker(source_synth)
        report_synth = checker_synth.audit_dataset(as_of=self.now)

        gate = DataRequiredGate(min_confirmed_fraud=50, min_confirmed_collusion=50)
        decision_synth = gate.evaluate_gate(report_synth)

        self.assertFalse(decision_synth.is_promotion_eligible)
        self.assertEqual(decision_synth.target_lifecycle_state, GraphSAGELifecycleState.REAL_DATA_REQUIRED)
        self.assertTrue(any("SYNTHETIC BOUNDARY" in r for r in decision_synth.blocking_reasons))

        # 2. Real dataset with only 12 confirmed fraud cases
        labels = [
            {"entity_id": f"txn_{i}", "maturity_state": LabelMaturityState.CONFIRMED_FRAUD.value, "label_type": "INTERNAL_COLLUSION"}
            for i in range(12)
        ]
        source_real = MockRealGraphDataSource(labels=labels, dataset_type=DatasetType.REAL_VALIDATION)
        checker_real = RealDataReadinessChecker(source_real)
        report_real = checker_real.audit_dataset(as_of=self.now)
        decision_real = gate.evaluate_gate(report_real)

        self.assertFalse(decision_real.is_promotion_eligible)
        self.assertEqual(decision_real.target_lifecycle_state, GraphSAGELifecycleState.REAL_DATA_REQUIRED)
        self.assertTrue(any("INSUFFICIENT REAL FRAUD LABELS" in r for r in decision_real.blocking_reasons))

    def test_06_collusion_dossier_generation(self):
        """Verifies creation of structured, privacy-preserving collusion dossiers."""
        emp = HeteroNode(id="emp_007", type=HeteroNodeType.EMPLOYEE, created_at=self.now, properties={"role": "customer_support"})
        acct = HeteroNode(id="acct_mule_99", type=HeteroNodeType.ACCOUNT, created_at=self.now)
        cons = HeteroNode(id="usr_mule_99", type=HeteroNodeType.CONSUMER, created_at=self.now)
        dev = HeteroNode(id="dev_colocated", type=HeteroNodeType.DEVICE, created_at=self.now)

        edge1 = HeteroEdge(id="e1", source_id="emp_007", target_id="acct_mule_99", type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT, timestamp=self.now)
        edge2 = HeteroEdge(id="e2", source_id="acct_mule_99", target_id="usr_mule_99", type=HeteroEdgeType.CONSUMER_OWNS_ACCOUNT, timestamp=self.now)

        shadow_out = ShadowInvestigationOutput(
            root_node_id="emp_007",
            root_node_type=HeteroNodeType.EMPLOYEE,
            graph_risk_score=0.88,
            collusion_risk_score=0.88,
            employee_risk_score=0.85,
            consumer_risk_score=0.40,
            neighborhood_anomaly_score=0.75,
            graph_poisoning_score=0.05,
            embedding=[0.0] * 64,
            top_related_entities=[{"id": "acct_mule_99", "type": "ACCOUNT"}],
            suspicious_paths=["emp_007 -> acct_mule_99 -> usr_mule_99"],
            relationship_evidence=["Repeated off-hours access before high-value transaction"],
            point_in_time_timestamp=self.now
        )

        gen = CollusionDossierGenerator()
        dossier = gen.generate_dossier(
            employee_node=emp,
            shadow_output=shadow_out,
            incident_edges=[edge1, edge2],
            associated_nodes=[acct, cons, dev],
            transaction_records=[{"transaction_id": "txn_01", "account_id": "acct_mule_99", "amount": 2500.0, "currency": "USD"}]
        )

        self.assertEqual(dossier.employee_id, "emp_007")
        self.assertEqual(dossier.employee_role, "customer_support")
        self.assertIn("acct_mule_99", dossier.affected_accounts)
        self.assertIn("usr_mule_99", dossier.affected_consumers)
        self.assertEqual(dossier.transaction_summary["total_amount_usd"], 2500.0)
        self.assertIn("STRICTLY NON-ENFORCING", dossier.governance_notice)

    def test_07_unseen_node_inductive_forward(self):
        """Verifies inductive forward inference on previously unseen nodes with no neighbors."""
        unseen_node = HeteroNode(id="acct_brand_new_999", type=HeteroNodeType.ACCOUNT, risk_score=0.05, created_at=self.now)
        emb, signals = self.model.forward(unseen_node, [], [])

        self.assertEqual(emb.shape, (64,))
        self.assertIn("relationship_collusion_risk", signals)
        self.assertIn("transaction_context_score", signals)
        self.assertLess(signals["transaction_context_score"], 0.20)


if __name__ == "__main__":
    unittest.main()
