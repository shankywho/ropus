package riskengine

import (
	"testing"
)

func TestCircuitBreaker_HealthyState(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxErrorRate:    0.05,
		MaxFallbackRate: 0.05,
		MinSampleCount:  10,
		FailureWindow:   3,
		CooldownSeconds: 60,
	}
	cb := NewCircuitBreaker(cfg)

	snapshot := map[string]interface{}{
		"candidate_requests_total": int64(20),
		"candidate_error_rate":     0.0,
		"candidate_fallback_rate":  0.0,
		"candidate_p95_latency_ms": 4.0,
		"candidate_p99_latency_ms": 8.0,
	}

	tripped, state, reason, violations := cb.EvaluateAndCheckTrip(snapshot)
	if tripped {
		t.Errorf("Expected tripped=false for healthy metrics, got true (%s)", reason)
	}
	if state != CircuitStateHealthy {
		t.Errorf("Expected state HEALTHY, got %v", state)
	}
	if len(violations) > 0 {
		t.Errorf("Expected 0 violations, got %v", violations)
	}
}

func TestCircuitBreaker_SustainedFailureTrip(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxErrorRate:    0.01, // 1%
		MaxFallbackRate: 0.01,
		MinSampleCount:  10,
		FailureWindow:   3, // 3 consecutive failures
		CooldownSeconds: 60,
	}
	cb := NewCircuitBreaker(cfg)

	// High fallback rate snapshot (20% > 1%)
	breachedSnapshot := map[string]interface{}{
		"candidate_requests_total": int64(25),
		"candidate_error_rate":     0.20,
		"candidate_fallback_rate":  0.20,
		"candidate_p95_latency_ms": 5.0,
		"candidate_p99_latency_ms": 10.0,
	}

	// 1st breach: state becomes FAILED, tripped=false
	tripped1, state1, _, _ := cb.EvaluateAndCheckTrip(breachedSnapshot)
	if tripped1 || state1 != CircuitStateFailed {
		t.Errorf("Iteration 1: Expected tripped=false, state=FAILED, got tripped=%v, state=%v", tripped1, state1)
	}

	// 2nd breach: still pending failure window
	tripped2, state2, _, _ := cb.EvaluateAndCheckTrip(breachedSnapshot)
	if tripped2 || state2 != CircuitStateFailed {
		t.Errorf("Iteration 2: Expected tripped=false, state=FAILED, got tripped=%v, state=%v", tripped2, state2)
	}

	// 3rd breach: REACHES FAILURE WINDOW -> TRIPS to ROLLED_BACK!
	tripped3, state3, reason3, violations3 := cb.EvaluateAndCheckTrip(breachedSnapshot)
	if !tripped3 {
		t.Errorf("Iteration 3: Expected circuit breaker to TRIP on 3rd consecutive breach")
	}
	if state3 != CircuitStateRolledBack {
		t.Errorf("Iteration 3: Expected state ROLLED_BACK, got %v", state3)
	}
	if len(violations3) == 0 || reason3 == "" {
		t.Errorf("Iteration 3: Expected violations and reason to be populated")
	}

	// Check status reflects ROLLED_BACK and in_cooldown=true
	status := cb.GetStatus()
	if status["state"] != string(CircuitStateRolledBack) {
		t.Errorf("Expected status state ROLLED_BACK, got %v", status["state"])
	}
	if status["in_cooldown"] != true {
		t.Errorf("Expected in_cooldown=true, got %v", status["in_cooldown"])
	}
}

