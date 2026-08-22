package riskengine

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/shankywho/ropus/backend/internal/audit"
)

//go:embed drift_baseline_25f.json
var embeddedBaseline25F []byte

// DriftDetector orchestrates periodic drift evaluation, feature distribution collection, and ClickHouse persistence.
type DriftDetector struct {
	config            DriftConfig
	baseline          ModelBaseline
	collector         *DriftCollector
	chClient          *audit.ClickHouseClient
	mu                sync.RWMutex
	currentStatus     ModelDriftStatus
	latestMeasurement *DriftMeasurement
	history                []DriftMeasurement
	maxHistory             int
	stopChan               chan struct{}
	onMeasurementEvaluated func(m *DriftMeasurement)
}

// SetOnMeasurementEvaluated registers a callback invoked whenever a drift measurement is completed.
func (dd *DriftDetector) SetOnMeasurementEvaluated(fn func(m *DriftMeasurement)) {
	dd.mu.Lock()
	defer dd.mu.Unlock()
	dd.onMeasurementEvaluated = fn
}

// NewDriftDetector initializes the drift detection subsystem with embedded baseline and collector.
func NewDriftDetector(cfg DriftConfig, chClient *audit.ClickHouseClient) (*DriftDetector, error) {
	if cfg.CalculationInterval <= 0 {
		cfg.CalculationInterval = 5 * time.Minute
	}
	if cfg.MaxWindowSize <= 0 {
		cfg.MaxWindowSize = 10000
	}
	if cfg.MinSamplesForDrift <= 0 {
		cfg.MinSamplesForDrift = 50
	}
	if cfg.PSIWarnThreshold <= 0 {
		cfg.PSIWarnThreshold = 0.10
	}
	if cfg.PSIHighThreshold <= 0 {
		cfg.PSIHighThreshold = 0.20
	}
	if cfg.PSICritThreshold <= 0 {
		cfg.PSICritThreshold = 0.30
	}
	if cfg.JSDWarnThreshold <= 0 {
		cfg.JSDWarnThreshold = 0.05
	}
	if cfg.JSDHighThreshold <= 0 {
		cfg.JSDHighThreshold = 0.10
	}
	if cfg.JSDCritThreshold <= 0 {
		cfg.JSDCritThreshold = 0.15
	}
	if cfg.KLWarnThreshold <= 0 {
		cfg.KLWarnThreshold = 0.10
	}
	if cfg.Epsilon <= 0 {
		cfg.Epsilon = 1e-4
	}

	var baseline ModelBaseline
	if err := json.Unmarshal(embeddedBaseline25F, &baseline); err != nil {
		return nil, fmt.Errorf("failed to parse embedded drift baseline: %w", err)
	}

	featureNames := make([]string, 0, len(baseline.Features))
	for name := range baseline.Features {
		featureNames = append(featureNames, name)
	}

	collector := NewDriftCollector(cfg.MaxWindowSize, featureNames)

	detector := &DriftDetector{
		config:        cfg,
		baseline:      baseline,
		collector:     collector,
		chClient:      chClient,
		currentStatus: DriftStatusHealthy,
		history:       make([]DriftMeasurement, 0, 50),
		maxHistory:    50,
		stopChan:      make(chan struct{}),
	}

	// Persist baseline metadata and BASELINE_CREATED event to ClickHouse on startup
	if chClient != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = chClient.InsertDriftBaseline(ctx, audit.DriftBaselineRecord{
				BaselineID:         baseline.BaselineID,
				ModelVersion:       baseline.ModelVersion,
				FeatureContract:    baseline.FeatureContract,
				CalibrationVersion: baseline.CalibrationVersion,
				DatasetVersion:     baseline.DatasetVersion,
				CreatedAt:          time.Now().UTC(),
				FeatureCount:       uint16(baseline.FeatureCount),
				Metadata:           fmt.Sprintf("source: %s, total_samples: %d", baseline.BaselineSource, baseline.TotalSamples),
			})

			_ = chClient.InsertDriftEvent(ctx, audit.DriftEventRecord{
				EventID:              fmt.Sprintf("evt_baseline_%d", time.Now().UnixNano()),
				Timestamp:            time.Now().UTC(),
				ModelVersion:         baseline.ModelVersion,
				BaselineID:           baseline.BaselineID,
				PreviousStatus:       "NONE",
				NewStatus:            string(DriftStatusHealthy),
				MaxPSI:               0.0,
				MaxJSD:               0.0,
				MaxKL:                0.0,
				AffectedFeatureCount: 0,
				CriticalFeatureCount: 0,
				Trigger:              "BASELINE_CREATED",
				Reason:               fmt.Sprintf("Loaded baseline %s for model %s with %d features", baseline.BaselineID, baseline.ModelVersion, baseline.FeatureCount),
			})
		}()
	}

	return detector, nil
}

