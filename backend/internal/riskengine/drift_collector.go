package riskengine

import (
	"sync"
)

// DriftCollector manages a thread-safe, bounded memory circular ring buffer of live feature vectors.
type DriftCollector struct {
	mu           sync.RWMutex
	maxCapacity  int
	featureNames []string
	buffers      map[string][]float64
	writeIndices map[string]int
	counts       map[string]int
	totalPushed  int64
}

// NewDriftCollector initializes a bounded memory feature collector.
func NewDriftCollector(maxCapacity int, featureNames []string) *DriftCollector {
	if maxCapacity <= 0 {
		maxCapacity = 10000
	}

	buffers := make(map[string][]float64, len(featureNames))
	writeIndices := make(map[string]int, len(featureNames))
	counts := make(map[string]int, len(featureNames))

	for _, name := range featureNames {
		buffers[name] = make([]float64, maxCapacity)
		writeIndices[name] = 0
		counts[name] = 0
	}

	return &DriftCollector{
		maxCapacity:  maxCapacity,
		featureNames: featureNames,
		buffers:      buffers,
		writeIndices: writeIndices,
		counts:       counts,
	}
}

// PushVector asynchronously stores a single feature vector in the circular buffer with minimal latency.
func (dc *DriftCollector) PushVector(featureMap map[string]float64) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.totalPushed++
	for name, val := range featureMap {
		buf, exists := dc.buffers[name]
		if !exists {
			continue
		}

		idx := dc.writeIndices[name]
		buf[idx] = val
		dc.writeIndices[name] = (idx + 1) % dc.maxCapacity
		if dc.counts[name] < dc.maxCapacity {
			dc.counts[name]++
		}
	}
}

// Snapshot returns a copy of the recent feature values up to the requested window size.
func (dc *DriftCollector) Snapshot(windowSize int) (map[string][]float64, int) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if windowSize <= 0 || windowSize > dc.maxCapacity {
		windowSize = dc.maxCapacity
	}

	result := make(map[string][]float64, len(dc.featureNames))
	minCount := windowSize

	for _, name := range dc.featureNames {
		count := dc.counts[name]
		if count == 0 {
			result[name] = []float64{}
			minCount = 0
			continue
		}

		actualLen := count
		if actualLen > windowSize {
			actualLen = windowSize
		}
		if actualLen < minCount {
			minCount = actualLen
		}

		samples := make([]float64, actualLen)
		buf := dc.buffers[name]
		writeIdx := dc.writeIndices[name]

		// Copy the most recent actualLen values
		startIdx := (writeIdx - actualLen + dc.maxCapacity) % dc.maxCapacity
		for i := 0; i < actualLen; i++ {
			samples[i] = buf[(startIdx+i)%dc.maxCapacity]
		}
		result[name] = samples
	}

	return result, minCount
}

// TotalCollected returns total number of feature vectors pushed since startup.
func (dc *DriftCollector) TotalCollected() int64 {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.totalPushed
}
