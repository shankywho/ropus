package riskengine

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

const (
	defaultSLOReservoirSize = 2000
)

// SLOEngine tracks rolling operational metrics and calculates real-time SLO compliance,
// error budget consumption, and burn rates.
type SLOEngine struct {
	mu sync.RWMutex

	windowDuration time.Duration

	// Request & latency sample buffer
	evalSamples    []evalSample
	evalIdx        int
	evalCount      int
	evalCapacity   int

	// Model divergence samples
	divergenceSamples []bool
	divIdx            int
	divCount          int

	// Drift metric cache
	lastMaxPSI     float64
	lastDriftState DriftStatus
	lastDriftTime  time.Time

	// Retraining outcomes
	retrainingTotal   int64
	retrainingSuccess int64

	// Canary rollbacks
	canaryRollbacks int64
	canaryTotal     int64

	// Dependency check counts
	depChecksTotal   map[string]int64
	depChecksSuccess map[string]int64

	definitions map[string]SLODefinition
}

type evalSample struct {
	timestamp    time.Time
	latencyMs    float64
	isSuccess    bool
	isFallback   bool
	isInferError bool
}

// NewSLOEngine initializes the SLO evaluation engine with standard production thresholds.
func NewSLOEngine(windowDuration time.Duration) *SLOEngine {
	if windowDuration <= 0 {
		windowDuration = 5 * time.Minute
	}

	engine := &SLOEngine{
		windowDuration:    windowDuration,
		evalCapacity:      defaultSLOReservoirSize,
		evalSamples:       make([]evalSample, defaultSLOReservoirSize),
		divergenceSamples: make([]bool, 500),
		depChecksTotal:    make(map[string]int64),
		depChecksSuccess:  make(map[string]int64),
		definitions:       make(map[string]SLODefinition),
	}

	engine.initDefaultDefinitions()
	return engine
}

func (e *SLOEngine) initDefaultDefinitions() {
	e.definitions = map[string]SLODefinition{
		"slo_availability": {
			ID:               "slo_availability",
			Name:             "Risk Evaluation Availability",
			Description:      "Percentage of risk evaluation requests processed successfully",
			Target:           0.999, // 99.9%
			WarningThreshold: 0.9995,
			BreachThreshold:  0.9990,
			WindowDuration:   e.windowDuration,
			IsUpperBound:     false,
			Unit:             "fraction",
		},
		"slo_p95_latency": {
			ID:               "slo_p95_latency",
			Name:             "Risk Evaluation P95 Latency",
			Description:      "95th percentile latency of synchronous risk evaluations",
			Target:           100.0, // <= 100 ms
			WarningThreshold: 75.0,
			BreachThreshold:  100.0,
			WindowDuration:   e.windowDuration,
			IsUpperBound:     true,
			Unit:             "ms",
		},
		"slo_p99_latency": {
			ID:               "slo_p99_latency",
			Name:             "Risk Evaluation P99 Latency",
			Description:      "99th percentile latency of synchronous risk evaluations",
			Target:           250.0, // <= 250 ms
			WarningThreshold: 150.0,
			BreachThreshold:  250.0,
			WindowDuration:   e.windowDuration,
			IsUpperBound:     true,
			Unit:             "ms",
		},
		"slo_inference_error_rate": {
			ID:               "slo_inference_error_rate",
			Name:             "Model Inference Error Rate",
			Description:      "Fraction of model predictions encountering execution errors",
			Target:           0.005, // <= 0.5%
			WarningThreshold: 0.002,
			BreachThreshold:  0.005,
			WindowDuration:   e.windowDuration,
			IsUpperBound:     true,
			Unit:             "fraction",
		},
		"slo_fallback_rate": {
			ID:               "slo_fallback_rate",
			Name:             "Emergency Fallback Execution Rate",
			Description:      "Fraction of transactions routed to heuristic fallback model",
			Target:           0.010, // <= 1.0%
			WarningThreshold: 0.005,
			BreachThreshold:  0.010,
			WindowDuration:   e.windowDuration,
			IsUpperBound:     true,
			Unit:             "fraction",
		},
		"slo_model_decision_divergence": {
			ID:               "slo_model_decision_divergence",
			Name:             "Shadow Model Decision Divergence",
			Description:      "Rate of differing decisions between candidate and production model",
			Target:           0.050, // <= 5.0%
			WarningThreshold: 0.030,
			BreachThreshold:  0.050,
			WindowDuration:   e.windowDuration,
			IsUpperBound:     true,
			Unit:             "fraction",
		},
		"slo_drift_health": {
			ID:               "slo_drift_health",
			Name:             "Feature Distribution Stability (Max PSI)",
			Description:      "Maximum population stability index across 25 production features",
			Target:           0.200, // Max PSI <= 0.20
			WarningThreshold: 0.100,
			BreachThreshold:  0.250,
			WindowDuration:   e.windowDuration,
			IsUpperBound:     true,
			Unit:             "psi",
		},
		"slo_retraining_success_rate": {
			ID:               "slo_retraining_success_rate",
			Name:             "Retraining Pipeline Success Rate",
			Description:      "Percentage of retraining candidate generation jobs completing cleanly",
			Target:           0.950, // >= 95.0%
			WarningThreshold: 0.980,
			BreachThreshold:  0.950,
			WindowDuration:   24 * time.Hour,
			IsUpperBound:     false,
			Unit:             "fraction",
		},
		"slo_canary_rollback_rate": {
			ID:               "slo_canary_rollback_rate",
			Name:             "Canary Rollback Rate",
			Description:      "Fraction of canary rollouts requiring safety gate rollbacks",
			Target:           0.050, // <= 5.0%
			WarningThreshold: 0.020,
			BreachThreshold:  0.050,
			WindowDuration:   24 * time.Hour,
			IsUpperBound:     true,
			Unit:             "fraction",
		},
		"slo_dependency_availability": {
			ID:               "slo_dependency_availability",
			Name:             "Upstream Dependency Availability",
			Description:      "Combined availability across PostgreSQL, Redis, ClickHouse, and ML Runtime",
			Target:           0.995, // >= 99.5%
			WarningThreshold: 0.998,
			BreachThreshold:  0.995,
			WindowDuration:   e.windowDuration,
			IsUpperBound:     false,
			Unit:             "fraction",
		},
	}
}

