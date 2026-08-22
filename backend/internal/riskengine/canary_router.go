package riskengine

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shankywho/ropus/backend/internal/audit"
)

// ModelRoute represents the routing decision for a risk evaluation.
type ModelRoute int

const (
	RouteLegacy ModelRoute = iota
	RouteCandidate
)

// String returns the canonical name of the model route.
func (r ModelRoute) String() string {
	switch r {
	case RouteCandidate:
		return "CANDIDATE"
	default:
		return "LEGACY"
	}
}

// CanaryRouterConfig defines configuration parameters for staged production rollout.
type CanaryRouterConfig struct {
	Enabled                  bool    `json:"enabled"`
	Percentage               int     `json:"percentage"`                 // 0 to 100
	CandidateModelVersion    string  `json:"candidate_model_version"`    // "fraud-xgb-25f-candidate-v1"
	CandidateFeatureContract string  `json:"candidate_feature_contract"` // "v2.5"
	MaxErrorRate             float64 `json:"max_error_rate"`             // Default: 0.01 (1%)
	MaxFallbackRate          float64 `json:"max_fallback_rate"`          // Default: 0.01 (1%)
	MaxP95LatencyMs          float64 `json:"max_p95_latency_ms"`         // Default: 15.0 ms
	MaxP99LatencyMs          float64 `json:"max_p99_latency_ms"`         // Default: 25.0 ms
	MaxDecisionChangeRate    float64 `json:"max_decision_change_rate"`   // Default: 0.10 (10%)
}

// DefaultCanaryRouterConfig returns the default safe configuration (0% candidate traffic).
func DefaultCanaryRouterConfig() CanaryRouterConfig {
	return CanaryRouterConfig{
		Enabled:                  false,
		Percentage:               0,
		CandidateModelVersion:    "fraud-xgb-25f-candidate-v1",
		CandidateFeatureContract: MLFeatureContractV25,
		MaxErrorRate:             0.01,
		MaxFallbackRate:          0.01,
		MaxP95LatencyMs:          15.0,
		MaxP99LatencyMs:          25.0,
		MaxDecisionChangeRate:    0.10,
	}
}

// Validate ensures configuration values are within safe, bounded limits.
func (cfg *CanaryRouterConfig) Validate() {
	if cfg.Percentage < 0 {
		log.Printf("Warning: Canary percentage %d is negative; failing safe to 0%%", cfg.Percentage)
		cfg.Percentage = 0
		cfg.Enabled = false
	} else if cfg.Percentage > 100 {
		log.Printf("Warning: Canary percentage %d exceeds 100%%; clamping to 100%%", cfg.Percentage)
		cfg.Percentage = 100
	}

	if cfg.CandidateModelVersion == "" {
		cfg.CandidateModelVersion = "fraud-xgb-25f-candidate-v1"
	}
	if cfg.CandidateFeatureContract == "" {
		cfg.CandidateFeatureContract = MLFeatureContractV25
	}
	if cfg.MaxErrorRate <= 0.0 {
		cfg.MaxErrorRate = 0.01
	}
	if cfg.MaxFallbackRate <= 0.0 {
		cfg.MaxFallbackRate = 0.01
	}
	if cfg.MaxP95LatencyMs <= 0.0 {
		cfg.MaxP95LatencyMs = 15.0
	}
	if cfg.MaxP99LatencyMs <= 0.0 {
		cfg.MaxP99LatencyMs = 25.0
	}
	if cfg.MaxDecisionChangeRate <= 0.0 {
		cfg.MaxDecisionChangeRate = 0.10
	}
}

// LatencyReservoir maintains a thread-safe circular sample of recent latencies for percentiles.
type LatencyReservoir struct {
	mu      sync.RWMutex
	samples []float64
	maxSize int
	idx     int
	count   int
}

// NewLatencyReservoir creates a reservoir with a given capacity.
func NewLatencyReservoir(maxSize int) *LatencyReservoir {
	if maxSize <= 0 {
		maxSize = 2000
	}
	return &LatencyReservoir{
		samples: make([]float64, maxSize),
		maxSize: maxSize,
	}
}

