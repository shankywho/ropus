"""
Unit Tests for Production Shadow Evaluation & Outcome Attribution Engine
Tests all 8 required edge cases:
1. valid outcome
2. early outcome (unmatured)
3. duplicate outcome (idempotent)
4. conflicting outcome (conflict resolution)
5. unknown evaluation ID (KeyError)
6. malformed outcome (ValueError)
7. future-dated / pre-dating outcome (ValueError)
8. immutable feature snapshot after outcome arrival
9. synthetic / non-production exclusion
10. promotion gate status enums & evidence basis
"""

import os
import sys
import time
import pytest
import numpy as np
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

def test_01_valid_outcome():
    """Valid outcome correctly attributes and tags as matured when >=60 days."""
    store = ShadowEvaluationTelemetryStore(min_maturation_seconds=60 * 86400)
    tx_time = 1000000.0
    store.record_prediction({
        "evaluation_id": "eval_valid_01",
        "tenant_id": "tenant_1",
        "timestamp": tx_time,
        "amount_usd": 150.0,
        "champion_probability": 0.05,
        "candidate_calibrated_probability": 0.04
    })

    res = store.attribute_outcome(
        "eval_valid_01",
        is_fraud=1,
        label_source="CHARGEBACK_FEEDBACK",
        attributed_at=tx_time + (61 * 86400)
    )
    assert res["status"] == "ATTRIBUTED"
    assert res["is_matured"] is True
    assert store.count_attributed_labels(require_matured=True) == 1

def test_02_early_unmatured_outcome():
    """Early outcome (<60 days) is stored but excluded when require_matured=True."""
    store = ShadowEvaluationTelemetryStore(min_maturation_seconds=60 * 86400)
    tx_time = 1000000.0
    store.record_prediction({"evaluation_id": "eval_early_01", "timestamp": tx_time})

    res = store.attribute_outcome(
        "eval_early_01",
        is_fraud=1,
        label_source="EARLY_DISPUTE",
        attributed_at=tx_time + (15 * 86400) # Only 15 days
    )
    assert res["status"] == "ATTRIBUTED"
    assert res["is_matured"] is False
    assert store.count_attributed_labels(require_matured=False) == 1
    assert store.count_attributed_labels(require_matured=True) == 0

def test_03_duplicate_outcome_idempotency():
    """Duplicate identical outcomes are handled gracefully as IDEMPOTENT_NOOP."""
    store = ShadowEvaluationTelemetryStore(min_maturation_seconds=0)
    store.record_prediction({"evaluation_id": "eval_dup_01", "timestamp": 1000.0})

    res1 = store.attribute_outcome("eval_dup_01", is_fraud=0, label_source="CLEAR")
    assert res1["status"] == "ATTRIBUTED"

    res2 = store.attribute_outcome("eval_dup_01", is_fraud=0, label_source="CLEAR")
    assert res2["status"] == "IDEMPOTENT_NOOP"
    assert store.count_attributed_labels() == 1

def test_04_conflicting_outcome_resolution():
    """Confirmed fraud/chargeback takes precedence over provisional clear."""
    store = ShadowEvaluationTelemetryStore(min_maturation_seconds=0)
    store.record_prediction({"evaluation_id": "eval_conflict_01", "timestamp": 1000.0})

    # Step 1: Provisional clear
    store.attribute_outcome("eval_conflict_01", is_fraud=0, label_source="PROVISIONAL_CLEAR")
    df1 = store.get_matched_evaluation_dataset()
    assert df1.iloc[0]["is_fraud"] == 0

    # Step 2: Chargeback dispute arrives
    res = store.attribute_outcome("eval_conflict_01", is_fraud=1, label_source="CHARGEBACK_DISPUTE")
    assert res["status"] == "CONFLICT_RESOLVED"
    assert res["resolved_is_fraud"] == 1
    df2 = store.get_matched_evaluation_dataset()
    assert df2.iloc[0]["is_fraud"] == 1

def test_05_unknown_evaluation_id_raises_key_error():
    """Attributing outcome to non-existent prediction raises KeyError."""
    store = ShadowEvaluationTelemetryStore()
    with pytest.raises(KeyError) as exc_info:
        store.attribute_outcome("non_existent_eval_id", is_fraud=1)
    assert "not found in shadow telemetry store" in str(exc_info.value)

def test_06_malformed_outcome_missing_eval_id_raises_value_error():
    """Recording telemetry without evaluation_id raises ValueError."""
    store = ShadowEvaluationTelemetryStore()
    with pytest.raises(ValueError) as exc_info:
        store.record_prediction({"timestamp": 1000.0})
    assert "evaluation_id is mandatory" in str(exc_info.value)

def test_07_invalid_outcome_predating_transaction_raises_value_error():
    """Outcome timestamp before transaction timestamp raises ValueError."""
    store = ShadowEvaluationTelemetryStore()
    tx_time = 1000000.0
    store.record_prediction({"evaluation_id": "eval_time_warp", "timestamp": tx_time})
    with pytest.raises(ValueError) as exc_info:
        store.attribute_outcome("eval_time_warp", is_fraud=1, attributed_at=tx_time - 500.0)
    assert "cannot pre-date transaction time" in str(exc_info.value)

