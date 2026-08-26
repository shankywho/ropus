"""
Real Label Maturation Engine & Dispute Lifecycle Tracker (Phase 60)
Tracks maturation timelines, dispute windows, and progress toward the >= 50 real internal collusion threshold.
"""

from datetime import datetime, timezone, timedelta
from typing import Dict, List, Any, Optional
from pydantic import BaseModel, Field
from .schema import LabelMaturityState


class MaturityRecord(BaseModel):
    record_id: str
    entity_id: str
    event_timestamp: datetime
    state: LabelMaturityState = LabelMaturityState.UNLABELED
    investigation_opened_at: Optional[datetime] = None
    investigation_closed_at: Optional[datetime] = None
    label_timestamp: Optional[datetime] = None
    label_source: str = "PENDING"
    label_confidence: float = 0.0
    case_id: Optional[str] = None
    is_collusion: bool = False
    notes: Optional[str] = None

    @property
    def maturation_duration_days(self) -> float:
        if self.label_timestamp and self.event_timestamp:
            return round((self.label_timestamp - self.event_timestamp).total_seconds() / 86400.0, 2)
        return 0.0


class MaturityReport(BaseModel):
    generated_at: datetime
    total_observations: int
    mature_labels_count: int
    immature_labels_count: int
    confirmed_fraud_count: int
    confirmed_legitimate_count: int
    confirmed_internal_collusion_count: int
    rejected_count: int
    target_collusion_threshold: int = 50
    progress_to_collusion_gate_pct: float
    average_maturation_days: float
    dispute_maturation_window_days: int = 90
    governance_status: str


class LabelMaturationEngine:
    """
    Manages point-in-time label transitions and tracks dispute maturation duration.
    """

    def __init__(self, dispute_window_days: int = 90):
        self.dispute_window_days = dispute_window_days
        self.records: Dict[str, MaturityRecord] = {}

    def register_transaction(self, entity_id: str, event_timestamp: datetime) -> MaturityRecord:
        """Registers a fresh transaction entering the unadjudicated 90-day dispute window."""
        rec = MaturityRecord(
            record_id=f"mat_{entity_id}",
            entity_id=entity_id,
            event_timestamp=event_timestamp,
            state=LabelMaturityState.UNLABELED
        )
        self.records[entity_id] = rec
        return rec

    def flag_suspected(self, entity_id: str, flagged_at: datetime, reason: str) -> Optional[MaturityRecord]:
        """Flags transaction as suspected anomaly (NON ground-truth)."""
        if entity_id in self.records:
            rec = self.records[entity_id]
            if rec.state == LabelMaturityState.UNLABELED:
                rec.state = LabelMaturityState.SUSPECTED
                rec.notes = f"Suspected flag: {reason} at {flagged_at.isoformat()}"
            return rec
        return None

    def open_investigation(self, entity_id: str, case_id: str, opened_at: datetime) -> Optional[MaturityRecord]:
        """Transitions case to UNDER_INVESTIGATION."""
        if entity_id in self.records:
            rec = self.records[entity_id]
            rec.state = LabelMaturityState.UNDER_INVESTIGATION
            rec.case_id = case_id
            rec.investigation_opened_at = opened_at
            return rec
        return None

    def confirm_label(
        self,
        entity_id: str,
        state: LabelMaturityState,
        source: str,
        confirmed_at: datetime,
        is_collusion: bool = False,
        confidence: float = 1.0,
        case_id: Optional[str] = None
    ) -> Optional[MaturityRecord]:
        """Formal confirmation of mature label (CONFIRMED_FRAUD, CONFIRMED_LEGITIMATE, or REJECTED)."""
        if entity_id not in self.records:
            # Cold entity registration
            self.register_transaction(entity_id, confirmed_at - timedelta(days=self.dispute_window_days))

        rec = self.records[entity_id]
        rec.state = state
        rec.label_source = source
        rec.label_timestamp = confirmed_at
        rec.investigation_closed_at = confirmed_at
        rec.is_collusion = is_collusion
        rec.label_confidence = confidence
        if case_id:
            rec.case_id = case_id
        return rec

    def auto_mature_clean_transactions(self, as_of: datetime) -> int:
        """Transitions transactions older than 90 days with zero disputes to CONFIRMED_LEGITIMATE."""
        matured_count = 0
        for rec in self.records.values():
            if rec.state == LabelMaturityState.UNLABELED:
                age_days = (as_of - rec.event_timestamp).total_seconds() / 86400.0
                if age_days >= self.dispute_window_days:
                    rec.state = LabelMaturityState.CONFIRMED_LEGITIMATE
                    rec.label_source = "DISPUTE_WINDOW_TIMEOUT"
                    rec.label_timestamp = rec.event_timestamp + timedelta(days=self.dispute_window_days)
                    rec.label_confidence = 0.99
                    matured_count += 1
        return matured_count

    def generate_report(self, as_of: Optional[datetime] = None) -> MaturityReport:
        now = as_of or datetime.now(timezone.utc)
        total = len(self.records)

        mature_fraud = sum(1 for r in self.records.values() if r.state == LabelMaturityState.CONFIRMED_FRAUD)
        mature_legit = sum(1 for r in self.records.values() if r.state == LabelMaturityState.CONFIRMED_LEGITIMATE)
        mature_collusion = sum(
            1 for r in self.records.values()
            if r.state == LabelMaturityState.CONFIRMED_FRAUD and r.is_collusion
        )
        rejected = sum(1 for r in self.records.values() if r.state == LabelMaturityState.REJECTED)

        mature_total = mature_fraud + mature_legit + rejected
        immature_total = total - mature_total

        durations = [r.maturation_duration_days for r in self.records.values() if r.label_timestamp]
        avg_dur = round(float(sum(durations) / max(len(durations), 1)), 2)

        progress_pct = round((mature_collusion / 50.0) * 100.0, 2)

        gov_status = "DATA_REQUIRED (0 / 50 Confirmed Real Collusion Cases)" if mature_collusion < 50 else "GATE_CRITERIA_MET"

        return MaturityReport(
            generated_at=now,
            total_observations=total,
            mature_labels_count=mature_total,
            immature_labels_count=immature_total,
            confirmed_fraud_count=mature_fraud,
            confirmed_legitimate_count=mature_legit,
            confirmed_internal_collusion_count=mature_collusion,
            rejected_count=rejected,
            target_collusion_threshold=50,
            progress_to_collusion_gate_pct=progress_pct,
            average_maturation_days=avg_dur,
            dispute_maturation_window_days=self.dispute_window_days,
            governance_status=gov_status
        )
