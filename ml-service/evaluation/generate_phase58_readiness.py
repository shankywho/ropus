#!/usr/bin/env python3
"""
Phase 58: Real-Data Shadow Readiness Audit Generator
Audits dataset readiness, evaluates the DATA_REQUIRED gate, and exports governance scorecard.
"""

import os
import sys
import json
import hashlib
from datetime import datetime, timezone

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import DatasetType, GraphSAGELifecycleState
from graphsage.graph_dataset import TemporalHeteroGraphDataset
from graphsage.data_source import MockRealGraphDataSource
from graphsage.readiness_checker import RealDataReadinessChecker
from graphsage.governance_gate import DataRequiredGate

PROD_CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
PROD_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"


def main():
    print("=" * 80)
    print("PHASE 58: REAL-DATA SHADOW READINESS AUDIT & GOVERNANCE GATE EVALUATION")
    print("=" * 80)

    # 1. Pre-execution champion checksum check
    hasher = hashlib.sha256()
    with open(PROD_CHAMPION_PATH, "rb") as f:
        while chunk := f.read(65536):
            hasher.update(chunk)
    champion_sha256 = hasher.hexdigest()
    assert champion_sha256 == PROD_EXPECTED_SHA256, "Production champion checksum mismatch!"

    data_dir = os.path.abspath(os.path.join(ML_SERVICE_DIR, "..", "synthetic_ropus", "data"))
    dataset = TemporalHeteroGraphDataset.load_from_directory(data_dir)

    # Convert in-memory dataset to MockRealGraphDataSource representing the audited dataset
    source = MockRealGraphDataSource(
        nodes=list(dataset.nodes.values()),
        edges=list(dataset.edges.values()),
        dataset_type=DatasetType.SYNTHETIC
    )

    # Load labels and transactions
    import pandas as pd
    txns_df = pd.read_csv(os.path.join(data_dir, "transactions", "transactions.csv"))
    labels_df = pd.read_csv(os.path.join(data_dir, "labels", "labels.csv"))
    source.txns_list = txns_df.to_dict("records")
    source.labels_list = labels_df.to_dict("records")

    checker = RealDataReadinessChecker(source)
    audit_report = checker.audit_dataset()

    gate = DataRequiredGate(min_confirmed_fraud=50, min_confirmed_collusion=50)
    gate_decision = gate.evaluate_gate(audit_report)

    output_payload = {
        "audit_metadata": {
            "phase": "PHASE_58_REAL_DATA_SHADOW_READINESS",
            "audited_at": audit_report.audited_at.isoformat(),
            "target_model": "GraphSAGE 2-Layer Heterogeneous Neural Network",
            "active_champion_model": "production_model_v8_bmr.joblib",
            "active_champion_sha256": champion_sha256,
            "operational_mode": "STRICTLY NON-ENFORCING SHADOW / OFFLINE RESEARCH ONLY",
            "customer_decision_routing": "UNCHANGED (Baseline Dynamic BMR Authoritative)"
        },
        "readiness_scorecard": {
            "dataset_type": audit_report.dataset_type.value,
            "is_ready_for_real_validation": audit_report.is_ready_for_real_validation,
            "governance_lifecycle_state": gate_decision.target_lifecycle_state.value,
            "is_promotion_eligible": gate_decision.is_promotion_eligible,
            "blocking_reasons": gate_decision.blocking_reasons,
            "enforcement_summary": gate_decision.enforcement_summary
        },
        "audit_dimensions": {
            "schema_completeness": audit_report.schema_completeness,
            "temporal_causality": audit_report.temporal_causality,
            "privacy_and_tokenization": audit_report.privacy_and_security,
            "graph_topology": audit_report.topology_statistics,
            "entity_coverage": audit_report.entity_coverage,
            "label_maturity_coverage": audit_report.label_maturity_coverage,
            "confirmed_fraud_labels_count": audit_report.confirmed_fraud_labels_count,
            "confirmed_internal_collusion_labels_count": audit_report.confirmed_internal_collusion_labels_count
        },
        "real_data_contract_specification": {
            "version": "2.0.0",
            "required_entities": ["EMPLOYEE", "CONSUMER", "ACCOUNT", "DEVICE", "IP", "PAYMENT_TOKEN", "TRANSACTION", "MERCHANT", "CASE"],
            "required_edge_types": [
                "EMPLOYEE_ACCESSES_ACCOUNT", "EMPLOYEE_REVIEWS_CASE", "EMPLOYEE_MODIFIES_TRANSACTION",
                "EMPLOYEE_APPROVES_TRANSACTION", "EMPLOYEE_USES_DEVICE", "EMPLOYEE_USES_IP",
                "CONSUMER_OWNS_ACCOUNT", "ACCOUNT_USES_DEVICE", "ACCOUNT_USES_IP",
                "ACCOUNT_USES_PAYMENT_TOKEN", "ACCOUNT_TRANSACTS_WITH_MERCHANT"
            ],
            "prohibited_attributes": ["raw_pan", "cvv", "pin", "ssn", "cleartext_password", "scenario_type"],
            "minimum_real_collusion_cases_for_promotion": 50,
            "minimum_real_fraud_cases_for_promotion": 50
        },
        "required_next_steps_for_real_readiness": [
            "1. Connect RealGraphDataSource adapter to passive Kafka audit event stream.",
            "2. Ingest shadow graph topology without altering production risk scoring.",
            "3. Accumulate >= 50 confirmed real internal collusion investigations.",
            "4. Mature chargeback labels for >= 90 days before historical holdout evaluation.",
            "5. Submit formal Model Risk Governance petition to Risk Council."
        ],
        "final_lifecycle_state": "SYNTHETICALLY_VALIDATED / REAL_DATA_REQUIRED / STRICTLY_NON_ENFORCING / PROMOTION_BLOCKED"
    }

    out_file = os.path.join(CURRENT_DIR, "phase_58_real_data_readiness.json")
    with open(out_file, "w") as f:
        json.dump(output_payload, f, indent=2)

    print(f"[Phase 58] Saved Readiness Scorecard: {out_file}")
    print(f"[Phase 58] Governance Decision: {gate_decision.target_lifecycle_state.value}")
    print(f"[Phase 58] Promotion Eligible:   {gate_decision.is_promotion_eligible}")
    print(f"[Phase 58] Blocking Reasons:     {len(gate_decision.blocking_reasons)}")
    for r in gate_decision.blocking_reasons:
        print(f"  - {r}")


if __name__ == "__main__":
    main()
