package graphsage

import (
	"time"
)

// HeteroNodeType classifies discrete entity types in the internal and consumer graph.
type HeteroNodeType string

const (
	NodeEmployee      HeteroNodeType = "EMPLOYEE"
	NodeConsumer      HeteroNodeType = "CONSUMER"
	NodeAccount       HeteroNodeType = "ACCOUNT"
	NodeDevice        HeteroNodeType = "DEVICE"
	NodeIP            HeteroNodeType = "IP"
	NodePaymentMethod HeteroNodeType = "PAYMENT_METHOD"
	NodeTransaction   HeteroNodeType = "TRANSACTION"
	NodeCase          HeteroNodeType = "CASE"
	NodeMerchant      HeteroNodeType = "MERCHANT"
	NodeSession       HeteroNodeType = "SESSION"
)

// HeteroEdgeType defines directed causal relationships between entities.
type HeteroEdgeType string

const (
	// Employee Operational & Administrative Relationships
	EdgeAccessesAccount  HeteroEdgeType = "ACCESSES_ACCOUNT"  // Employee -> Account
	EdgeAccessesConsumer HeteroEdgeType = "ACCESSES_CONSUMER" // Employee -> Consumer
	EdgeModifiesTx       HeteroEdgeType = "MODIFIES"          // Employee -> Transaction
	EdgeReviewsCase      HeteroEdgeType = "REVIEWS"           // Employee -> Case
	EdgeApprovesTx       HeteroEdgeType = "APPROVES"          // Employee -> Transaction
	EdgeAccesses         HeteroEdgeType = "ACCESSES"          // Employee -> Consumer / Account (Generic)
	EdgeManages          HeteroEdgeType = "MANAGES"           // Employee -> Account

	// Employee Telemetry Linkages
	EdgeEmployeeUsesDevice HeteroEdgeType = "USES_DEVICE" // Employee -> Device
	EdgeEmployeeUsesIP     HeteroEdgeType = "USES_IP"     // Employee -> IP

	// Consumer, Account & Financial Linkages
	EdgeOwns          HeteroEdgeType = "OWNS"           // Consumer -> Account
	EdgeUsesDevice    HeteroEdgeType = "USES_DEVICE"    // Account / Consumer -> Device
	EdgeUsesIP        HeteroEdgeType = "USES_IP"        // Device / Account -> IP
	EdgeUsesPayment   HeteroEdgeType = "USES_PAYMENT"   // Account -> PaymentMethod
	EdgeCreatesTx     HeteroEdgeType = "CREATES"        // Account / Consumer -> Transaction
	EdgeTransactsWith HeteroEdgeType = "TRANSACTS_WITH" // Account -> Merchant or Account -> Account
	EdgeTargets       HeteroEdgeType = "TARGETS"        // Transaction -> Merchant
	EdgeHasSession    HeteroEdgeType = "HAS_SESSION"    // Account -> Session

	// Shared Infrastructure Linkages
	EdgeSharesDevice HeteroEdgeType = "SHARES_DEVICE" // Account <-> Account or Employee <-> Account
	EdgeSharesIP     HeteroEdgeType = "SHARES_IP"     // Account <-> Account or Employee <-> Account
)