def test_08_immutable_feature_snapshot_after_outcome_arrival():
    """Attributing an outcome cannot modify the original prediction snapshot fields."""
    store = ShadowEvaluationTelemetryStore(min_maturation_seconds=0)
    orig_pred = {
        "evaluation_id": "eval_immutability",
        "tenant_id": "tenant_xyz",
        "timestamp": 5000.0,
        "champion_probability": 0.035,
        "candidate_calibrated_probability": 0.025,
        "amount_usd": 99.99
    }
    store.record_prediction(orig_pred)

    # Attribute outcome
    store.attribute_outcome("eval_immutability", is_fraud=1)

    # Direct inspection of internal store
    stored_pred = store._predictions["eval_immutability"]
    assert stored_pred["champion_probability"] == 0.035
    assert stored_pred["candidate_calibrated_probability"] == 0.025
    assert stored_pred["amount_usd"] == 99.99
    # Label is not present in prediction snapshot
    assert "is_fraud" not in stored_pred

def test_09_synthetic_test_exclusion():
    """Synthetic or non-production events are strictly excluded from production promotion evaluations."""
    store = ShadowEvaluationTelemetryStore(min_maturation_seconds=0)

    # 10 synthetic records
    for i in range(10):
        store.record_prediction({
            "evaluation_id": f"eval_synth_{i}",
            "is_synthetic": True,
            "environment": "test",
            "timestamp": 1000.0 + i
        })
        store.attribute_outcome(f"eval_synth_{i}", is_fraud=1)

    assert store.count_predictions(exclude_synthetic=True) == 0
    assert store.count_predictions(exclude_synthetic=False) == 10
    df_prod = store.get_matched_evaluation_dataset(exclude_synthetic=True)
    assert len(df_prod) == 0

def test_10_promotion_gate_status_enums_and_evidence_basis():
    """Verifies that the promotion evaluator outputs correct status enums and evidence tiers."""
    store = ShadowEvaluationTelemetryStore()

    # Empty store -> NOT_POPULATED
    res_empty = ShadowPromotionGateEvaluator.evaluate_gates(store, is_live_production=True)
    assert res_empty["promotion_gates"]["gate_1_live_volume"]["status"] == "NOT_POPULATED"
    assert res_empty["promotion_gates"]["gate_1_live_volume"]["evidence_basis"] == "NOT_POPULATED"
    assert res_empty["overall_promotion_decision"] == "REMAIN_IN_SHADOW_MODE"
    assert res_empty["all_promotion_gates_passed"] is False

def test_11_provenance_isolation_and_gate1_counter():
    """Only LIVE_PRODUCTION records increment Gate 1 counter."""
    store = ShadowEvaluationTelemetryStore(min_maturation_seconds=0)

    store.record_prediction({"evaluation_id": "prod_1", "provenance": "LIVE_PRODUCTION"})
    store.record_prediction({"evaluation_id": "prod_2", "provenance": "LIVE_PRODUCTION"})
    store.record_prediction({"evaluation_id": "synth_1", "provenance": "SYNTHETIC"})
    store.record_prediction({"evaluation_id": "replay_1", "provenance": "REPLAY"})
    store.record_prediction({"evaluation_id": "offline_1", "provenance": "OFFLINE_TEST"})

    # Deduplicate check: duplicate evaluation_id
    store.record_prediction({"evaluation_id": "prod_1", "provenance": "LIVE_PRODUCTION"})

    assert store.count_predictions(provenance_filter="LIVE_PRODUCTION") == 2
    assert store.count_predictions(provenance_filter=None) == 5
    counts = store.get_provenance_counts()
    assert counts["LIVE_PRODUCTION"] == 2
    assert counts["SYNTHETIC"] == 1
    assert counts["REPLAY"] == 1
    assert counts["OFFLINE_TEST"] == 1

def test_12_failed_inference_still_counts_as_production_tx():
    """Failed candidate inference still counts as a genuine production shadow transaction."""
    store = ShadowEvaluationTelemetryStore(min_maturation_seconds=0)
    store.record_prediction({
        "evaluation_id": "eval_failed_candidate_01",
        "provenance": "LIVE_PRODUCTION",
        "shadow_execution_status": "FAILED",
        "error_category": "TIMEOUT"
    })
    assert store.count_predictions(provenance_filter="LIVE_PRODUCTION") == 1

def test_13_conditional_promotion_governance_flags():
    """Verifies that the promotion evaluator outputs explicit 5-way governance flags and conditional canary eligibility."""
    store = ShadowEvaluationTelemetryStore()
    res = ShadowPromotionGateEvaluator.evaluate_gates(store, is_live_production=True)

    flags = res["governance_status_flags"]
    assert flags["MODEL_VALIDATED"] is True
    assert flags["PRODUCTION_SHADOW_READY"] is True
    assert flags["LIVE_EVIDENCE_UNAVAILABLE"] is True
    assert flags["CONDITIONAL_PROMOTION_ELIGIBLE"] is True
    assert flags["PRODUCTION_PROMOTION_BLOCKED"] is True

    cond = res["conditional_promotion"]
    assert cond["eligible"] is True
    assert cond["status"] == "CONDITIONAL_PROMOTION_ELIGIBLE"
    assert cond["recommended_canary_percentage"] == 5
    assert cond["maker_checker_required"] is True

    uncond = res["unconditional_promotion"]
    assert uncond["allowed"] is False
    assert uncond["status"] == "PRODUCTION_PROMOTION_BLOCKED"
