import requests
import hmac
import hashlib
import time
from typing import Dict, Any, Optional

class RopusClient:
    """
    Ropus AI Risk Manager Official Python SDK.
    Provides sub-10ms real-time risk decisioning, case management, and explainability.
    """
    def __init__(self, api_key: str, base_url: str = "http://localhost:8080"):
        self.api_key = api_key
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()
        self.session.headers.update({
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json",
            "X-Client-SDK": "ropus-python-v3.34",
        })

    def evaluate_risk(self, transaction: Dict[str, Any]) -> Dict[str, Any]:
        """
        Evaluate real-time risk score and actionable decision for a financial transaction.
        """
        url = f"{self.base_url}/v1/risk/evaluate"
        response = self.session.post(url, json=transaction, timeout=2.0)
        response.raise_for_status()
        return response.json()

    def create_case(self, case_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Manually or programmatically open an investigation case.
        """
        url = f"{self.base_url}/v1/cases/create"
        response = self.session.post(url, json=case_data, timeout=5.0)
        response.raise_for_status()
        return response.json()

    def get_case(self, case_id: str) -> Dict[str, Any]:
        """
        Retrieve case timeline, evidentiary graph, and AI investigator report.
        """
        url = f"{self.base_url}/v1/cases/{case_id}"
        response = self.session.get(url, timeout=5.0)
        response.raise_for_status()
        return response.json()

    @staticmethod
    def verify_webhook_signature(payload: bytes, signature: str, secret: str) -> bool:
        """
        Verify incoming webhook HMAC-SHA256 signature.
        """
        expected_sig = "sha256=" + hmac.new(secret.encode(), payload, hashlib.sha256).hexdigest()
        return hmac.compare_digest(expected_sig, signature)
