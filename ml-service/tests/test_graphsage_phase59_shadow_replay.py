"""
Phase 59 Test Suite: Real-Data Shadow Replay, Employee Behavioral Baselines, Path Extraction & Governance
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
    ShadowInvestigationOutput, RelationshipRiskLevel
)
from graphsage.shadow_replay import generate_realistic_shadow_fixture
from graphsage.data_source import PointInTimeReplayHarness
from graphsage.path_extractor import RelationshipPathExtractor
from graphsage.relationship_features import RelationshipFeatureExtractor
from graphsage.employee_baseline import EmployeeBehavioralBaselineEngine
from graphsage.calibration import ShadowCollusionCalibrator
from graphsage.drift_monitor import ShadowGraphDriftMonitor
from graphsage.dossier_generator import CollusionDossierGenerator
from graphsage.model import GraphSAGEModel

CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"


class TestGraphSAGEPhase59ShadowReplay(unittest.TestCase):

    def setUp(self):
        self.fixture = generate_realistic_shadow_fixture()
        self.model = GraphSAGEModel(input_dim=32, hidden_dim=64, output_dim=64, seed=42)
        self.eval_time = datetime(2026, 8, 3, 12, 0, 0, tzinfo=timezone.utc)

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

    def test_02_point_in_time_replay_and_future_leakage_rejection(self):
        """Proves that future-dated edges injected at T + 1h are strictly excluded from replay at T."""
        t_eval = datetime(2026, 8, 2, 10, 0, 0, tzinfo=timezone.utc)
        t_future = t_eval + timedelta(hours=1)

        # Inject an adversarial future edge
        future_edge = HeteroEdge(
            id="e_future_collusion",
            source_id="emp_support_alice",
            target_id="acct_mule_01",
            type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
            timestamp=t_future
        )
        self.fixture.add_edge(future_edge)

        nodes_as_of = self.fixture.get_nodes(as_of=t_eval)
        edges_as_of = self.fixture.get_edges(as_of=t_eval)

        edge_ids = [e.id for e in edges_as_of]
        self.assertNotIn("e_future_collusion", edge_ids, "Adversarial future edge leaked into point-in-time slice!")

    def test_03_multi_hop_path_extraction(self):
        """Verifies extraction of directed relationship chains (Employee -> Account -> Consumer)."""
        nodes = {n.id: n for n in self.fixture.get_nodes()}
        edges = self.fixture.get_edges()

        extractor = RelationshipPathExtractor()
        paths = extractor.find_paths(
            source_id="emp_rogue_charlie",
            target_id="usr_mule_01",
            nodes=nodes,
            edges=edges,
            max_depth=3
        )

        self.assertGreater(len(paths), 0, "Failed to discover multi-hop path between rogue employee and mule consumer")
        first_path = paths[0]
        self.assertIn("emp_rogue_charlie", first_path["path_string"])
        self.assertIn("usr_mule_01", first_path["path_string"])
        self.assertGreaterEqual(first_path["hops"], 2)

    def test_04_relationship_features_and_hard_negative_discrimination(self):
        """Verifies that legitimate support staff has low concentration vs high concentration for rogue staff."""
        nodes = {n.id: n for n in self.fixture.get_nodes()}
        edges = self.fixture.get_edges()
        txns = self.fixture.get_transactions()

        extractor = RelationshipFeatureExtractor()

        # Alice: Legitimate Support (20 accounts touched)
        alice_edges = [e for e in edges if e.source_id == "emp_support_alice"]
        alice_feats = extractor.extract_features("emp_support_alice", nodes, alice_edges, txns)

        # Charlie: Rogue Support (3 mule accounts repeatedly touched)
        charlie_edges = [e for e in edges if e.source_id == "emp_rogue_charlie"]
        charlie_feats = extractor.extract_features("emp_rogue_charlie", nodes, charlie_edges, txns)

        # Concentration check: Charlie's HHI concentration must be significantly higher than Alice's
        self.assertLess(alice_feats["employee_account_concentration"], 0.15, "Support agent should have low account concentration")
        self.assertGreater(charlie_feats["employee_account_concentration"], 0.30, "Rogue employee should have high account concentration")
        self.assertGreater(charlie_feats["employee_approval_after_access_count"], 0, "Rogue employee should have self-approval events")

    def test_05_employee_behavioral_baseline_detection(self):
        """Verifies that David's 02:30 AM access triggers personal anomaly deviation."""
        edges = self.fixture.get_edges()
        engine = EmployeeBehavioralBaselineEngine()

        # Generate 10 days of normal 10:00 AM history for David
        hist_edges = []
        for d in range(10, 1, -1):
            hist_edges.append(HeteroEdge(
                id=f"e_hist_{d}",
                source_id="emp_rogue_david",
                target_id="acct_legit_001",
                type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
                timestamp=datetime(2026, 7, 20 + d, 10, 0, 0, tzinfo=timezone.utc)
            ))

        david_current_edges = [e for e in edges if e.source_id == "emp_rogue_david"]
        baseline_res = engine.compute_baseline_and_deviation(
            employee_id="emp_rogue_david",
            current_events=david_current_edges,
            historical_edges=hist_edges,
            as_of=self.eval_time
        )

        self.assertTrue(baseline_res["has_historical_baseline"])
        self.assertTrue(baseline_res["is_unusual_hour_for_employee"], "02:30 AM access should be flagged as unusual for employee")
        self.assertGreater(baseline_res["personal_anomaly_delta"], 0.30)

    def test_06_dossier_supporting_and_counter_evidence(self):
        """Verifies balanced evidence generation: counter-evidence for support role, risk evidence for collusion."""
        nodes = {n.id: n for n in self.fixture.get_nodes()}
        edges = self.fixture.get_edges()
        txns = self.fixture.get_transactions()

        dossier_gen = CollusionDossierGenerator()

        # 1. Generate dossier for Alice (Legitimate Support)
        emp_alice = nodes["emp_support_alice"]
        alice_edges = [e for e in edges if e.source_id == "emp_support_alice"]
        alice_shadow_out = ShadowInvestigationOutput(
            root_node_id="emp_support_alice",
            root_node_type=HeteroNodeType.EMPLOYEE,
            graph_risk_score=0.12,
            collusion_risk_score=0.10,
            employee_risk_score=0.10,
            consumer_risk_score=0.05,
            neighborhood_anomaly_score=0.08,
            graph_poisoning_score=0.01,
            embedding=[0.0] * 64,
            top_related_entities=[],
            suspicious_paths=[],
            relationship_evidence=[],
            point_in_time_timestamp=self.eval_time
        )

        dossier_alice = dossier_gen.generate_dossier(
            employee_node=emp_alice,
            shadow_output=alice_shadow_out,
            incident_edges=alice_edges,
            associated_nodes=[nodes[f"acct_legit_{i:03d}"] for i in range(1, 21)],
            transaction_records=txns
        )

        self.assertIn("counter_evidence", dossier_alice.temporal_evidence)
        counter_texts = " ".join(dossier_alice.temporal_evidence["counter_evidence"])
        self.assertIn("customer_support", counter_texts)
        self.assertEqual(dossier_alice.risk_level, RelationshipRiskLevel.LOW)

        # 2. Generate dossier for Charlie (Rogue Support)
        emp_charlie = nodes["emp_rogue_charlie"]
        charlie_edges = [e for e in edges if e.source_id == "emp_rogue_charlie"]
        charlie_shadow_out = ShadowInvestigationOutput(
            root_node_id="emp_rogue_charlie",
            root_node_type=HeteroNodeType.EMPLOYEE,
            graph_risk_score=0.88,
            collusion_risk_score=0.88,
            employee_risk_score=0.85,
            consumer_risk_score=0.75,
            neighborhood_anomaly_score=0.80,
            graph_poisoning_score=0.05,
            embedding=[0.0] * 64,
            top_related_entities=[],
            suspicious_paths=[],
            relationship_evidence=[],
            point_in_time_timestamp=self.eval_time
        )

        dossier_charlie = dossier_gen.generate_dossier(
            employee_node=emp_charlie,
            shadow_output=charlie_shadow_out,
            incident_edges=charlie_edges,
            associated_nodes=[nodes["acct_mule_01"], nodes["acct_mule_02"], nodes["acct_mule_03"]],
            transaction_records=txns
        )

        self.assertIn("supporting_evidence", dossier_charlie.temporal_evidence)
        support_texts = " ".join(dossier_charlie.temporal_evidence["supporting_evidence"])
        self.assertIn("concentration", support_texts.lower())
        self.assertEqual(dossier_charlie.risk_level, RelationshipRiskLevel.HIGH)

    def test_07_shadow_calibration_governance_gate(self):
        """Verifies that score calibration is DATA_REQUIRED when confirmed labels < 50."""
        calibrator = ShadowCollusionCalibrator(confirmed_labels_count=1)
        res = calibrator.calibrate(raw_score=0.85, relationship_score=0.80)

        self.assertEqual(res.calibration_status, "DATA_REQUIRED")
        self.assertIsNone(res.calibrated_score)
        self.assertIn("CALIBRATION INACTIVE", res.governance_notice)

    def test_08_shadow_graph_drift_monitoring(self):
        """Verifies computation of data drift, topology drift, score drift, and label drift."""
        nodes = self.fixture.get_nodes()
        edges = self.fixture.get_edges()
        scores = [0.05] * 25 + [0.85] * 5

        monitor = ShadowGraphDriftMonitor()
        report = monitor.compute_drift_report(
            current_nodes=nodes,
            current_edges=edges,
            current_scores=scores,
            baseline_stats={"average_degree": 4.5, "mean_score": 0.15}
        )

        self.assertIn("topology_drift", report)
        self.assertIn("score_drift", report)
        self.assertEqual(report["governance_action"], "NO_RETRAINING_TRIGGERED (Shadow Monitoring Only)")


if __name__ == "__main__":
    unittest.main()
