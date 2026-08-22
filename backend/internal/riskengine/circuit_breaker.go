package riskengine

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// CircuitBreakerState represents the operational state of the automatic rollback safety gate.
type CircuitBreakerState string

const (
	CircuitStateHealthy     CircuitBreakerState = "HEALTHY"
	CircuitStateWarning     CircuitBreakerState = "WARNING"
	CircuitStateFailed      CircuitBreakerState = "FAILED"
	CircuitStateRollingBack CircuitBreakerState = "ROLLING_BACK"
	CircuitStateRolledBack  CircuitBreakerState = "ROLLED_BACK"
)

// CircuitBreakerConfig defines operational parameters and thresholds for automated rollback.
type CircuitBreakerConfig struct {
	MaxErrorRate          float64 `json:"max_error_rate"`             // Default: 0.01 (1%)
	MaxFallbackRate       float64 `json:"max_fallback_rate"`          // Default: 0.01 (1%)
	MaxP95LatencyMs       float64 `json:"max_p95_latency_ms"`         // Default: 15.0 ms
	MaxP99LatencyMs       float64 `json:"max_p99_latency_ms"`         // Default: 25.0 ms
	MaxDecisionChangeRate float64 `json:"max_decision_change_rate"`   // Default: 0.10 (10%)
	MinSampleCount        int     `json:"min_sample_count"`           // Default: 10
	FailureWindow         int     `json:"failure_window"`             // Default: 3 consecutive breaches
	CooldownSeconds       int     `json:"cooldown_seconds"`           // Default: 300 seconds (5 min)
}

// DefaultCircuitBreakerConfig returns conservative production safety thresholds.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxErrorRate:          0.01,
		MaxFallbackRate:       0.01,
		MaxP95LatencyMs:       15.0,
		MaxP99LatencyMs:       25.0,
		MaxDecisionChangeRate: 0.10,
		MinSampleCount:        10,
		FailureWindow:         3,
		CooldownSeconds:       300,
	}
}

// CircuitBreaker automatically monitors rollout metrics and trips to 0% on sustained degradation.
type CircuitBreaker struct {
	config             CircuitBreakerConfig
	mu                 sync.RWMutex
	state              CircuitBreakerState
	consecutiveFailures int
	lastTrippedAt      time.Time
	lastTripReason     string
	trippedCount       int
}

// NewCircuitBreaker initializes an automated rollback circuit breaker.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
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
	if cfg.MinSampleCount <= 0 {
		cfg.MinSampleCount = 10
	}
	if cfg.FailureWindow <= 0 {
		cfg.FailureWindow = 3
	}
	if cfg.CooldownSeconds <= 0 {
		cfg.CooldownSeconds = 300
	}

	return &CircuitBreaker{
		config: cfg,
		state:  CircuitStateHealthy,
	}
}

// EvaluateAndCheckTrip evaluates the point-in-time metrics snapshot against thresholds.
// Returns tripped=true and the reason if sustained failures breach the failure window.
func (cb *CircuitBreaker) EvaluateAndCheckTrip(snapshot map[string]interface{}) (tripped bool, state CircuitBreakerState, reason string, violations []string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	violations = make([]string, 0)

	// If already in ROLLED_BACK state, check if cooldown has elapsed
	if cb.state == CircuitStateRolledBack {
		if time.Since(cb.lastTrippedAt) < time.Duration(cb.config.CooldownSeconds)*time.Second {
			return false, CircuitStateRolledBack, cb.lastTripReason, violations
		}
	}

	candidateTotal, _ := snapshot["candidate_requests_total"].(int64)
	if candidateTotal < int64(cb.config.MinSampleCount) {
		// Not enough samples to reliably trip the circuit breaker
		return false, cb.state, "", violations
	}

	errorRate, _ := snapshot["candidate_error_rate"].(float64)
	fallbackRate, _ := snapshot["candidate_fallback_rate"].(float64)
	p95, _ := snapshot["candidate_p95_latency_ms"].(float64)
	p99, _ := snapshot["candidate_p99_latency_ms"].(float64)

	isBreached := false

	if errorRate > cb.config.MaxErrorRate {
		violations = append(violations, fmt.Sprintf("Error rate %.2f%% > max %.2f%%", errorRate*100, cb.config.MaxErrorRate*100))
		isBreached = true
	}
	if fallbackRate > cb.config.MaxFallbackRate {
		violations = append(violations, fmt.Sprintf("Fallback rate %.2f%% > max %.2f%%", fallbackRate*100, cb.config.MaxFallbackRate*100))
		isBreached = true
	}
	if p95 > cb.config.MaxP95LatencyMs {
		violations = append(violations, fmt.Sprintf("p95 latency %.2fms > max %.2fms", p95, cb.config.MaxP95LatencyMs))
		isBreached = true
	}
	if p99 > cb.config.MaxP99LatencyMs {
		violations = append(violations, fmt.Sprintf("p99 latency %.2fms > max %.2fms", p99, cb.config.MaxP99LatencyMs))
		isBreached = true
	}

	if isBreached {
		cb.consecutiveFailures++
		if cb.consecutiveFailures >= cb.config.FailureWindow {
			cb.state = CircuitStateRolledBack
			cb.lastTrippedAt = time.Now().UTC()
			cb.trippedCount++
			cb.lastTripReason = fmt.Sprintf("Sustained safety breach (%d consecutive evaluations): %v",
				cb.consecutiveFailures, violations)
			log.Printf("AUTOMATIC CIRCUIT BREAKER TRIPPED: %s", cb.lastTripReason)
			return true, CircuitStateRolledBack, cb.lastTripReason, violations
		}
		cb.state = CircuitStateFailed
		return false, CircuitStateFailed, fmt.Sprintf("Breach detected (%d/%d): %v", cb.consecutiveFailures, cb.config.FailureWindow, violations), violations
	}

	// Healthy evaluation resets consecutive failures
	cb.consecutiveFailures = 0
	cb.state = CircuitStateHealthy
	return false, CircuitStateHealthy, "", violations
}

