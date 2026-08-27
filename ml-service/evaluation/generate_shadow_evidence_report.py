"""
AI Risk Manager — Live Shadow Evaluation Evidence Reporter
Generates machine-readable live shadow evidence report adhering to strict status enums.
"""

import os
import sys
import json
import pandas as pd

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from evaluation.shadow_evaluator import (
    ShadowEvaluationTelemetryStore,
    ShadowPromotionGateEvaluator,
    CostPolicyConfig
)

def generate_live_shadow_report():
    print("=" * 80)
    print("GENERATING LIVE SHADOW EVALUATION EVIDENCE REPORT")
    print("=" * 80)

    # Instantiate empty live store representing current zero live traffic state
    live_store = ShadowEvaluationTelemetryStore(min_maturation_seconds=60 * 86400)
    cost_cfg = CostPolicyConfig()

    eval_result = ShadowPromotionGateEvaluator.evaluate_gates(
        telemetry_store=live_store,
        cost_config=cost_cfg,
        candidate_decision_threshold=0.220,
        significance_alpha=0.05,
        require_matured=True,
        is_live_production=True
    )

    report_payload = {
        "report_type": "live_shadow_evaluation_evidence_report",
        "generated_at": pd.Timestamp.now("UTC").isoformat(),
        "models": {
            "production_champion": {
                "name": "fraud-xgb-25f-v3.0",
                "customer_decision_authority_pct": 100.0,
                "feature_contract": "MLFeatureContractV15",
                "status": "ACTIVE_PRODUCTION"
            },
            "shadow_candidate": {
                "name": "extended_catboost_58f",
                "customer_decision_authority_pct": 0.0,
                "feature_contract": "MLFeatureContractV25",
                "status": "SHADOW_MODE_NON_ENFORCING"
            }
        },
        "dataset_availability": "FULL_DATASET_UNAVAILABLE",
        "frozen_holdout_sha256": "a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44",
        "live_telemetry_summary": {
            "total_live_transactions": live_store.count_predictions(),
            "total_matured_chargebacks": live_store.count_attributed_labels(require_matured=True),
            "maturation_window_days": 60
        },
        "promotion_governance": eval_result,
        "dual_control_signatures": {
            "risk_director_approval": "PENDING_LIVE_EVIDENCE",
            "mlops_lead_approval": "PENDING_LIVE_EVIDENCE"
        },
        "final_verdict": "REMAIN_IN_SHADOW_MODE"
    }

    out_path = os.path.join(CURRENT_DIR, "live_shadow_evidence_report.json")
    with open(out_path, "w") as f:
        json.dump(report_payload, f, indent=2)
    print(f"[SUCCESS] Machine-readable report saved to: {out_path}")
    return report_payload

if __name__ == "__main__":
    generate_live_shadow_report()
