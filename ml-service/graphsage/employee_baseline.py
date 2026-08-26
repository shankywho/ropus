"""
Employee Behavioral Baseline Engine (Phase 59)
Constructs personal historical behavioral baselines to detect "unusual for this employee" patterns point-in-time.
"""

import numpy as np
from datetime import datetime, timezone, timedelta
from typing import Dict, List, Any, Optional, Set
from collections import Counter, defaultdict

from .schema import HeteroEdge, HeteroNode


class EmployeeBehavioralBaselineEngine:
    """
    Computes per-employee baseline profiles point-in-time (strictly events < T).
    """

    def compute_baseline_and_deviation(
        self,
        employee_id: str,
        current_events: List[HeteroEdge],
        historical_edges: List[HeteroEdge],
        as_of: datetime
    ) -> Dict[str, Any]:
        """
        Calculates personal baseline profile and deviation scores.
        """
        # Filter historical edges strictly before evaluation window
        prior_edges = [
            e for e in historical_edges
            if (e.source_id == employee_id or e.target_id == employee_id)
            and e.timestamp < as_of
        ]

        if not prior_edges:
            # Cold-start employee fallback
            return {
                "has_historical_baseline": False,
                "prior_interaction_count": 0,
                "mean_daily_volume": 0.0,
                "normal_active_hours": [9, 10, 11, 12, 13, 14, 15, 16, 17],
                "trusted_devices": [],
                "trusted_ips": [],
                "volume_z_score": 0.0,
                "is_unusual_hour_for_employee": False,
                "is_new_device_for_employee": True,
                "personal_anomaly_delta": 0.15,
                "baseline_notes": "Zero historical telemetry for this employee; default role baseline applied."
            }

        # 1. Historical hours distribution
        hours_hist = [e.timestamp.hour for e in prior_edges]
        hour_counts = Counter(hours_hist)
        frequent_hours = [h for h, cnt in hour_counts.items() if (cnt / len(prior_edges)) >= 0.05]

        # 2. Historical daily volumes
        day_buckets = defaultdict(int)
        for e in prior_edges:
            day_key = e.timestamp.strftime("%Y-%m-%d")
            day_buckets[day_key] += 1

        daily_vols = list(day_buckets.values())
        mean_vol = float(np.mean(daily_vols)) if daily_vols else 1.0
        std_vol = float(np.std(daily_vols)) + 1e-5

        # 3. Trusted infrastructure
        trusted_devs = set()
        trusted_ips = set()
        for e in prior_edges:
            etype = e.type.value if hasattr(e.type, "value") else str(e.type)
            if "DEVICE" in etype:
                trusted_devs.add(e.target_id if e.source_id == employee_id else e.source_id)
            if "IP" in etype:
                trusted_ips.add(e.target_id if e.source_id == employee_id else e.source_id)

        # 4. Compare Current Activity against Historical Baseline
        curr_hours = [e.timestamp.hour for e in current_events]
        curr_unusual_hours = sum(1 for h in curr_hours if h not in frequent_hours)
        is_unusual_hour = (curr_unusual_hours / max(len(current_events), 1)) >= 0.50 if current_events else False

        curr_vol = float(len(current_events))
        vol_z = (curr_vol - mean_vol) / std_vol

        # Check new devices
        curr_devs = set()
        for e in current_events:
            etype = e.type.value if hasattr(e.type, "value") else str(e.type)
            if "DEVICE" in etype:
                curr_devs.add(e.target_id if e.source_id == employee_id else e.source_id)

        is_new_device = any(d not in trusted_devs for d in curr_devs) if curr_devs else False

        # Combined personal deviation metric
        anomaly_delta = 0.0
        if is_unusual_hour:
            anomaly_delta += 0.35
        if vol_z > 2.0:
            anomaly_delta += 0.30
        if is_new_device and len(trusted_devs) > 0:
            anomaly_delta += 0.25

        return {
            "has_historical_baseline": True,
            "prior_interaction_count": len(prior_edges),
            "mean_daily_volume": round(mean_vol, 2),
            "std_daily_volume": round(std_vol, 2),
            "normal_active_hours": sorted(frequent_hours),
            "trusted_devices": list(trusted_devs),
            "trusted_ips": list(trusted_ips),
            "volume_z_score": round(float(vol_z), 2),
            "is_unusual_hour_for_employee": bool(is_unusual_hour),
            "is_new_device_for_employee": bool(is_new_device),
            "personal_anomaly_delta": round(float(min(1.0, anomaly_delta)), 4),
            "baseline_notes": f"Baseline constructed from {len(prior_edges)} events across {len(daily_vols)} active days."
        }