// RecordEvaluation records a single risk evaluation outcome into the rolling reservoir.
func (e *SLOEngine) RecordEvaluation(latencyMs float64, isSuccess, isFallback, isInferError bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	sample := evalSample{
		timestamp:    time.Now().UTC(),
		latencyMs:    latencyMs,
		isSuccess:    isSuccess,
		isFallback:   isFallback,
		isInferError: isInferError,
	}

	e.evalSamples[e.evalIdx] = sample
	e.evalIdx = (e.evalIdx + 1) % e.evalCapacity
	if e.evalCount < e.evalCapacity {
		e.evalCount++
	}
}

// RecordModelDivergence records whether a shadow prediction diverged from production.
func (e *SLOEngine) RecordModelDivergence(diverged bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cap := len(e.divergenceSamples)
	e.divergenceSamples[e.divIdx] = diverged
	e.divIdx = (e.divIdx + 1) % cap
	if e.divCount < cap {
		e.divCount++
	}
}

// RecordDriftMetrics updates the current drift measurement cache.
func (e *SLOEngine) RecordDriftMetrics(maxPSI float64, state DriftStatus) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.lastMaxPSI = maxPSI
	e.lastDriftState = state
	e.lastDriftTime = time.Now().UTC()
}

// RecordRetrainingOutcome records the completion status of a retraining job.
func (e *SLOEngine) RecordRetrainingOutcome(isSuccess bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.retrainingTotal++
	if isSuccess {
		e.retrainingSuccess++
	}
}

// RecordCanaryRollback records an automated canary rollback event.
func (e *SLOEngine) RecordCanaryRollback() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.canaryRollbacks++
	e.canaryTotal++
}

// RecordCanaryCompleted records a successful canary promotion event.
func (e *SLOEngine) RecordCanaryCompleted() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.canaryTotal++
}

// RecordDependencyCheck records the outcome of a periodic dependency health probe.
func (e *SLOEngine) RecordDependencyCheck(depName string, isHealthy bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.depChecksTotal[depName]++
	if isHealthy {
		e.depChecksSuccess[depName]++
	}
}

