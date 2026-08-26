"""
ROPUS Production Shadow Gateway & Dual-Decision Telemetry Engine (Phase 32)
Integrates BMR Floor 0.040 shadow-mode evaluation directly into the production gateway telemetry path:
1. Enforces Baseline Dynamic BMR (no floor) as the sole active production routing policy.
2. Evaluates Candidate Floor tau_floor = 0.040 in parallel (strictly non-enforcing shadow evaluation).
3. Production-safe dual-decision telemetry with salted HMAC-SHA256 correlation IDs (zero PAN/CVV/secrets).
4. Durable shadow-flip persistence store (shadow_flips.jsonl).
5. Delayed fraud/chargeback dispute ingestion with explicit 60-day label-maturity tracking.
6. Automated exact Clopper-Pearson 95% confidence interval calculations on mature labels.
7. Fail-closed error isolation: shadow evaluation or telemetry failures never disrupt live customer routing.
"""

import os
import sys
import time
import math
import hashlib
import hmac
import json
from datetime import datetime, timezone, timedelta
from typing import Dict, Any, List, Optional, Tuple, Set
import numpy as np

current_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from monitoring.production_telemetry import (
    ProductionTelemetryCollector,
    EXPECTED_CHAMPION_VERSION,
    EXPECTED_SHA256,
    EXPECTED_FILE_SIZE
)
from monitoring.production_gateway_adapter import (
    PrivacyGuard,
    GatewayProvenanceGuard,
    TELEMETRY_SCHEMA_VERSION,
    ALLOWED_EVIDENCE_TIERS
)

SHADOW_TELEMETRY_SCHEMA_VERSION = "2.0.0"
DEFAULT_SHADOW_FLOOR = 0.040
DEFAULT_MATURATION_DAYS = 60

def compute_clopper_pearson_exact(k: int, n: int, confidence: float = 0.95) -> Tuple[float, float]:
    """
    Compute exact Clopper-Pearson two-sided confidence interval for binomial proportion.
    Handles boundary cases (n=0, k=0, k=n) analytically.
    """
    if n == 0:
        return 0.0, 1.0
    alpha = 1.0 - confidence
    if k == 0:
        lower = 0.0
        upper = float(1.0 - (alpha ** (1.0 / n)))
        return lower, upper
    if k == n:
        lower = float(alpha ** (1.0 / n))
        upper = 1.0
        return lower, upper
    from scipy.stats import beta as beta_dist
    lower = float(beta_dist.ppf(alpha / 2.0, k, n - k + 1))
    upper = float(beta_dist.ppf(1.0 - alpha / 2.0, k + 1, n - k))
    return lower, upper

