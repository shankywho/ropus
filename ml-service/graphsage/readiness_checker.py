"""
Real-Data Readiness Checker & Audit Engine (Phase 58)
Evaluates candidate datasets across schema, privacy, temporal causality, topology, and label maturity.
"""

import re
from datetime import datetime, timezone
from typing import Dict, List, Any, Optional, Set
from collections import Counter

from .schema import (
    HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType,
    LabelMaturityState, DatasetType, GraphSAGELifecycleState,
    ReadinessAuditReport
)
from .data_source import RealGraphDataSource


class RealDataReadinessChecker:
    """
    Comprehensive Readiness & Governance Audit Engine for GraphSAGE real-data readiness.
    """

    def __init__(self, data_source: RealGraphDataSource):
        self.data_source = data_source

    def audit_dataset(self, as_of: Optional[datetime] = None) -> ReadinessAuditReport:
        """
        Executes complete multi-dimensional audit of the data source.
        """
        audited_at = datetime.now(timezone.utc)
        dataset_type = self.data_source.get_dataset_type()
        blocking_reasons: List[str] = []

        nodes = self.data_source.get_nodes(as_of=as_of)
        edges = self.data_source.get_edges(as_of=as_of)
        txns = self.data_source.get_transactions()
        labels = self.data_source.get_labels(mature_only=False)
        investigations = self.data_source.get_investigation_events()

        node_map = {n.id: n for n in nodes}

        # 1. Schema Completeness & Referential Integrity
        missing_fields_count = 0
        invalid_edge_refs = 0
        missing_ts_count = 0

        for n in nodes:
            if not n.id or not n.type:
                missing_fields_count += 1
            if not n.created_at:
                missing_ts_count += 1

        for e in edges:
            if not e.id or not e.source_id or not e.target_id or not e.type:
                missing_fields_count += 1
            if not e.timestamp:
                missing_ts_count += 1
            if e.source_id not in node_map or e.target_id not in node_map:
                invalid_edge_refs += 1

        schema_status = {
            "total_nodes": len(nodes),
            "total_edges": len(edges),
            "total_transactions": len(txns),
            "missing_required_fields": missing_fields_count,
            "missing_timestamps": missing_ts_count,
            "invalid_edge_references": invalid_edge_refs,
            "status": "PASSED" if (missing_fields_count == 0 and invalid_edge_refs == 0) else "FAILED"
        }
        if invalid_edge_refs > 0:
            blocking_reasons.append(f"Referential integrity failure: {invalid_edge_refs} edges reference missing nodes")

        # 2. Temporal Causality Audit
        future_edges_count = 0
        premature_labels_count = 0

        if as_of:
            for e in edges:
                if e.timestamp > as_of:
                    future_edges_count += 1

        # Check label timestamps vs transaction event timestamps
        txn_ts_map = {}
        for t in txns:
            ts = t.get("event_timestamp")
            if isinstance(ts, str):
                ts = datetime.fromisoformat(ts)
            txn_ts_map[t["transaction_id"]] = ts

        for l in labels:
            eid = l.get("entity_id")
            l_ts = l.get("label_timestamp")
            if isinstance(l_ts, str):
                l_ts = datetime.fromisoformat(l_ts)
            if eid in txn_ts_map and l_ts:
                t_ts = txn_ts_map[eid]
                if t_ts and l_ts <= t_ts:
                    premature_labels_count += 1

        temporal_status = {
            "future_edges_detected": future_edges_count,
            "premature_labels_detected": premature_labels_count,
            "temporal_causality_violations": future_edges_count + premature_labels_count,
            "status": "PASSED" if (future_edges_count == 0 and premature_labels_count == 0) else "FAILED"
        }
        if premature_labels_count > 0:
            blocking_reasons.append(f"Temporal causality violation: {premature_labels_count} labels timestamped <= event timestamp")

        # 3. Privacy & Security Audit (Zero Raw PAN, CVV, PII)
        pan_regex = re.compile(r"\b(?:\d[ -]*?){13,19}\b")
        cvv_regex = re.compile(r"\b\d{3,4}\b")
        pan_violations = 0
        cvv_violations = 0
        pii_violations = 0

        for n in nodes:
            props_str = str(n.properties) + str(n.id)
            if "raw_pan" in props_str.lower() or "card_number" in props_str.lower():
                pan_violations += 1
            if "cvv" in props_str.lower() or "cvv2" in props_str.lower():
                cvv_violations += 1
            if "ssn" in props_str.lower() or "password" in props_str.lower():
                pii_violations += 1

        privacy_status = {
            "raw_pan_violations": pan_violations,
            "raw_cvv_violations": cvv_violations,
            "pii_violations": pii_violations,
            "tokenization_compliance": "PASSED" if (pan_violations == 0 and cvv_violations == 0 and pii_violations == 0) else "FAILED"
        }
        if pan_violations > 0 or cvv_violations > 0:
            blocking_reasons.append(f"Security/Privacy violation: Raw payment card data detected in graph properties")

        # 4. Graph Topology & Density Analysis
        degrees: Dict[str, int] = Counter()
        for e in edges:
            degrees[e.source_id] += 1
            degrees[e.target_id] += 1

        excessive_degree_nodes = [nid for nid, deg in degrees.items() if deg > 500]
        isolated_nodes_count = sum(1 for n in nodes if degrees[n.id] == 0)

        topology_status = {
            "total_nodes": len(nodes),
            "total_edges": len(edges),
            "isolated_nodes": isolated_nodes_count,
            "excessive_degree_nodes_count": len(excessive_degree_nodes),
            "max_degree_observed": max(degrees.values()) if degrees else 0,
            "average_degree": round(sum(degrees.values()) / max(len(nodes), 1), 2)
        }

        # 5. Entity Type Coverage
        type_counts = Counter(n.type.value if hasattr(n.type, "value") else str(n.type) for n in nodes)
        entity_coverage = {
            "EMPLOYEE_count": type_counts.get("EMPLOYEE", 0),
            "CONSUMER_count": type_counts.get("CONSUMER", 0),
            "ACCOUNT_count": type_counts.get("ACCOUNT", 0),
            "DEVICE_count": type_counts.get("DEVICE", 0),
            "IP_count": type_counts.get("IP", 0),
            "PAYMENT_TOKEN_count": type_counts.get("PAYMENT_TOKEN", 0),
            "MERCHANT_count": type_counts.get("MERCHANT", 0),
            "CASE_count": type_counts.get("CASE", 0),
        }

        # 6. Label Maturity Coverage
        maturity_counts = Counter(l.get("maturity_state", "UNLABELED") for l in labels)
        confirmed_fraud = maturity_counts.get(LabelMaturityState.CONFIRMED_FRAUD.value, 0)
        confirmed_legit = maturity_counts.get(LabelMaturityState.CONFIRMED_LEGITIMATE.value, 0)
        suspected = maturity_counts.get(LabelMaturityState.SUSPECTED.value, 0)
        under_inv = maturity_counts.get(LabelMaturityState.UNDER_INVESTIGATION.value, 0)
        unlabeled = maturity_counts.get(LabelMaturityState.UNLABELED.value, 0)

        # Count confirmed internal collusion specifically
        collusion_count = sum(
            1 for l in labels
            if l.get("maturity_state") == LabelMaturityState.CONFIRMED_FRAUD.value
            and (l.get("label_type") == "INTERNAL_COLLUSION" or l.get("is_collusion", False))
        )

        label_coverage = {
            "UNLABELED": unlabeled,
            "SUSPECTED": suspected,
            "UNDER_INVESTIGATION": under_inv,
            "CONFIRMED_FRAUD": confirmed_fraud,
            "CONFIRMED_LEGITIMATE": confirmed_legit,
            "CONFIRMED_INTERNAL_COLLUSION": collusion_count,
            "total_labels": len(labels)
        }

        # 7. Promotion Gate Evaluation
        min_required_fraud = 50
        min_required_collusion = 50

        if dataset_type == DatasetType.SYNTHETIC:
            blocking_reasons.append("PROMOTION BLOCKED: Dataset is SYNTHETIC. Real-world validation required.")
        elif confirmed_fraud < min_required_fraud:
            blocking_reasons.append(f"PROMOTION BLOCKED: Insufficient real confirmed fraud cases ({confirmed_fraud}/{min_required_fraud})")
        elif collusion_count < min_required_collusion:
            blocking_reasons.append(f"PROMOTION BLOCKED: Insufficient real internal collusion cases ({collusion_count}/{min_required_collusion})")

        is_ready = len(blocking_reasons) == 0

        governance_state = (
            GraphSAGELifecycleState.REAL_DATA_REQUIRED
            if not is_ready
            else GraphSAGELifecycleState.SHADOW_EVALUATION
        )

        return ReadinessAuditReport(
            audited_at=audited_at,
            dataset_type=dataset_type,
            is_ready_for_real_validation=is_ready,
            governance_state=governance_state,
            blocking_reasons=blocking_reasons,
            schema_completeness=schema_status,
            temporal_causality=temporal_status,
            privacy_and_security=privacy_status,
            topology_statistics=topology_status,
            entity_coverage=entity_coverage,
            label_maturity_coverage=label_coverage,
            confirmed_fraud_labels_count=confirmed_fraud,
            confirmed_internal_collusion_labels_count=collusion_count
        )