// Add appends a latency sample thread-safely.
func (r *LatencyReservoir) Add(val float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.samples[r.idx] = val
	r.idx = (r.idx + 1) % r.maxSize
	if r.count < r.maxSize {
		r.count++
	}
}

// Percentiles calculates p50, p95, and p99 from the recorded samples.
func (r *LatencyReservoir) Percentiles() (p50, p95, p99 float64) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.count == 0 {
		return 0, 0, 0
	}

	copied := make([]float64, r.count)
	copy(copied, r.samples[:r.count])
	sort.Float64s(copied)

	calcPct := func(pct float64) float64 {
		if len(copied) == 1 {
			return copied[0]
		}
		idx := int(math.Ceil(pct*float64(len(copied)))) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(copied) {
			idx = len(copied) - 1
		}
		return copied[idx]
	}

	return calcPct(0.50), calcPct(0.95), calcPct(0.99)
}

// CanaryMetrics tracks live operational counters and distributions atomically.
type CanaryMetrics struct {
	LegacyRequestsTotal           atomic.Int64
	CandidateRequestsTotal        atomic.Int64
	CandidateSuccessTotal         atomic.Int64
	CandidateErrorTotal           atomic.Int64
	CandidateFallbackTotal        atomic.Int64
	CandidateDecisionAllowTotal   atomic.Int64
	CandidateDecisionReviewTotal  atomic.Int64
	CandidateDecisionDeclineTotal atomic.Int64
	DecisionChangedTotal          atomic.Int64
	TotalCandidateLatencyUs       atomic.Int64
	LatencyReservoir              *LatencyReservoir
}

// Snapshot returns a thread-safe point-in-time dictionary of current metrics.
func (m *CanaryMetrics) Snapshot() map[string]interface{} {
	candidateTotal := m.CandidateRequestsTotal.Load()
	candidateSuccess := m.CandidateSuccessTotal.Load()
	candidateErrors := m.CandidateErrorTotal.Load()
	candidateFallbacks := m.CandidateFallbackTotal.Load()
	legacyTotal := m.LegacyRequestsTotal.Load()
	totalRequests := legacyTotal + candidateTotal

	errorRate := 0.0
	fallbackRate := 0.0
	if candidateTotal > 0 {
		errorRate = float64(candidateErrors) / float64(candidateTotal)
		fallbackRate = float64(candidateFallbacks) / float64(candidateTotal)
	}

	actualCanaryPct := 0.0
	if totalRequests > 0 {
		actualCanaryPct = (float64(candidateTotal) / float64(totalRequests)) * 100.0
	}

	p50, p95, p99 := 0.0, 0.0, 0.0
	if m.LatencyReservoir != nil {
		p50, p95, p99 = m.LatencyReservoir.Percentiles()
	}

	avgLatencyMs := 0.0
	if candidateSuccess > 0 {
		avgLatencyMs = float64(m.TotalCandidateLatencyUs.Load()) / (float64(candidateSuccess) * 1000.0)
	}

	return map[string]interface{}{
		"total_requests":                  totalRequests,
		"legacy_requests_total":           legacyTotal,
		"candidate_requests_total":        candidateTotal,
		"candidate_success_total":         candidateSuccess,
		"candidate_error_total":           candidateErrors,
		"candidate_fallback_total":        candidateFallbacks,
		"candidate_error_rate":            math.Round(errorRate*10000) / 10000,
		"candidate_fallback_rate":         math.Round(fallbackRate*10000) / 10000,
		"candidate_decision_allow_total":  m.CandidateDecisionAllowTotal.Load(),
		"candidate_decision_review_total": m.CandidateDecisionReviewTotal.Load(),
		"candidate_decision_decline_total": m.CandidateDecisionDeclineTotal.Load(),
		"decision_changed_total":          m.DecisionChangedTotal.Load(),
		"actual_canary_percentage":        math.Round(actualCanaryPct*100) / 100,
		"candidate_avg_latency_ms":        math.Round(avgLatencyMs*100) / 100,
		"candidate_p50_latency_ms":        math.Round(p50*100) / 100,
		"candidate_p95_latency_ms":        math.Round(p95*100) / 100,
		"candidate_p99_latency_ms":        math.Round(p99*100) / 100,
	}
}

// SafetyGateStatus represents the operational health rating of the canary rollout.
type SafetyGateStatus string

