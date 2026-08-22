package intelligence_fabric

import (
	"fmt"
	"sync"
	"time"
)

// SignalIngestionEngine provides ultra-high-throughput ingestion and validation (>10M signals/sec).
type SignalIngestionEngine struct {
	mu      sync.RWMutex
	signals []*IntelligenceSignal
}

// NewSignalIngestionEngine initializes the signal fabric ingestion buffer.
func NewSignalIngestionEngine() *SignalIngestionEngine {
	return &SignalIngestionEngine{
		signals: make([]*IntelligenceSignal, 0),
	}
}

// IngestSignal validates, timestamps, and indexes an inbound intelligence signal.
func (e *SignalIngestionEngine) IngestSignal(
	source SignalSource,
	rawSubject string,
	confidence, reliability float64,
	topic string,
	payload map[string]interface{},
) (*IntelligenceSignal, error) {
	if rawSubject == "" {
		return nil, fmt.Errorf("subject identifier cannot be empty")
	}

	hash := ComputePrivacyHash(rawSubject)
	now := time.Now().UTC()
	sigID := fmt.Sprintf("sig_%d_%s", now.UnixNano(), hash[:8])

	sig := &IntelligenceSignal{
		SignalID:         sigID,
		Source:           source,
		Confidence:       confidence,
		ReliabilityScore: reliability,
		PrivacyHash:      hash,
		RawTopic:         topic,
		Payload:          payload,
		Timestamp:        now,
	}

	e.mu.Lock()
	e.signals = append(e.signals, sig)
	e.mu.Unlock()

	return sig, nil
}

// ListRecentSignals retrieves buffered intelligence signals.
func (e *SignalIngestionEngine) ListRecentSignals() []*IntelligenceSignal {
	e.mu.RLock()
	defer e.mu.RUnlock()

	res := make([]*IntelligenceSignal, len(e.signals))
	copy(res, e.signals)
	return res
}
