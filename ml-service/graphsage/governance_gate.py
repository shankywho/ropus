"""
DATA_REQUIRED Governance Gate & Promotion Policy Enforcer (Phase 58)
Enforces the mandatory boundary between synthetic validation and real production readiness.
"""

from typing import Dict, List, Any
from pydantic import BaseModel
from .schema import GraphSAGELifecycleState, DatasetType, ReadinessAuditReport


class GovernanceGateDecision(BaseModel):
    is_promotion_eligible: bool
    current_lifecycle_state: GraphSAGELifecycleState
    target_lifecycle_state: GraphSAGELifecycleState
    blocking_reasons: List[str]
    enforcement_summary: str
    can_serve_shadow_traffic: bool


class DataRequiredGate:
    """
    Formal Governance Gate blocking GraphSAGE from claiming production validation
    until all real-data prerequisite gates are satisfied.
    """

    def __init__(
        self,
        min_confirmed_fraud: int = 50,
        min_confirmed_collusion: int = 50
    ):
        self.min_confirmed_fraud = min_confirmed_fraud
        self.min_confirmed_collusion = min_confirmed_collusion

    def evaluate_gate(self, report: ReadinessAuditReport) -> GovernanceGateDecision:
        blocking_reasons: List[str] = []

        # 1. Dataset Type Check
        if report.dataset_type == DatasetType.SYNTHETIC:
            blocking_reasons.append(
                "SYNTHETIC BOUNDARY: The evaluated dataset is synthetic. "
                "Synthetic experiments establish algorithmic viability but CANNOT be cited as production validation."
            )

        # 2. Confirmed Real Fraud Count Gate
        if report.confirmed_fraud_labels_count < self.min_confirmed_fraud:
            blocking_reasons.append(
                f"INSUFFICIENT REAL FRAUD LABELS: Found {report.confirmed_fraud_labels_count} confirmed real fraud cases; "
                f"minimum required for promotion review is {self.min_confirmed_fraud}."
            )

        # 3. Confirmed Internal Collusion Count Gate
        if report.confirmed_internal_collusion_labels_count < self.min_confirmed_collusion:
            blocking_reasons.append(
                f"INSUFFICIENT REAL COLLUSION LABELS: Found {report.confirmed_internal_collusion_labels_count} confirmed internal collusion cases; "
                f"minimum required for promotion review is {self.min_confirmed_collusion}."
            )

        # 4. Privacy & Security Gate
        if report.privacy_and_security.get("tokenization_compliance") != "PASSED":
            blocking_reasons.append("SECURITY VIOLATION: Unsanitized or cleartext payment identifiers detected in graph store.")

        # 5. Temporal Causality Gate
        if report.temporal_causality.get("status") != "PASSED":
            blocking_reasons.append("CAUSALITY VIOLATION: Graph store contains future-dated edges or premature ground-truth labels.")

        is_eligible = len(blocking_reasons) == 0

        current_state = GraphSAGELifecycleState.SYNTHETIC_VALIDATED
        target_state = (
            GraphSAGELifecycleState.PROMOTION_ELIGIBLE
            if is_eligible
            else GraphSAGELifecycleState.REAL_DATA_REQUIRED
        )

        summary = (
            "PROMOTION BLOCKED: GraphSAGE remains strictly NON-ENFORCING in SHADOW/RESEARCH mode. "
            "Real-world data ingestion and label maturation are required before any customer routing."
            if not is_eligible else
            "PROMOTION ELIGIBLE: All real-data and governance criteria satisfied."
        )

        return GovernanceGateDecision(
            is_promotion_eligible=is_eligible,
            current_lifecycle_state=current_state,
            target_lifecycle_state=target_state,
            blocking_reasons=blocking_reasons,
            enforcement_summary=summary,
            can_serve_shadow_traffic=True  # Shadow mode is always safe and enabled
        )
