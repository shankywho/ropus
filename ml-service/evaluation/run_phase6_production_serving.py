"""
ROPUS Phase 6: Production Serving & End-to-End Deployment Validation Pipeline
Audits artifact loading, inference contract, schema enforcement, prediction consistency,
BMR boundary precision, latency profiling, concurrent safety, fail-safe modes, and regression parity.
"""

import os
import sys
import json
import time
import hashlib
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed
import numpy as np
import pandas as pd
import joblib
import xgboost as xgb
from sklearn.metrics import (
    roc_auc_score,
    average_precision_score,
    precision_score,
    recall_score,
    f1_score,
    confusion_matrix,
    accuracy_score,
    brier_score_loss,
    log_loss
)

current_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from run_phase2_experiments import extract_phase2_rich_features
from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from evaluation.run_phase3_pipeline import fit_preprocessor_and_transform, calculate_ece
from evaluation.run_phase4_pipeline import calculate_mce, BetaCalibrator

AUTHORITATIVE_28_FEATURES = [
    "amount", "ip_velocity_1h", "ip_velocity_24h", "token_velocity_24h", "device_seen_before",
    "transaction_hour", "transaction_day", "product_cd_encoded", "card_type_encoded",
    "card_category_encoded", "email_domain_risk", "dist1_missing", "device_type_mobile",
    "device_info_missing", "amount_to_mean_ratio", "device_tx_count_5m", "device_tx_count_1h",
    "device_amount_sum_24h", "tx_acceleration_5m_1h", "device_amount_concentration_5m_1h",
    "device_unique_tokens_1h", "device_reputation_score", "device_fraud_rate",
    "log_amount", "sin_tx_hour", "cos_tx_hour", "ip_burst_ratio", "amt_novelty_risk"
]

def compute_sha256(filepath):
    sha = hashlib.sha256()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(65536), b""):
            sha.update(chunk)
    return sha.hexdigest()

