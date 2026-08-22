package resilience

import (
	"context"
	"sync"
	"time"
)

// ServiceHealthStatus represents individual subsystem status.
type ServiceHealthStatus struct {
	ServiceName string        `json:"service_name"`
	IsHealthy   bool          `json:"is_healthy"`
	Latency     time.Duration `json:"latency"`
	LastError   string        `json:"last_error,omitempty"`
	LastChecked time.Time     `json:"last_checked"`
}

// SystemHealthReport aggregates all dependent services.
type SystemHealthReport struct {
	OverallStatus string                         `json:"overall_status"` // "HEALTHY", "DEGRADED", "CRITICAL"
	Services      map[string]ServiceHealthStatus `json:"services"`
	CheckedAt     time.Time                      `json:"checked_at"`
}

// HealthCheckFn is a function that pings a service.
type HealthCheckFn func(ctx context.Context) error

// HealthManager monitors multi-component system reliability.
type HealthManager struct {
	mu           sync.RWMutex
	checkers     map[string]HealthCheckFn
	statuses     map[string]ServiceHealthStatus
	fallbackBuffer []map[string]interface{}
}

// NewHealthManager initializes the health manager.
func NewHealthManager() *HealthManager {
	return &HealthManager{
		checkers:       make(map[string]HealthCheckFn),
		statuses:       make(map[string]ServiceHealthStatus),
		fallbackBuffer: make([]map[string]interface{}, 0),
	}
}

// RegisterChecker registers a ping check for a subsystem.
func (h *HealthManager) RegisterChecker(serviceName string, fn HealthCheckFn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[serviceName] = fn
}

// RunHealthChecks executes all registered health checks.
func (h *HealthManager) RunHealthChecks(ctx context.Context) *SystemHealthReport {
	h.mu.Lock()
	defer h.mu.Unlock()

	overall := "HEALTHY"
	now := time.Now().UTC()

	for name, fn := range h.checkers {
		start := time.Now()
		err := fn(ctx)
		latency := time.Since(start)

		status := ServiceHealthStatus{
			ServiceName: name,
			IsHealthy:   err == nil,
			Latency:     latency,
			LastChecked: now,
		}

		if err != nil {
			status.LastError = err.Error()
			overall = "DEGRADED"
		}

		h.statuses[name] = status
	}

	report := &SystemHealthReport{
		OverallStatus: overall,
		Services:      make(map[string]ServiceHealthStatus),
		CheckedAt:     now,
	}

	for k, v := range h.statuses {
		report.Services[k] = v
	}

	return report
}

// BufferFallbackEvent stores events locally when Kafka/Streaming is degraded.
func (h *HealthManager) BufferFallbackEvent(event map[string]interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.fallbackBuffer = append(h.fallbackBuffer, event)
}

// FlushFallbackBuffer retrieves buffered events after recovery.
func (h *HealthManager) FlushFallbackBuffer() []map[string]interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()
	flushed := make([]map[string]interface{}, len(h.fallbackBuffer))
	copy(flushed, h.fallbackBuffer)
	h.fallbackBuffer = make([]map[string]interface{}, 0)
	return flushed
}
