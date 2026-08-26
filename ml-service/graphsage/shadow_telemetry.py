"""
Production Shadow Telemetry Engine (Phase 61)
Coordinates the end-to-end non-blocking shadow telemetry plane:
Stream Ingestion -> Privacy Sanitizer -> Point-in-Time Graph -> GNN Forward -> Evidence Ledger -> Label Maturation.
"""

import time
import threading
from queue import Queue, Empty, Full
from datetime import datetime, timezone, timedelta
from typing import Dict, List, Any, Optional, Tuple
from pydantic import BaseModel, Field

from .schema import (
    HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType,
    LabelMaturityState, DatasetType, GraphSAGELifecycleState,
    RelationshipRiskLevel, CollusionInvestigationDossier
)
from .privacy_sanitizer import PrivacySanitizer
from .label_maturation import LabelMaturationEngine, MaturityReport
from .evidence_ledger import ShadowEvidenceLedger, EvidenceLedgerEntry
from .collusion_tracker import CollusionAccumulationTracker, RealCollusionSummaryView
from .passive_ingestion import PassiveShadowIngestionAdapter, ShadowIngestionMetrics
from .drift_monitor import ShadowGraphDriftMonitor
from .employee_baseline import EmployeeBehavioralBaselineEngine
from .path_extractor import RelationshipPathExtractor
from .relationship_features import RelationshipFeatureExtractor
from .model import GraphSAGEModel
from .data_source import MockRealGraphDataSource


class ShadowTelemetryStatus(BaseModel):
    is_running: bool
    real_data_connectivity: str = "NOT_CONNECTED"
    buffered_events_count: int
    queue_capacity: int
    total_events_processed: int
    total_dossiers_generated: int
    total_dlq_events: int
    graph_node_count: int
    graph_edge_count: int
    drift_status: str
    drift_summary: Dict[str, Any] = Field(default_factory=dict)
    governance_state: str = "REAL_DATA_REQUIRED"
    operational_mode: str = "STRICTLY_NON_ENFORCING_SHADOW"
    customer_enforcement_authority: str = "0% (Baseline Dynamic BMR 100% Authoritative)"