func TestCircuitBreaker_IntermittentRecovery(t *testing.T) {
	cfg := CircuitBreakerConfig{
		MaxErrorRate:   0.01,
		MinSampleCount: 10,
		FailureWindow:  3,
	}
	cb := NewCircuitBreaker(cfg)

	breached := map[string]interface{}{
		"candidate_requests_total": int64(20),
		"candidate_error_rate":     0.15,
		"candidate_fallback_rate":  0.15,
		"candidate_p95_latency_ms": 5.0,
		"candidate_p99_latency_ms": 10.0,
	}
	healthy := map[string]interface{}{
		"candidate_requests_total": int64(20),
		"candidate_error_rate":     0.0,
		"candidate_fallback_rate":  0.0,
		"candidate_p95_latency_ms": 5.0,
		"candidate_p99_latency_ms": 10.0,
	}

	// 2 breaches followed by 1 healthy evaluation resets failure counter
	cb.EvaluateAndCheckTrip(breached)
	cb.EvaluateAndCheckTrip(breached)
	tripped, state, _, _ := cb.EvaluateAndCheckTrip(healthy)

	if tripped || state != CircuitStateHealthy {
		t.Errorf("Expected recovery to HEALTHY, got tripped=%v, state=%v", tripped, state)
	}
	if cb.consecutiveFailures != 0 {
		t.Errorf("Expected consecutiveFailures=0 after healthy evaluation, got %d", cb.consecutiveFailures)
	}
}

func TestCanaryRouter_HotReloadAndAdminControl(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:    false,
		Percentage: 0,
	}
	router := NewCanaryRouter(cfg, nil)

	// Initial status is disabled 0%
	if router.Route("tenant_admin", "tx_01") != RouteLegacy {
		t.Errorf("Expected RouteLegacy at initial 0%%")
	}

	// Hot-reload to 25% via UpdateConfig
	err := router.UpdateConfig(true, 25, "admin_user_01", "Stage 4 promotion")
	if err != nil {
		t.Fatalf("Failed to hot-reload config: %v", err)
	}

	status := router.GetStatus()
	if status["enabled"] != true {
		t.Errorf("Expected enabled=true after hot-reload, got %v", status["enabled"])
	}
	if status["target_percentage"] != 25 {
		t.Errorf("Expected target_percentage=25 after hot-reload, got %v", status["target_percentage"])
	}

	// Invalid percentage validation
	if err := router.UpdateConfig(true, -5, "admin", "test"); err == nil {
		t.Errorf("Expected error for negative percentage, got nil")
	}
	if err := router.UpdateConfig(true, 105, "admin", "test"); err == nil {
		t.Errorf("Expected error for percentage > 100, got nil")
	}
}

func TestCanaryRouter_AutoRollbackOnSustainedBreach(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:         true,
		Percentage:      50,
		MaxFallbackRate: 0.01,
	}
	router := NewCanaryRouter(cfg, nil)

	// Simulate 12 requests with 6 fallbacks (50% fallback rate > 1%)
	for i := 0; i < 6; i++ {
		router.RecordCandidateRequest()
		router.RecordCandidateSuccess(2.0, "ALLOW_RECOMMENDATION")
	}
	for i := 0; i < 6; i++ {
		router.RecordCandidateRequest()
		router.RecordCandidateFallback(50.0, "connection timeout")
	}

	// Verify circuit breaker tripped and rolled back canary to 0%
	status := router.GetStatus()
	cbStatus := status["circuit_breaker"].(map[string]interface{})

	if cbStatus["state"] != string(CircuitStateRolledBack) {
		t.Errorf("Expected circuit breaker state ROLLED_BACK, got %v", cbStatus["state"])
	}
	if status["target_percentage"] != 0 {
		t.Errorf("Expected target_percentage reset to 0%% after rollback, got %v", status["target_percentage"])
	}
	if status["enabled"] != false {
		t.Errorf("Expected enabled=false after rollback, got %v", status["enabled"])
	}

	// Verify all subsequent routes return RouteLegacy even if percentage was previously 50%
	for i := 0; i < 50; i++ {
		r := router.Route("tenant_test", "tx_post_rollback")
		if r != RouteLegacy {
			t.Fatalf("Expected RouteLegacy post-rollback, got %v", r)
		}
	}
}