// Start launches the background periodic drift detection worker.
func (dd *DriftDetector) Start(ctx context.Context) {
	ticker := time.NewTicker(dd.config.CalculationInterval)
	defer ticker.Stop()

	log.Printf("Drift detection worker started: interval=%v, window_size=%d, baseline=%s (source=%s)",
		dd.config.CalculationInterval, dd.config.MaxWindowSize, dd.baseline.BaselineID, dd.baseline.BaselineSource)

	for {
		select {
		case <-ctx.Done():
			log.Println("Drift detection worker stopping...")
			return
		case <-dd.stopChan:
			log.Println("Drift detection worker stopped.")
			return
		case <-ticker.C:
			dd.EvaluateLiveDrift(ctx)
		}
	}
}

// Stop terminates the background worker loop.
func (dd *DriftDetector) Stop() {
	select {
	case <-dd.stopChan:
		// already closed
	default:
		close(dd.stopChan)
	}
}

// IngestVector asynchronously feeds a live feature vector into the drift collector.
func (dd *DriftDetector) IngestVector(featureMap map[string]float64) {
	if dd != nil && dd.collector != nil {
		dd.collector.PushVector(featureMap)
	}
}

// EvaluateLiveDrift performs a drift evaluation over the current live sample window.
func (dd *DriftDetector) EvaluateLiveDrift(ctx context.Context) *DriftMeasurement {
	liveSamples, sampleCount := dd.collector.Snapshot(dd.config.MaxWindowSize)

	measurement := &DriftMeasurement{
		MeasurementID:    fmt.Sprintf("meas_%d", time.Now().UnixNano()),
		Timestamp:        time.Now().UTC(),
		ModelVersion:     dd.baseline.ModelVersion,
		BaselineID:       dd.baseline.BaselineID,
		EvaluationWindow: dd.config.MaxWindowSize,
		SampleCount:      sampleCount,
		OverallStatus:    DriftStatusHealthy,
		FeatureResults:   make(map[string]FeatureDriftResult, len(dd.baseline.Features)),
	}

	if sampleCount < dd.config.MinSamplesForDrift {
		measurement.OverallStatus = DriftStatusHealthy
		dd.updateState(measurement)
		return measurement
	}

	var maxPSI, maxJSD, maxKL float64
	var driftedCount, criticalCount int

	featureRecords := make([]audit.DriftFeatureMeasurementRecord, 0, len(dd.baseline.Features))

	for featureName, baseDist := range dd.baseline.Features {
		vals := liveSamples[featureName]
		driftRes := CalculateFeatureDrift(vals, baseDist, dd.config)
		measurement.FeatureResults[featureName] = driftRes

		if driftRes.PSI > maxPSI {
			maxPSI = driftRes.PSI
		}
		if driftRes.JSD > maxJSD {
			maxJSD = driftRes.JSD
		}
		if driftRes.KL > maxKL {
			maxKL = driftRes.KL
		}

		if driftRes.Severity == SeverityWarning || driftRes.Severity == SeverityHigh {
			driftedCount++
		} else if driftRes.Severity == SeverityCritical {
			driftedCount++
			criticalCount++
		}

		featureRecords = append(featureRecords, audit.DriftFeatureMeasurementRecord{
			MeasurementID: measurement.MeasurementID,
			Timestamp:     measurement.Timestamp,
			FeatureName:   featureName,
			SampleCount:   uint32(len(vals)),
			PSI:           driftRes.PSI,
			JSD:           driftRes.JSD,
			KL:            driftRes.KL,
			BaselineMean:  driftRes.BaselineMean,
			LiveMean:      driftRes.LiveMean,
			MeanShift:     driftRes.MeanShift,
			BaselineStd:   driftRes.BaselineStd,
			LiveStd:       driftRes.LiveStd,
			StdShift:      driftRes.StdShift,
			MissingRate:   driftRes.MissingRate,
			Severity:      string(driftRes.Severity),
		})
	}

	measurement.MaxPSI = maxPSI
	measurement.MaxJSD = maxJSD
	measurement.MaxKL = maxKL
	measurement.DriftedFeatureCount = driftedCount
	measurement.CriticalFeatureCount = criticalCount

	// Determine overall model drift status
	if criticalCount >= 2 || maxPSI >= dd.config.PSICritThreshold {
		measurement.OverallStatus = DriftStatusCritical
	} else if criticalCount >= 1 || driftedCount >= 4 || maxPSI >= dd.config.PSIHighThreshold {
		measurement.OverallStatus = DriftStatusDegraded
	} else if driftedCount >= 1 || maxPSI >= dd.config.PSIWarnThreshold {
		measurement.OverallStatus = DriftStatusWarning
	} else {
		measurement.OverallStatus = DriftStatusHealthy
	}

	dd.updateState(measurement)

	// Persist to ClickHouse asynchronously
	if dd.chClient != nil {
		go func() {
			insCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			_ = dd.chClient.InsertDriftMeasurement(insCtx, audit.DriftMeasurementRecord{
				MeasurementID:        measurement.MeasurementID,
				Timestamp:            measurement.Timestamp,
				ModelVersion:         measurement.ModelVersion,
				BaselineID:           measurement.BaselineID,
				EvaluationWindow:     uint32(measurement.EvaluationWindow),
				SampleCount:          uint32(measurement.SampleCount),
				OverallStatus:        string(measurement.OverallStatus),
				MaxPSI:               measurement.MaxPSI,
				MaxJSD:               measurement.MaxJSD,
				MaxKL:                measurement.MaxKL,
				DriftedFeatureCount:  uint16(measurement.DriftedFeatureCount),
				CriticalFeatureCount: uint16(measurement.CriticalFeatureCount),
			})

			_ = dd.chClient.InsertDriftFeatureMeasurements(insCtx, featureRecords)
		}()
	}

	// Notify registered listener (e.g. RetrainingCoordinator)
	dd.mu.RLock()
	cb := dd.onMeasurementEvaluated
	dd.mu.RUnlock()
	if cb != nil {
		cb(measurement)
	}

	return measurement
}

