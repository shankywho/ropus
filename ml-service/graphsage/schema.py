"""
GraphSAGE Heterogeneous Graph Schema, Real-Data Contract & Governance Types (Phase 58)
"""

from enum import Enum
from typing import Dict, List, Optional, Any, Set
from pydantic import BaseModel, Field
from datetime import datetime


class HeteroNodeType(str, Enum):
    EMPLOYEE = "EMPLOYEE"
    CONSUMER = "CONSUMER"
    ACCOUNT = "ACCOUNT"
    DEVICE = "DEVICE"
    IP = "IP"
    PAYMENT_TOKEN = "PAYMENT_TOKEN"
    TRANSACTION = "TRANSACTION"
    MERCHANT = "MERCHANT"
    CASE = "CASE"
    SESSION = "SESSION"


class HeteroEdgeType(str, Enum):
    ACCESSES = "ACCESSES"                               # Employee -> Consumer
    MANAGES = "MANAGES"                                 # Employee -> Account
    PERFORMS = "PERFORMS"                               # Employee -> Transaction / Action
    INTERACTS_WITH_ACCOUNT = "INTERACTS_WITH_ACCOUNT"   # Employee -> Account
    INTERACTS_WITH_CONSUMER = "INTERACTS_WITH_CONSUMER" # Employee -> Consumer
    OWNS = "OWNS"                                       # Consumer -> Account
    USES_DEVICE = "USES_DEVICE"                         # Account -> Device / Consumer -> Device
    USES_IP = "USES_IP"                                 # Device -> IP / Account -> IP
    USES_PAYMENT_TOKEN = "USES_PAYMENT_TOKEN"           # Account -> PaymentToken
    CREATES = "CREATES"                                 # Consumer / Account -> Transaction
    TARGETS = "TARGETS"                                 # Transaction -> Merchant
    SHARES_DEVICE = "SHARES_DEVICE"                     # Account <-> Account
    SHARES_IP = "SHARES_IP"                             # Account <-> Account
    EMPLOYEE_ACCESSES_ACCOUNT = "EMPLOYEE_ACCESSES_ACCOUNT"
    EMPLOYEE_REVIEWS_CASE = "EMPLOYEE_REVIEWS_CASE"
    EMPLOYEE_MODIFIES_TRANSACTION = "EMPLOYEE_MODIFIES_TRANSACTION"
    EMPLOYEE_APPROVES_TRANSACTION = "EMPLOYEE_APPROVES_TRANSACTION"
    EMPLOYEE_USES_DEVICE = "EMPLOYEE_USES_DEVICE"
    EMPLOYEE_USES_IP = "EMPLOYEE_USES_IP"
    CONSUMER_OWNS_ACCOUNT = "CONSUMER_OWNS_ACCOUNT"
    ACCOUNT_USES_DEVICE = "ACCOUNT_USES_DEVICE"
    ACCOUNT_USES_IP = "ACCOUNT_USES_IP"
    ACCOUNT_USES_PAYMENT_TOKEN = "ACCOUNT_USES_PAYMENT_TOKEN"
    ACCOUNT_TRANSACTS_WITH_MERCHANT = "ACCOUNT_TRANSACTS_WITH_MERCHANT"
    CASE_REFERENCES_ACCOUNT = "CASE_REFERENCES_ACCOUNT"
    CASE_REFERENCES_CONSUMER = "CASE_REFERENCES_CONSUMER"
    CASE_REFERENCES_EMPLOYEE = "CASE_REFERENCES_EMPLOYEE"


class RelationshipRiskLevel(str, Enum):
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"
    CRITICAL = "CRITICAL"


class GraphSAGELifecycleState(str, Enum):
    DESIGNED = "DESIGNED"
    DATA_REQUIRED = "DATA_REQUIRED"
    SYNTHETIC_TRAINED = "SYNTHETIC_TRAINED"
    SYNTHETIC_VALIDATED = "SYNTHETIC_VALIDATED"
    REAL_DATA_REQUIRED = "REAL_DATA_REQUIRED"
    SHADOW_EVALUATION = "SHADOW_EVALUATION"
    HUMAN_REVIEW = "HUMAN_REVIEW"
    PROMOTION_ELIGIBLE = "PROMOTION_ELIGIBLE"
    # Legacy alias support
    TRAINED_OFFLINE = "TRAINED_OFFLINE"
    OFFLINE_VALIDATED = "OFFLINE_VALIDATED"
    SHADOW = "SHADOW"
    SHADOW_VALIDATED = "SHADOW_VALIDATED"
    GOVERNANCE_REVIEW = "GOVERNANCE_REVIEW"