class PassiveShadowTelemetryEngine:
    """
    Production-grade shadow telemetry runtime.
    Processes stream events asynchronously without impacting the customer transaction critical path.
    """

    def __init__(
        self,
        queue_size: int = 10000,
        is_live_connected: bool = False,
        data_source: Optional[MockRealGraphDataSource] = None
    ):
        self.queue_size = queue_size
        self.is_live_connected = is_live_connected
        self.event_queue: Queue = Queue(maxsize=queue_size)
        self.is_running: bool = False

        self.data_source = data_source or MockRealGraphDataSource(dataset_type=DatasetType.REAL_SHADOW)
        self.adapter = PassiveShadowIngestionAdapter(
            target_data_source=self.data_source,
            is_live_connected=is_live_connected
        )
        self.maturation_engine = LabelMaturationEngine(dispute_window_days=90)
        self.evidence_ledger = ShadowEvidenceLedger()
        self.collusion_tracker = CollusionAccumulationTracker()
        self.drift_monitor = ShadowGraphDriftMonitor()
        self.baseline_engine = EmployeeBehavioralBaselineEngine()
        self.path_extractor = RelationshipPathExtractor()
        self.feature_extractor = RelationshipFeatureExtractor()
        self.gnn_model = GraphSAGEModel()
        self.recorded_scores: List[float] = []

        self.total_processed: int = 0
        self.total_dossiers: int = 0
        self._worker_thread: Optional[threading.Thread] = None

    def start(self):
        """Starts the background telemetry worker thread."""
        if not self.is_running:
            self.is_running = True
            self._worker_thread = threading.Thread(target=self._process_queue_loop, daemon=True)
            self._worker_thread.start()

    def stop(self, timeout: float = 2.0):
        """Gracefully stops background queue processing."""
        self.is_running = False
        if self._worker_thread and self._worker_thread.is_alive():
            self._worker_thread.join(timeout=timeout)

    def enqueue_event(
        self,
        event_id: str,
        topic: str,
        payload: Dict[str, Any],
        event_timestamp: datetime
    ) -> bool:
        """
        Non-blocking enqueue for incoming production stream events.
        Never blocks the production caller; drops to backpressure counter if queue is full.
        """
        try:
            self.event_queue.put_nowait((event_id, topic, payload, event_timestamp))
            return True
        except Full:
            # Backpressure handling: record dropped event in metrics without raising
            self.adapter.metrics.events_rejected += 1
            return False

    def process_event_sync(
        self,
        event_id: str,
        topic: str,
        payload: Dict[str, Any],
        event_timestamp: datetime
    ) -> Tuple[bool, str]:
        """Synchronously processes an event (for offline evaluation/replay/testing)."""
        ok, msg = self.adapter.consume_event(event_id, topic, payload, event_timestamp)
        if ok:
            self.total_processed += 1
            if topic == "audit.events":
                emp_id = payload.get("employee_id")
                acct_id = payload.get("account_id")
                if emp_id and acct_id:
                    edge = HeteroEdge(
                        id=event_id,
                        source_id=emp_id,
                        target_id=acct_id,
                        type=HeteroEdgeType.EMPLOYEE_ACCESSES_ACCOUNT,
                        timestamp=event_timestamp
                    )
                    self.collusion_tracker.process_edge(edge)
            elif topic == "transactions.created":
                txn_id = payload.get("transaction_id")
                if txn_id:
                    self.maturation_engine.register_transaction(txn_id, event_timestamp)
        return ok, msg

    def _process_queue_loop(self):
        """Worker loop draining event queue."""
        while self.is_running:
            try:
                event_id, topic, payload, ts = self.event_queue.get(timeout=0.1)
                self.process_event_sync(event_id, topic, payload, ts)
                self.event_queue.task_done()
            except Empty:
                continue

    def evaluate_shadow_investigation(
        self,
        employee_id: str,
        eval_timestamp: datetime,
        employee_role: str = "customer_support"
    ) -> Optional[CollusionInvestigationDossier]:
        """
        Executes a point-in-time shadow investigation and records the dossier in the evidence ledger.
        STRICTLY NON-ENFORCING.
        """
        # 1. Point-in-time subgraph extraction (t <= eval_timestamp)
        subgraph = self.data_source.get_point_in_time_subgraph(
            center_node_id=employee_id,
            as_of=eval_timestamp,
            depth=2
        )
        if not subgraph or employee_id not in subgraph.nodes:
            return None

        # 2. Extract multi-hop explainability paths
        paths = self.path_extractor.find_paths(
            source_id=employee_id,
            target_id=None,
            nodes=subgraph.nodes,
            edges=subgraph.edges,
            max_depth=3
        )

        # 3. Extract explicit relationship features
        features = self.feature_extractor.extract_features(
            employee_id=employee_id,
            nodes=subgraph.nodes,
            incident_edges=subgraph.edges
        )

        # 4. GNN Forward pass
        root_node = subgraph.nodes[employee_id]
        l1_nodes = [subgraph.nodes[e.target_id] for e in subgraph.edges if e.source_id == employee_id and e.target_id in subgraph.nodes]
        l2_nodes = []
        for l1 in l1_nodes:
            l2_nodes.extend([subgraph.nodes[e.target_id] for e in subgraph.edges if e.source_id == l1.id and e.target_id in subgraph.nodes])

        embedding, signals = self.gnn_model.forward(root_node, l1_nodes, l2_nodes)
        gnn_score = signals.get("relationship_collusion_risk", 0.05)

        # Combine structural and GNN risk score
        collusion_struct_score = features["neighborhood_anomaly_score"]
        risk_score = round(float(0.4 * gnn_score + 0.6 * collusion_struct_score), 4)
        risk_level = (
            RelationshipRiskLevel.CRITICAL if risk_score >= 0.85 else
            RelationshipRiskLevel.HIGH if risk_score >= 0.70 else
            RelationshipRiskLevel.MEDIUM if risk_score >= 0.40 else
            RelationshipRiskLevel.LOW
        )

        # 5. Build Dossier with supporting and counter evidence
        supporting = []
        counter = []

        hhi = features["employee_account_concentration"]
        self_apprs = int(features["employee_approval_after_access_count"])
        off_hours = features["off_hours_ratio"]
        unique_accts = int(features["employee_consumer_unique_count"])

        if hhi > 0.25:
            supporting.append(f"Elevated account concentration (HHI: {hhi:.2f}).")
        if self_apprs > 0:
            supporting.append(f"Direct segregation-of-duties violation: {self_apprs} self-approvals.")
        if off_hours > 0.4:
            supporting.append(f"Abnormal off-hours activity ratio ({off_hours * 100:.1f}%).")

        if employee_role in ["customer_support", "fraud_operations"]:
            counter.append(f"Authorized business role ({employee_role}) explains elevated account contact volume.")
        if hhi < 0.10 and unique_accts >= 10:
            counter.append("High account dispersion matches typical support queue distribution.")

        dossier = CollusionInvestigationDossier(
            dossier_id=f"dos_{eval_timestamp.strftime('%Y%m%d%H%M%S')}_{employee_id}",
            generated_at=eval_timestamp,
            overall_risk_score=risk_score,
            risk_level=risk_level,
            employee_id=employee_id,
            employee_role=employee_role,
            affected_accounts=[e.target_id for e in subgraph.edges if e.source_id == employee_id and e.target_id.startswith("acct_")],
            affected_consumers=[n.id for n in subgraph.nodes.values() if n.type == HeteroNodeType.CONSUMER],
            relationship_paths=[p.get("path_string", str(p)) for p in paths[:5]],
            temporal_evidence={"supporting_evidence": supporting, "counter_evidence": counter},
            transaction_summary={"total_transactions_count": int(features["employee_account_access_count"]), "total_amount_usd": 0.0},
            shared_infrastructure={"device_overlap_count": int(features["shared_device_overlap_count"]), "ip_overlap_count": int(features["shared_ip_overlap_count"])},
            confidence_score=0.90,
            human_explanation=f"Shadow investigation dossier for {employee_id}. Risk level: {risk_level.value}.",
            data_provenance="passive_shadow_stream",
            model_version="graphsage-v1.0-shadow",
            governance_notice="INVESTIGATION INTELLIGENCE ONLY - STRICTLY NON-ENFORCING"
        )

        # 6. Record to persistent evidence ledger
        self.evidence_ledger.record_dossier(dossier)
        self.recorded_scores.append(risk_score)
        self.total_dossiers += 1

        return dossier

    def get_telemetry_status(self) -> ShadowTelemetryStatus:
        conn = "CONNECTED_SHADOW" if self.is_live_connected else "NOT_CONNECTED"
        drift_rep = self.drift_monitor.compute_drift_report(
            current_nodes=list(self.data_source.nodes.values()),
            current_edges=self.data_source.edges,
            current_scores=self.recorded_scores
        )
        ing_metrics = self.adapter.get_metrics()

        return ShadowTelemetryStatus(
            is_running=self.is_running,
            real_data_connectivity=conn,
            buffered_events_count=self.event_queue.qsize(),
            queue_capacity=self.queue_size,
            total_events_processed=self.total_processed,
            total_dossiers_generated=self.total_dossiers,
            total_dlq_events=ing_metrics.dlq_size,
            graph_node_count=len(self.data_source.nodes),
            graph_edge_count=len(self.data_source.edges),
            drift_status=drift_rep.get("governance_action", "NO_RETRAINING_TRIGGERED"),
            drift_summary=drift_rep,
            governance_state="REAL_DATA_REQUIRED (0/50 Real Collusion Cases)"
        )