func (dd *DriftDetector) updateState(m *DriftMeasurement) {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	prevStatus := dd.currentStatus
	dd.currentStatus = m.OverallStatus
	dd.latestMeasurement = m

	// Maintain bounded history
	if len(dd.history) >= dd.maxHistory {
		dd.history = dd.history[1:]
	}
	dd.history = append(dd.history, *m)

	// Emit operational audit event if state changed
	if prevStatus != m.OverallStatus && dd.chClient != nil {
		eventType := "DRIFT_STATUS_CHANGED"
		if m.OverallStatus == DriftStatusCritical {
			eventType = "DRIFT_CRITICAL"
		} else if m.OverallStatus == DriftStatusDegraded {
			eventType = "DRIFT_HIGH"
		} else if m.OverallStatus == DriftStatusWarning {
			eventType = "DRIFT_WARNING"
		} else if m.OverallStatus == DriftStatusHealthy {
			eventType = "DRIFT_RECOVERED"
		}

		go func() {
			evCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = dd.chClient.InsertDriftEvent(evCtx, audit.DriftEventRecord{
				EventID:              fmt.Sprintf("evt_drift_%d", time.Now().UnixNano()),
				Timestamp:            time.Now().UTC(),
				ModelVersion:         m.ModelVersion,
				BaselineID:           m.BaselineID,
				PreviousStatus:       string(prevStatus),
				NewStatus:            string(m.OverallStatus),
				MaxPSI:               m.MaxPSI,
				MaxJSD:               m.MaxJSD,
				MaxKL:                m.MaxKL,
				AffectedFeatureCount: uint16(m.DriftedFeatureCount),
				CriticalFeatureCount: uint16(m.CriticalFeatureCount),
				Trigger:              eventType,
				Reason:               fmt.Sprintf("Overall drift transitioned from %s to %s (max PSI: %.4f)", prevStatus, m.OverallStatus, m.MaxPSI),
			})
		}()
	}
}