class LabelMaturityState(str, Enum):
    """
    Formal label maturity lifecycle state machine.
    Ensures speculative flags are NEVER treated as ground truth.
    """
    UNLABELED = "UNLABELED"                         # Recent transaction within chargeback/dispute window (0-90 days)
    SUSPECTED = "SUSPECTED"                         # Flagged by heuristics/rules/models, unconfirmed
    UNDER_INVESTIGATION = "UNDER_INVESTIGATION"     # Active case open by human fraud analyst
    CONFIRMED_FRAUD = "CONFIRMED_FRAUD"             # Legal chargeback, customer affidavit, or confirmed internal fraud
    CONFIRMED_LEGITIMATE = "CONFIRMED_LEGITIMATE"   # Mature transaction past dispute window (>90 days) with no chargeback
    REJECTED = "REJECTED"                           # False dispute / chargeback won / dismissed investigation


class DatasetType(str, Enum):
    """
    Strict dataset taxonomy to prevent synthetic/real metric confusion.
    """
    SYNTHETIC = "SYNTHETIC"
    REAL_SHADOW = "REAL_SHADOW"
    REAL_VALIDATION = "REAL_VALIDATION"


class HeteroNode(BaseModel):
    id: str
    type: HeteroNodeType
    features: List[float] = Field(default_factory=list)
    risk_score: float = 0.05
    is_known_bad: bool = False
    created_at: datetime
    properties: Dict[str, Any] = Field(default_factory=dict)


class HeteroEdge(BaseModel):
    id: str
    source_id: str
    target_id: str
    type: HeteroEdgeType
    weight: float = 1.0
    confidence: float = 1.0
    timestamp: datetime
    provenance: str = "audit_stream"
    properties: Dict[str, Any] = Field(default_factory=dict)


class HeteroSubGraph(BaseModel):
    center_node_id: str
    as_of: datetime
    nodes: Dict[str, HeteroNode]
    edges: List[HeteroEdge]


class RealEntityIdentifierContract(BaseModel):
    """Specification of tokenization and hashing rules for production entities."""
    entity_type: HeteroNodeType
    identifier_format: str
    tokenization_method: str
    allowed_attributes: List[str]
    prohibited_attributes: List[str]
    source_system: str
    retention_days: int


class RealGraphDataContract(BaseModel):
    """
    Formal REAL-DATA GraphSAGE Ingestion & Governance Contract
    """
    version: str = "2.0.0"
    target_architecture: str = "ROPUS Multi-Defense Relationship Intelligence"
    governance_status: GraphSAGELifecycleState = GraphSAGELifecycleState.REAL_DATA_REQUIRED
    allowed_dataset_types: List[DatasetType] = [DatasetType.REAL_SHADOW, DatasetType.REAL_VALIDATION]
    minimum_genuine_internal_fraud_cases_for_promotion: int = 50
    minimum_genuine_external_fraud_cases_for_promotion: int = 50
    temporal_guarantee: str = "Strict Point-in-Time: edge_timestamp <= evaluation_time T; label_timestamp <= evaluation_time T (mature only)"
    privacy_contract: str = "Zero Raw PAN, CVV, PIN, Passwords, Cleartext PII (SSN, National ID, Plaintext Email/Phone). Salted SHA-256 / Vault Tokens Only."
    prohibited_feature_fields: List[str] = [
        "scenario_type", "is_adversarial", "is_hard_negative",
        "raw_pan", "cvv", "pin", "ssn", "password_hash",
        "label_source", "label_confidence", "label_timestamp"
    ]
    entity_specifications: Dict[str, Dict[str, Any]] = {
        "EMPLOYEE": {
            "id_format": "emp_<salted_sha256_hex8>",
            "allowed_attributes": ["role", "department", "employment_status", "hire_date"],
            "source": "idp_okta_audit_log"
        },
        "CONSUMER": {
            "id_format": "usr_<tokenized_uuid>",
            "allowed_attributes": ["kyc_status", "country_code", "created_at"],
            "source": "core_banking_customer_service"
        },
        "ACCOUNT": {
            "id_format": "acct_<tokenized_uuid>",
            "allowed_attributes": ["consumer_id", "status", "account_type", "currency"],
            "source": "core_ledger"
        },
        "DEVICE": {
            "id_format": "dev_<hardware_fingerprint_sha256>",
            "allowed_attributes": ["device_type", "os_family", "first_seen_at"],
            "source": "sdk_telemetry"
        },
        "IP": {
            "id_format": "ip_<anonymized_subnet_hash>",
            "allowed_attributes": ["ip_type", "asn", "country_code"],
            "source": "api_gateway_waf"
        },
        "PAYMENT_TOKEN": {
            "id_format": "tok_<pci_vault_token>",
            "allowed_attributes": ["token_type", "created_at", "last_four_only"],
            "source": "pci_tokenization_vault"
        },
        "TRANSACTION": {
            "id_format": "txn_<uuid>",
            "allowed_attributes": ["account_id", "device_id", "ip_id", "payment_token_id", "merchant_id", "amount", "currency", "event_timestamp"],
            "source": "transaction_processing_engine"
        },
        "MERCHANT": {
            "id_format": "mch_<tokenized_id>",
            "allowed_attributes": ["category", "mcc_code", "country_code"],
            "source": "merchant_directory"
        },
        "CASE": {
            "id_format": "case_<uuid>",
            "allowed_attributes": ["account_id", "consumer_id", "assigned_employee_id", "opened_at", "closed_at", "status", "disposition"],
            "source": "case_management_system"
        }
    }