// HeteroNode represents an entity in the temporal heterogeneous graph.
type HeteroNode struct {
	ID         string                 `json:"id"`
	Type       HeteroNodeType         `json:"type"`
	Features   []float64              `json:"features"`   // Baseline numerical node features
	RiskScore  float64                `json:"risk_score"` // Known historical prior risk [0.0, 1.0]
	IsKnownBad bool                   `json:"is_known_bad"`
	CreatedAt  time.Time              `json:"created_at"` // Exact point-in-time creation
	UpdatedAt  time.Time              `json:"updated_at"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// HeteroEdge represents a directed, time-stamped relationship between two nodes.
type HeteroEdge struct {
	ID         string                 `json:"id"`
	SourceID   string                 `json:"source_id"`
	TargetID   string                 `json:"target_id"`
	Type       HeteroEdgeType         `json:"type"`
	Weight     float64                `json:"weight"`     // Frequency, interaction count, or amount
	Confidence float64                `json:"confidence"` // [0.0, 1.0]
	Timestamp  time.Time              `json:"timestamp"`  // Point-in-time creation (MUST be <= eval_time)
	Provenance string                 `json:"provenance"` // Source system (e.g., "audit_log", "api_gateway", "core_banking")
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// RelationshipRiskLevel categorizes the severity of internal collusion or relationship anomalies.
type RelationshipRiskLevel string

const (
	RiskLevelLow      RelationshipRiskLevel = "LOW"      // Score < 0.30: Normal baseline interactions
	RiskLevelMedium   RelationshipRiskLevel = "MEDIUM"   // Score 0.30 - 0.69: Elevated scrutiny
	RiskLevelHigh     RelationshipRiskLevel = "HIGH"     // Score 0.70 - 0.89: Suspicious collusion / clustering
	RiskLevelCritical RelationshipRiskLevel = "CRITICAL" // Score >= 0.90: Immediate security review
)

// GraphSAGELifecycleState represents the formal governance state of the model.
type GraphSAGELifecycleState string

const (
	StateDesigned           GraphSAGELifecycleState = "DESIGNED"
	StateDataRequired       GraphSAGELifecycleState = "DATA_REQUIRED"
	StateSyntheticTrained   GraphSAGELifecycleState = "SYNTHETIC_TRAINED"
	StateSyntheticValidated GraphSAGELifecycleState = "SYNTHETIC_VALIDATED"
	StateRealDataRequired   GraphSAGELifecycleState = "REAL_DATA_REQUIRED"
	StateTrainedOffline     GraphSAGELifecycleState = "TRAINED_OFFLINE"
	StateOfflineValidated   GraphSAGELifecycleState = "OFFLINE_VALIDATED"
	StateShadow             GraphSAGELifecycleState = "SHADOW"
	StateShadowValidated    GraphSAGELifecycleState = "SHADOW_VALIDATED"
	StateGovernanceReview   GraphSAGELifecycleState = "GOVERNANCE_REVIEW"
	StatePromotionEligible  GraphSAGELifecycleState = "PRODUCTION_ELIGIBLE"
)

// LabelMaturityState classifies the confirmation lifecycle of fraud and dispute labels.
type LabelMaturityState string

const (
	LabelUnlabeled           LabelMaturityState = "UNLABELED"
	LabelSuspected           LabelMaturityState = "SUSPECTED"
	LabelUnderInvestigation  LabelMaturityState = "UNDER_INVESTIGATION"
	LabelConfirmedFraud      LabelMaturityState = "CONFIRMED_FRAUD"
	LabelConfirmedLegitimate LabelMaturityState = "CONFIRMED_LEGITIMATE"
	LabelRejected            LabelMaturityState = "REJECTED"
)

// DatasetType establishes formal separation between synthetic, shadow, and validation data.
type DatasetType string

const (
	DatasetSynthetic      DatasetType = "SYNTHETIC"
	DatasetRealShadow     DatasetType = "REAL_SHADOW"
	DatasetRealValidation DatasetType = "REAL_VALIDATION"
)

// ExplainableCollusionSignal provides structured, machine-readable evidence for compliance and investigations.
type ExplainableCollusionSignal struct {
	RiskSignal        string   `json:"risk_signal"`
	Score             float64  `json:"score"`
	Evidence          []string `json:"evidence"`
	TemporalWindow    string   `json:"temporal_window"`
	AffectedEntities  []string `json:"affected_entities"` // Anonymized/tokenized IDs
	Confidence        float64  `json:"confidence"`
	OperationalReason string   `json:"operational_reason,omitempty"`
}

// RelationshipIntelligenceSignals holds granular sub-scores from the GraphSAGE heads.
type RelationshipIntelligenceSignals struct {
	EmployeeRisk              float64 `json:"employee_risk"`
	ConsumerRisk              float64 `json:"consumer_risk"`
	RelationshipCollusionRisk float64 `json:"relationship_collusion_risk"`
	EntityNeighborhoodRisk    float64 `json:"entity_neighborhood_risk"`
	GraphAnomalyScore         float64 `json:"graph_anomaly_score"`
	TransactionContextScore   float64 `json:"transaction_context_score"`

	// Structural Topology Counts
	ConcentrationRatio   float64 `json:"concentration_ratio"`
	SharedDeviceOverlap  int     `json:"shared_device_overlap"`
	SharedIPOverlap      int     `json:"shared_ip_overlap"`
	TemporalAnomalyDelta float64 `json:"temporal_anomaly_delta_sec"`
	MultiHopPathLength   int     `json:"multi_hop_path_length"`
	OffHoursAccessCount  int     `json:"off_hours_access_count"`
	PreTxAccessProximity float64 `json:"pre_tx_access_proximity_sec"`
}

// RelationshipIntelligenceReport encapsulates the non-enforcing output of the GraphSAGE layer.
type RelationshipIntelligenceReport struct {
	RootNodeID         string                          `json:"root_node_id"`
	RootNodeType       HeteroNodeType                  `json:"root_node_type"`
	OverallRiskScore   float64                         `json:"overall_risk_score"` // [0.0, 1.0]
	RiskLevel          RelationshipRiskLevel           `json:"risk_level"`
	LifecycleState     GraphSAGELifecycleState         `json:"lifecycle_state"`
	Signals            RelationshipIntelligenceSignals `json:"signals"`
	ExplainableSignals []ExplainableCollusionSignal    `json:"explainable_signals"`
	Embedding          []float64                       `json:"embedding"` // 64-dim representation
	ObservedFacts      []string                        `json:"observed_facts"`
	InferredPatterns   []string                        `json:"inferred_patterns"`
	InvestigatorNotes  string                          `json:"investigator_notes"`
	EvaluatedAt        time.Time                       `json:"evaluated_at"`
	LatencyMs          float64                         `json:"latency_ms"`
	IsDegraded         bool                            `json:"is_degraded"`
	DegradeReason      string                          `json:"degrade_reason,omitempty"`
	ModelVersion       string                          `json:"model_version"`
	OperationalMode    string                          `json:"operational_mode"` // "NON_ENFORCING"
	IntendedUse        string                          `json:"intended_use"`     // "INVESTIGATION_ONLY"
}

// CollusionInvestigationDossier provides complete evidence for fraud and compliance operations.
type CollusionInvestigationDossier struct {
	DossierID            string                 `json:"dossier_id"`
	GeneratedAt          time.Time              `json:"generated_at"`
	OverallRiskScore     float64                `json:"overall_risk_score"`
	RiskLevel            RelationshipRiskLevel  `json:"risk_level"`
	EmployeeID           string                 `json:"employee_id"`
	EmployeeRole         string                 `json:"employee_role"`
	AffectedAccounts     []string               `json:"affected_accounts"`
	AffectedConsumers    []string               `json:"affected_consumers"`
	RelationshipPaths    []string               `json:"relationship_paths"`
	TemporalEvidence     map[string]interface{} `json:"temporal_evidence"`
	TransactionSummary   map[string]interface{} `json:"transaction_summary"`
	SharedInfrastructure map[string]interface{} `json:"shared_infrastructure"`
	ConfidenceScore      float64                `json:"confidence_score"`
	HumanExplanation     string                 `json:"human_explanation"`
	DataProvenance       string                 `json:"data_provenance"`
	ModelVersion         string                 `json:"model_version"`
	GovernanceNotice     string                 `json:"governance_notice"`
}

// ReadinessAuditReport captures multi-dimensional dataset readiness in Go.
type ReadinessAuditReport struct {
	AuditedAt                             time.Time               `json:"audited_at"`
	DatasetType                           DatasetType             `json:"dataset_type"`
	IsReadyForRealValidation              bool                    `json:"is_ready_for_real_validation"`
	GovernanceState                       GraphSAGELifecycleState `json:"governance_state"`
	BlockingReasons                       []string                `json:"blocking_reasons"`
	TotalNodes                            int                     `json:"total_nodes"`
	TotalEdges                            int                     `json:"total_edges"`
	ConfirmedFraudLabelsCount             int                     `json:"confirmed_fraud_labels_count"`
	ConfirmedInternalCollusionLabelsCount int                     `json:"confirmed_internal_collusion_labels_count"`
}