// Evaluate computes the real-time SLOSummary across all configured SLOs.
func (e *SLOEngine) Evaluate(now time.Time) SLOSummary {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if now.IsZero() {
		now = time.Now().UTC()
	}

	cutoff := now.Add(-e.windowDuration)

	// Filter valid samples within rolling window
	var validSamples []evalSample
	var successCount int64
	var fallbackCount int64
	var errorCount int64
	latencies := make([]float64, 0, e.evalCount)

	for i := 0; i < e.evalCount; i++ {
		s := e.evalSamples[i]
		if s.timestamp.After(cutoff) {
			validSamples = append(validSamples, s)
			if s.isSuccess {
				successCount++
			}
			if s.isFallback {
				fallbackCount++
			}
			if s.isInferError {
				errorCount++
			}
			latencies = append(latencies, s.latencyMs)
		}
	}

	totalEvals := int64(len(validSamples))

	// 1. Availability
	var avail float64 = 1.0
	if totalEvals > 0 {
		avail = float64(successCount) / float64(totalEvals)
	}

	// 2. Latencies
	var p95Latency float64 = 0.0
	var p99Latency float64 = 0.0
	if len(latencies) > 0 {
		sort.Float64s(latencies)
		p95Latency = percentile(latencies, 0.95)
		p99Latency = percentile(latencies, 0.99)
	}

	// 3. Error Rate
	var inferErrRate float64 = 0.0
	if totalEvals > 0 {
		inferErrRate = float64(errorCount) / float64(totalEvals)
	}

	// 4. Fallback Rate
	var fallbackRate float64 = 0.0
	if totalEvals > 0 {
		fallbackRate = float64(fallbackCount) / float64(totalEvals)
	}

	// 5. Decision Divergence
	var divRate float64 = 0.0
	if e.divCount > 0 {
		var divTrue int64
		for i := 0; i < e.divCount; i++ {
			if e.divergenceSamples[i] {
				divTrue++
			}
		}
		divRate = float64(divTrue) / float64(e.divCount)
	}

	// 6. Drift Max PSI
	maxPSI := e.lastMaxPSI

	// 7. Retraining Success Rate
	var retrainSuccessRate float64 = 1.0
	if e.retrainingTotal > 0 {
		retrainSuccessRate = float64(e.retrainingSuccess) / float64(e.retrainingTotal)
	}

	// 8. Canary Rollback Rate
	var canaryRbRate float64 = 0.0
	if e.canaryTotal > 0 {
		canaryRbRate = float64(e.canaryRollbacks) / float64(e.canaryTotal)
	}

	// 9. Dependency Availability
	var depAvail float64 = 1.0
	var depTotalSum, depSuccSum int64
	for dep, tot := range e.depChecksTotal {
		depTotalSum += tot
		depSuccSum += e.depChecksSuccess[dep]
	}
	if depTotalSum > 0 {
		depAvail = float64(depSuccSum) / float64(depTotalSum)
	}

	// Compute SLOMetricRecords
	measurements := make(map[string]SLOMetricRecord)

	e.evalSingleSLO("slo_availability", avail, totalEvals, totalEvals-successCount, measurements, now)
	e.evalSingleSLO("slo_p95_latency", p95Latency, int64(len(latencies)), 0, measurements, now)
	e.evalSingleSLO("slo_p99_latency", p99Latency, int64(len(latencies)), 0, measurements, now)
	e.evalSingleSLO("slo_inference_error_rate", inferErrRate, totalEvals, errorCount, measurements, now)
	e.evalSingleSLO("slo_fallback_rate", fallbackRate, totalEvals, fallbackCount, measurements, now)
	e.evalSingleSLO("slo_model_decision_divergence", divRate, int64(e.divCount), 0, measurements, now)
	e.evalSingleSLO("slo_drift_health", maxPSI, 1, 0, measurements, now)
	e.evalSingleSLO("slo_retraining_success_rate", retrainSuccessRate, e.retrainingTotal, e.retrainingTotal-e.retrainingSuccess, measurements, now)
	e.evalSingleSLO("slo_canary_rollback_rate", canaryRbRate, e.canaryTotal, e.canaryRollbacks, measurements, now)
	e.evalSingleSLO("slo_dependency_availability", depAvail, depTotalSum, depTotalSum-depSuccSum, measurements, now)

	// Summarize counts
	var healthyCount, warningCount, breachedCount int
	overall := SLOStatusHealthy

	for _, m := range measurements {
		switch m.Status {
		case SLOStatusHealthy:
			healthyCount++
		case SLOStatusWarning:
			warningCount++
			if overall != SLOStatusBreached {
				overall = SLOStatusWarning
			}
		case SLOStatusBreached:
			breachedCount++
			overall = SLOStatusBreached
		}
	}

	return SLOSummary{
		OverallStatus: overall,
		TotalSLOs:     len(measurements),
		HealthyCount:  healthyCount,
		WarningCount:  warningCount,
		BreachedCount: breachedCount,
		CalculatedAt:  now,
		Measurements:  measurements,
	}
}