const (
	GateStatusHealthy SafetyGateStatus = "HEALTHY"
	GateStatusWarning SafetyGateStatus = "WARNING"
	GateStatusFailed  SafetyGateStatus = "FAILED"
	GateStatusIdle    SafetyGateStatus = "IDLE"
)

// SafetyGateResult details the outcome of automated gate evaluation.
type SafetyGateResult struct {
	Status          SafetyGateStatus       `json:"status"`
	Violations      []string               `json:"violations"`
	Warnings        []string               `json:"warnings"`
	MetricsSnapshot map[string]interface{} `json:"metrics_snapshot"`
	EvaluatedAt     time.Time              `json:"evaluated_at"`
}

// CanaryRouter manages deterministic request routing, safety evaluation, circuit breakers, and fallback tracking.
type CanaryRouter struct {
	config         CanaryRouterConfig
	metrics        CanaryMetrics
	circuitBreaker *CircuitBreaker
	chClient       *audit.ClickHouseClient
	mu             sync.RWMutex
}

// NewCanaryRouter instantiates the CanaryRouter with validated configuration and circuit breaker.
func NewCanaryRouter(cfg CanaryRouterConfig, chClient *audit.ClickHouseClient) *CanaryRouter {
	cfg.Validate()

	cbCfg := CircuitBreakerConfig{
		MaxErrorRate:          cfg.MaxErrorRate,
		MaxFallbackRate:       cfg.MaxFallbackRate,
		MaxP95LatencyMs:       cfg.MaxP95LatencyMs,
		MaxP99LatencyMs:       cfg.MaxP99LatencyMs,
		MaxDecisionChangeRate: cfg.MaxDecisionChangeRate,
		MinSampleCount:        10,
		FailureWindow:         3,
		CooldownSeconds:       300,
	}

	router := &CanaryRouter{
		config: cfg,
		metrics: CanaryMetrics{
			LatencyReservoir: NewLatencyReservoir(2000),
		},
		circuitBreaker: NewCircuitBreaker(cbCfg),
		chClient:       chClient,
	}

	if cfg.Enabled && cfg.Percentage > 0 {
		log.Printf("Canary rollout initialized: enabled=true, candidate_percentage=%d%%, candidate_model=%s",
			cfg.Percentage, cfg.CandidateModelVersion)
	} else {
		log.Printf("Canary rollout initialized: enabled=false, candidate_percentage=%d%%, candidate_model=%s (100%% legacy path active)",
			cfg.Percentage, cfg.CandidateModelVersion)
	}

	return router
}

// Route deterministically decides whether a transaction should use LEGACY or CANDIDATE.
// It computes a stable SHA-256 hash over tenantID:transactionID modulo 100.
func (cr *CanaryRouter) Route(tenantID, transactionID string) ModelRoute {
	cr.mu.RLock()
	enabled := cr.config.Enabled
	percentage := cr.config.Percentage
	cr.mu.RUnlock()

	// If circuit breaker is tripped/rolled back, force RouteLegacy
	if cr.circuitBreaker != nil {
		status := cr.circuitBreaker.GetStatus()
		if status["state"] == string(CircuitStateRolledBack) {
			return RouteLegacy
		}
	}

	if !enabled || percentage <= 0 {
		return RouteLegacy
	}
	if percentage >= 100 {
		return RouteCandidate
	}

	key := tenantID + ":" + transactionID
	if tenantID == "" || transactionID == "" {
		key = fmt.Sprintf("anon_%s", transactionID)
	}

	h := sha256.Sum256([]byte(key))
	val := binary.BigEndian.Uint32(h[:4])
	bucket := int(val % 100) // 0 to 99

	if bucket < percentage {
		return RouteCandidate
	}
	return RouteLegacy
}

// RecordLegacyRequest records a request routed to the legacy model.
func (cr *CanaryRouter) RecordLegacyRequest() {
	cr.metrics.LegacyRequestsTotal.Add(1)
}

// RecordCandidateRequest records a request routed to the candidate model.
func (cr *CanaryRouter) RecordCandidateRequest() {
	cr.metrics.CandidateRequestsTotal.Add(1)
}

