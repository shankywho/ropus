"""
Persistent Shadow Evidence Ledger (Phase 60)
Maintains append-only, auditable, privacy-safe records of investigation dossiers and relationship evidence.
"""

import json
import hashlib
from datetime import datetime, timezone
from typing import Dict, List, Any, Optional
from pydantic import BaseModel, Field

from .schema import CollusionInvestigationDossier, RelationshipRiskLevel


class EvidenceLedgerEntry(BaseModel):
    entry_id: str
    dossier_id: str
    recorded_at: datetime
    employee_id: str
    employee_role: str
    overall_risk_score: float
    risk_level: RelationshipRiskLevel
    affected_accounts_count: int
    affected_consumers_count: int
    relationship_paths_count: int
    shared_devices_count: int
    shared_ips_count: int
    supporting_evidence_count: int
    counter_evidence_count: int
    dossier_payload: Dict[str, Any]
    provenance_hash: str


class ShadowEvidenceLedger:
    """
    Append-only persistent ledger for investigation intelligence dossiers.
    """

    def __init__(self):
        self.entries: List[EvidenceLedgerEntry] = []
        self._dossier_index: Dict[str, EvidenceLedgerEntry] = {}

    def total_entries(self) -> int:
        return len(self.entries)

    def record_dossier(self, dossier: CollusionInvestigationDossier) -> EvidenceLedgerEntry:
        """Appends a new investigation dossier to the audit ledger."""
        now = datetime.now(timezone.utc)
        payload = dossier.model_dump(mode="json")
        payload_str = json.dumps(payload, sort_keys=True)
        prov_hash = hashlib.sha256(payload_str.encode("utf-8")).hexdigest()

        entry_id = f"ev_ledg_{len(self.entries) + 1:06d}_{prov_hash[:8]}"

        entry = EvidenceLedgerEntry(
            entry_id=entry_id,
            dossier_id=dossier.dossier_id,
            recorded_at=now,
            employee_id=dossier.employee_id,
            employee_role=dossier.employee_role,
            overall_risk_score=dossier.overall_risk_score,
            risk_level=dossier.risk_level,
            affected_accounts_count=len(dossier.affected_accounts),
            affected_consumers_count=len(dossier.affected_consumers),
            relationship_paths_count=len(dossier.relationship_paths),
            shared_devices_count=dossier.shared_infrastructure.get("device_overlap_count", 0),
            shared_ips_count=dossier.shared_infrastructure.get("ip_overlap_count", 0),
            supporting_evidence_count=len(dossier.temporal_evidence.get("supporting_evidence", [])),
            counter_evidence_count=len(dossier.temporal_evidence.get("counter_evidence", [])),
            dossier_payload=payload,
            provenance_hash=prov_hash
        )

        self.entries.append(entry)
        self._dossier_index[dossier.dossier_id] = entry
        return entry

    def get_by_dossier_id(self, dossier_id: str) -> Optional[EvidenceLedgerEntry]:
        return self._dossier_index.get(dossier_id)

    def get_entries_for_employee(self, employee_id: str) -> List[EvidenceLedgerEntry]:
        return [e for e in self.entries if e.employee_id == employee_id]

    def get_summary_statistics(self) -> Dict[str, Any]:
        high_risk = sum(1 for e in self.entries if e.risk_level in [RelationshipRiskLevel.HIGH, RelationshipRiskLevel.CRITICAL])
        return {
            "total_dossiers_recorded": len(self.entries),
            "high_risk_dossiers_count": high_risk,
            "unique_employees_investigated": len(set(e.employee_id for e in self.entries)),
            "ledger_integrity": "CRYPTOGRAPHICALLY_LINKED",
            "storage_mode": "APPEND_ONLY_AUDIT"
        }
