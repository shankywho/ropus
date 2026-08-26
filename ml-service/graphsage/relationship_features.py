"""
Relationship-Level Feature Extraction Engine (Phase 59)
Computes structural, temporal, and topological relationship metrics decoupled from raw IDs.
"""

import numpy as np
from collections import Counter
from datetime import datetime
from typing import Dict, List, Any, Optional

from .schema import HeteroNode, HeteroEdge, HeteroNodeType, HeteroEdgeType


class RelationshipFeatureExtractor:
    """
    Extracts 13 explicit relationship signals separate from entity IDs.
    """

    def extract_features(
        self,
        employee_id: str,
        nodes: Dict[str, HeteroNode],
        incident_edges: List[HeteroEdge],
        associated_txns: Optional[List[Dict[str, Any]]] = None
    ) -> Dict[str, float]:
        if not incident_edges:
            return {
                "employee_account_access_count": 0.0,
                "employee_consumer_unique_count": 0.0,
                "employee_account_concentration": 0.0,
                "employee_transaction_proximity_sec": 999999.0,
                "employee_approval_after_access_count": 0.0,
                "off_hours_ratio": 0.0,
                "shared_device_overlap_count": 0.0,
                "shared_ip_overlap_count": 0.0,
                "employee_to_consumer_path_count": 0.0,
                "employee_to_account_path_count": 0.0,
                "repeated_access_burst_score": 0.0,
                "low_and_slow_relationship_score": 0.0,
                "neighborhood_anomaly_score": 0.0,
            }

        # 1. Accounts and Consumers touched
        acct_touches = []
        appr_count = 0
        off_hours_count = 0
        device_touches = set()
        ip_touches = set()

        for e in incident_edges:
            etype = e.type.value if hasattr(e.type, "value") else str(e.type)
            if "ACCOUNT" in etype or e.target_id.startswith("acct_"):
                acct_touches.append(e.target_id)
            if "APPROVE" in etype:
                appr_count += 1
            if "DEVICE" in etype:
                device_touches.add(e.target_id)
            if "IP" in etype:
                ip_touches.add(e.target_id)
            if e.timestamp.hour < 7 or e.timestamp.hour >= 19 or e.timestamp.weekday() >= 5:
                off_hours_count += 1

        total_access = len(incident_edges)
        unique_accts = list(set(acct_touches))

        # 2. Herfindahl-Hirschman Index (Concentration Ratio)
        if acct_touches:
            counts = Counter(acct_touches)
            total_t = len(acct_touches)
            hhi = sum((cnt / total_t) ** 2 for cnt in counts.values())
        else:
            hhi = 0.0

        # 3. Transaction Proximity
        txns = associated_txns or []
        min_delta_sec = 999999.0
        for e in incident_edges:
            for t in txns:
                t_ts = t.get("event_timestamp")
                if isinstance(t_ts, str):
                    t_ts = datetime.fromisoformat(t_ts)
                if t_ts and t_ts >= e.timestamp:
                    delta = (t_ts - e.timestamp).total_seconds()
                    if delta < min_delta_sec:
                        min_delta_sec = delta

        # 4. Burst and Low-and-Slow Scores
        timestamps = sorted([e.timestamp for e in incident_edges])
        burst_score = 0.0
        low_slow_score = 0.0

        if len(timestamps) >= 2:
            intervals = [(timestamps[i] - timestamps[i-1]).total_seconds() for i in range(1, len(timestamps))]
            if min(intervals) < 120.0:  # Repeated access within 2 minutes
                burst_score = 0.85
            span_days = (timestamps[-1] - timestamps[0]).total_seconds() / 86400.0
            if span_days > 14.0 and len(timestamps) <= 5:  # Low frequency across weeks
                low_slow_score = 0.70

        # 5. Anomaly Score
        off_hours_ratio = off_hours_count / max(total_access, 1)
        anomaly_score = (hhi * 0.40) + (off_hours_ratio * 0.30) + (min(1.0, (100.0 / max(min_delta_sec, 1.0))) * 0.30)

        return {
            "employee_account_access_count": float(len(acct_touches)),
            "employee_consumer_unique_count": float(len(unique_accts)),
            "employee_account_concentration": round(float(hhi), 4),
            "employee_transaction_proximity_sec": round(float(min_delta_sec), 1),
            "employee_approval_after_access_count": float(appr_count),
            "off_hours_ratio": round(float(off_hours_ratio), 4),
            "shared_device_overlap_count": float(len(device_touches)),
            "shared_ip_overlap_count": float(len(ip_touches)),
            "employee_to_consumer_path_count": float(len(unique_accts)),
            "employee_to_account_path_count": float(len(acct_touches)),
            "repeated_access_burst_score": round(float(burst_score), 4),
            "low_and_slow_relationship_score": round(float(low_slow_score), 4),
            "neighborhood_anomaly_score": round(float(anomaly_score), 4),
        }