// Trip forces the circuit breaker into ROLLED_BACK state with a specific reason.
func (cb *CircuitBreaker) Trip(reason string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitStateRolledBack
	cb.lastTrippedAt = time.Now().UTC()
	cb.lastTripReason = reason
	cb.trippedCount++
}

// Reset clears the circuit breaker state to Healthy (e.g. after manual administrative hot-reload).
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitStateHealthy
	cb.consecutiveFailures = 0
	cb.lastTripReason = ""
}

// GetState returns the current operational state of the circuit breaker.
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStatus returns the current circuit breaker status dictionary.
func (cb *CircuitBreaker) GetStatus() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	inCooldown := false
	var remainingCooldownSec float64
	if cb.state == CircuitStateRolledBack {
		elapsed := time.Since(cb.lastTrippedAt).Seconds()
		cooldownTotal := float64(cb.config.CooldownSeconds)
		if elapsed < cooldownTotal {
			inCooldown = true
			remainingCooldownSec = cooldownTotal - elapsed
		}
	}

	return map[string]interface{}{
		"state":                  string(cb.state),
		"consecutive_failures":   cb.consecutiveFailures,
		"failure_window":         cb.config.FailureWindow,
		"min_sample_count":       cb.config.MinSampleCount,
		"tripped_count":          cb.trippedCount,
		"last_trip_reason":       cb.lastTripReason,
		"last_tripped_at":        cb.lastTrippedAt.Format(time.RFC3339),
		"in_cooldown":            inCooldown,
		"remaining_cooldown_sec": int(remainingCooldownSec),
	}
}

// RunSelfTest executes an isolated self-test suite validating state transitions, trips, and resets.
func (cb *CircuitBreaker) RunSelfTest() error {
	testCfg := CircuitBreakerConfig{
		MaxErrorRate:    0.05,
		MaxP95LatencyMs: 50.0,
		MinSampleCount:  10,
		FailureWindow:   2,
		CooldownSeconds: 60,
	}
	testCB := NewCircuitBreaker(testCfg)

	// 1. Initial State
	if testCB.GetState() != CircuitStateHealthy {
		return fmt.Errorf("self-test failed: expected initial state HEALTHY, got %s", testCB.GetState())
	}

	// 2. Normal evaluation does not trip
	healthySnap := map[string]interface{}{
		"candidate_requests_total": int64(100),
		"candidate_error_rate":     0.0,
		"candidate_p95_latency_ms": 10.0,
	}
	tripped, st, _, _ := testCB.EvaluateAndCheckTrip(healthySnap)
	if tripped || st != CircuitStateHealthy {
		return fmt.Errorf("self-test failed: healthy evaluation caused trip")
	}

	// 3. Breaching evaluation increments failure counter
	failingSnap := map[string]interface{}{
		"candidate_requests_total": int64(100),
		"candidate_error_rate":     0.50, // Breached
		"candidate_p95_latency_ms": 10.0,
	}
	tripped, st, _, _ = testCB.EvaluateAndCheckTrip(failingSnap)
	if tripped || st != CircuitStateFailed {
		return fmt.Errorf("self-test failed: first failure did not enter FAILED state (got %s)", st)
	}

	// 4. Second breach trips circuit breaker
	tripped, st, reason, _ := testCB.EvaluateAndCheckTrip(failingSnap)
	if !tripped || st != CircuitStateRolledBack {
		return fmt.Errorf("self-test failed: second consecutive failure did not trip (got %s, tripped=%v)", st, tripped)
	}
	if reason == "" {
		return fmt.Errorf("self-test failed: trip reason empty")
	}

	// 5. Reset restores healthy state
	testCB.Reset()
	if testCB.GetState() != CircuitStateHealthy {
		return fmt.Errorf("self-test failed: Reset did not restore HEALTHY state")
	}

	return nil
}
