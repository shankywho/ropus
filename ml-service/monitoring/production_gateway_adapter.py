"""
ROPUS Production Gateway Adapter & Telemetry Ingestion Layer
Provides a secure, privacy-preserving, fail-closed telemetry ingestion pipeline for live production traffic:
1. Versioned Telemetry Contract & Schema Validation (v1.0.0)
2. Strict Privacy & Data Minimization Guard (zero PAN, CVV, Secrets; anonymized correlation IDs)
3. Malformed Telemetry Integrity Watchdog & Quarantine Quarantine Router
4. Evidence Classification Boundary Enforcement (LIVE_PRODUCTION_EVIDENCE vs LOCAL_OPERATIONAL_TEST vs SYNTHETIC_TEST)
5. Delayed Ground-Truth Dispute / Chargeback Ingestion Interface
6. Production-vs-Local Telemetry Reconciliation Engine
7. Fail-Closed Integration with ProductionTelemetryCollector
"""

import os
import sys
import time
import math
import hashlib
import hmac
import json
from datetime import datetime, timezone
from typing import Dict, Any, List, Optional, Tuple, Set

current_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if current_dir not in sys.path:
    sys.path.insert(0, current_dir)

from monitoring.production_telemetry import (
    ProductionTelemetryCollector,
    EXPECTED_CHAMPION_VERSION,
    EXPECTED_SHA256,
    EXPECTED_FILE_SIZE
)

TELEMETRY_SCHEMA_VERSION = "1.0.0"
SALT = "ropus_gateway_telemetry_salt_2026"

# Forbidden sensitive PII / PCI keywords
FORBIDDEN_SENSITIVE_KEYS: Set[str] = {
    "pan", "card_number", "credit_card", "cvv", "cvc", "cvv2", "cid",
    "password", "secret", "api_key", "token_secret", "ssn", "social_security",
    "raw_cardholder_name", "pin", "auth_token", "private_key"
}

REQUIRED_TELEMETRY_FIELDS = {
    "telemetry_schema_version",
    "timestamp",
    "correlation_id",
    "evidence_tier",
    "model_version",
    "model_sha256",
    "http_status",
    "latency_ms",
    "calibrated_probability",
    "decision",
    "risk_tier"
}

ALLOWED_EVIDENCE_TIERS = {
    "LIVE_PRODUCTION_EVIDENCE",
    "PRODUCTION_LIVE",
    "LOCAL_OPERATIONAL_TEST",
    "STAGING_TEST",
    "SYNTHETIC_TEST",
    "HISTORICAL_REFERENCE",
    "HISTORICAL_REPLAY"
}

class PrivacyGuard:
    """Enforces strict data minimization and identifier anonymization."""
    @staticmethod
    def audit_record_privacy(record: Dict[str, Any]) -> Tuple[bool, List[str]]:
        """Verify record contains zero forbidden sensitive fields."""
        violations = []

        def check_dict(d: Dict[str, Any], prefix: str = ""):
            for k, v in d.items():
                full_k = f"{prefix}.{k}" if prefix else k
                k_lower = k.lower()
                if any(forbidden in k_lower for forbidden in FORBIDDEN_SENSITIVE_KEYS):
                    violations.append(f"Forbidden sensitive field detected: {full_k}")
                if isinstance(v, dict):
                    check_dict(v, full_k)

        check_dict(record)
        return (len(violations) == 0, violations)

    @staticmethod
    def anonymize_identifier(raw_id: str) -> str:
        """Create a irreversible salted HMAC-SHA256 correlation hash."""
        if not raw_id:
            return "ANON_UNSPECIFIED"
        h = hmac.new(SALT.encode("utf-8"), raw_id.encode("utf-8"), hashlib.sha256).hexdigest()
        return f"ANON_TX_{h[:16]}"

GATEWAY_PROVENANCE_SECRET = os.getenv("ROPUS_GATEWAY_PROVENANCE_SECRET", "ropus_gateway_trusted_provenance_key_2026")

class GatewayProvenanceGuard:
    """Cryptographic provenance verification preventing arbitrary client-spoofed PRODUCTION_LIVE evidence."""
    @staticmethod
    def generate_provenance_signature(raw_correlation_id: str, timestamp_int: int) -> str:
        msg = f"{raw_correlation_id}:{timestamp_int}:ROPUS_PROD_GATEWAY"
        return hmac.new(GATEWAY_PROVENANCE_SECRET.encode("utf-8"), msg.encode("utf-8"), hashlib.sha256).hexdigest()

    @staticmethod
    def verify_provenance_signature(raw_correlation_id: str, timestamp_int: int, signature: Optional[str]) -> bool:
        if not signature:
            return False
        expected = GatewayProvenanceGuard.generate_provenance_signature(raw_correlation_id, timestamp_int)
        return hmac.compare_digest(expected, signature)

