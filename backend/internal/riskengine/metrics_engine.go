package riskengine

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

// MetricsEngine manages thread-safe production telemetry counters, gauges, and histograms.
type MetricsEngine struct {
	mu sync.RWMutex

	// Atomic Request Counters
	requestsTotal   int64
	requestsSuccess int64
	requestsFailed  int64
	inferenceErrors int64
	fallbacks       int64

	// Decisions
	decisionsAllow  int64
	decisionsReview int64
	decisionsReject int64

	// Score Distribution (10 buckets: 0-9, 10-19, ..., 90-100)
	scoreBuckets [10]int64

	// Latency Histogram Buckets (<=5ms, <=10ms, <=25ms, <=50ms, <=100ms, <=250ms, <=500ms, <=1000ms, >1000ms)
	latencyLe5ms    int64
	latencyLe10ms   int64
	latencyLe25ms   int64
	latencyLe50ms   int64
	latencyLe100ms  int64
	latencyLe250ms  int64
	latencyLe500ms  int64
	latencyLe1000ms int64
	latencyInf      int64
	latencySumMs    float64

	// Operational Events
	driftEvaluationsTotal int64
	retrainingJobsTotal   int64
	retrainingFailures    int64
	canaryRollbacksTotal  int64
	modelPromotionsTotal  int64
	circuitBreakerTrips   int64
	dependencyFailures    int64

	// Latency Reservoir for Percentiles
	latencyReservoir []float64
	resIdx           int
	resCount         int
	resCap           int
}

// NewMetricsEngine initializes the global telemetry subsystem.
func NewMetricsEngine() *MetricsEngine {
	return &MetricsEngine{
		latencyReservoir: make([]float64, 1000),
		resCap:           1000,
	}
}

// RecordRequest tracks a synchronous risk evaluation request.
func (m *MetricsEngine) RecordRequest(latencyMs float64, decision string, score int, isSuccess, isFallback, isInferError bool) {
	atomic.AddInt64(&m.requestsTotal, 1)
	if isSuccess {
		atomic.AddInt64(&m.requestsSuccess, 1)
	} else {
		atomic.AddInt64(&m.requestsFailed, 1)
	}
	if isFallback {
		atomic.AddInt64(&m.fallbacks, 1)
	}
	if isInferError {
		atomic.AddInt64(&m.inferenceErrors, 1)
	}

	// Record Decision
	switch decision {
	case "ALLOW_RECOMMENDATION", "ALLOW":
		atomic.AddInt64(&m.decisionsAllow, 1)
	case "MANUAL_REVIEW", "REVIEW":
		atomic.AddInt64(&m.decisionsReview, 1)
	case "REJECT_RECOMMENDATION", "REJECT", "BLOCK":
		atomic.AddInt64(&m.decisionsReject, 1)
	}

	// Record Score Distribution
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}
	bIdx := score / 10
	if bIdx >= 10 {
		bIdx = 9
	}
	atomic.AddInt64(&m.scoreBuckets[bIdx], 1)

	// Record Latency Histogram
	if latencyMs <= 5.0 {
		atomic.AddInt64(&m.latencyLe5ms, 1)
	}
	if latencyMs <= 10.0 {
		atomic.AddInt64(&m.latencyLe10ms, 1)
	}
	if latencyMs <= 25.0 {
		atomic.AddInt64(&m.latencyLe25ms, 1)
	}
	if latencyMs <= 50.0 {
		atomic.AddInt64(&m.latencyLe50ms, 1)
	}
	if latencyMs <= 100.0 {
		atomic.AddInt64(&m.latencyLe100ms, 1)
	}
	if latencyMs <= 250.0 {
		atomic.AddInt64(&m.latencyLe250ms, 1)
	}
	if latencyMs <= 500.0 {
		atomic.AddInt64(&m.latencyLe500ms, 1)
	}
	if latencyMs <= 1000.0 {
		atomic.AddInt64(&m.latencyLe1000ms, 1)
	}
	atomic.AddInt64(&m.latencyInf, 1)

	m.mu.Lock()
	m.latencySumMs += latencyMs
	m.latencyReservoir[m.resIdx] = latencyMs
	m.resIdx = (m.resIdx + 1) % m.resCap
	if m.resCount < m.resCap {
		m.resCount++
	}
	m.mu.Unlock()
}

