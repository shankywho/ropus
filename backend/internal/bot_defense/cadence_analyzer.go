package bot_defense

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// CadenceTracker keeps historical sliding-window arrival timestamps for a client key.
type CadenceTracker struct {
	timestamps []time.Time
}

// CadenceAnalyzer detects robotic timing, deterministic loops, and burst pacing.
type CadenceAnalyzer struct {
	mu           sync.Mutex
	history      map[string]*CadenceTracker
	maxWindowLen int
}

// NewCadenceAnalyzer initializes the cadence and entropy analyzer.
func NewCadenceAnalyzer(maxWindowLen int) *CadenceAnalyzer {
	if maxWindowLen <= 0 {
		maxWindowLen = 12
	}
	return &CadenceAnalyzer{
		history:      make(map[string]*CadenceTracker),
		maxWindowLen: maxWindowLen,
	}
}

// AnalyzeCadence records the arrival of a request and calculates statistical timing metrics.
func (a *CadenceAnalyzer) AnalyzeCadence(entityKey string, arrivalTime time.Time, userAgent string) (signals BotSignals, rules []string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if arrivalTime.IsZero() {
		arrivalTime = time.Now().UTC()
	}

	// 1. Inspect User-Agent for known automation signatures (Headless browsers, cURL, Python)
	uaLower := strings.ToLower(userAgent)
	if strings.Contains(uaLower, "headlesschrome") ||
		strings.Contains(uaLower, "phantomjs") ||
		strings.Contains(uaLower, "selenium") ||
		strings.Contains(uaLower, "puppeteer") ||
		strings.Contains(uaLower, "playwright") ||
		(strings.HasPrefix(uaLower, "python-requests") && !strings.Contains(uaLower, "merchant-sdk")) {
		signals.IsHeadlessUA = true
		rules = append(rules, fmt.Sprintf("Client User-Agent signature matches headless automation framework (%s)", userAgent))
	}

	tracker, exists := a.history[entityKey]
	if !exists {
		a.history[entityKey] = &CadenceTracker{
			timestamps: []time.Time{arrivalTime},
		}
		signals.CoefficientOfVariation = 1.0 // Initial neutral value
		return signals, rules
	}

	// Append timestamp and maintain fixed sliding window
	tracker.timestamps = append(tracker.timestamps, arrivalTime)
	if len(tracker.timestamps) > a.maxWindowLen {
		tracker.timestamps = tracker.timestamps[len(tracker.timestamps)-a.maxWindowLen:]
	}

	n := len(tracker.timestamps)
	if n < 3 {
		signals.CoefficientOfVariation = 1.0
		return signals, rules
	}

	// Calculate Inter-Arrival Times (IAT) in milliseconds
	var iats []float64
	var sum float64
	burstCount := 0

	for i := 1; i < n; i++ {
		diffMs := float64(tracker.timestamps[i].Sub(tracker.timestamps[i-1]).Microseconds()) / 1000.0
		if diffMs < 0 {
			diffMs = 0 // Timestamp monotonicity guard
		}
		iats = append(iats, diffMs)
		sum += diffMs
		if diffMs < 200.0 { // Sub-200ms consecutive requests
			burstCount++
		}
	}

	mean := sum / float64(len(iats))
	signals.InterArrivalTimeMeanMs = mean

	// Calculate Standard Deviation
	var varianceSum float64
	for _, diff := range iats {
		varianceSum += math.Pow(diff-mean, 2)
	}
	stdDev := math.Sqrt(varianceSum / float64(len(iats)))
	signals.InterArrivalTimeStdDev = stdDev

	// Coefficient of Variation (CV = StdDev / Mean)
	cv := 1.0
	if mean > 0 {
		cv = stdDev / mean
	}
	signals.CoefficientOfVariation = cv

	// Instantaneous Cadence (Hz)
	if mean > 0 {
		signals.RequestCadenceHz = 1000.0 / mean
	}

	// Rule 1: Deterministic Timing / Robotic Pacing
	// If standard deviation is extremely low relative to mean interval (CV < 0.15 with >= 4 requests)
	if n >= 4 && mean < 5000.0 && cv < 0.15 {
		signals.IsDeterministicCadence = true
		rules = append(rules, fmt.Sprintf("Deterministic inter-arrival cadence detected (mean: %.1fms, CV: %.3f < 0.15)", mean, cv))
	}

	// Rule 2: Burst Pacing (Rapid consecutive programmatic calls)
	if burstCount >= 3 {
		signals.IsBurstPacing = true
		rules = append(rules, fmt.Sprintf("High-frequency burst pacing detected (%d sub-200ms inter-arrival intervals)", burstCount))
	}

	return signals, rules
}
