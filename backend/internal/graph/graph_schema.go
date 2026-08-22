package graph

import (
	"time"
)

// NodeType represents the categorical classification of an entity in the fraud graph.
type NodeType string

const (
	NodeUser        NodeType = "USER"
	NodeAccount     NodeType = "ACCOUNT"
	NodeCard        NodeType = "CARD"
	NodeDevice      NodeType = "DEVICE"
	NodeIPAddress   NodeType = "IP_ADDRESS"
	NodeEmail       NodeType = "EMAIL"
	NodePhone       NodeType = "PHONE"
	NodeMerchant    NodeType = "MERCHANT"
	NodeTransaction NodeType = "TRANSACTION"
	NodeLocation    NodeType = "LOCATION"
)

// EdgeType represents the relational link between entities.
type EdgeType string

const (
	EdgeOwns           EdgeType = "OWNS"
	EdgeUsedBy         EdgeType = "USED_BY"
	EdgeConnectedTo    EdgeType = "CONNECTED_TO"
	EdgeTransferredTo  EdgeType = "TRANSFERRED_TO"
	EdgeLoggedInFrom   EdgeType = "LOGGED_IN_FROM"
	EdgeSharesDevice   EdgeType = "SHARES_DEVICE"
	EdgeSharesIP       EdgeType = "SHARES_IP"
	EdgeTransactedWith EdgeType = "TRANSACTED_WITH"
)

// Node represents an entity in the fraud knowledge graph.
type Node struct {
	ID         string                 `json:"id"`
	Type       NodeType               `json:"type"`
	RiskScore  float64                `json:"risk_score"`  // 0.0 to 1.0 known risk
	IsKnownBad bool                   `json:"is_known_bad"` // Confirmed fraudster/compromised entity
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Edge represents a directed, weighted, time-stamped relationship between two nodes.
type Edge struct {
	ID         string                 `json:"id"`
	SourceID   string                 `json:"source_id"`
	TargetID   string                 `json:"target_id"`
	Type       EdgeType               `json:"type"`
	Weight     float64                `json:"weight"`     // Frequency or amount
	Confidence float64                `json:"confidence"` // 0.0 to 1.0
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Path represents a sequential traversal path between entities.
type Path struct {
	Nodes  []*Node  `json:"nodes"`
	Edges  []*Edge  `json:"edges"`
	Length int      `json:"length"`
	Weight float64  `json:"weight"`
}
