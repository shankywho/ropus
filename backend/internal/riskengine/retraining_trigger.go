package riskengine

import (
	"fmt"
	"sync"
	"time"
)

// RetrainingTriggerEngine determines when drift conditions warrant triggering an automated retraining job.
type RetrainingTriggerEngine struct {
	mu                         sync.RWMutex
	config                     RetrainingConfig
	consecutiveDriftCount      int
	lastTriggerTime            time.Time
	lastTriggerReason          string
	lastSuccessfulTrainingTime time.Time
	lastFailureTime            time.Time
	activeJobID                string
}

// NewRetrainingTriggerEngine initializes the trigger evaluation engine.
func NewRetrainingTriggerEngine(cfg RetrainingConfig) *RetrainingTriggerEngine {
	cfg.Validate()
	return &RetrainingTriggerEngine{
		config: cfg,
	}
}

// UpdateConfig dynamically updates the trigger configuration.
func (e *RetrainingTriggerEngine) UpdateConfig(cfg RetrainingConfig) {
	cfg.Validate()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = cfg
}

// EvaluateDrift checks whether a drift measurement qualifies for automated retraining triggering.
func (e *RetrainingTriggerEngine) EvaluateDrift(
	measurement *DriftMeasurement,
	cbState CircuitBreakerState,
) (shouldTrigger bool, triggerType string, triggerReason string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.config.Enabled {
		return false, "NONE", "Retraining subsystem is disabled"
	}

	if measurement == nil {
		return false, "NONE", "No drift measurement available"
	}

	// 1. Check Circuit Breaker health (do not trigger retraining during active emergency rollback)
	if cbState != CircuitStateHealthy && cbState != "" {
		return false, "NONE", fmt.Sprintf("Circuit breaker is not healthy (%s); retraining suppressed", cbState)
	}

	// 2. Check active job concurrency
	if e.activeJobID != "" {
		return false, "NONE", fmt.Sprintf("Retraining job %s is currently active; concurrent triggers rejected", e.activeJobID)
	}

	// 3. Check sample count quorum
	if uint32(measurement.SampleCount) < e.config.MinSamples {
		return false, "NONE", fmt.Sprintf("Sample count (%d) below minimum required quorum (%d)", measurement.SampleCount, e.config.MinSamples)
	}

	// 4. Check cooldown window
	now := time.Now()
	if !e.lastTriggerTime.IsZero() && now.Sub(e.lastTriggerTime) < e.config.CooldownDuration {
		remaining := e.config.CooldownDuration - now.Sub(e.lastTriggerTime)
		return false, "NONE", fmt.Sprintf("In cooldown window; %.1f minutes remaining", remaining.Minutes())
	}

	// 5. Evaluate drift severity and consecutive window persistence
	isDrifted := measurement.MaxPSI >= e.config.DriftThreshold ||
		measurement.OverallStatus == DriftStatusCritical ||
		measurement.OverallStatus == DriftStatusDegraded

	if !isDrifted {
		// Reset consecutive count on healthy traffic
		e.consecutiveDriftCount = 0
		return false, "NONE", "Traffic distribution is stable"
	}

	e.consecutiveDriftCount++

	// Immediate trigger on CRITICAL drift with high PSI
	if measurement.OverallStatus == DriftStatusCritical && measurement.MaxPSI >= 0.30 {
		reason := fmt.Sprintf("CRITICAL drift detected (max PSI: %.4f >= 0.30, critical features: %d)",
			measurement.MaxPSI, measurement.CriticalFeatureCount)
		return true, "DRIFT_CRITICAL", reason
	}

	// Sustained trigger on HIGH / DEGRADED drift across consecutive windows
	if e.consecutiveDriftCount >= e.config.RequiredConsecutiveWindows {
		reason := fmt.Sprintf("Sustained drift across %d consecutive evaluation windows (max PSI: %.4f >= %.2f)",
			e.consecutiveDriftCount, measurement.MaxPSI, e.config.DriftThreshold)
		return true, "DRIFT_SUSTAINED", reason
	}

	return false, "NONE", fmt.Sprintf("Drift observed (%d/%d required consecutive windows)",
		e.consecutiveDriftCount, e.config.RequiredConsecutiveWindows)
}

// SetActiveJob marks a job as active in the trigger engine.
func (e *RetrainingTriggerEngine) SetActiveJob(jobID string, reason string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.activeJobID = jobID
	e.lastTriggerTime = time.Now().UTC()
	e.lastTriggerReason = reason
}

// ClearActiveJob clears active job state upon completion or failure.
func (e *RetrainingTriggerEngine) ClearActiveJob(success bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.activeJobID = ""
	e.consecutiveDriftCount = 0
	now := time.Now().UTC()
	if success {
		e.lastSuccessfulTrainingTime = now
	} else {
		e.lastFailureTime = now
	}
}

// CanTriggerManual evaluates eligibility for an operator-initiated manual retraining trigger.
func (e *RetrainingTriggerEngine) CanTriggerManual(cbState CircuitBreakerState) (bool, string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.config.Enabled {
		return false, "Retraining subsystem is disabled"
	}
	if cbState != CircuitStateHealthy && cbState != "" {
		return false, fmt.Sprintf("Circuit breaker is not healthy (%s)", cbState)
	}
	if e.activeJobID != "" {
		return false, fmt.Sprintf("Retraining job %s is currently active", e.activeJobID)
	}
	return true, ""
}

// GetSummary returns an operational summary for system status.
func (e *RetrainingTriggerEngine) GetSummary(state JobState, candidateModel *string) RetrainingStatusSummary {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var remainingCooldownSec int64 = 0
	if !e.lastTriggerTime.IsZero() {
		elapsed := time.Since(e.lastTriggerTime)
		if elapsed < e.config.CooldownDuration {
			remainingCooldownSec = int64((e.config.CooldownDuration - elapsed).Seconds())
		}
	}

	var activeJobPtr *string
	if e.activeJobID != "" {
		j := e.activeJobID
		activeJobPtr = &j
	}

	lastTriggerStr := "never"
	if !e.lastTriggerTime.IsZero() {
		lastTriggerStr = e.lastTriggerTime.Format(time.RFC3339)
	}

	lastSuccessStr := "never"
	if !e.lastSuccessfulTrainingTime.IsZero() {
		lastSuccessStr = e.lastSuccessfulTrainingTime.Format(time.RFC3339)
	}

	lastFailureStr := "never"
	if !e.lastFailureTime.IsZero() {
		lastFailureStr = e.lastFailureTime.Format(time.RFC3339)
	}

	return RetrainingStatusSummary{
		Enabled:               e.config.Enabled,
		State:                 state,
		ActiveJobID:           activeJobPtr,
		CandidateModel:        candidateModel,
		LastTrigger:           lastTriggerStr,
		LastSuccessfulTrain:   lastSuccessStr,
		LastFailure:           lastFailureStr,
		CooldownRemainingSec:  remainingCooldownSec,
		TrainingAdapterStatus: "LOCAL_EXECUTION_ADAPTER",
	}
}
