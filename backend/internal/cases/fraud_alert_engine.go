package cases

import (
	"fmt"
	"sync"
	"time"
)

// AlertSeverity defines the notification urgency.
type AlertSeverity string

const (
	AlertLow      AlertSeverity = "LOW"
	AlertMedium   AlertSeverity = "MEDIUM"
	AlertHigh     AlertSeverity = "HIGH"
	AlertCritical AlertSeverity = "CRITICAL"
)

// FraudAlert represents a dispatched security notification.
type FraudAlert struct {
	AlertID   string        `json:"alert_id"`
	CaseID    string        `json:"case_id"`
	Severity  AlertSeverity `json:"severity"`
	Title     string        `json:"title"`
	Message   string        `json:"message"`
	Channels  []string      `json:"channels"` // "WEBHOOK", "SLACK", "EMAIL", "PAGERDUTY"
	Timestamp time.Time     `json:"timestamp"`
}

// FraudAlertEngine routes critical incidents to designated notification channels.
type FraudAlertEngine struct {
	mu     sync.RWMutex
	alerts []*FraudAlert
}

// NewFraudAlertEngine initializes the alert dispatcher.
func NewFraudAlertEngine() *FraudAlertEngine {
	return &FraudAlertEngine{
		alerts: make([]*FraudAlert, 0),
	}
}

// DispatchAlert routes an alert to appropriate channels based on severity.
func (e *FraudAlertEngine) DispatchAlert(caseID, title, message string, severity AlertSeverity) *FraudAlert {
	e.mu.Lock()
	defer e.mu.Unlock()

	var channels []string
	switch severity {
	case AlertCritical:
		channels = []string{"PAGERDUTY", "SLACK", "WEBHOOK", "EMAIL"}
	case AlertHigh:
		channels = []string{"SLACK", "WEBHOOK", "EMAIL"}
	case AlertMedium:
		channels = []string{"SLACK", "WEBHOOK"}
	default:
		channels = []string{"WEBHOOK"}
	}

	alert := &FraudAlert{
		AlertID:   fmt.Sprintf("alt_%d", time.Now().UnixNano()),
		CaseID:    caseID,
		Severity:  severity,
		Title:     title,
		Message:   message,
		Channels:  channels,
		Timestamp: time.Now().UTC(),
	}

	e.alerts = append(e.alerts, alert)
	return alert
}

// ListAlerts returns dispatched alerts.
func (e *FraudAlertEngine) ListAlerts() []*FraudAlert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	res := make([]*FraudAlert, len(e.alerts))
	copy(res, e.alerts)
	return res
}