// RecordCandidateSuccess records a successful candidate model inference.
func (cr *CanaryRouter) RecordCandidateSuccess(latencyMs float64, decision string) {
	cr.metrics.CandidateSuccessTotal.Add(1)
	cr.metrics.TotalCandidateLatencyUs.Add(int64(latencyMs * 1000.0))
	if cr.metrics.LatencyReservoir != nil {
		cr.metrics.LatencyReservoir.Add(latencyMs)
	}

	switch decision {
	case "ALLOW_RECOMMENDATION", "ALLOW":
		cr.metrics.CandidateDecisionAllowTotal.Add(1)
	case "MANUAL_REVIEW":
		cr.metrics.CandidateDecisionReviewTotal.Add(1)
	case "DECLINE_RECOMMENDATION", "DECLINE":
		cr.metrics.CandidateDecisionDeclineTotal.Add(1)
	}

	cr.checkCircuitBreaker()
}

// RecordCandidateFallback records a candidate inference failure and legacy fallback.
func (cr *CanaryRouter) RecordCandidateFallback(latencyMs float64, errReason string) {
	cr.metrics.CandidateErrorTotal.Add(1)
	cr.metrics.CandidateFallbackTotal.Add(1)
	if latencyMs > 0 && cr.metrics.LatencyReservoir != nil {
		cr.metrics.LatencyReservoir.Add(latencyMs)
	}

	cr.checkCircuitBreaker()
}

// RecordCandidateError records a candidate error without fallback.
func (cr *CanaryRouter) RecordCandidateError(errReason string) {
	cr.metrics.CandidateErrorTotal.Add(1)
	cr.checkCircuitBreaker()
}

// RecordDecisionChanged records when candidate decision diverges from legacy baseline.
func (cr *CanaryRouter) RecordDecisionChanged() {
	cr.metrics.DecisionChangedTotal.Add(1)
}

// checkCircuitBreaker inspects recent metrics and automatically rolls back if breached.
func (cr *CanaryRouter) checkCircuitBreaker() {
	if cr.circuitBreaker == nil {
		return
	}

	snapshot := cr.metrics.Snapshot()
	tripped, state, reason, violations := cr.circuitBreaker.EvaluateAndCheckTrip(snapshot)

	if tripped {
		cr.mu.Lock()
		prevPct := cr.config.Percentage
		cr.config.Percentage = 0
		cr.config.Enabled = false
		cr.mu.Unlock()

		log.Printf("AUTOMATIC ROLLBACK TRIGGERED: Candidate traffic reset to 0%% (was %d%%). Reason: %s (violations: %v)",
			prevPct, reason, violations)

		// Persist audit event to ClickHouse
		cr.LogRolloutEvent(context.Background(), audit.CanaryRolloutEvent{
			EventID:              fmt.Sprintf("evt_auto_%d", time.Now().UnixNano()),
			Timestamp:            time.Now().UTC(),
			EventType:            "AUTOMATIC_ROLLBACK",
			PreviousPercentage:   uint8(prevPct),
			NewPercentage:        0,
			PreviousModelVersion: cr.config.CandidateModelVersion,
			NewModelVersion:      "fraud-xgb-25f-v3.0",
			Trigger:              "AUTOMATIC_CIRCUIT_BREAKER",
			SafetyStatus:         string(state),
			ErrorRate:            snapshot["candidate_error_rate"].(float64),
			FallbackRate:         snapshot["candidate_fallback_rate"].(float64),
			P95LatencyMs:         snapshot["candidate_p95_latency_ms"].(float64),
			P99LatencyMs:         snapshot["candidate_p99_latency_ms"].(float64),
			Actor:                "circuit_breaker_daemon",
			Reason:               reason,
		})
	}
}

