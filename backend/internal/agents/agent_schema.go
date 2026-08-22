package agents

import (
	"time"
)

// AgentType defines the specialized role of an AI agent in the autonomous organization.
type AgentType string

const (
	AgentFraudInvestigator AgentType = "FraudInvestigatorAgent"
	AgentThreatHunter      AgentType = "ThreatHunterAgent"
	AgentRiskOptimizer     AgentType = "RiskOptimizerAgent"
	AgentCompliance        AgentType = "ComplianceAgent"
	AgentResponse          AgentType = "ResponseAgent"
	AgentDataQuality       AgentType = "DataQualityAgent"
)

// AgentStatus tracks the lifecycle phase of an agent execution.
type AgentStatus string

const (
	StatusCreated      AgentStatus = "CREATED"
	StatusInitializing AgentStatus = "INITIALIZING"
	StatusRunning      AgentStatus = "RUNNING"
	StatusWaiting      AgentStatus = "WAITING"
	StatusCompleted    AgentStatus = "COMPLETED"
	StatusFailed       AgentStatus = "FAILED"
)

// AgentTask represents a delegated work unit assigned to an autonomous agent.
type AgentTask struct {
	TaskID          string                 `json:"task_id"`
	TraceID         string                 `json:"trace_id"`
	TargetAgentType AgentType              `json:"target_agent_type"`
	Payload         map[string]interface{} `json:"payload"`
	CreatedAt       time.Time              `json:"created_at"`
	Deadline        time.Time              `json:"deadline"`
}

// AgentResult encapsulates the output, findings, and confidence of a finished agent execution.
type AgentResult struct {
	TaskID          string                 `json:"task_id"`
	TraceID         string                 `json:"trace_id"`
	AgentID         string                 `json:"agent_id"`
	AgentType       AgentType              `json:"agent_type"`
	Status          AgentStatus            `json:"status"`
	Confidence      float64                `json:"confidence"`
	ReasoningOutput map[string]interface{} `json:"reasoning_output"`
	ExecutionMs     float64                `json:"execution_ms"`
	CompletedAt     time.Time              `json:"completed_at"`
	Error           string                 `json:"error,omitempty"`
}
