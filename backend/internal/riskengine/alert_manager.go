package riskengine

import (
	"context"
	"log"
	"sync"
	"time"
)

// Alert represents an operational alert notification emitted to sinks.
type Alert struct {
	AlertID       string           `json:"alert_id"`
	Timestamp     time.Time        `json:"timestamp"`
	Severity      IncidentSeverity `json:"severity"`
	Title         string           `json:"title"`
	Message       string           `json:"message"`
	IncidentID    string           `json:"incident_id,omitempty"`
	Subsystem     string           `json:"subsystem"`
	CorrelationID string           `json:"correlation_id,omitempty"`
}

// AlertSink defines the interface for downstream notification targets.
type AlertSink interface {
	Emit(ctx context.Context, alert Alert) error
}

// LogAlertSink outputs alerts as structured log messages.
type LogAlertSink struct{}

func (s *LogAlertSink) Emit(ctx context.Context, alert Alert) error {
	log.Printf("[ALERT] [%s] %s | Subsystem: %s | Incident: %s | Message: %s",
		alert.Severity, alert.Title, alert.Subsystem, alert.IncidentID, alert.Message)
	return nil
}

// InMemoryAlertSink stores recent alerts in a bounded memory buffer for inspection.
type InMemoryAlertSink struct {
	mu      sync.RWMutex
	alerts  []Alert
	maxSize int
}

// NewInMemoryAlertSink creates an in-memory alert buffer.
func NewInMemoryAlertSink(maxSize int) *InMemoryAlertSink {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &InMemoryAlertSink{
		alerts:  make([]Alert, 0, maxSize),
		maxSize: maxSize,
	}
}

func (s *InMemoryAlertSink) Emit(ctx context.Context, alert Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.alerts = append(s.alerts, alert)
	if len(s.alerts) > s.maxSize {
		s.alerts = s.alerts[len(s.alerts)-s.maxSize:]
	}
	return nil
}

// GetRecentAlerts returns a copy of stored alerts in reverse order.
func (s *InMemoryAlertSink) GetRecentAlerts() []Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]Alert, len(s.alerts))
	for i, idx := 0, len(s.alerts)-1; idx >= 0; i, idx = i+1, idx-1 {
		res[i] = s.alerts[idx]
	}
	return res
}

// AlertManager dispatches alerts asynchronously to registered sinks.
type AlertManager struct {
	sinks    []AlertSink
	queue    chan Alert
	stopCh   chan struct{}
	wg       sync.WaitGroup
	closed   bool
	mu       sync.Mutex
}

// NewAlertManager initializes an asynchronous non-blocking alert manager.
func NewAlertManager(sinks ...AlertSink) *AlertManager {
	if len(sinks) == 0 {
		sinks = []AlertSink{&LogAlertSink{}}
	}

	mgr := &AlertManager{
		sinks:  sinks,
		queue:  make(chan Alert, 500),
		stopCh: make(chan struct{}),
	}

	mgr.wg.Add(1)
	go mgr.worker()
	return mgr
}

// Emit enqueues an alert for asynchronous non-blocking delivery.
func (am *AlertManager) Emit(alert Alert) {
	am.mu.Lock()
	if am.closed {
		am.mu.Unlock()
		return
	}
	am.mu.Unlock()

	select {
	case am.queue <- alert:
	default:
		// Drop if buffer is full under extreme overload to protect memory and inference
		log.Printf("[ALERT_DROP] Alert queue full, dropping alert: %s", alert.Title)
	}
}

func (am *AlertManager) worker() {
	defer am.wg.Done()
	ctx := context.Background()

	for {
		select {
		case alert, ok := <-am.queue:
			if !ok {
				return
			}
			for _, sink := range am.sinks {
				_ = sink.Emit(ctx, alert)
			}
		case <-am.stopCh:
			// Drain remaining alerts
			for len(am.queue) > 0 {
				alert := <-am.queue
				for _, sink := range am.sinks {
					_ = sink.Emit(ctx, alert)
				}
			}
			return
		}
	}
}

// Close gracefully stops the alert manager.
func (am *AlertManager) Close() {
	am.mu.Lock()
	if am.closed {
		am.mu.Unlock()
		return
	}
	am.closed = true
	am.mu.Unlock()

	close(am.stopCh)
	am.wg.Wait()
}
