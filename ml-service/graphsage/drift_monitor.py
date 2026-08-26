"""
Shadow Graph Drift Monitoring Engine (Phase 59)
Tracks Data Drift, Topology Drift, Score Drift, and Label Drift across sliding shadow windows.
"""

import numpy as np
from collections import Counter
from datetime import datetime, timezone
from typing import Dict, List, Any, Optional

from .schema import HeteroNode, HeteroEdge, HeteroNodeType, LabelMaturityState


class ShadowGraphDriftMonitor:
    """
    Evaluates drift across 4 distinct shadow dimensions without automated retraining.
    """

    def compute_drift_report(
        self,
        current_nodes: List[HeteroNode],
        current_edges: List[HeteroEdge],
        current_scores: List[float],
        baseline_stats: Optional[Dict[str, Any]] = None,
        current_labels: Optional[List[Dict[str, Any]]] = None
    ) -> Dict[str, Any]:
        now = datetime.now(timezone.utc)

        # 1. Topology metrics
        degrees = Counter()
        for e in current_edges:
            degrees[e.source_id] += 1
            degrees[e.target_id] += 1

        deg_values = list(degrees.values()) if degrees else [0]
        avg_deg = float(np.mean(deg_values))
        p95_deg = float(np.percentile(deg_values, 95))
        p99_deg = float(np.percentile(deg_values, 99))
        max_deg = int(max(deg_values))
        isolated_nodes = sum(1 for n in current_nodes if degrees[n.id] == 0)

        # 2. Entity counts
        type_counts = Counter(n.type.value if hasattr(n.type, "value") else str(n.type) for n in current_nodes)

        # 3. Score distribution
        scores = np.array(current_scores) if current_scores else np.array([0.05])
        mean_score = float(np.mean(scores))
        p95_score = float(np.percentile(scores, 95))
        high_risk_ratio = float(np.mean(scores >= 0.70))

        # 4. Label distribution
        labels = current_labels or []
        label_states = Counter(l.get("maturity_state", "UNLABELED") for l in labels)

        # Baseline comparison
        base = baseline_stats or {}
        base_avg_deg = base.get("average_degree", avg_deg)
        base_mean_score = base.get("mean_score", mean_score)

        deg_shift_pct = round(((avg_deg - base_avg_deg) / max(base_avg_deg, 1e-5)) * 100.0, 2)
        score_shift_pct = round(((mean_score - base_mean_score) / max(base_mean_score, 1e-5)) * 100.0, 2)

        return {
            "monitored_at": now.isoformat(),
            "window_sample_size": {
                "nodes_count": len(current_nodes),
                "edges_count": len(current_edges),
                "scores_count": len(scores),
                "labels_count": len(labels)
            },
            "data_drift": {
                "entity_type_counts": dict(type_counts),
                "isolated_nodes_count": isolated_nodes,
                "data_drift_status": "STABLE" if isolated_nodes < len(current_nodes) * 0.10 else "ELEVATED_ISOLATION"
            },
            "topology_drift": {
                "average_degree": round(avg_deg, 2),
                "p95_degree": round(p95_deg, 2),
                "p99_degree": round(p99_deg, 2),
                "max_degree": max_deg,
                "degree_shift_pct_from_baseline": deg_shift_pct,
                "topology_drift_status": "STABLE" if abs(deg_shift_pct) < 30.0 else "TOPOLOGY_SHIFT"
            },
            "score_drift": {
                "mean_score": round(mean_score, 4),
                "p95_score": round(p95_score, 4),
                "high_risk_entity_ratio_pct": round(high_risk_ratio * 100.0, 2),
                "score_shift_pct_from_baseline": score_shift_pct,
                "score_drift_status": "STABLE" if abs(score_shift_pct) < 25.0 else "SCORE_SHIFT"
            },
            "label_drift": {
                "maturity_counts": dict(label_states),
                "label_drift_status": "MONITORED"
            },
            "governance_action": "NO_RETRAINING_TRIGGERED (Shadow Monitoring Only)"
        }