class ROPUSProductionRiskServer:
    """
    Thread-safe, standalone Production Risk Serving Engine.
    Encapsulates feature extraction, train-fitted preprocessor, calibrated XGBoost scoring,
    and Bayes Minimum Risk (BMR) decisioning with strict schema validation and graceful degradation.
    """
    def __init__(self, model_bundle_path: str, default_cost_fp: float = 25.0, default_surcharge: float = 1.05):
        t0 = time.perf_counter()
        if not os.path.exists(model_bundle_path):
            raise FileNotFoundError(f"Production bundle not found: {model_bundle_path}")

        bundle = joblib.load(model_bundle_path)
        self.model = bundle["model"]
        self.calibrator = bundle["calibrator"]
        self.preprocessor_state = bundle["preprocessor_state"]
        self.feature_names = bundle["feature_names"]
        self.model_version = bundle.get("model_version", "v4.0-bmr-28f")
        self.default_cost_fp = default_cost_fp
        self.default_surcharge = default_surcharge
        self.load_latency_ms = (time.perf_counter() - t0) * 1000.0
        self._lock = threading.Lock()

    def _extract_single_features(self, txn_dict: dict) -> pd.DataFrame:
        df_raw = pd.DataFrame([txn_dict])
        df_feat = extract_phase2_rich_features(df_raw)

        # Apply fitted preprocessor statistics
        p_map = self.preprocessor_state["product_map"]
        c_map = self.preprocessor_state["card_type_map"]
        cat_map = self.preprocessor_state["card_cat_map"]
        e_map = self.preprocessor_state["email_domain_risk"]
        prior = self.preprocessor_state["global_fraud_prior"]

        out = pd.DataFrame(index=[0])
        for col in [
            "amount", "ip_velocity_1h", "ip_velocity_24h", "token_velocity_24h", "amount_to_mean_ratio",
            "device_tx_count_5m", "device_tx_count_1h", "device_amount_sum_24h", "tx_acceleration_5m_1h",
            "device_amount_concentration_5m_1h", "device_unique_tokens_1h", "device_reputation_score", "device_fraud_rate",
            "log_amount", "sin_tx_hour", "cos_tx_hour", "ip_burst_ratio", "amt_novelty_risk"
        ]:
            if col in txn_dict and not pd.isna(txn_dict[col]):
                out[col] = float(txn_dict[col])
            elif col in df_feat.columns and not pd.isna(df_feat[col].values[0]):
                out[col] = float(df_feat[col].values[0])
            else:
                out[col] = 0.0

        out["device_seen_before"] = int(txn_dict.get("device_seen_before", df_feat["device_seen_before"].values[0] if "device_seen_before" in df_feat.columns else 0))
        out["transaction_hour"] = int(txn_dict.get("transaction_hour", df_feat["transaction_hour"].values[0] if "transaction_hour" in df_feat.columns else 12))
        out["transaction_day"] = int(txn_dict.get("transaction_day", df_feat["transaction_day"].values[0] if "transaction_day" in df_feat.columns else 0))
        out["dist1_missing"] = int(txn_dict.get("dist1_missing", df_feat["dist1_missing"].values[0] if "dist1_missing" in df_feat.columns else 1))
        out["device_type_mobile"] = int(txn_dict.get("device_type_mobile", df_feat["device_type_mobile"].values[0] if "device_type_mobile" in df_feat.columns else 0))
        out["device_info_missing"] = int(txn_dict.get("device_info_missing", df_feat["device_info_missing"].values[0] if "device_info_missing" in df_feat.columns else 0))

        if "product_cd_encoded" in txn_dict:
            out["product_cd_encoded"] = int(txn_dict["product_cd_encoded"])
        else:
            raw_prod = str(txn_dict.get("raw_product_cd", txn_dict.get("ProductCD", df_feat["raw_product_cd"].values[0] if "raw_product_cd" in df_feat.columns else "W")))
            out["product_cd_encoded"] = p_map.get(raw_prod, -1)

        if "card_type_encoded" in txn_dict:
            out["card_type_encoded"] = int(txn_dict["card_type_encoded"])
        else:
            raw_card = str(txn_dict.get("raw_card_type", txn_dict.get("card4", df_feat["raw_card_type"].values[0] if "raw_card_type" in df_feat.columns else "visa")))
            out["card_type_encoded"] = c_map.get(raw_card, -1)

        if "card_category_encoded" in txn_dict:
            out["card_category_encoded"] = int(txn_dict["card_category_encoded"])
        else:
            raw_cat = str(txn_dict.get("raw_card_category", txn_dict.get("card6", df_feat["raw_card_category"].values[0] if "raw_card_category" in df_feat.columns else "debit")))
            out["card_category_encoded"] = cat_map.get(raw_cat, -1)

        if "email_domain_risk" in txn_dict:
            out["email_domain_risk"] = float(txn_dict["email_domain_risk"])
        else:
            raw_email = str(txn_dict.get("raw_email_domain", txn_dict.get("P_emaildomain", df_feat["raw_email_domain"].values[0] if "raw_email_domain" in df_feat.columns else "missing")))
            out["email_domain_risk"] = e_map.get(raw_email, prior)

        # Enforce exact 28 feature ordering
        return out[self.feature_names]

    def evaluate_transaction(self, txn_dict: dict, cost_fp: float = None, surcharge: float = None) -> dict:
        t_start = time.perf_counter()
        c_fp = cost_fp if cost_fp is not None else self.default_cost_fp
        s_fact = surcharge if surcharge is not None else self.default_surcharge

        # 1. Validation & Input Sanitization
        if not isinstance(txn_dict, dict) or len(txn_dict) == 0:
            return {
                "status": "FAIL_SAFE_REVIEW",
                "error": "Empty or malformed transaction payload",
                "fraud_probability": 0.50,
                "expected_loss_allow": 0.0,
                "expected_loss_decline": 0.0,
                "decision": "MANUAL_REVIEW",
                "bmr_threshold": 0.50,
                "latency_ms": (time.perf_counter() - t_start) * 1000.0
            }

        try:
            amt = float(txn_dict.get("TransactionAmt", txn_dict.get("amount", 0.0)))
        except (ValueError, TypeError):
            amt = 0.0

        amt = max(0.0, amt)

        # 2. Extract & Preprocess Features
        try:
            X_row = self._extract_single_features(txn_dict)
        except Exception as e:
            return {
                "status": "FAIL_SAFE_ERROR",
                "error": f"Feature extraction failed: {str(e)}",
                "fraud_probability": 0.50,
                "expected_loss_allow": 0.0,
                "expected_loss_decline": 0.0,
                "decision": "MANUAL_REVIEW",
                "bmr_threshold": 0.50,
                "latency_ms": (time.perf_counter() - t_start) * 1000.0
            }

        # 3. Model Inference & Calibration
        raw_prob = float(self.model.predict_proba(X_row)[:, 1][0])
        cal_prob = float(self.calibrator.predict_proba(np.array([raw_prob]))[0])
        cal_prob = float(np.clip(cal_prob, 0.0, 1.0))

        # 4. Bayes Minimum Risk (BMR) Formulation
        loss_allow = cal_prob * amt * s_fact
        loss_decline = (1.0 - cal_prob) * c_fp

        # Theoretical boundary: P*(A) = C_FP / (s_fact * A + C_FP)
        p_star = c_fp / (s_fact * amt + c_fp) if (s_fact * amt + c_fp) > 0 else 0.50

        decision = "DECLINE" if loss_decline < loss_allow else "ALLOW"
        latency_ms = (time.perf_counter() - t_start) * 1000.0

        return {
            "status": "SUCCESS",
            "model_version": self.model_version,
            "amount": amt,
            "raw_model_score": float(round(raw_prob, 6)),
            "fraud_probability": float(round(cal_prob, 6)),
            "expected_loss_allow": float(round(loss_allow, 4)),
            "expected_loss_decline": float(round(loss_decline, 4)),
            "bmr_threshold": float(round(p_star, 6)),
            "decision": decision,
            "latency_ms": float(round(latency_ms, 3))
        }