class GraphSAGEDataContract(BaseModel):
    """Legacy compatibility alias."""
    version: str = "1.1.0"
    target_relationship: str = "Employee ↔ Consumer Collusion & Synthetic Mule Networks"
    required_entity_types: List[str] = [
        "EMPLOYEE", "CONSUMER", "ACCOUNT", "DEVICE", "IP", "PAYMENT_TOKEN", "TRANSACTION", "MERCHANT", "CASE", "SESSION"
    ]
    required_edge_types: List[str] = [
        "ACCESSES", "MANAGES", "PERFORMS", "OWNS", "USES_DEVICE", "USES_IP", "USES_PAYMENT_TOKEN",
        "CREATES", "TARGETS", "INTERACTS_WITH_ACCOUNT", "INTERACTS_WITH_CONSUMER"
    ]
    temporal_guarantee: str = "Strict Point-in-Time (Edges where timestamp <= evaluation_time T)"
    privacy_contract: str = "Zero Raw PAN, CVV, Passwords or Direct Customer PII. Opaque Hashed Tokens Only."
    lifecycle_status: GraphSAGELifecycleState = GraphSAGELifecycleState.REAL_DATA_REQUIRED
    minimum_genuine_internal_fraud_cases_for_promotion: int = 50


class ShadowInvestigationOutput(BaseModel):
    """
    Machine-readable relationship intelligence output.
    Explicitly marked NON_ENFORCING / INVESTIGATION_ONLY.
    """
    root_node_id: str
    root_node_type: HeteroNodeType
    graph_risk_score: float
    collusion_risk_score: float
    employee_risk_score: float
    consumer_risk_score: float
    neighborhood_anomaly_score: float
    graph_poisoning_score: float
    embedding: List[float]
    top_related_entities: List[Dict[str, Any]]
    suspicious_paths: List[str]
    relationship_evidence: List[str]
    point_in_time_timestamp: datetime
    model_version: str = "graphsage-hetero-v1.0-shadow"
    data_version: str = "real-shadow-v1.0"
    operational_mode: str = "NON_ENFORCING"
    intended_use: str = "INVESTIGATION_ONLY"


class CollusionInvestigationDossier(BaseModel):
    """
    Comprehensive, privacy-preserving collusion dossier for fraud investigators.
    """
    dossier_id: str
    generated_at: datetime
    overall_risk_score: float
    risk_level: RelationshipRiskLevel
    employee_id: str
    employee_role: str
    affected_accounts: List[str]
    affected_consumers: List[str]
    relationship_paths: List[str]
    temporal_evidence: Dict[str, Any]
    transaction_summary: Dict[str, Any]
    shared_infrastructure: Dict[str, Any]
    confidence_score: float
    human_explanation: str
    data_provenance: str = "shadow_graph_stream"
    model_version: str = "graphsage-v1.0-shadow"
    governance_notice: str = "INVESTIGATION INTELLIGENCE ONLY - STRICTLY NON-ENFORCING"


class ReadinessAuditReport(BaseModel):
    """
    Structured scorecard produced by the Real-Data Readiness Checker.
    """
    audited_at: datetime
    dataset_type: DatasetType
    is_ready_for_real_validation: bool
    governance_state: GraphSAGELifecycleState
    blocking_reasons: List[str]
    schema_completeness: Dict[str, Any]
    temporal_causality: Dict[str, Any]
    privacy_and_security: Dict[str, Any]
    topology_statistics: Dict[str, Any]
    entity_coverage: Dict[str, Any]
    label_maturity_coverage: Dict[str, Any]
    confirmed_fraud_labels_count: int
    confirmed_internal_collusion_labels_count: int
