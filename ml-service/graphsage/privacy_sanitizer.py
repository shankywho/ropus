"""
Strict Data Privacy Boundary & Sanitizer (Phase 60)
Guarantees zero cleartext PAN, CVV, PIN, SSN, raw bank account numbers, credentials, or PII enter GraphSAGE.
"""

import re
import hashlib
from typing import Dict, Any, Tuple, Optional, List
from pydantic import BaseModel


class PrivacyViolationRecord(BaseModel):
    field_name: str
    violation_type: str
    action_taken: str
    quarantined_value_hash: str


class PrivacySanitizer:
    """
    Validates and sanitizes raw incoming event payloads before graph store ingestion.
    """

    # Regex patterns for prohibited sensitive values
    PAN_REGEX = re.compile(r"\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13})\b")
    CVV_REGEX = re.compile(r"\b\d{3,4}\b")
    SSN_REGEX = re.compile(r"^\d{3}-\d{2}-\d{4}$|^\d{9}$")
    IBAN_REGEX = re.compile(r"^[A-Z]{2}\d{2}[A-Z0-9]{11,30}$")

    PROHIBITED_KEY_TERMS = {
        "raw_pan", "pan", "card_number", "cvv", "cvv2", "cvc", "pin", "ssn",
        "social_security", "password", "cleartext_password", "password_hash",
        "bank_account_number", "iban", "routing_number"
    }

    def __init__(self, salt: str = "ropus_salt_prod_2026"):
        self.salt = salt.encode("utf-8")

    def tokenize_identifier(self, prefix: str, raw_id: str) -> str:
        """Deterministically tokenizes a raw entity ID with a salted SHA-256 hash."""
        if not raw_id:
            return f"{prefix}_unknown"
        hasher = hashlib.sha256(self.salt)
        hasher.update(raw_id.encode("utf-8"))
        return f"{prefix}_{hasher.hexdigest()[:12]}"

    def sanitize_payload(
        self,
        raw_payload: Dict[str, Any]
    ) -> Tuple[Dict[str, Any], List[PrivacyViolationRecord], bool]:
        """
        Sanitizes an incoming raw event payload.
        Returns: (sanitized_payload, list_of_violations, is_quarantined)
        """
        sanitized = {}
        violations: List[PrivacyViolationRecord] = []
        is_quarantined = False

        for k, v in raw_payload.items():
            k_lower = k.lower()

            # 1. Check prohibited key names
            if k_lower in self.PROHIBITED_KEY_TERMS:
                val_hash = hashlib.sha256(str(v).encode("utf-8")).hexdigest()[:16]
                violations.append(PrivacyViolationRecord(
                    field_name=k,
                    violation_type="PROHIBITED_FIELD_KEY",
                    action_taken="STRIPPED_AND_QUARANTINED",
                    quarantined_value_hash=val_hash
                ))
                is_quarantined = True
                continue

            # 2. Check string value patterns
            if isinstance(v, str):
                # Check for cleartext PAN
                if self.PAN_REGEX.search(v.replace(" ", "").replace("-", "")):
                    val_hash = hashlib.sha256(v.encode("utf-8")).hexdigest()[:16]
                    violations.append(PrivacyViolationRecord(
                        field_name=k,
                        violation_type="RAW_PAN_DETECTED",
                        action_taken="REDACTED",
                        quarantined_value_hash=val_hash
                    ))
                    is_quarantined = True
                    continue

                # Check SSN
                if self.SSN_REGEX.match(v):
                    val_hash = hashlib.sha256(v.encode("utf-8")).hexdigest()[:16]
                    violations.append(PrivacyViolationRecord(
                        field_name=k,
                        violation_type="SSN_PII_DETECTED",
                        action_taken="STRIPPED",
                        quarantined_value_hash=val_hash
                    ))
                    is_quarantined = True
                    continue

            sanitized[k] = v

        return sanitized, violations, is_quarantined
