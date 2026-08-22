package streaming

import (
	"fmt"
	"sync"
	"time"
)

// StreamDetectionAlert represents an immediate real-time threat signal detected on the event stream.
type StreamDetectionAlert struct {
	AlertID        string    `json:"alert_id"`
	PatternType    string    `json:"pattern_type"` // "VELOCITY_ATTACK", "CARD_TESTING_BURST", "ACCOUNT_TAKEOVER_WAVE", "MERCHANT_COMPROMISE"
	EntityID       string    `json:"entity_id"`
	Confidence     float64   `json:"confidence"`
	EventCount     int       `json:"event_count"`
	Message        string    `json:"message"`
	DetectedAt     time.Time `json:"detected_at"`
}

// StreamFraudDetector maintains sliding windows across high-velocity entity streams.
type StreamFraudDetector struct {
	mu           sync.RWMutex
	deviceWindow map[string][]time.Time
	ipWindow     map[string][]time.Time
	userLogins   map[string][]time.Time
}

// NewStreamFraudDetector initializes the stream pattern detector.
func NewStreamFraudDetector() *StreamFraudDetector {
	return &StreamFraudDetector{
		deviceWindow: make(map[string][]time.Time),
		ipWindow:     make(map[string][]time.Time),
		userLogins:   make(map[string][]time.Time),
	}
}

// ProcessTransactionEvent evaluates an inbound transaction event in real-time sliding windows (<20ms).
func (d *StreamFraudDetector) ProcessTransactionEvent(event *StreamingEvent) *StreamDetectionAlert {
	if event == nil {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	cutoff := now.Add(-5 * time.Minute)

	// Extract device & IP
	device, _ := event.Payload["device_fingerprint"].(string)
	ip, _ := event.Payload["ip_address"].(string)

	if device != "" {
		// Prune stale timestamps
		var valid []time.Time
		for _, ts := range d.deviceWindow[device] {
			if ts.After(cutoff) {
				valid = append(valid, ts)
			}
		}
		valid = append(valid, now)
		d.deviceWindow[device] = valid

		// Threshold: > 20 events on device within 5 minutes -> VELOCITY_ATTACK
		if len(valid) >= 20 {
			return &StreamDetectionAlert{
				AlertID:     fmt.Sprintf("str_alt_%d_%s", now.UnixNano(), device),
				PatternType: "VELOCITY_ATTACK",
				EntityID:    device,
				Confidence:  0.96,
				EventCount:  len(valid),
				Message:     fmt.Sprintf("High frequency transaction surge: %d events on device %s in 5m", len(valid), device),
				DetectedAt:  now,
			}
		}
	}

	if ip != "" {
		var validIP []time.Time
		for _, ts := range d.ipWindow[ip] {
			if ts.After(cutoff) {
				validIP = append(validIP, ts)
			}
		}
		validIP = append(validIP, now)
		d.ipWindow[ip] = validIP

		if len(validIP) >= 30 {
			return &StreamDetectionAlert{
				AlertID:     fmt.Sprintf("str_alt_%d_%s", now.UnixNano(), ip),
				PatternType: "CARD_TESTING_BURST",
				EntityID:    ip,
				Confidence:  0.98,
				EventCount:  len(validIP),
				Message:     fmt.Sprintf("Distributed card testing burst from IP %s (%d events)", ip, len(validIP)),
				DetectedAt:  now,
			}
		}
	}

	return nil
}