// IncrementDriftEvaluations increments the drift evaluation counter.
func (m *MetricsEngine) IncrementDriftEvaluations() {
	atomic.AddInt64(&m.driftEvaluationsTotal, 1)
}

// IncrementRetrainingJobs increments retraining job counters.
func (m *MetricsEngine) IncrementRetrainingJobs(failed bool) {
	atomic.AddInt64(&m.retrainingJobsTotal, 1)
	if failed {
		atomic.AddInt64(&m.retrainingFailures, 1)
	}
}

// IncrementCanaryRollbacks increments the canary rollback counter.
func (m *MetricsEngine) IncrementCanaryRollbacks() {
	atomic.AddInt64(&m.canaryRollbacksTotal, 1)
}

// IncrementModelPromotions increments the model promotion counter.
func (m *MetricsEngine) IncrementModelPromotions() {
	atomic.AddInt64(&m.modelPromotionsTotal, 1)
}

// IncrementCircuitBreakerTrips increments the circuit breaker trip counter.
func (m *MetricsEngine) IncrementCircuitBreakerTrips() {
	atomic.AddInt64(&m.circuitBreakerTrips, 1)
}

// IncrementDependencyFailures increments the dependency failure counter.
func (m *MetricsEngine) IncrementDependencyFailures() {
	atomic.AddInt64(&m.dependencyFailures, 1)
}

// GetSnapshot returns a map representation of all collected metrics.
func (m *MetricsEngine) GetSnapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var p50, p95, p99 float64
	if m.resCount > 0 {
		sorted := make([]float64, m.resCount)
		copy(sorted, m.latencyReservoir[:m.resCount])
		sort.Float64s(sorted)
		p50 = percentile(sorted, 0.50)
		p95 = percentile(sorted, 0.95)
		p99 = percentile(sorted, 0.99)
	}

	scores := make(map[string]int64)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%d-%d", i*10, (i+1)*10)
		if i == 9 {
			key = "90-100"
		}
		scores[key] = atomic.LoadInt64(&m.scoreBuckets[i])
	}

	return map[string]interface{}{
		"requests_total":          atomic.LoadInt64(&m.requestsTotal),
		"requests_success":        atomic.LoadInt64(&m.requestsSuccess),
		"requests_failed":         atomic.LoadInt64(&m.requestsFailed),
		"inference_errors_total":  atomic.LoadInt64(&m.inferenceErrors),
		"fallbacks_total":         atomic.LoadInt64(&m.fallbacks),
		"decisions": map[string]int64{
			"allow":  atomic.LoadInt64(&m.decisionsAllow),
			"review": atomic.LoadInt64(&m.decisionsReview),
			"reject": atomic.LoadInt64(&m.decisionsReject),
		},
		"latency_percentiles_ms": map[string]float64{
			"p50": p50,
			"p95": p95,
			"p99": p99,
		},
		"score_distribution":        scores,
		"drift_evaluations_total":   atomic.LoadInt64(&m.driftEvaluationsTotal),
		"retraining_jobs_total":     atomic.LoadInt64(&m.retrainingJobsTotal),
		"retraining_failures_total": atomic.LoadInt64(&m.retrainingFailures),
		"canary_rollbacks_total":    atomic.LoadInt64(&m.canaryRollbacksTotal),
		"model_promotions_total":    atomic.LoadInt64(&m.modelPromotionsTotal),
		"circuit_breaker_trips":     atomic.LoadInt64(&m.circuitBreakerTrips),
		"dependency_failures_total": atomic.LoadInt64(&m.dependencyFailures),
	}
}

