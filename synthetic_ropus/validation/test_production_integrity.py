"""
Automated Pytest Suite: Production Integrity & Non-Interference
"""
import os
import hashlib
import pytest

EXPECTED_PRODUCTION_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"
REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))

def test_production_champion_checksum_unchanged():
    model_path = os.path.join(REPO_ROOT, "ml-service", "model", "candidates", "production_model_v8_bmr.joblib")
    assert os.path.exists(model_path), f"Production champion model missing at {model_path}"

    hasher = hashlib.sha256()
    with open(model_path, "rb") as f:
        while chunk := f.read(8192):
            hasher.update(chunk)

    computed_sha256 = hasher.hexdigest()
    assert computed_sha256 == EXPECTED_PRODUCTION_SHA256, (
        f"CRITICAL: Production model checksum mismatch!\n"
        f"Expected: {EXPECTED_PRODUCTION_SHA256}\n"
        f"Actual:   {computed_sha256}"
    )
