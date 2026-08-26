"""
Collusion-Specific Investigation Dossier Generator & Explainability Engine (Phase 59)
Produces structured, tokenized, machine-readable investigation dossiers with balanced supporting and counter-evidence.
"""

import uuid
from datetime import datetime, timezone
from typing import Dict, List, Any, Optional

from .schema import (
    HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType,
    RelationshipRiskLevel, CollusionInvestigationDossier,
    ShadowInvestigationOutput
)
from .path_extractor import RelationshipPathExtractor
from .relationship_features import RelationshipFeatureExtractor
from .employee_baseline import EmployeeBehavioralBaselineEngine


class CollusionDossierGenerator:
    """
    Generates privacy-preserving, explainable collusion dossiers with supporting and counter-evidence.
    Strictly marked NON_ENFORCING / INVESTIGATION_ONLY.
    """

    def __init__(self):
        self.path_extractor = RelationshipPathExtractor()
        self.feature_extractor = RelationshipFeatureExtractor()
        self.baseline_engine = EmployeeBehavioralBaselineEngine()

    def generate_dossier(
        self,
        employee_node: HeteroNode,
        shadow_output: ShadowInvestigationOutput,
        incident_edges: List[HeteroEdge],
        associated_nodes: List[HeteroNode],
        historical_edges: Optional[List[HeteroEdge]] = None,
        transaction_records: Optional[List[Dict[str, Any]]] = None
    ) -> CollusionInvestigationDossier:
        dossier_id = f"dos_collusion_{uuid.uuid4().hex[:12]}"
        now = datetime.now(timezone.utc)

        # 1. Identify affected accounts, consumers, and infrastructure
        node_map = {n.id: n for n in associated_nodes}
        node_map[employee_node.id] = employee_node

        affected_accts: List[str] = []
        affected_cons: List[str] = []
        shared_devices: List[str] = []
        shared_ips: List[str] = []

        for n in associated_nodes:
            if n.type == HeteroNodeType.ACCOUNT:
                affected_accts.append(n.id)
            elif n.type == HeteroNodeType.CONSUMER:
                affected_cons.append(n.id)
            elif n.type == HeteroNodeType.DEVICE:
                shared_devices.append(n.id)
            elif n.type == HeteroNodeType.IP:
                shared_ips.append(n.id)

        # 2. Extract Multi-Hop Relationship Paths
        paths = self.path_extractor.find_paths(
            source_id=employee_node.id,
            target_id=None,
            nodes=node_map,
            edges=incident_edges,
            max_depth=3
        )
        path_strings = [p["path_string"] for p in paths] if paths else [
            f"({employee_node.id}) -[INTERACTS]-> ({n.id})" for n in associated_nodes[:5]
        ]

        # 3. Extract Explicit Relationship Features
        rel_feats = self.feature_extractor.extract_features(
            employee_id=employee_node.id,
            nodes=node_map,
            incident_edges=incident_edges,
            associated_txns=transaction_records
        )

        # 4. Compute Employee Behavioral Baseline Deviation
        hist = historical_edges or []
        baseline_info = self.baseline_engine.compute_baseline_and_deviation(
            employee_id=employee_node.id,
            current_events=incident_edges,
            historical_edges=hist,
            as_of=now
        )

        # 5. Formulate Supporting Evidence vs Counter-Evidence
        supporting_evidence: List[str] = []
        counter_evidence: List[str] = []

        # Supporting evidence rules
        if rel_feats["employee_account_concentration"] > 0.30:
            supporting_evidence.append(
                f"Elevated account concentration (HHI: {rel_feats['employee_account_concentration']:.2f}): "
                f"access is tightly clustered around {len(affected_accts)} specific account(s)."
            )
        if rel_feats["off_hours_ratio"] > 0.20:
            supporting_evidence.append(
                f"Unusual temporal access: {rel_feats['off_hours_ratio']*100:.1f}% of interactions occurred outside normal business hours."
            )
        if rel_feats["employee_transaction_proximity_sec"] < 120.0:
            supporting_evidence.append(
                f"Immediate pre-transaction proximity: account access occurred {rel_feats['employee_transaction_proximity_sec']:.0f}s before transaction creation."
            )
        if rel_feats["employee_approval_after_access_count"] > 0:
            supporting_evidence.append(
                f"Self-approval / Segregation-of-Duties anomaly: employee approved {int(rel_feats['employee_approval_after_access_count'])} transaction(s) on previously accessed accounts."
            )
        if rel_feats["shared_device_overlap_count"] > 0:
            supporting_evidence.append(
                f"Shared hardware infrastructure: {int(rel_feats['shared_device_overlap_count'])} device(s) co-located between employee and consumer sessions."
            )
        if baseline_info["personal_anomaly_delta"] > 0.30:
            supporting_evidence.append(
                f"Deviation from personal historical baseline: anomaly delta of {baseline_info['personal_anomaly_delta']:.2f} compared to past working habits."
            )

        # Counter-evidence rules (Benign Explanations)
        role = str(employee_node.properties.get("role", "unknown"))
        if role in ["customer_support", "fraud_analyst", "risk_operations"]:
            counter_evidence.append(
                f"Employee role ({role}) routinely accesses customer accounts as part of standard support and case review workflows."
            )
        if rel_feats["employee_account_concentration"] < 0.10 and len(affected_accts) > 10:
            counter_evidence.append(
                f"High account dispersion ({len(affected_accts)} accounts touched): behavior is consistent with standard high-volume customer service queue handling."
            )
        if rel_feats["off_hours_ratio"] == 0.0:
            counter_evidence.append("All access occurred strictly within standard business operating hours.")
        if rel_feats["shared_device_overlap_count"] == 0 and rel_feats["shared_ip_overlap_count"] == 0:
            counter_evidence.append("Zero hardware or IP colocation detected between employee and consumer entities.")
        if not supporting_evidence:
            supporting_evidence.append("Baseline interactions detected with no structural anomalies.")

        # 6. Transaction summary
        txns = transaction_records or []
        total_amt = sum(float(t.get("amount", 0.0)) for t in txns)
        currencies = list(set(t.get("currency", "USD") for t in txns))

        txn_summary = {
            "total_transactions_count": len(txns),
            "total_amount_usd": round(total_amt, 2),
            "currencies": currencies,
            "account_ids": list(set(t.get("account_id") for t in txns if t.get("account_id")))
        }

        # 7. Risk Tiering & Narrative
        risk_score = shadow_output.collusion_risk_score
        if risk_score >= 0.90:
            risk_level = RelationshipRiskLevel.CRITICAL
        elif risk_score >= 0.70:
            risk_level = RelationshipRiskLevel.HIGH
        elif risk_score >= 0.30:
            risk_level = RelationshipRiskLevel.MEDIUM
        else:
            risk_level = RelationshipRiskLevel.LOW

        explanation = (
            f"Investigation Dossier for Employee {employee_node.id} ({role}). "
            f"Overall Relationship Risk Score: {risk_score:.2f} (Tier: {risk_level.value}). "
            f"Identified {len(supporting_evidence)} risk indicator(s) and {len(counter_evidence)} mitigating counter-evidence factor(s). "
            f"Evaluated point-in-time for shadow review only."
        )

        return CollusionInvestigationDossier(
            dossier_id=dossier_id,
            generated_at=now,
            overall_risk_score=risk_score,
            risk_level=risk_level,
            employee_id=employee_node.id,
            employee_role=role,
            affected_accounts=affected_accts,
            affected_consumers=affected_cons,
            relationship_paths=path_strings[:10],
            temporal_evidence={
                "total_interaction_events": len(incident_edges),
                "relationship_features": rel_feats,
                "behavioral_baseline": baseline_info,
                "supporting_evidence": supporting_evidence,
                "counter_evidence": counter_evidence
            },
            transaction_summary=txn_summary,
            shared_infrastructure={
                "device_nodes": shared_devices,
                "ip_nodes": shared_ips,
                "device_overlap_count": len(shared_devices),
                "ip_overlap_count": len(shared_ips)
            },
            confidence_score=round(min(1.0, 0.50 + len(incident_edges) * 0.05), 2),
            human_explanation=explanation,
            data_provenance="shadow_graph_stream",
            model_version=shadow_output.model_version,
            governance_notice="INVESTIGATION INTELLIGENCE ONLY - STRICTLY NON-ENFORCING"
        )
