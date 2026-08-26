package graphsage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// IngestionMetrics holds health and throughput counters in Go.
type IngestionMetrics struct {
	RealDataConnectivity string `json:"real_data_connectivity"`
	EventsConsumed       int    `json:"events_consumed"`
	EventsSanitized      int    `json:"events_sanitized"`
	EventsRejected       int    `json:"events_rejected"`
	EventsQuarantined    int    `json:"events_quarantined"`
	EventsDuplicated     int    `json:"events_duplicated"`
	EventsOutOfOrder     int    `json:"events_out_of_order"`
	DLQSize              int    `json:"dlq_size"`
}

// PassiveShadowIngestionAdapter safely consumes stream events into the shadow graph store in Go.
type PassiveShadowIngestionAdapter struct {
	mu           sync.RWMutex
	store        *TemporalHeteroGraphStore
	sanitizer    *PrivacySanitizer
	seenEvents   map[string]bool
	metrics      IngestionMetrics
	lastEventTS  time.Time
	isLiveActive bool
}

// NewPassiveShadowIngestionAdapter creates a passive shadow ingestion adapter.
func NewPassiveShadowIngestionAdapter(store *TemporalHeteroGraphStore, isLiveActive bool) *PassiveShadowIngestionAdapter {
	if store == nil {
		store = NewTemporalHeteroGraphStore()
	}
	connStatus := "NOT_CONNECTED"
	if isLiveActive {
		connStatus = "CONNECTED_SHADOW"
	}
	return &PassiveShadowIngestionAdapter{
		store:        store,
		sanitizer:    NewPrivacySanitizer(""),
		seenEvents:   make(map[string]bool),
		isLiveActive: isLiveActive,
		metrics: IngestionMetrics{
			RealDataConnectivity: connStatus,
		},
	}
}

// ConsumeEvent asynchronously processes an incoming stream event.
func (a *PassiveShadowIngestionAdapter) ConsumeEvent(
	ctx context.Context,
	eventID string,
	topic string,
	rawPayload map[string]interface{},
	eventTS time.Time,
) (bool, string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.metrics.EventsConsumed++

	// 1. Deduplication
	if a.seenEvents[eventID] {
		a.metrics.EventsDuplicated++
		return false, "DUPLICATE_DROPPED"
	}
	a.seenEvents[eventID] = true

	// 2. Out-of-order tracking
	if !a.lastEventTS.IsZero() && eventTS.Before(a.lastEventTS) {
		a.metrics.EventsOutOfOrder++
	} else {
		a.lastEventTS = eventTS
	}

	// 3. Privacy Sanitization
	sanitized, violations, isQuarantined := a.sanitizer.SanitizePayload(rawPayload)
	if isQuarantined {
		a.metrics.EventsQuarantined++
		a.metrics.EventsRejected++
		a.metrics.DLQSize++
		return false, fmt.Sprintf("QUARANTINED_PRIVACY: %v", violations)
	}

	// 4. Ingest into graph store
	if topic == "audit.events" {
		empID, _ := sanitized["employee_id"].(string)
		acctID, _ := sanitized["account_id"].(string)
		if empID != "" && acctID != "" {
			a.store.AddEdge(&HeteroEdge{
				ID:        fmt.Sprintf("e_stream_%s", eventID),
				SourceID:  empID,
				TargetID:  acctID,
				Type:      EdgeAccessesAccount,
				Timestamp: eventTS,
			})
		}
	}

	a.metrics.EventsSanitized++
	return true, "INGESTED_TO_SHADOW_GRAPH"
}

// GetMetrics returns snapshot of ingestion metrics.
func (a *PassiveShadowIngestionAdapter) GetMetrics() IngestionMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.metrics
}
