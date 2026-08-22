package riskengine

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// SecurityAuditAction constants representing all privileged operations.
const (
	ActionAuthFailure          = "AUTH_FAILURE"
	ActionPrivilegedAPICall    = "PRIVILEGED_API_CALL"
	ActionModelPromoted        = "MODEL_PROMOTED"
	ActionModelRollback        = "MODEL_ROLLBACK"
	ActionCandidateApproved    = "CANDIDATE_APPROVED"
	ActionCandidateRejected    = "CANDIDATE_REJECTED"
	ActionRetrainingTriggered  = "RETRAINING_TRIGGERED"
	ActionMaintenanceEnabled   = "MAINTENANCE_ENABLED"
	ActionMaintenanceDisabled  = "MAINTENANCE_DISABLED"
	ActionModelFrozen          = "MODEL_FROZEN"
	ActionModelUnfrozen        = "MODEL_UNFROZEN"
	ActionRecoveryTriggered    = "RECOVERY_TRIGGERED"
	ActionCanaryUpdated        = "CANARY_UPDATED"
	ActionCircuitBreakerTrip   = "CIRCUIT_BREAKER_TRIPPED"
	ActionInvariantBreach      = "INVARIANT_BREACH"
)

// SecurityAuditEvent models an immutable, PII-sanitized audit log entry.
type SecurityAuditEvent struct {
	EventID       string            `json:"event_id"`
	Timestamp     time.Time         `json:"timestamp"`
	Actor         string            `json:"actor"`
	Role          Role              `json:"role"`
	Action        string            `json:"action"`
	Resource      string            `json:"resource"`
	Result        string            `json:"result"` // "SUCCESS", "DENIED", "FAILED"
	CorrelationID string            `json:"correlation_id"`
	Reason        string            `json:"reason,omitempty"`
	ClientIP      string            `json:"client_ip,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// SecurityAuditLogger manages structured audit event recording and forensic buffer retention.
type SecurityAuditLogger struct {
	mu           sync.RWMutex
	recentEvents []SecurityAuditEvent
	maxRetention int
	outbox       *DurableOutbox
}

// NewSecurityAuditLogger initializes a thread-safe audit logger with bounded circular retention.
func NewSecurityAuditLogger(maxRetention int, outbox *DurableOutbox) *SecurityAuditLogger {
	if maxRetention <= 0 {
		maxRetention = 1000
	}
	return &SecurityAuditLogger{
		recentEvents: make([]SecurityAuditEvent, 0, maxRetention),
		maxRetention: maxRetention,
		outbox:       outbox,
	}
}

// LogEvent records a sanitized audit event to memory and optionally enqueues to the durable outbox.
func (l *SecurityAuditLogger) LogEvent(evt SecurityAuditEvent) {
	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now().UTC()
	}
	if evt.EventID == "" {
		evt.EventID = fmt.Sprintf("audit_%d_%s", evt.Timestamp.UnixNano(), evt.Action)
	}

	// Sanitize metadata to guarantee NO secrets or payment PII enter audit logs
	evt.Metadata = sanitizeMetadata(evt.Metadata)

	l.mu.Lock()
	if len(l.recentEvents) >= l.maxRetention {
		// Circular buffer eviction
		l.recentEvents = l.recentEvents[1:]
	}
	l.recentEvents = append(l.recentEvents, evt)
	l.mu.Unlock()

	// Structured stdout logging
	log.Printf("[SECURITY_AUDIT] event_id=%s action=%s actor=%s role=%s result=%s resource=%s reason=%q",
		evt.EventID, evt.Action, evt.Actor, evt.Role, evt.Result, evt.Resource, evt.Reason)

	// Async non-blocking outbox enqueue
	if l.outbox != nil {
		payloadBytes, _ := json.Marshal(evt)
		l.outbox.Enqueue(OutboxEvent{
			EventID:       evt.EventID,
			EventType:     "SECURITY_AUDIT_EVENT",
			Timestamp:     evt.Timestamp,
			CorrelationID: evt.CorrelationID,
			Payload:       payloadBytes,
		})
	}
}

// GetRecentEvents retrieves a slice of recent audit events.
func (l *SecurityAuditLogger) GetRecentEvents(limit int) []SecurityAuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if limit <= 0 || limit > len(l.recentEvents) {
		limit = len(l.recentEvents)
	}

	start := len(l.recentEvents) - limit
	res := make([]SecurityAuditEvent, limit)
	copy(res, l.recentEvents[start:])
	return res
}

// sanitizeMetadata strips any accidental sensitive credentials or card details.
func sanitizeMetadata(m map[string]string) map[string]string {
	if m == nil {
		return make(map[string]string)
	}
	clean := make(map[string]string, len(m))
	sensitiveSubstrings := []string{"key", "secret", "password", "token", "cvv", "pan", "card", "auth"}

	for k, v := range m {
		kLower := strings.ToLower(k)
		isSensitive := false
		for _, s := range sensitiveSubstrings {
			if strings.Contains(kLower, s) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			clean[k] = "[REDACTED]"
		} else {
			clean[k] = v
		}
	}
	return clean
}