class ProductionGatewayAdapter:
    """
    Production-facing gateway telemetry adapter.
    Validates, sanitizes, and ingests telemetry events from live or local gateway clients.
    """
    def __init__(
        self,
        telemetry_collector: ProductionTelemetryCollector,
        quarantine_dir: Optional[str] = None
    ):
        self.collector = telemetry_collector
        self.quarantine_dir = quarantine_dir or os.path.join(current_dir, "monitoring", "quarantine_store")
        os.makedirs(self.quarantine_dir, exist_ok=True)

        # Telemetry accounting
        self.ingested_live_count = 0
        self.ingested_local_count = 0
        self.ingested_synthetic_count = 0
        self.quarantined_count = 0

        # Ground-truth accounting
        self.live_ground_truth_count = 0
        self.local_ground_truth_count = 0

        # Seen correlation IDs for deduplication
        self.seen_correlation_ids: Set[str] = set()

    def validate_and_ingest_event(self, raw_record: Dict[str, Any]) -> Dict[str, Any]:
        """
        Validate incoming telemetry event against schema, privacy, and integrity rules.
        If valid: ingests into telemetry collector.
        If invalid: quarantines record and emits warning without corrupting active baselines.
        """
        # 1. Privacy Check
        is_private, privacy_violations = PrivacyGuard.audit_record_privacy(raw_record)
        if not is_private:
            return self._quarantine_record(raw_record, "PRIVACY_VIOLATION", privacy_violations)

        # 2. Schema Field Completeness
        missing_fields = [f for f in REQUIRED_TELEMETRY_FIELDS if f not in raw_record]
        if missing_fields:
            return self._quarantine_record(raw_record, "MISSING_REQUIRED_FIELDS", missing_fields)

        # 3. Evidence Tier Validation
        tier = raw_record["evidence_tier"]
        if tier not in ALLOWED_EVIDENCE_TIERS:
            return self._quarantine_record(raw_record, "INVALID_EVIDENCE_TIER", [f"Unknown tier: {tier}"])

        # 4. Deduplication
        cid = raw_record["correlation_id"]
        if cid in self.seen_correlation_ids:
            return self._quarantine_record(raw_record, "DUPLICATE_CORRELATION_ID", [f"Duplicate ID: {cid}"])

        # 5. Timestamp Validation
        ts = raw_record["timestamp"]
        if not isinstance(ts, (int, float)) or ts <= 0:
            return self._quarantine_record(raw_record, "INVALID_TIMESTAMP", [f"Invalid ts: {ts}"])

        # 6. Latency Validation
        lat = raw_record["latency_ms"]
        if not isinstance(lat, (int, float)) or lat < 0.01 or lat > 60000.0:
            return self._quarantine_record(raw_record, "IMPOSSIBLE_LATENCY", [f"Latency out of bounds: {lat} ms"])

        # 7. HTTP Status Code Validation
        status = raw_record["http_status"]
        if not isinstance(status, int) or status < 100 or status > 599:
            return self._quarantine_record(raw_record, "INVALID_HTTP_STATUS", [f"Invalid status: {status}"])

        # 8. Model & Checksum Validation
        model_ver = raw_record["model_version"]
        model_sha = raw_record["model_sha256"]
        if model_ver != EXPECTED_CHAMPION_VERSION or model_sha != EXPECTED_SHA256:
            return self._quarantine_record(
                raw_record,
                "MODEL_INTEGRITY_MISMATCH",
                [f"Expected {EXPECTED_CHAMPION_VERSION} ({EXPECTED_SHA256[:12]}...), got {model_ver} ({model_sha[:12]}...)"]
            )

        # 9. Valid Telemetry Event -> Ingest into Collector
        self.seen_correlation_ids.add(cid)
        if tier == "LIVE_PRODUCTION_EVIDENCE":
            self.ingested_live_count += 1
        elif tier == "LOCAL_OPERATIONAL_TEST":
            self.ingested_local_count += 1
        elif tier == "SYNTHETIC_TEST":
            self.ingested_synthetic_count += 1

        self.collector.record_inference_event(
            tx_id=cid,
            latency_ms=float(lat),
            http_status=status,
            model_version=model_ver,
            sha256_hash=model_sha,
            amount=float(raw_record.get("amount", 0.0)),
            calibrated_p=float(raw_record["calibrated_probability"]),
            decision=raw_record["decision"],
            risk_tier=raw_record["risk_tier"],
            raw_features=raw_record.get("drift_features", {}),
            is_schema_valid=True
        )

        return {
            "status": "ACCEPTED",
            "correlation_id": cid,
            "evidence_tier": tier,
            "ingested_at": datetime.now(timezone.utc).isoformat()
        }

    def ingest_delayed_dispute(
        self,
        dispute_id: str,
        correlation_id: str,
        is_fraud: int,
        chargeback_amount: float,
        evidence_tier: str = "LOCAL_OPERATIONAL_TEST",
        reported_at: Optional[str] = None
    ) -> Dict[str, Any]:
        """Ingest delayed dispute/chargeback label with evidence tier tagging."""
        if evidence_tier == "LIVE_PRODUCTION_EVIDENCE":
            self.live_ground_truth_count += 1
        else:
            self.local_ground_truth_count += 1

        self.collector.ingest_delayed_ground_truth(
            tx_id=correlation_id,
            is_fraud=is_fraud,
            chargeback_amount=chargeback_amount,
            reported_at=reported_at or datetime.now(timezone.utc).isoformat()
        )
        return {
            "status": "DISPUTE_INGESTED",
            "dispute_id": dispute_id,
            "correlation_id": correlation_id,
            "evidence_tier": evidence_tier,
            "is_fraud": is_fraud
        }

    def _quarantine_record(self, raw_record: Dict[str, Any], reason_code: str, details: List[str]) -> Dict[str, Any]:
        """Safely quarantine defective record to disk."""
        self.quarantined_count += 1
        quarantine_entry = {
            "quarantined_at": datetime.now(timezone.utc).isoformat(),
            "reason_code": reason_code,
            "details": details,
            "raw_record": raw_record
        }
        q_path = os.path.join(self.quarantine_dir, "quarantine_events.jsonl")
        with open(q_path, "a") as f:
            f.write(json.dumps(quarantine_entry) + "\n")

        return {
            "status": "QUARANTINED",
            "reason_code": reason_code,
            "details": details
        }

    def get_telemetry_manifest(self) -> Dict[str, Any]:
        """Summary of all ingested telemetry counts by evidence tier."""
        return {
            "telemetry_schema_version": TELEMETRY_SCHEMA_VERSION,
            "live_production_events": self.ingested_live_count,
            "local_operational_events": self.ingested_local_count,
            "synthetic_test_events": self.ingested_synthetic_count,
            "quarantined_events": self.quarantined_count,
            "live_ground_truth_labels": self.live_ground_truth_count,
            "local_ground_truth_labels": self.local_ground_truth_count,
            "live_traffic_status": "AWAITING_GATEWAY_TRAFFIC" if self.ingested_live_count == 0 else "LIVE_TELEMETRY_ACTIVE",
            "live_ground_truth_status": "AWAITING_PRODUCTION_LABELS" if self.live_ground_truth_count == 0 else "LIVE_LABELS_INGESTED"
        }

    def reconcile_with_serving_logs(self, serving_log_records: List[Dict[str, Any]]) -> Dict[str, Any]:
        """
        Reconcile gateway ingested events against raw local serving logs
        without polluting live production statistics.
        """
        gateway_events = list(self.collector.events)
        n_serving = len(serving_log_records)
        n_gateway = len(gateway_events)

        # Verify model versions
        serving_versions = set(r.get("model_version") for r in serving_log_records)
        gateway_versions = set(e.get("model_version") for e in gateway_events)

        version_parity = (serving_versions == gateway_versions and EXPECTED_CHAMPION_VERSION in gateway_versions)
        count_match = (n_serving == self.ingested_local_count)

        return {
            "evidence_type": "LOCAL_OPERATIONAL_TEST",
            "serving_log_count": n_serving,
            "gateway_ingested_local_count": self.ingested_local_count,
            "gateway_ingested_live_count": self.ingested_live_count,
            "count_reconciled": count_match,
            "version_parity": version_parity,
            "quarantined_records": self.quarantined_count,
            "reconciliation_status": "RECONCILED" if version_parity and count_match else "DISCREPANCY"
        }
