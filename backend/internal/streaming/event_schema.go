package streaming

import (
	"time"
)

// StreamEventType defines the categorical classification of a real-time event.
type StreamEventType string

const (
	EventTransactionCreated StreamEventType = "transaction_created"
	EventLoginAttempt        StreamEventType = "login_attempt"
	EventDeviceRegistered    StreamEventType = "device_registered"
	EventAccountCreated      StreamEventType = "account_created"
	EventPaymentFailed       StreamEventType = "payment_failed"
	EventChargebackReceived  StreamEventType = "chargeback_received"
	EventFraudConfirmed      StreamEventType = "fraud_confirmed"
)

// StreamingEvent represents an immutable, ordered event in the streaming intelligence mesh.
type StreamingEvent struct {
	EventID        string                 `json:"event_id"`
	TenantID       string                 `json:"tenant_id"`
	Type           StreamEventType        `json:"type"`
	IdempotencyKey string                 `json:"idempotency_key"`
	Offset         int64                  `json:"offset"`
	Timestamp      time.Time              `json:"timestamp"`
	Payload        map[string]interface{} `json:"payload"`
}