func (e *SLOEngine) evalSingleSLO(
	id string,
	currentVal float64,
	sampleCount, violationsCount int64,
	dest map[string]SLOMetricRecord,
	now time.Time,
) {
	def, exists := e.definitions[id]
	if !exists {
		return
	}

	status := SLOStatusHealthy
	var violations []string

	if def.IsUpperBound {
		// e.g. Latency, Error Rate (Lower is better)
		if currentVal > def.BreachThreshold {
			status = SLOStatusBreached
			violations = append(violations, fmt.Sprintf("Value %.4f breaches threshold %.4f", currentVal, def.BreachThreshold))
		} else if currentVal > def.WarningThreshold {
			status = SLOStatusWarning
			violations = append(violations, fmt.Sprintf("Value %.4f exceeds warning threshold %.4f", currentVal, def.WarningThreshold))
		}
	} else {
		// e.g. Availability, Success Rate (Higher is better)
		if currentVal < def.BreachThreshold {
			status = SLOStatusBreached
			violations = append(violations, fmt.Sprintf("Value %.4f falls below breach target %.4f", currentVal, def.BreachThreshold))
		} else if currentVal < def.WarningThreshold {
			status = SLOStatusWarning
			violations = append(violations, fmt.Sprintf("Value %.4f is in warning zone below %.4f", currentVal, def.WarningThreshold))
		}
	}

	// Calculate Error Budget & Burn Rate
	var errorBudget float64 = 0.001
	var remainingBudget float64 = 100.0
	var burnRate float64 = 0.0

	if !def.IsUpperBound {
		errorBudget = 1.0 - def.Target
		if errorBudget <= 0 {
			errorBudget = 0.0001
		}
		actualDeficit := 1.0 - currentVal
		if actualDeficit <= 0 {
			remainingBudget = 100.0
			burnRate = 0.0
		} else {
			remainingFraction := 1.0 - (actualDeficit / errorBudget)
			if remainingFraction < 0 {
				remainingFraction = 0
			}
			remainingBudget = remainingFraction * 100.0
			burnRate = actualDeficit / errorBudget
		}
	} else {
		errorBudget = def.Target
		if errorBudget <= 0 {
			errorBudget = 1.0
		}
		if currentVal <= 0 {
			remainingBudget = 100.0
			burnRate = 0.0
		} else {
			remainingFraction := 1.0 - (currentVal / errorBudget)
			if remainingFraction < 0 {
				remainingFraction = 0
			}
			remainingBudget = remainingFraction * 100.0
			burnRate = currentVal / errorBudget
		}
	}

	dest[id] = SLOMetricRecord{
		SLOID:                def.ID,
		Name:                 def.Name,
		CurrentValue:         currentVal,
		Target:               def.Target,
		Status:               status,
		ErrorBudget:          errorBudget,
		ErrorBudgetRemaining: math.Round(remainingBudget*100) / 100,
		BurnRate:             math.Round(burnRate*100) / 100,
		MeasurementWindow:    def.WindowDuration.String(),
		SampleCount:          sampleCount,
		ViolationsCount:      violationsCount,
		CalculatedAt:         now,
		Violations:           violations,
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	idx := int(math.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