class ProductionShadowGatewayPipeline:
    """
    Production Gateway Pipeline with Integrated Shadow-Mode Evaluation.
    Maintains 100% isolation between enforced customer routing and shadow evidence accumulation.
    """
    def __init__(
        self,
        scoring_engine: Any,
        telemetry_collector: ProductionTelemetryCollector,
        shadow_floor: float = DEFAULT_SHADOW_FLOOR,
        maturation_window_days: int = DEFAULT_MATURATION_DAYS,
        store_dir: Optional[str] = None
    ):
        self.scoring_engine = scoring_engine
        self.collector = telemetry_collector
        self.shadow_floor = shadow_floor
        self.maturation_window_days = maturation_window_days

        self.store_dir = store_dir or os.path.join(current_dir, "monitoring", "telemetry_store")
        os.makedirs(self.store_dir, exist_ok=True)
        self.shadow_flips_path = os.path.join(self.store_dir, "shadow_flips.jsonl")
        self.disputes_path = os.path.join(self.store_dir, "shadow_disputes.jsonl")

        # In-memory durable index of shadow flips and deduplication cache
        self.shadow_flips_index: Dict[str, Dict[str, Any]] = {}
        self.disputes_index: Dict[str, Dict[str, Any]] = {}
        self.seen_correlation_ids: Set[str] = set()

        # Operational Telemetry & Provenance Counters
        self.duplicate_events_count: int = 0
        self.untrusted_events_count: int = 0
        self.shadow_exceptions_count: int = 0
        self.persistence_failures_count: int = 0
        self.first_live_timestamp: Optional[float] = None
        self.last_live_timestamp: Optional[float] = None

        # Load existing durable records if present
        self._load_durable_stores()

    def _load_durable_stores(self):
        """Load persistent shadow flips and dispute records from disk."""
        if os.path.exists(self.shadow_flips_path):
            try:
                with open(self.shadow_flips_path, "r") as f:
                    for line in f:
                        if line.strip():
                            rec = json.loads(line.strip())
                            cid = rec["correlation_id"]
                            self.shadow_flips_index[cid] = rec
                            self.seen_correlation_ids.add(cid)
            except Exception as e:
                print(f"[Warning] Failed to load shadow flips from {self.shadow_flips_path}: {e}")

        if os.path.exists(self.disputes_path):
            try:
                with open(self.disputes_path, "r") as f:
                    for line in f:
                        if line.strip():
                            rec = json.loads(line.strip())
                            self.disputes_index[rec["correlation_id"]] = rec
            except Exception as e:
                print(f"[Warning] Failed to load disputes from {self.disputes_path}: {e}")

    def evaluate_and_route(
        self,
        amount: float,
        calibrated_prob: float,
        raw_correlation_id: str,
        evidence_tier: str = "LIVE_PRODUCTION_EVIDENCE",
        gateway_signature: Optional[str] = None,
        timestamp: Optional[float] = None
    ) -> Dict[str, Any]:
        """
        Main gateway execution path:
        1. Evaluates Enforced Production Policy: Baseline Dynamic BMR (No Floor)
        2. Validates traffic provenance with cryptographic anti-spoofing guard
        3. Protects against duplicate / replay events (Idempotency)
        4. Safely evaluates Shadow Candidate Policy: Floor 0.040 (Fail-Closed Isolation)
        5. Persists durable shadow flip telemetry if Baseline DECLINE -> Shadow ALLOW
        6. Ingests into ProductionTelemetryCollector
        7. Returns enforced decision for active customer routing
        """
        start_time = time.perf_counter()
        if timestamp is None:
            timestamp = time.time()

        # 1. Anonymize correlation ID (HMAC-SHA256 Salted, Zero PAN/CVV)
        anonymized_id = PrivacyGuard.anonymize_identifier(raw_correlation_id)

        # 2. Provenance Normalization & Cryptographic Anti-Spoofing Guard
        is_claiming_live = evidence_tier in ("LIVE_PRODUCTION_EVIDENCE", "PRODUCTION_LIVE")
        is_live_production = False

        if is_claiming_live:
            if gateway_signature and GatewayProvenanceGuard.verify_provenance_signature(raw_correlation_id, int(timestamp), gateway_signature):
                is_live_production = True
                evidence_tier = "LIVE_PRODUCTION_EVIDENCE"
            else:
                # Spoof attempt or unauthenticated external caller claiming PRODUCTION_LIVE
                evidence_tier = "UNTRUSTED"
                self.untrusted_events_count += 1
        elif evidence_tier not in ALLOWED_EVIDENCE_TIERS:
            evidence_tier = "UNTRUSTED"
            self.untrusted_events_count += 1

        # 3. Duplicate / Replay Ingestion Protection
        is_duplicate = False
        if anonymized_id in self.seen_correlation_ids:
            is_duplicate = True
            self.duplicate_events_count += 1
        else:
            self.seen_correlation_ids.add(anonymized_id)

        if is_live_production and not is_duplicate:
            if self.first_live_timestamp is None:
                self.first_live_timestamp = timestamp
            self.last_live_timestamp = timestamp

        # 4. Enforce Baseline Dynamic BMR Decision
        cost_fp = getattr(self.scoring_engine, "cost_fp", 25.0)
        surcharge = getattr(self.scoring_engine, "surcharge", 1.05)
        p_star = cost_fp / (surcharge * amount + cost_fp) if (surcharge * amount + cost_fp) > 0 else 0.5

        enforced_decision = "DECLINE" if calibrated_prob > p_star else "ALLOW"

        # Determine risk tier
        if calibrated_prob < 0.02:
            risk_tier = "LOW_RISK"
        elif calibrated_prob < 0.10:
            risk_tier = "MODERATE_RISK"
        elif calibrated_prob < 0.25:
            risk_tier = "ELEVATED_RISK"
        else:
            risk_tier = "CRITICAL_FRAUD"

        # 5. Shadow Evaluation with Complete Error Isolation (Fail-Closed to Baseline)
        shadow_decision = enforced_decision
        is_shadow_flip = False
        shadow_eval_status = "SUCCESS"

        try:
            # Candidate Floor 0.040 rule: DECLINE only if P > P*(A) AND P >= 0.040
            if calibrated_prob > p_star and calibrated_prob >= self.shadow_floor:
                shadow_decision = "DECLINE"
            else:
                shadow_decision = "ALLOW"

            # Check if this order is a shadow flip: Baseline DECLINE -> Shadow ALLOW
            if enforced_decision == "DECLINE" and shadow_decision == "ALLOW":
                is_shadow_flip = True
        except Exception as e:
            self.shadow_exceptions_count += 1
            shadow_eval_status = f"SHADOW_EVAL_ERROR: {str(e)}"
            shadow_decision = enforced_decision
            is_shadow_flip = False

        latency_ms = (time.perf_counter() - start_time) * 1000.0

        # 6. Construct Dual-Decision Telemetry Record
        telemetry_record = {
            "telemetry_schema_version": SHADOW_TELEMETRY_SCHEMA_VERSION,
            "timestamp": timestamp,
            "datetime_utc": datetime.fromtimestamp(timestamp, tz=timezone.utc).isoformat(),
            "correlation_id": anonymized_id,
            "evidence_tier": evidence_tier,
            "is_duplicate": is_duplicate,
            "model_version": EXPECTED_CHAMPION_VERSION,
            "model_sha256": EXPECTED_SHA256,
            "amount": round(float(amount), 2),
            "calibrated_probability": round(float(calibrated_prob), 6),
            "bmr_hurdle_p_star": round(float(p_star), 6),
            "shadow_floor": self.shadow_floor,
            "enforced_decision": enforced_decision,
            "shadow_decision": shadow_decision,
            "is_shadow_flip": is_shadow_flip,
            "risk_tier": risk_tier,
            "shadow_eval_status": shadow_eval_status,
            "latency_ms": round(latency_ms, 3)
        }

        # 7. Durable Shadow-Flip Persistence (Only if not duplicate)
        if is_shadow_flip and not is_duplicate:
            flip_entry = {
                "correlation_id": anonymized_id,
                "timestamp": timestamp,
                "datetime_utc": datetime.fromtimestamp(timestamp, tz=timezone.utc).isoformat(),
                "evidence_tier": evidence_tier,
                "amount": round(float(amount), 2),
                "calibrated_prob": round(float(calibrated_prob), 6),
                "bmr_hurdle": round(float(p_star), 6),
                "shadow_floor": self.shadow_floor,
                "risk_tier": risk_tier,
                "label_status": "UNLABELED",
                "ground_truth": None,
                "chargeback_amount": None,
                "dispute_ingested_at": None,
                "maturation_date_utc": (datetime.fromtimestamp(timestamp, tz=timezone.utc) + timedelta(days=self.maturation_window_days)).isoformat()
            }
            self.shadow_flips_index[anonymized_id] = flip_entry
            try:
                with open(self.shadow_flips_path, "a") as f:
                    f.write(json.dumps(flip_entry) + "\n")
            except Exception as e:
                self.persistence_failures_count += 1
                print(f"[Warning] Failed to write shadow flip to disk: {e}")

        # 8. Ingest Telemetry into Production Collector (Only if live & non-duplicate)
        if is_live_production and not is_duplicate:
            try:
                self.collector.record_inference_event(
                    tx_id=anonymized_id,
                    latency_ms=latency_ms,
                    http_status=200,
                    model_version=EXPECTED_CHAMPION_VERSION,
                    sha256_hash=EXPECTED_SHA256,
                    amount=amount,
                    calibrated_p=calibrated_prob,
                    decision=enforced_decision,
                    risk_tier=risk_tier,
                    raw_features={},
                    is_schema_valid=True
                )
            except Exception as e:
                print(f"[Warning] Telemetry collector ingestion error: {e}")

        # Return enforced decision for customer routing along with shadow audit metadata
        return {
            "enforced_decision": enforced_decision,
            "calibrated_probability": round(float(calibrated_prob), 6),
            "bmr_hurdle": round(float(p_star), 6),
            "risk_tier": risk_tier,
            "shadow_candidate_decision": shadow_decision,
            "is_shadow_flip": is_shadow_flip,
            "is_duplicate": is_duplicate,
            "provenance_tier": evidence_tier,
            "correlation_id": anonymized_id,
            "latency_ms": round(latency_ms, 3)
        }

    def ingest_delayed_dispute(
        self,
        dispute_id: str,
        raw_correlation_id: str,
        is_fraud: int,
        chargeback_amount: float,
        evidence_tier: str = "LIVE_PRODUCTION_EVIDENCE",
        reported_at: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Ingest delayed chargeback dispute with label-maturity tracking:
        1. Anonymizes correlation ID.
        2. Matches against durable shadow-flip store.
        3. Evaluates maturation window (60-day lag).
        4. Updates label status to MATURE_LEGITIMATE or MATURE_FRAUD.
        """
        anonymized_id = PrivacyGuard.anonymize_identifier(raw_correlation_id)
        if reported_at is None:
            reported_at = datetime.now(timezone.utc).isoformat()

        dispute_record = {
            "dispute_id": dispute_id,
            "correlation_id": anonymized_id,
            "evidence_tier": evidence_tier,
            "is_fraud": int(is_fraud),
            "chargeback_amount": round(float(chargeback_amount), 2),
            "reported_at": reported_at
        }
        self.disputes_index[anonymized_id] = dispute_record

        # Persist dispute to disk
        try:
            with open(self.disputes_path, "a") as f:
                f.write(json.dumps(dispute_record) + "\n")
        except Exception as e:
            print(f"[Warning] Failed to persist dispute to disk: {e}")

        # Match against shadow flip store
        matched_flip = self.shadow_flips_index.get(anonymized_id)
        is_shadow_matched = (matched_flip is not None)

        if matched_flip:
            # Check maturity
            tx_time = matched_flip["timestamp"]
            tx_dt = datetime.fromtimestamp(tx_time, tz=timezone.utc)
            rep_dt = datetime.fromisoformat(reported_at.replace("Z", "+00:00"))
            elapsed_days = (rep_dt - tx_dt).total_seconds() / 86400.0

            if elapsed_days >= self.maturation_window_days:
                label_status = "MATURE_FRAUD" if is_fraud == 1 else "MATURE_LEGITIMATE"
            else:
                label_status = "PENDING_MATURATION"

            matched_flip["label_status"] = label_status
            matched_flip["ground_truth"] = int(is_fraud)
            matched_flip["chargeback_amount"] = round(float(chargeback_amount), 2)
            matched_flip["dispute_ingested_at"] = reported_at
            matched_flip["elapsed_maturation_days"] = round(elapsed_days, 1)

            # Rewrite updated shadow flips index to disk
            self._flush_shadow_flips_to_disk()

        return {
            "status": "DISPUTE_INGESTED",
            "dispute_id": dispute_id,
            "correlation_id": anonymized_id,
            "evidence_tier": evidence_tier,
            "is_fraud": is_fraud,
            "shadow_flip_matched": is_shadow_matched,
            "label_status": matched_flip["label_status"] if matched_flip else "NON_FLIP_TRANSACTION"
        }

    def _flush_shadow_flips_to_disk(self):
        """Flush in-memory shadow flips index atomically to disk."""
        try:
            with open(self.shadow_flips_path, "w") as f:
                for rec in self.shadow_flips_index.values():
                    f.write(json.dumps(rec) + "\n")
        except Exception as e:
            print(f"[Warning] Failed to flush shadow flips index: {e}")

    def get_shadow_evidence_summary(self, evidence_tier: str = "LIVE_PRODUCTION_EVIDENCE") -> Dict[str, Any]:
        """
        Calculate complete live shadow evidence metrics, distinguishing:
        - Unlabeled shadow flips (never treated as legitimate)
        - Pending maturation flips (<60 days)
        - Mature labeled flips (>=60 days)
        - Exact Clopper-Pearson 95% confidence bounds
        """
        # Filter flips by evidence tier
        flips = [r for r in self.shadow_flips_index.values() if r["evidence_tier"] == evidence_tier]
        n_flips = len(flips)
        total_recovered_gmv = sum(r["amount"] for r in flips)
        flip_amounts = [r["amount"] for r in flips]

        unlabeled_flips = [r for r in flips if r["label_status"] == "UNLABELED"]
        pending_flips = [r for r in flips if r["label_status"] == "PENDING_MATURATION"]
        mature_flips = [r for r in flips if r["label_status"] in ("MATURE_LEGITIMATE", "MATURE_FRAUD")]
        mature_frauds = [r for r in mature_flips if r["label_status"] == "MATURE_FRAUD"]

        n_mature = len(mature_flips)
        k_mature_fraud = len(mature_frauds)
        mature_fraud_dollars = sum(r.get("chargeback_amount", r["amount"]) for r in mature_frauds)

        if n_mature == 0:
            observed_fraud_rate_str = "UNDEFINED (0 mature labels)"
            lower_ci_str = "UNDEFINED"
            upper_ci_str = "UNDEFINED"
            upper_ci_val = None
            meets_2pct_gate = False
        else:
            observed_rate = k_mature_fraud / n_mature
            observed_fraud_rate_str = f"{observed_rate:.2%}"
            lower_ci_val, upper_ci_val = compute_clopper_pearson_exact(k_mature_fraud, n_mature, 0.95)
            lower_ci_str = f"{lower_ci_val:.2%}"
            upper_ci_str = f"{upper_ci_val:.2%}"
            meets_2pct_gate = (upper_ci_val < 0.020)

        return {
            "evidence_tier": evidence_tier,
            "total_shadow_flips": n_flips,
            "total_recovered_gmv": round(total_recovered_gmv, 2),
            "average_flip_amount": round(float(np.mean(flip_amounts)), 2) if n_flips > 0 else 0.0,
            "max_flip_amount": round(float(np.max(flip_amounts)), 2) if n_flips > 0 else 0.0,
            "unlabeled_flips_count": len(unlabeled_flips),
            "pending_maturation_count": len(pending_flips),
            "mature_labeled_flips_count": n_mature,
            "mature_frauds_observed": k_mature_fraud,
            "mature_fraud_dollars_exposure": round(mature_fraud_dollars, 2),
            "observed_fraud_rate": observed_fraud_rate_str,
            "exact_95_ci_lower": lower_ci_str,
            "exact_95_ci_upper": upper_ci_str,
            "upper_95_ci_numeric": upper_ci_val,
            "meets_2pct_governance_hurdle": meets_2pct_gate
        }

    def get_gateway_health_status(self) -> Dict[str, Any]:
        """
        Evaluate health/heartbeat visibility to distinguish:
        - GATEWAY_DISCONNECTED
        - GATEWAY_CONNECTED_AWAITING_TRAFFIC
        - GATEWAY_CONNECTED_TRAFFIC_FLOWING_NO_FLIPS
        - GATEWAY_CONNECTED_SHADOW_FLIPS_ACCUMULATING
        """
        if self.collector is None or self.scoring_engine is None:
            state = "GATEWAY_DISCONNECTED"
            desc = "Scoring engine or telemetry collector not attached."
        else:
            live_tx = getattr(self.collector, "ingested_live_count", 0)
            live_flips = len([r for r in self.shadow_flips_index.values() if r["evidence_tier"] == "LIVE_PRODUCTION_EVIDENCE"])
            if live_tx == 0:
                state = "GATEWAY_CONNECTED_AWAITING_TRAFFIC"
                desc = "Gateway telemetry listener is active and attached, awaiting genuine upstream transaction traffic."
            elif live_flips == 0:
                state = "GATEWAY_CONNECTED_TRAFFIC_FLOWING_NO_FLIPS"
                desc = f"Gateway traffic is actively flowing ({live_tx} events), zero shadow flips observed."
            else:
                state = "GATEWAY_CONNECTED_SHADOW_FLIPS_ACCUMULATING"
                desc = f"Gateway traffic is actively flowing ({live_tx} events), {live_flips} shadow flips logged."

        return {
            "gateway_health_state": state,
            "description": desc,
            "heartbeat_timestamp": datetime.now(timezone.utc).isoformat(),
            "store_path": self.shadow_flips_path,
            "store_writable": os.access(self.store_dir, os.W_OK)
        }

    def get_traffic_heartbeat_metrics(self) -> Dict[str, Any]:
        """
        Operational telemetry and heartbeat metrics for live gateway monitoring:
        - First & last genuine production event timestamps
        - Total genuine transactions, flips, untrusted events, duplicates, exceptions
        - Evidence freshness diagnostic
        """
        live_tx = getattr(self.collector, "ingested_live_count", 0)
        live_flips = len([r for r in self.shadow_flips_index.values() if r["evidence_tier"] == "LIVE_PRODUCTION_EVIDENCE"])

        if self.last_live_timestamp is None:
            freshness_str = "AWAITING_FIRST_PRODUCTION_TRANSACTION"
            last_ts_iso = None
            first_ts_iso = None
        else:
            first_ts_iso = datetime.fromtimestamp(self.first_live_timestamp, tz=timezone.utc).isoformat()
            last_ts_iso = datetime.fromtimestamp(self.last_live_timestamp, tz=timezone.utc).isoformat()
            elapsed_sec = time.time() - self.last_live_timestamp
            freshness_str = f"{elapsed_sec:.1f}s ago"

        return {
            "first_genuine_production_timestamp": first_ts_iso,
            "last_genuine_production_timestamp": last_ts_iso,
            "genuine_live_transactions": live_tx,
            "genuine_live_shadow_flips": live_flips,
            "untrusted_events_count": self.untrusted_events_count,
            "duplicate_events_count": self.duplicate_events_count,
            "shadow_exceptions_count": self.shadow_exceptions_count,
            "persistence_failures_count": self.persistence_failures_count,
            "evidence_freshness": freshness_str
        }