// GetStatus returns the current live drift status payload for GET /v1/drift/status.
func (dd *DriftDetector) GetStatus() map[string]interface{} {
	if dd == nil {
		return map[string]interface{}{
			"status":  "UNAVAILABLE",
			"message": "drift detection subsystem is not initialized",
		}
	}

	dd.mu.RLock()
	defer dd.mu.RUnlock()

	status := dd.currentStatus
	latest := dd.latestMeasurement

	totalSamples := int64(0)
	if dd.collector != nil {
		totalSamples = dd.collector.TotalCollected()
	}

	maxPSI, maxJSD, maxKL := 0.0, 0.0, 0.0
	drifted, crit := 0, 0
	lastCalculated := "never"
	sampleCount := 0
	recommendation := "NONE"
	featureList := make([]map[string]interface{}, 0)

	if latest != nil {
		maxPSI = latest.MaxPSI
		maxJSD = latest.MaxJSD
		maxKL = latest.MaxKL
		drifted = latest.DriftedFeatureCount
		crit = latest.CriticalFeatureCount
		sampleCount = latest.SampleCount
		lastCalculated = latest.Timestamp.Format(time.RFC3339)

		if status == DriftStatusCritical {
			recommendation = "CRITICAL: Multiple features exhibiting major distribution shift. Model inspection and retraining recommended."
		} else if status == DriftStatusDegraded {
			recommendation = "DEGRADED: Moderate feature drift observed. Review incoming traffic profiles."
		} else if status == DriftStatusWarning {
			recommendation = "WARNING: Slight divergence in feature distributions. Continue automated monitoring."
		}

		for name, res := range latest.FeatureResults {
			featureList = append(featureList, map[string]interface{}{
				"name":                 name,
				"sample_count":         res.SampleCount,
				"psi":                  res.PSI,
				"jsd":                  res.JSD,
				"kl":                   res.KL,
				"baseline_mean":        res.BaselineMean,
				"live_mean":            res.LiveMean,
				"mean_shift":           res.MeanShift,
				"baseline_std":         res.BaselineStd,
				"live_std":             res.LiveStd,
				"std_shift":            res.StdShift,
				"live_min":             res.LiveMin,
				"live_max":             res.LiveMax,
				"live_p50":             res.LiveP50,
				"live_p95":             res.LiveP95,
				"missing_rate":         res.MissingRate,
				"unseen_category_rate": res.UnseenCategoryRate,
				"severity":             res.Severity,
				"recommendation":       res.Recommendation,
			})
		}
	}

	return map[string]interface{}{
		"status":                 status,
		"model_version":          dd.baseline.ModelVersion,
		"feature_contract":       dd.baseline.FeatureContract,
		"baseline_id":            dd.baseline.BaselineID,
		"baseline_source":        dd.baseline.BaselineSource,
		"total_samples_ingested": totalSamples,
		"window": map[string]interface{}{
			"sample_count": sampleCount,
			"max_capacity": dd.config.MaxWindowSize,
		},
		"metrics": map[string]interface{}{
			"max_psi": maxPSI,
			"max_jsd": maxJSD,
			"max_kl":  maxKL,
		},
		"max_psi":            maxPSI,
		"max_jsd":            maxJSD,
		"max_kl":             maxKL,
		"drifted_features":   drifted,
		"critical_features":  crit,
		"last_calculated_at": lastCalculated,
		"recommendation":     recommendation,
		"features":           featureList,
	}
}

// GetSummary returns a concise summary of the drift status for GET /v1/system/status.
func (dd *DriftDetector) GetSummary() map[string]interface{} {
	if dd == nil {
		return map[string]interface{}{
			"status": "DEGRADED",
		}
	}

	dd.mu.RLock()
	defer dd.mu.RUnlock()

	latest := dd.latestMeasurement
	if latest == nil {
		return map[string]interface{}{
			"status":             string(dd.currentStatus),
			"max_psi":            0.0,
			"max_jsd":            0.0,
			"drifted_features":   0,
			"critical_features":  0,
			"last_calculated_at": "never",
			"baseline_source":    dd.baseline.BaselineSource,
		}
	}

	return map[string]interface{}{
		"status":             string(dd.currentStatus),
		"max_psi":            latest.MaxPSI,
		"max_jsd":            latest.MaxJSD,
		"drifted_features":   latest.DriftedFeatureCount,
		"critical_features":  latest.CriticalFeatureCount,
		"last_calculated_at": latest.Timestamp.Format(time.RFC3339),
		"baseline_source":    dd.baseline.BaselineSource,
	}
}

// GetHistory returns the recent historical drift measurements.
func (dd *DriftDetector) GetHistory() []DriftMeasurement {
	if dd == nil {
		return []DriftMeasurement{}
	}

	dd.mu.RLock()
	defer dd.mu.RUnlock()

	histCopy := make([]DriftMeasurement, len(dd.history))
	copy(histCopy, dd.history)
	return histCopy
}