// EvaluateSafetyGates evaluates live metrics against operational safety thresholds.
func (cr *CanaryRouter) EvaluateSafetyGates() SafetyGateResult {
	cr.mu.RLock()
	cfg := cr.config
	cr.mu.RUnlock()

	snapshot := cr.metrics.Snapshot()
	candidateTotal := snapshot["candidate_requests_total"].(int64)

	result := SafetyGateResult{
		Status:          GateStatusHealthy,
		Violations:      make([]string, 0),
		Warnings:        make([]string, 0),
		MetricsSnapshot: snapshot,
		EvaluatedAt:     time.Now().UTC(),
	}

	if !cfg.Enabled || cfg.Percentage == 0 || candidateTotal == 0 {
		if cr.circuitBreaker != nil {
			cbStatus := cr.circuitBreaker.GetStatus()
			if cbStatus["state"] == string(CircuitStateRolledBack) || cbStatus["state"] == string(CircuitStateFailed) {
				result.Status = GateStatusFailed
				reason, _ := cbStatus["last_trip_reason"].(string)
				if reason != "" {
					result.Violations = append(result.Violations, reason)
				} else {
					result.Violations = append(result.Violations, fmt.Sprintf("Circuit breaker state: %v", cbStatus["state"]))
				}
				return result
			}
		}
		result.Status = GateStatusIdle
		return result
	}

	errorRate := snapshot["candidate_error_rate"].(float64)
	fallbackRate := snapshot["candidate_fallback_rate"].(float64)
	p95 := snapshot["candidate_p95_latency_ms"].(float64)
	p99 := snapshot["candidate_p99_latency_ms"].(float64)

	// Gate 1: Candidate Error Rate
	if errorRate > cfg.MaxErrorRate && candidateTotal >= 10 {
		msg := fmt.Sprintf("Candidate error rate %.2f%% exceeds max allowed %.2f%%", errorRate*100, cfg.MaxErrorRate*100)
		result.Violations = append(result.Violations, msg)
		result.Status = GateStatusFailed
	} else if errorRate > cfg.MaxErrorRate*0.5 && candidateTotal >= 5 {
		msg := fmt.Sprintf("Candidate error rate %.2f%% is approaching max allowed %.2f%%", errorRate*100, cfg.MaxErrorRate*100)
		result.Warnings = append(result.Warnings, msg)
	}

	// Gate 2: Fallback Rate
	if fallbackRate > cfg.MaxFallbackRate && candidateTotal >= 10 {
		msg := fmt.Sprintf("Candidate fallback rate %.2f%% exceeds max allowed %.2f%%", fallbackRate*100, cfg.MaxFallbackRate*100)
		result.Violations = append(result.Violations, msg)
		result.Status = GateStatusFailed
	}

	// Gate 3: Latency p95
	if p95 > cfg.MaxP95LatencyMs && candidateTotal >= 20 {
		msg := fmt.Sprintf("Candidate p95 latency %.2fms exceeds threshold %.2fms", p95, cfg.MaxP95LatencyMs)
		result.Violations = append(result.Violations, msg)
		if result.Status != GateStatusFailed {
			result.Status = GateStatusWarning
		}
	}

	// Gate 4: Latency p99
	if p99 > cfg.MaxP99LatencyMs && candidateTotal >= 20 {
		msg := fmt.Sprintf("Candidate p99 latency %.2fms exceeds threshold %.2fms", p99, cfg.MaxP99LatencyMs)
		result.Violations = append(result.Violations, msg)
		if result.Status != GateStatusFailed {
			result.Status = GateStatusWarning
		}
	}

	return result
}

// LogEvaluation records a rollout evaluation event asynchronously to ClickHouse.
func (cr *CanaryRouter) LogEvaluation(ctx context.Context, eval audit.CanaryRolloutEvaluation) {
	if cr.chClient == nil {
		return
	}
	go func() {
		insertCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := cr.chClient.InsertCanaryEvaluation(insertCtx, eval); err != nil {
			log.Printf("Warning: Failed to persist canary rollout evaluation to ClickHouse: %v", err)
		}
	}()
}

// LogRolloutEvent persists a manual or automated rollout change to ClickHouse.
func (cr *CanaryRouter) LogRolloutEvent(ctx context.Context, event audit.CanaryRolloutEvent) {
	if cr.chClient == nil {
		return
	}
	go func() {
		insertCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := cr.chClient.InsertCanaryRolloutEvent(insertCtx, event); err != nil {
			log.Printf("Warning: Failed to persist canary rollout event to ClickHouse: %v", err)
		}
	}()
}