// ExportPrometheus renders all system metrics into standard Prometheus exposition format.
func (m *MetricsEngine) ExportPrometheus(
	sloSummary *SLOSummary,
	prodModelVer, fbModelVer string,
	driftStatus DriftStatus,
	maxPSI, maxJSD float64,
	canaryStage int,
	cbState CircuitBreakerState,
	retrainActive bool,
) string {
	var buf bytes.Buffer

	// 1. Risk Evaluation Metrics
	buf.WriteString("# HELP risk_evaluations_total Total number of risk evaluation requests.\n")
	buf.WriteString("# TYPE risk_evaluations_total counter\n")
	fmt.Fprintf(&buf, "risk_evaluations_total %d\n", atomic.LoadInt64(&m.requestsTotal))

	buf.WriteString("# HELP risk_evaluation_success_total Total successful risk evaluation requests.\n")
	buf.WriteString("# TYPE risk_evaluation_success_total counter\n")
	fmt.Fprintf(&buf, "risk_evaluation_success_total %d\n", atomic.LoadInt64(&m.requestsSuccess))

	buf.WriteString("# HELP risk_evaluation_errors_total Total failed risk evaluation requests.\n")
	buf.WriteString("# TYPE risk_evaluation_errors_total counter\n")
	fmt.Fprintf(&buf, "risk_evaluation_errors_total %d\n", atomic.LoadInt64(&m.requestsFailed))

	buf.WriteString("# HELP risk_evaluation_fallback_total Total requests routed to emergency fallback.\n")
	buf.WriteString("# TYPE risk_evaluation_fallback_total counter\n")
	fmt.Fprintf(&buf, "risk_evaluation_fallback_total %d\n", atomic.LoadInt64(&m.fallbacks))

	buf.WriteString("# HELP model_inference_errors_total Total ML inference prediction errors.\n")
	buf.WriteString("# TYPE model_inference_errors_total counter\n")
	fmt.Fprintf(&buf, "model_inference_errors_total %d\n", atomic.LoadInt64(&m.inferenceErrors))

	// 2. Latency Histogram
	buf.WriteString("# HELP risk_evaluation_latency_ms Latency of risk evaluation requests in milliseconds.\n")
	buf.WriteString("# TYPE risk_evaluation_latency_ms histogram\n")
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_bucket{le=\"5\"} %d\n", atomic.LoadInt64(&m.latencyLe5ms))
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_bucket{le=\"10\"} %d\n", atomic.LoadInt64(&m.latencyLe10ms))
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_bucket{le=\"25\"} %d\n", atomic.LoadInt64(&m.latencyLe25ms))
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_bucket{le=\"50\"} %d\n", atomic.LoadInt64(&m.latencyLe50ms))
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_bucket{le=\"100\"} %d\n", atomic.LoadInt64(&m.latencyLe100ms))
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_bucket{le=\"250\"} %d\n", atomic.LoadInt64(&m.latencyLe250ms))
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_bucket{le=\"500\"} %d\n", atomic.LoadInt64(&m.latencyLe500ms))
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_bucket{le=\"1000\"} %d\n", atomic.LoadInt64(&m.latencyLe1000ms))
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_bucket{le=\"+Inf\"} %d\n", atomic.LoadInt64(&m.latencyInf))
	m.mu.RLock()
	sum := m.latencySumMs
	m.mu.RUnlock()
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_sum %.2f\n", sum)
	fmt.Fprintf(&buf, "risk_evaluation_latency_ms_count %d\n", atomic.LoadInt64(&m.requestsTotal))

	// 3. Model Information
	buf.WriteString("# HELP model_active_info Active production model metadata.\n")
	buf.WriteString("# TYPE model_active_info gauge\n")
	fmt.Fprintf(&buf, "model_active_info{version=\"%s\",role=\"production\"} 1\n", prodModelVer)
	fmt.Fprintf(&buf, "model_active_info{version=\"%s\",role=\"fallback\"} 1\n", fbModelVer)

	// 4. Decisions
	buf.WriteString("# HELP risk_model_decisions_total Total decisions categorized by action.\n")
	buf.WriteString("# TYPE risk_model_decisions_total counter\n")
	fmt.Fprintf(&buf, "risk_model_decisions_total{action=\"ALLOW\"} %d\n", atomic.LoadInt64(&m.decisionsAllow))
	fmt.Fprintf(&buf, "risk_model_decisions_total{action=\"MANUAL_REVIEW\"} %d\n", atomic.LoadInt64(&m.decisionsReview))
	fmt.Fprintf(&buf, "risk_model_decisions_total{action=\"REJECT\"} %d\n", atomic.LoadInt64(&m.decisionsReject))

	// 5. Drift Metrics
	buf.WriteString("# HELP drift_status Current drift status (0=HEALTHY, 1=WARNING, 2=DEGRADED, 3=CRITICAL).\n")
	buf.WriteString("# TYPE drift_status gauge\n")
	var driftCode int
	switch driftStatus {
	case DriftStatusDegraded:
		driftCode = 2
	case DriftStatusCritical:
		driftCode = 3
	default:
		driftCode = 0
	}
	fmt.Fprintf(&buf, "drift_status %d\n", driftCode)
	fmt.Fprintf(&buf, "drift_max_psi %.4f\n", maxPSI)
	fmt.Fprintf(&buf, "drift_max_jsd %.4f\n", maxJSD)

	// 6. Retraining & Canary Metrics
	buf.WriteString("# HELP retraining_jobs_total Total candidate model retraining jobs.\n")
	buf.WriteString("# TYPE retraining_jobs_total counter\n")
	fmt.Fprintf(&buf, "retraining_jobs_total %d\n", atomic.LoadInt64(&m.retrainingJobsTotal))
	fmt.Fprintf(&buf, "retraining_jobs_failed_total %d\n", atomic.LoadInt64(&m.retrainingFailures))
	activeVal := 0
	if retrainActive {
		activeVal = 1
	}
	fmt.Fprintf(&buf, "retraining_active %d\n", activeVal)

	buf.WriteString("# HELP canary_stage Current candidate traffic percentage (0-100).\n")
	buf.WriteString("# TYPE canary_stage gauge\n")
	fmt.Fprintf(&buf, "canary_stage %d\n", canaryStage)
	fmt.Fprintf(&buf, "canary_rollbacks_total %d\n", atomic.LoadInt64(&m.canaryRollbacksTotal))
	fmt.Fprintf(&buf, "model_promotions_total %d\n", atomic.LoadInt64(&m.modelPromotionsTotal))

	// 7. Circuit Breaker
	buf.WriteString("# HELP circuit_breaker_state Circuit breaker status (0=HEALTHY, 1=WARNING, 2=FAILED, 3=ROLLED_BACK).\n")
	buf.WriteString("# TYPE circuit_breaker_state gauge\n")
	var cbCode int
	switch cbState {
	case CircuitStateWarning:
		cbCode = 1
	case CircuitStateFailed:
		cbCode = 2
	case CircuitStateRolledBack, CircuitStateRollingBack:
		cbCode = 3
	default:
		cbCode = 0
	}
	fmt.Fprintf(&buf, "circuit_breaker_state %d\n", cbCode)
	fmt.Fprintf(&buf, "circuit_breaker_trips_total %d\n", atomic.LoadInt64(&m.circuitBreakerTrips))

	// 8. SLO Metrics
	if sloSummary != nil {
		buf.WriteString("# HELP slo_availability Current availability ratio.\n")
		buf.WriteString("# TYPE slo_availability gauge\n")
		if mAvail, ok := sloSummary.Measurements["slo_availability"]; ok {
			fmt.Fprintf(&buf, "slo_availability %.6f\n", mAvail.CurrentValue)
			fmt.Fprintf(&buf, "slo_error_budget_remaining{slo=\"slo_availability\"} %.2f\n", mAvail.ErrorBudgetRemaining)
			fmt.Fprintf(&buf, "slo_burn_rate{slo=\"slo_availability\"} %.2f\n", mAvail.BurnRate)
		}

		if mP95, ok := sloSummary.Measurements["slo_p95_latency"]; ok {
			fmt.Fprintf(&buf, "slo_latency_p95 %.2f\n", mP95.CurrentValue)
		}
		if mP99, ok := sloSummary.Measurements["slo_p99_latency"]; ok {
			fmt.Fprintf(&buf, "slo_latency_p99 %.2f\n", mP99.CurrentValue)
		}
	}

	return buf.String()
}
