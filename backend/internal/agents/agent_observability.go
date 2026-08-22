package agents

import (
	"sync"
	"time"
)

// AgentExecutionTelemetry records performance, latency, and decision confidence for an agent run.
type AgentExecutionTelemetry struct {
	TraceID     string    `json:"trace_id"`
	AgentType   AgentType `json:"agent_type"`
	DurationMs  float64   `json:"duration_ms"`
	Confidence  float64   `json:"confidence"`
	Status      string    `json:"status"`
	ExecutedAt  time.Time `json:"executed_at"`
}

// AgentObservabilityPlatform monitors multi-agent performance and decision health.
type AgentObservabilityPlatform struct {
	mu           sync.RWMutex
	executions   []*AgentExecutionTelemetry
	overrideCount int
}

// NewAgentObservabilityPlatform initializes the agent observability aggregator.
func NewAgentObservabilityPlatform() *AgentObservabilityPlatform {
	return &AgentObservabilityPlatform{
		executions: make([]*AgentExecutionTelemetry, 0),
	}
}

// RecordExecution logs an agent telemetry event.
func (o *AgentObservabilityPlatform) RecordExecution(traceID string, agentType AgentType, durationMs, confidence float64, status string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.executions = append(o.executions, &AgentExecutionTelemetry{
		TraceID:    traceID,
		AgentType:  agentType,
		DurationMs: durationMs,
		Confidence: confidence,
		Status:     status,
		ExecutedAt: time.Now().UTC(),
	})
}

// RecordHumanOverride increments analyst overrides over agent suggestions.
func (o *AgentObservabilityPlatform) RecordHumanOverride() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.overrideCount++
}

// GetMetricsSummary computes aggregate health metrics.
func (o *AgentObservabilityPlatform) GetMetricsSummary() (int, float64, int) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	total := len(o.executions)
	avgMs := 0.0
	if total > 0 {
		sum := 0.0
		for _, e := range o.executions {
			sum += e.DurationMs
		}
		avgMs = sum / float64(total)
	}

	return total, avgMs, o.overrideCount
}