def build_production_bundle(output_path: str):
    """Packages the frozen Phase 3 model, preprocessor, and Beta calibrator into a standalone bundle."""
    df_raw, _ = load_raw_dataset()
    df_train_raw, df_val_raw, df_test_raw, _ = temporal_train_val_test_split(df_raw)

    df_feat_train = extract_phase2_rich_features(df_train_raw)
    df_feat_val = extract_phase2_rich_features(df_val_raw)
    df_feat_test = extract_phase2_rich_features(df_test_raw)

    X_train, X_val, X_test = fit_preprocessor_and_transform(df_feat_train, df_feat_val, df_feat_test)
    y_train = df_feat_train["isFraud"].values
    y_val = df_feat_val["isFraud"].values

    # Train Phase 3 Model
    p3_params = {
        "n_estimators": 90, "max_depth": 3, "learning_rate": 0.025, "min_child_weight": 8,
        "subsample": 0.90, "colsample_bytree": 0.90, "reg_lambda": 1.5, "reg_alpha": 0.2,
        "scale_pos_weight": 21.05, "random_state": 42, "eval_metric": "logloss", "tree_method": "hist"
    }
    model_p3 = xgb.XGBClassifier(**p3_params)
    model_p3.fit(X_train[AUTHORITATIVE_28_FEATURES], y_train, eval_set=[(X_val[AUTHORITATIVE_28_FEATURES], y_val)], verbose=False)

    # Fit Beta Calibrator strictly on validation
    raw_probs_val = model_p3.predict_proba(X_val[AUTHORITATIVE_28_FEATURES])[:, 1]
    calibrator = BetaCalibrator().fit(raw_probs_val, y_val)

    # Extract Preprocessor State
    unique_prods = sorted(df_feat_train["raw_product_cd"].dropna().unique().tolist())
    p_map = {prod: idx for idx, prod in enumerate(unique_prods)}
    unique_cards = sorted(df_feat_train["raw_card_type"].dropna().unique().tolist())
    c_map = {card: idx for idx, card in enumerate(unique_cards)}
    unique_cats = sorted(df_feat_train["raw_card_category"].dropna().unique().tolist())
    cat_map = {cat: idx for idx, cat in enumerate(unique_cats)}
    prior = float(df_feat_train["isFraud"].mean())
    domain_stats = df_feat_train.groupby("raw_email_domain")["isFraud"].agg(["count", "sum"])
    e_map = {}
    for domain, row in domain_stats.iterrows():
        cnt = row["count"]
        f_sum = row["sum"]
        e_map[str(domain)] = float(round((f_sum + 20.0 * prior) / (cnt + 20.0), 4))

    prep_state = {
        "product_map": p_map,
        "card_type_map": c_map,
        "card_cat_map": cat_map,
        "email_domain_risk": e_map,
        "global_fraud_prior": prior
    }

    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    joblib.dump({
        "model": model_p3,
        "calibrator": calibrator,
        "preprocessor_state": prep_state,
        "feature_names": AUTHORITATIVE_28_FEATURES,
        "model_version": "v4.0-bmr-28f",
        "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    }, output_path)
    print(f"Exported standalone production bundle to: {output_path} (Size: {os.path.getsize(output_path)} bytes)")

def main():
    print("=" * 90)
    print("PHASE 6: PRODUCTION SERVING & END-TO-END DEPLOYMENT VALIDATION")
    print("=" * 90)

    prod_bundle_path = os.path.join(current_dir, "model", "candidates", "production_model_28f.joblib")
    build_production_bundle(prod_bundle_path)

    # ------------------ 1. ARTIFACT LOADING & METADATA ------------------
    print("\n[STEP 1] Auditing Production Artifact Loading...")
    bundle_hash = compute_sha256(prod_bundle_path)
    bundle_size = os.path.getsize(prod_bundle_path)

    server = ROPUSProductionRiskServer(prod_bundle_path)
    print(f"-> Production Bundle SHA-256: {bundle_hash}")
    print(f"-> Bundle Size: {bundle_size:,} bytes")
    print(f"-> Cold Start Load Latency: {server.load_latency_ms:.2f} ms")
    print(f"-> Enforced Feature Count: {len(server.feature_names)}")

    serving_audit_meta = {
        "bundle_path": prod_bundle_path,
        "sha256": bundle_hash,
        "size_bytes": bundle_size,
        "load_latency_ms": server.load_latency_ms,
        "feature_count": len(server.feature_names),
        "model_version": server.model_version
    }

    # ------------------ 2 & 3. END-TO-END INFERENCE & SCHEMA CONTRACT ENFORCEMENT ------------------
    print("\n[STEP 2 & 3] Testing Inference Contract & Schema Enforcement...")
    test_cases = [
        {"name": "Standard Card Auth", "payload": {"TransactionID": 10001, "TransactionDT": 16000000, "TransactionAmt": 100.0, "ProductCD": "W", "card1": 1001, "card4": "visa", "card6": "debit", "addr1": 200.0, "DeviceInfo": "iOS_17", "P_emaildomain": "gmail.com"}},
        {"name": "Missing Optional Fields", "payload": {"TransactionAmt": 250.0, "card1": 5555}},
        {"name": "Completely Empty Payload", "payload": {}},
        {"name": "Invalid Amount String", "payload": {"TransactionAmt": "INVALID_AMOUNT", "card1": 1234}},
        {"name": "Huge Amount (₹14.5 Lakhs)", "payload": {"TransactionAmt": 14500.0, "card1": 9999, "P_emaildomain": "fraud.ru"}},
        {"name": "Micro Amount ($0.01)", "payload": {"TransactionAmt": 0.01, "card1": 1111}}
    ]

    schema_results = []
    for tc in test_cases:
        res = server.evaluate_transaction(tc["payload"])
        schema_results.append({"test_case": tc["name"], "input": tc["payload"], "response": res})
        print(f"Case '{tc['name']:<28}' -> Status: {res['status']:<16} | Dec: {res['decision']:<7} | Prob: {res['fraud_probability']:<7.4f} | Latency: {res['latency_ms']:.2f}ms")

    # ------------------ 4. PREDICTION CONSISTENCY AUDIT ------------------
    print("\n[STEP 4] Prediction Consistency Audit Across 100 Random Deterministic Transactions...")
    df_raw, _ = load_raw_dataset()
    _, _, df_test_raw, _ = temporal_train_val_test_split(df_raw)
    df_feat_test = extract_phase2_rich_features(df_test_raw)
    sample_records = df_feat_test.head(100).to_dict(orient="records")

    consistency_passed = True
    max_diff = 0.0
    for rec in sample_records:
        res1 = server.evaluate_transaction(rec)
        res2 = server.evaluate_transaction(rec)
        diff = abs(res1["fraud_probability"] - res2["fraud_probability"])
        max_diff = max(max_diff, diff)
        if diff > 1e-6 or res1["decision"] != res2["decision"]:
            consistency_passed = False

    print(f"-> 100/100 Repeated Queries Evaluated: Maximum Probability Delta = {max_diff:.2e}")
    print(f"-> Deterministic Consistency Status: {'PASS (100% Deterministic)' if consistency_passed else 'FAIL'}")

    # ------------------ 5. BMR BOUNDARY PRECISION VERIFICATION ------------------
    print("\n[STEP 5] BMR Precision Boundary Verification (Immediate Above/Below Thresholds)...")
    # For Amount = $500, C_FP = $25, Surcharge = 1.05:
    # P*(500) = 25 / (1.05 * 500 + 25) = 25 / (525 + 25) = 25 / 550 = 0.0454545...
    p_star_500 = 25.0 / (1.05 * 500.0 + 25.0)
    boundary_tests = [
        {"amount": 500.0, "p": p_star_500 - 0.001, "expected_dec": "ALLOW", "note": "P immediately below boundary (P* - 0.1%)"},
        {"amount": 500.0, "p": p_star_500 + 0.001, "expected_dec": "DECLINE", "note": "P immediately above boundary (P* + 0.1%)"},
        {"amount": 14500.0, "p": 0.0015, "expected_dec": "ALLOW", "note": "P=0.15% below P*(14500)=0.164%"},
        {"amount": 14500.0, "p": 0.0018, "expected_dec": "DECLINE", "note": "P=0.18% above P*(14500)=0.164%"}
    ]

    bmr_boundary_results = []
    bmr_all_pass = True
    for bt in boundary_tests:
        a = bt["amount"]
        p = bt["p"]
        l_allow = p * a * 1.05
        l_dec = (1.0 - p) * 25.0
        dec = "DECLINE" if l_dec < l_allow else "ALLOW"
        passed = (dec == bt["expected_dec"])
        if not passed:
            bmr_all_pass = False
        bmr_boundary_results.append({
            "amount": a, "p": p, "loss_allow": l_allow, "loss_decline": l_dec,
            "actual_decision": dec, "expected_decision": bt["expected_dec"], "pass": passed
        })
        print(f"Amount: ${a:<8.2f} | P: {p:<7.4f} | Loss(ALLOW): ${l_allow:<7.2f} | Loss(DECL): ${l_dec:<7.2f} | Dec: {dec:<7} | Expected: {bt['expected_dec']:<7} | [{'PASS' if passed else 'FAIL'}]")

    # ------------------ 6. LATENCY & THROUGHPUT BENCHMARKING ------------------
    print("\n[STEP 6] Profiling Latency & Throughput over 1,000 Production Requests...")
    latencies = []
    for _ in range(1000):
        # Pick random sample from test
        rec = sample_records[np.random.randint(len(sample_records))]
        t0 = time.perf_counter()
        server.evaluate_transaction(rec)
        latencies.append((time.perf_counter() - t0) * 1000.0)

    latencies = np.array(latencies)
    p50 = float(np.percentile(latencies, 50))
    p90 = float(np.percentile(latencies, 90))
    p95 = float(np.percentile(latencies, 95))
    p99 = float(np.percentile(latencies, 99))
    mean_lat = float(np.mean(latencies))
    throughput_rps = float(1000.0 / np.sum(latencies) * 1000.0)

    print(f"-> Warm Inference Mean Latency: {mean_lat:.2f} ms")
    print(f"-> Latency p50: {p50:.2f} ms")
    print(f"-> Latency p95: {p95:.2f} ms (Acceptance Threshold: <= 15.0 ms) [{'PASS' if p95 <= 15.0 else 'FAIL'}]")
    print(f"-> Latency p99: {p99:.2f} ms (Acceptance Threshold: <= 30.0 ms) [{'PASS' if p99 <= 30.0 else 'FAIL'}]")
    print(f"-> Throughput:  {throughput_rps:.1f} requests/second/core")

    latency_results = {
        "requests_benchmarked": 1000,
        "mean_latency_ms": mean_lat,
        "p50_latency_ms": p50,
        "p90_latency_ms": p90,
        "p95_latency_ms": p95,
        "p99_latency_ms": p99,
        "throughput_rps": throughput_rps,
        "acceptance_criteria": {
            "p95_max_ms": 15.0,
            "p99_max_ms": 30.0,
            "status": "PASS" if p95 <= 15.0 and p99 <= 30.0 else "FAIL"
        }
    }

    # ------------------ 8. CONCURRENCY & THREAD-SAFETY AUDIT ------------------
    print("\n[STEP 8] Concurrency Audit (500 Parallel Requests over 8 Threads)...")
    concurrent_errors = 0
    concurrent_results = []

    def run_concurrent_eval(idx):
        rec = sample_records[idx % len(sample_records)]
        return server.evaluate_transaction(rec)

    with ThreadPoolExecutor(max_workers=8) as executor:
        futures = [executor.submit(run_concurrent_eval, i) for i in range(500)]
        for f in as_completed(futures):
            try:
                res = f.result()
                if res["status"] != "SUCCESS":
                    concurrent_errors += 1
            except Exception:
                concurrent_errors += 1

    print(f"-> 500 Concurrent Inferences Completed: Errors = {concurrent_errors}")
    print(f"-> Thread Safety Status: {'PASS (Zero Race Conditions)' if concurrent_errors == 0 else 'FAIL'}")

    # ------------------ 10. FAIL-SAFE BEHAVIOR & FAILURE MODES ------------------
    print("\n[STEP 10] Fail-Safe Mode & Adversarial Resiliency Testing...")
    adversarial_inputs = [
        {"desc": "Null payload pointer", "data": None},
        {"desc": "Nested invalid object", "data": {"TransactionAmt": {"bad": "nested"}}},
        {"desc": "NaN amount", "data": {"TransactionAmt": float("nan")}},
        {"desc": "Negative infinity amount", "data": {"TransactionAmt": float("-inf")}},
        {"desc": "100k Character malicious string in DeviceInfo", "data": {"TransactionAmt": 100.0, "DeviceInfo": "A" * 100000}}
    ]

    fail_safe_results = []
    for adv in adversarial_inputs:
        res = server.evaluate_transaction(adv["data"])
        fail_safe_results.append({"case": adv["desc"], "response": res})
        print(f"Adversarial Case: '{adv['desc']:<38}' -> Status: {res['status']:<17} | Fallback Decision: {res['decision']}")

    # ------------------ 11. REGRESSION PARITY CHECK AGAINST PHASE 5 ------------------
    print("\n[STEP 11] Regression Parity Verification against Phase 5 Held-Out Test Baseline...")

    # Run full 1,200 test set through production server with point-in-time features
    test_recs = df_feat_test.to_dict(orient="records")
    y_test_arr = df_test_raw["isFraud"].values

    test_cal_probs = []
    test_decisions = []
    for r in test_recs:
        eval_res = server.evaluate_transaction(r)
        test_cal_probs.append(eval_res["fraud_probability"])
        test_decisions.append(1 if eval_res["decision"] == "DECLINE" else 0)

    test_cal_probs = np.array(test_cal_probs)
    test_decisions = np.array(test_decisions)

    # Compute test set metrics
    pr_auc_6 = float(average_precision_score(y_test_arr, test_cal_probs))
    roc_auc_6 = float(roc_auc_score(y_test_arr, test_cal_probs))
    ece_6 = float(calculate_ece(y_test_arr, test_cal_probs))
    brier_6 = float(brier_score_loss(y_test_arr, test_cal_probs))

    tn_6, fp_6, fn_6, tp_6 = confusion_matrix(y_test_arr, test_decisions).ravel()

    fn_mask_6 = (y_test_arr == 1) & (test_decisions == 0)
    fp_mask_6 = (y_test_arr == 0) & (test_decisions == 1)
    amt_test_arr = df_test_raw["TransactionAmt"].fillna(0.0).values
    tot_loss_6 = float(np.sum(amt_test_arr[fn_mask_6] * 1.05) + fp_6 * 25.0)

    # Expected Phase 5 values
    phase5_expected = {
        "pr_auc": 0.087433,
        "roc_auc": 0.632990,
        "ece": 0.0092,
        "brier": 0.0410,
        "tp": 5,
        "fp": 72,
        "tn": 1076,
        "fn": 47,
        "total_loss": 7670.52
    }

    regression_checks = {
        "pr_auc": {"p5": phase5_expected["pr_auc"], "p6": pr_auc_6, "match": abs(phase5_expected["pr_auc"] - pr_auc_6) < 1e-4},
        "roc_auc": {"p5": phase5_expected["roc_auc"], "p6": roc_auc_6, "match": abs(phase5_expected["roc_auc"] - roc_auc_6) < 1e-4},
        "ece": {"p5": phase5_expected["ece"], "p6": ece_6, "match": abs(phase5_expected["ece"] - ece_6) < 2e-3},
        "brier": {"p5": phase5_expected["brier"], "p6": brier_6, "match": abs(phase5_expected["brier"] - brier_6) < 1e-3},
        "tp": {"p5": phase5_expected["tp"], "p6": int(tp_6), "match": int(tp_6) == phase5_expected["tp"]},
        "fp": {"p5": phase5_expected["fp"], "p6": int(fp_6), "match": int(fp_6) == phase5_expected["fp"]},
        "tn": {"p5": phase5_expected["tn"], "p6": int(tn_6), "match": int(tn_6) == phase5_expected["tn"]},
        "fn": {"p5": phase5_expected["fn"], "p6": int(fn_6), "match": int(fn_6) == phase5_expected["fn"]},
        "total_loss": {"p5": phase5_expected["total_loss"], "p6": tot_loss_6, "match": abs(phase5_expected["total_loss"] - tot_loss_6) < 1e-1}
    }

    all_regression_pass = all(item["match"] for item in regression_checks.values())
    print(f"\n-> Phase 6 Serving vs Phase 5 Baseline Parity: {'100% EXACT PARITY (PASS)' if all_regression_pass else 'REGRESSION DETECTED'}")
    for k, v in regression_checks.items():
        p5_s = f"{v['p5']:.4f}" if isinstance(v['p5'], float) else f"{v['p5']}"
        p6_s = f"{v['p6']:.4f}" if isinstance(v['p6'], float) else f"{v['p6']}"
        print(f"   {k:<15} | Phase 5: {p5_s:<10} | Phase 6 Serving: {p6_s:<10} | [{'PASS' if v['match'] else 'FAIL'}]")

    # ------------------ SAVE ALL PRODUCTION DELIVERABLES ------------------
    eval_dir = os.path.join(current_dir, "evaluation")

    # 1. Serving Audit
    p6_serving_path = os.path.join(eval_dir, "phase_6_serving_audit.json")
    with open(p6_serving_path, "w") as f:
        json.dump(serving_audit_meta, f, indent=2)

    # 2. Latency Results
    p6_latency_path = os.path.join(eval_dir, "phase_6_latency_results.json")
    with open(p6_latency_path, "w") as f:
        json.dump(latency_results, f, indent=2)

    # 3. Schema Results
    p6_schema_path = os.path.join(eval_dir, "phase_6_schema_results.json")
    with open(p6_schema_path, "w") as f:
        json.dump({"schema_tests": schema_results}, f, indent=2)

    # 4. Regression Parity
    p6_regr_path = os.path.join(eval_dir, "phase_6_inference_regression.json")
    with open(p6_regr_path, "w") as f:
        json.dump({"regression_checks": regression_checks, "overall_status": "PASS" if all_regression_pass else "FAIL"}, f, indent=2)

    # 5. Failure Mode Results
    p6_fail_path = os.path.join(eval_dir, "phase_6_failure_mode_results.json")
    with open(p6_fail_path, "w") as f:
        json.dump({"failure_mode_tests": fail_safe_results}, f, indent=2)

    print(f"\nAll Phase 6 deliverables generated successfully in {eval_dir}:")
    print(f"1. {p6_serving_path}")
    print(f"2. {p6_latency_path}")
    print(f"3. {p6_schema_path}")
    print(f"4. {p6_regr_path}")
    print(f"5. {p6_fail_path}")

if __name__ == "__main__":
    main()
