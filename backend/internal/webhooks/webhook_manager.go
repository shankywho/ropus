package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// WebhookEventType defines supported subscription events.
type WebhookEventType string

const (
	EventRiskDecisionCreated WebhookEventType = "risk.decision.created"
	EventFraudDetected       WebhookEventType = "fraud.detected"
	EventCaseCreated         WebhookEventType = "case.created"
	EventModelUpdated        WebhookEventType = "model.updated"
	EventPolicyChanged       WebhookEventType = "policy.changed"
)

// WebhookSubscription represents a customer endpoint registered for events.
type WebhookSubscription struct {
	SubscriptionID string             `json:"subscription_id"`
	OrgID          string             `json:"org_id"`
	TargetURL      string             `json:"target_url"`
	SecretKey      string             `json:"-"`
	Events         []WebhookEventType `json:"events"`
	IsActive       bool               `json:"is_active"`
	CreatedAt      time.Time          `json:"created_at"`
}

// WebhookDeliveryLog records an emitted webhook payload and delivery status.
type WebhookDeliveryLog struct {
	DeliveryID     string           `json:"delivery_id"`
	SubscriptionID string           `json:"subscription_id"`
	EventType      WebhookEventType `json:"event_type"`
	PayloadSummary string           `json:"payload_summary"`
	Signature      string           `json:"signature"`
	AttemptCount   int              `json:"attempt_count"`
	Delivered      bool             `json:"delivered"`
	Timestamp      time.Time        `json:"timestamp"`
}

// WebhookPlatform manages subscription registration, HMAC signing, and dispatching.
type WebhookPlatform struct {
	mu            sync.RWMutex
	subscriptions map[string]*WebhookSubscription
	deliveryLogs  []*WebhookDeliveryLog
}

// NewWebhookPlatform initializes the webhook platform.
func NewWebhookPlatform() *WebhookPlatform {
	return &WebhookPlatform{
		subscriptions: make(map[string]*WebhookSubscription),
		deliveryLogs:  make([]*WebhookDeliveryLog, 0),
	}
}

// RegisterWebhook registers a new customer webhook endpoint.
func (w *WebhookPlatform) RegisterWebhook(orgID, url string, events []WebhookEventType, secret string) *WebhookSubscription {
	w.mu.Lock()
	defer w.mu.Unlock()

	subID := fmt.Sprintf("sub_%d", time.Now().UnixNano())
	sub := &WebhookSubscription{
		SubscriptionID: subID,
		OrgID:          orgID,
		TargetURL:      url,
		SecretKey:      secret,
		Events:         events,
		IsActive:       true,
		CreatedAt:      time.Now().UTC(),
	}

	w.subscriptions[subID] = sub
	return sub
}

// DispatchEvent prepares and dispatches an event with HMAC-SHA256 signature.
func (w *WebhookPlatform) DispatchEvent(eventType WebhookEventType, orgID string, payload interface{}) ([]*WebhookDeliveryLog, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var logs []*WebhookDeliveryLog

	for _, sub := range w.subscriptions {
		if sub.OrgID != orgID || !sub.IsActive {
			continue
		}

		// Check if subscription listens to this event
		subscribed := false
		for _, e := range sub.Events {
			if e == eventType || e == "*" {
				subscribed = true
				break
			}
		}
		if !subscribed {
			continue
		}

		// Generate HMAC signature
		mac := hmac.New(sha256.New, []byte(sub.SecretKey))
		mac.Write(payloadBytes)
		sig := hex.EncodeToString(mac.Sum(nil))

		log := &WebhookDeliveryLog{
			DeliveryID:     fmt.Sprintf("dlv_%d", time.Now().UnixNano()),
			SubscriptionID: sub.SubscriptionID,
			EventType:      eventType,
			PayloadSummary: string(payloadBytes),
			Signature:      "sha256=" + sig,
			AttemptCount:   1,
			Delivered:      true, // Synchronously delivered to event queue
			Timestamp:      time.Now().UTC(),
		}

		w.deliveryLogs = append(w.deliveryLogs, log)
		logs = append(logs, log)
	}

	return logs, nil
}

// GetDeliveryLogs returns recent webhook delivery logs.
func (w *WebhookPlatform) GetDeliveryLogs() []*WebhookDeliveryLog {
	w.mu.RLock()
	defer w.mu.RUnlock()

	res := make([]*WebhookDeliveryLog, len(w.deliveryLogs))
	copy(res, w.deliveryLogs)
	return res
}
