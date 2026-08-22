package crime_intelligence

import (
	"time"
)

// CrimeEntityType defines the classification of nodes in the global financial crime intelligence graph.
type CrimeEntityType string

const (
	EntityCriminalGroup         CrimeEntityType = "CRIMINAL_GROUP"
	EntityFraudCampaign         CrimeEntityType = "FRAUD_CAMPAIGN"
	EntityAttackInfrastructure  CrimeEntityType = "ATTACK_INFRASTRUCTURE"
	EntityMoneyFlow             CrimeEntityType = "MONEY_FLOW"
	EntityToolProvider          CrimeEntityType = "TOOL_PROVIDER"
	EntityBotNetwork            CrimeEntityType = "BOT_NETWORK"
	EntityIdentityFactory       CrimeEntityType = "IDENTITY_FACTORY"
	EntityMalwareCluster        CrimeEntityType = "MALWARE_CLUSTER"
)

// CrimeRelationshipType defines directed relationships between threat entities.
type CrimeRelationshipType string

const (
	RelOperates     CrimeRelationshipType = "OPERATES"
	RelFunds        CrimeRelationshipType = "FUNDS"
	RelSupplies     CrimeRelationshipType = "SUPPLIES"
	RelTargets      CrimeRelationshipType = "TARGETS"
	RelEvolvedFrom  CrimeRelationshipType = "EVOLVED_FROM"
	RelConnectedTo  CrimeRelationshipType = "CONNECTED_TO"
)

// CrimeNode represents an entity in the threat intelligence graph (zero raw PII).
type CrimeNode struct {
	EntityID     string                 `json:"entity_id"`
	Type         CrimeEntityType        `json:"type"`
	HashedAlias  string                 `json:"hashed_alias"`
	RiskScore    float64                `json:"risk_score"`
	ThreatLevel  string                 `json:"threat_level"` // "CRITICAL", "HIGH", "MEDIUM"
	FirstSeenAt  time.Time              `json:"first_seen_at"`
	LastSeenAt   time.Time              `json:"last_seen_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// CrimeEdge represents an intelligence connection between two crime nodes.
type CrimeEdge struct {
	EdgeID       string                `json:"edge_id"`
	SourceID     string                `json:"source_id"`
	TargetID     string                `json:"target_id"`
	Type         CrimeRelationshipType `json:"type"`
	Confidence   float64               `json:"confidence"`
	ObservedAt   time.Time             `json:"observed_at"`
}
