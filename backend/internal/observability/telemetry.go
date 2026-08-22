package observability

import (
	"sync"
	"sync/atomic"
	"time"
)

// PlatformMetricsTracker records real-time Prometheus / OpenTelemetry telemetry counters.
type PlatformMetricsTracker struct {
	mu                   sync.RWMutex
	RiskRequestsTotal    uint64
	FraudDetectedTotal   uint64
	TotalLatencyMs       float64
	ModelInferenceCount  uint64
	AgentEvaluationsCount uint64
	CasesResolvedTotal   uint64
}

// GlobalMetrics provides an active singleton metrics instance.
var GlobalMetrics = &PlatformMetricsTracker{}

// RecordRiskEvaluation logs a completed decision request and updates latency meters.
func (m *PlatformMetricsTracker) RecordRiskEvaluation(latencyMs float64, isFraud bool) {
	atomic.AddUint64(&m.RiskRequestsTotal, 1)
	if isFraud {
		atomic.AddUint64(&m.FraudDetectedTotal, 1)
	}

	m.mu.Lock()
	m.TotalLatencyMs += latencyMs
	m.mu.Unlock()
}

// RecordAgentExecution logs an autonomous agent investigation.
func (m *PlatformMetricsTracker) RecordAgentExecution() {
	atomic.AddUint64(&m.AgentEvaluationsCount, 1)
}

// RecordCaseResolved logs case resolution time.
func (m *PlatformMetricsTracker) RecordCaseResolved() {
	atomic.AddUint64(&m.CasesResolvedTotal, 1)
}

// GetSnapshot returns atomic summary values.
func (m *PlatformMetricsTracker) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reqs := atomic.LoadUint64(&m.RiskRequestsTotal)
	frauds := atomic.LoadUint64(&m.FraudDetectedTotal)
	agents := atomic.LoadUint64(&m.AgentEvaluationsCount)
	cases := atomic.LoadUint64(&m.CasesResolvedTotal)

	avgLatency := 0.0
	if reqs > 0 {
		avgLatency = m.TotalLatencyMs / float64(reqs)
	}

	return map[string]interface{}{
		"risk_requests_total":    reqs,
		"fraud_detected_total":   frauds,
		"avg_latency_ms":         avgLatency,
		"agent_evaluations_total": agents,
		"cases_resolved_total":   cases,
		"timestamp":              time.Now().UTC(),
	}
}
