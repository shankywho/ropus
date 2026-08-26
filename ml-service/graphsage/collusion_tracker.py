"""
Real-World Employee-Consumer Collusion Accumulation Tracker (Phase 60)
Aggregates structural relationship telemetry and monitors progress toward genuine real-data requirements.
"""

from datetime import datetime, timezone
from typing import Dict, List, Any, Optional, Set
from collections import defaultdict, Counter
from pydantic import BaseModel

from .schema import HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType, GraphSAGELifecycleState


class RealCollusionSummaryView(BaseModel):
    tracked_at: datetime
    total_employees_observed: int
    employees_with_account_access: int
    employees_with_concentrated_access: int
    total_self_approval_events: int
    total_off_hours_access_events: int
    total_shared_device_overlaps: int
    total_shared_ip_overlaps: int
    confirmed_real_internal_collusion_labels: int
    required_collusion_labels_for_promotion: int = 50
    collusion_gate_status: str
    operational_mode: str = "NON_ENFORCING_INVESTIGATION_INTELLIGENCE"
    legal_notice: str = "Investigation intelligence only. High risk scores denote structural anomaly, not legal guilt."


class CollusionAccumulationTracker:
    """
    Accumulates structural employee-consumer graph telemetry over sliding production shadow windows.
    """

    def __init__(self):
        self.observed_employees: Set[str] = set()
        self.account_access_map: Dict[str, List[str]] = defaultdict(list)
        self.self_approvals_count: int = 0
        self.off_hours_count: int = 0
        self.shared_devices_count: int = 0
        self.shared_ips_count: int = 0
        self.confirmed_collusion_labels: int = 0

    def process_edge(self, edge: HeteroEdge):
        """Updates accumulation counters for an incoming edge."""
        etype = edge.type.value if hasattr(edge.type, "value") else str(edge.type)

        if edge.source_id.startswith("emp_"):
            self.observed_employees.add(edge.source_id)

        if "ACCESS" in etype and edge.target_id.startswith("acct_"):
            self.account_access_map[edge.source_id].append(edge.target_id)
            if edge.timestamp.hour < 7 or edge.timestamp.hour >= 19:
                self.off_hours_count += 1

        if "APPROVE" in etype:
            self.self_approvals_count += 1

        if "DEVICE" in etype:
            self.shared_devices_count += 1

        if "IP" in etype:
            self.shared_ips_count += 1

    def record_confirmed_collusion(self, count: int = 1):
        self.confirmed_collusion_labels += count

    def get_summary_view(self) -> RealCollusionSummaryView:
        now = datetime.now(timezone.utc)

        # Count concentrated employees (HHI > 0.25)
        concentrated_count = 0
        for emp_id, accts in self.account_access_map.items():
            if len(accts) >= 3:
                counts = Counter(accts)
                total = len(accts)
                hhi = sum((c / total) ** 2 for c in counts.values())
                if hhi > 0.25:
                    concentrated_count += 1

        gate_status = (
            f"DATA_REQUIRED: Found {self.confirmed_collusion_labels}/50 confirmed real internal collusion cases."
            if self.confirmed_collusion_labels < 50 else
            "GATE_CRITERIA_MET: Ready for Model Risk Council Review."
        )

        return RealCollusionSummaryView(
            tracked_at=now,
            total_employees_observed=len(self.observed_employees),
            employees_with_account_access=len(self.account_access_map),
            employees_with_concentrated_access=concentrated_count,
            total_self_approval_events=self.self_approvals_count,
            total_off_hours_access_events=self.off_hours_count,
            total_shared_device_overlaps=self.shared_devices_count,
            total_shared_ip_overlaps=self.shared_ips_count,
            confirmed_real_internal_collusion_labels=self.confirmed_collusion_labels,
            required_collusion_labels_for_promotion=50,
            collusion_gate_status=gate_status
        )
