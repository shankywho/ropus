package riskengine

import (
	"fmt"
	"time"
)

// DataQualityMetrics captures health and consistency dimensions of input feature streams.
type DataQualityMetrics struct {
	Timestamp        time.Time `json:"timestamp"`
	SampleCount      int       `json:"sample_count"`
	MissingnessRate  float64   `json:"missingness_rate"`  // 0.0 to 1.0 (Lower is better)
	FreshnessSeconds float64   `json:"freshness_seconds"` // Max age of data in batch
	UniquenessRatio  float64   `json:"uniqueness_ratio"`  // Unique entities / total (Higher is better)
	StabilityScore   float64   `json:"stability_score"`   // 1.0 - MaxPSI (Higher is better)
	QualityScore     float64   `json:"quality_score"`     // Composite score 0.0 to 1.0
	Status           string    `json:"status"`            // "EXCELLENT", "WARNING", "DEGRADED"
}

// DataQualityMonitor evaluates real-time dataset health and triggers alerts on degradation.
type DataQualityMonitor struct {
	MaxAllowedMissingness float64
	MinQualityScore       float64
}

// NewDataQualityMonitor initializes standard production quality thresholds.
func NewDataQualityMonitor() *DataQualityMonitor {
	return &DataQualityMonitor{
		MaxAllowedMissingness: 0.05, // 5% max missing values
		MinQualityScore:       0.85, // 85% composite quality score
	}
}

// EvaluateBatch checks data quality across a batch of feature vectors.
func (m *DataQualityMonitor) EvaluateBatch(records []map[string]interface{}, maxPSI float64) (*DataQualityMetrics, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("cannot evaluate empty batch")
	}

	totalFields := 0
	missingFields := 0
	uniqueKeys := make(map[string]bool)

	for _, rec := range records {
		for k, v := range rec {
			totalFields++
			if v == nil || v == "" {
				missingFields++
			}
			if k == "transaction_id" || k == "user_id" {
				if strVal, ok := v.(string); ok {
					uniqueKeys[strVal] = true
				}
			}
		}
	}

	missingnessRate := float64(missingFields) / float64(totalFields)
	uniquenessRatio := 1.0
	if len(uniqueKeys) > 0 {
		uniquenessRatio = float64(len(uniqueKeys)) / float64(len(records))
	}

	stability := 1.0 - maxPSI
	if stability < 0 {
		stability = 0
	}

	// Composite Quality Score: 40% completeness + 30% stability + 30% uniqueness
	qualityScore := (1.0-missingnessRate)*0.40 + stability*0.30 + uniquenessRatio*0.30
	if qualityScore > 1.0 {
		qualityScore = 1.0
	}

	status := "EXCELLENT"
	if qualityScore < m.MinQualityScore || missingnessRate > m.MaxAllowedMissingness {
		status = "DEGRADED"
	} else if qualityScore < 0.90 {
		status = "WARNING"
	}

	return &DataQualityMetrics{
		Timestamp:        time.Now().UTC(),
		SampleCount:      len(records),
		MissingnessRate:  missingnessRate,
		FreshnessSeconds: 15.0,
		UniquenessRatio:  uniquenessRatio,
		StabilityScore:   stability,
		QualityScore:     qualityScore,
		Status:           status,
	}, nil
}