// UpdateConfig atomically updates the rollout configuration at runtime without restart.
func (cr *CanaryRouter) UpdateConfig(enabled bool, percentage int, actor, reason string) error {
	if percentage < 0 || percentage > 100 {
		return fmt.Errorf("invalid percentage: %d (must be between 0 and 100)", percentage)
	}

	cr.mu.Lock()
	prevPct := cr.config.Percentage
	prevEnabled := cr.config.Enabled
	cr.config.Percentage = percentage
	cr.config.Enabled = enabled && percentage > 0
	currCandidateModel := cr.config.CandidateModelVersion
	currEnabled := cr.config.Enabled
	cr.mu.Unlock()

	// Reset circuit breaker on manual admin reconfiguration or rollback to 0%
	if cr.circuitBreaker != nil && (percentage == 0 || strings.HasPrefix(actor, "admin_") || actor == "ADMIN_OPERATOR") {
		cr.circuitBreaker.Reset()
	}

	eventType := "MANUAL_ROLLOUT"
	if percentage == 0 && prevPct > 0 {
		eventType = "MANUAL_ROLLBACK"
	} else if percentage == 100 {
		eventType = "MODEL_PROMOTION"
	}

	snapshot := cr.metrics.Snapshot()
	event := audit.CanaryRolloutEvent{
		EventID:              fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Timestamp:            time.Now().UTC(),
		EventType:            eventType,
		PreviousPercentage:   uint8(prevPct),
		NewPercentage:        uint8(percentage),
		PreviousModelVersion: currCandidateModel,
		NewModelVersion:      "fraud-xgb-25f-v3.0",
		Trigger:              "ADMIN_API",
		SafetyStatus:         "HEALTHY",
		ErrorRate:            snapshot["candidate_error_rate"].(float64),
		FallbackRate:         snapshot["candidate_fallback_rate"].(float64),
		P95LatencyMs:         snapshot["candidate_p95_latency_ms"].(float64),
		P99LatencyMs:         snapshot["candidate_p99_latency_ms"].(float64),
		Actor:                actor,
		Reason:               reason,
	}
	cr.LogRolloutEvent(context.Background(), event)

	log.Printf("Canary configuration hot-reloaded: enabled=%v (was %v), percentage=%d%% (was %d%%), actor=%s, reason=%s",
		currEnabled, prevEnabled, percentage, prevPct, actor, reason)
	return nil
}

// GetStatus returns the complete operational status of the canary router and circuit breaker.
func (cr *CanaryRouter) GetStatus() map[string]interface{} {
	cr.mu.RLock()
	cfg := cr.config
	cr.mu.RUnlock()

	safety := cr.EvaluateSafetyGates()
	var cbStatus map[string]interface{}
	if cr.circuitBreaker != nil {
		cbStatus = cr.circuitBreaker.GetStatus()
	}

	return map[string]interface{}{
		"enabled":                  cfg.Enabled,
		"target_percentage":        cfg.Percentage,
		"production_model":         "fraud-xgb-25f-v3.0",
		"candidate_model":          cfg.CandidateModelVersion,
		"feature_contract":         "fraud-risk-25f-v2.5",
		"calibration_version":      "beta-calibrated-v2.5",
		"safety_gate_status":       safety.Status,
		"safety_violations":        safety.Violations,
		"safety_warnings":          safety.Warnings,
		"circuit_breaker":          cbStatus,
		"thresholds": map[string]interface{}{
			"max_error_rate":           cfg.MaxErrorRate,
			"max_fallback_rate":        cfg.MaxFallbackRate,
			"max_p95_latency_ms":       cfg.MaxP95LatencyMs,
			"max_p99_latency_ms":       cfg.MaxP99LatencyMs,
			"max_decision_change_rate": cfg.MaxDecisionChangeRate,
		},
		"metrics": safety.MetricsSnapshot,
	}
}

// SetPercentage dynamically updates the target percentage (used in controlled staged rollout testing).
func (cr *CanaryRouter) SetPercentage(pct int) {
	_ = cr.UpdateConfig(pct > 0, pct, "test_harness", "direct SetPercentage call")
}

// GetPercentage returns the current canary target percentage.
func (cr *CanaryRouter) GetPercentage() int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.config.Percentage
}

// GetCircuitBreaker returns the active circuit breaker instance.
func (cr *CanaryRouter) GetCircuitBreaker() *CircuitBreaker {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.circuitBreaker
}

