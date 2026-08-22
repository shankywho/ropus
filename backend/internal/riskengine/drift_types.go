package riskengine

import "time"

// DriftSeverity defines the severity level of data distribution drift for a feature.
type DriftSeverity string

const (
	SeverityStable   DriftSeverity = "STABLE"
	SeverityWarning  DriftSeverity = "WARNING"
	SeverityHigh     DriftSeverity = "HIGH"
	SeverityCritical DriftSeverity = "CRITICAL"
)

// ModelDriftStatus defines the overall health status of the model based on feature drift.
type ModelDriftStatus string
type DriftStatus = ModelDriftStatus

const (
	DriftStatusHealthy  ModelDriftStatus = "HEALTHY"
	DriftStatusWarning  ModelDriftStatus = "WARNING"
	DriftStatusDegraded ModelDriftStatus = "DEGRADED"
	DriftStatusCritical ModelDriftStatus = "CRITICAL"
)

// DriftConfig holds configurable thresholds and operational parameters for drift detection.
type DriftConfig struct {
	PSIWarnThreshold    float64       `json:"psi_warn_threshold"`    // Default: 0.10
	PSIHighThreshold    float64       `json:"psi_high_threshold"`    // Default: 0.20
	PSICritThreshold    float64       `json:"psi_crit_threshold"`    // Default: 0.30
	JSDWarnThreshold    float64       `json:"jsd_warn_threshold"`    // Default: 0.05
	JSDHighThreshold    float64       `json:"jsd_high_threshold"`    // Default: 0.10
	JSDCritThreshold    float64       `json:"jsd_crit_threshold"`    // Default: 0.15
	KLWarnThreshold     float64       `json:"kl_warn_threshold"`     // Default: 0.10
	MinSamplesForDrift  int           `json:"min_samples_for_drift"` // Default: 50
	MaxWindowSize       int           `json:"max_window_size"`       // Default: 10000 (supports 1k, 10k, 100k)
	CalculationInterval time.Duration `json:"calculation_interval"`  // Default: 5m
	Epsilon             float64       `json:"epsilon"`               // Default: 1e-6
}

// DefaultDriftConfig returns production-standard drift detection parameters.
func DefaultDriftConfig() DriftConfig {
	return DriftConfig{
		PSIWarnThreshold:    0.10,
		PSIHighThreshold:    0.20,
		PSICritThreshold:    0.30,
		JSDWarnThreshold:    0.05,
		JSDHighThreshold:    0.10,
		JSDCritThreshold:    0.15,
		KLWarnThreshold:     0.10,
		MinSamplesForDrift:  50,
		MaxWindowSize:       10000,
		CalculationInterval: 5 * time.Minute,
		Epsilon:             1e-6,
	}
}

// FeatureDistribution captures the baseline distribution statistics and histogram/category bins for a feature.
type FeatureDistribution struct {
	Name          string             `json:"name"`
	DataType      string             `json:"data_type"`
	IsCategorical bool               `json:"is_categorical,omitempty"`
	Count         int                `json:"count"`
	Mean          float64            `json:"mean"`
	Std           float64            `json:"std"`
	Min           float64            `json:"min"`
	Max           float64            `json:"max"`
	P01           float64            `json:"p01"`
	P05           float64            `json:"p05"`
	P25           float64            `json:"p25"`
	P50           float64            `json:"p50"`
	P75           float64            `json:"p75"`
	P95           float64            `json:"p95"`
	P99           float64            `json:"p99"`
	BinEdges      []float64          `json:"bin_edges,omitempty"`
	BinProbs      []float64          `json:"bin_probs,omitempty"`
	Categories    []string           `json:"categories,omitempty"`
	CategoryProbs map[string]float64 `json:"category_probs,omitempty"`
}

// ModelBaseline represents the reference baseline dataset statistics for a model version.
type ModelBaseline struct {
	BaselineID         string                         `json:"baseline_id"`
	ModelVersion       string                         `json:"model_version"`
	FeatureContract    string                         `json:"feature_contract"`
	CalibrationVersion string                         `json:"calibration_version"`
	BaselineSource     string                         `json:"baseline_source"` // "training" or "development_fixture"
	DatasetVersion     string                         `json:"dataset_version"`
	CreatedAt          string                         `json:"created_at"`
	TotalSamples       int                            `json:"total_samples"`
	FeatureCount       int                            `json:"feature_count"`
	Features           map[string]FeatureDistribution `json:"features"`
}

// FeatureDriftResult holds the drift analysis metrics for a single feature.
type FeatureDriftResult struct {
	FeatureName        string        `json:"feature_name"`
	SampleCount        int           `json:"sample_count"`
	PSI                float64       `json:"psi"`
	JSD                float64       `json:"jsd"`
	KL                 float64       `json:"kl"`
	BaselineMean       float64       `json:"baseline_mean"`
	LiveMean           float64       `json:"live_mean"`
	MeanShift          float64       `json:"mean_shift"`
	BaselineStd        float64       `json:"baseline_std"`
	LiveStd            float64       `json:"live_std"`
	StdShift           float64       `json:"std_shift"`
	LiveMin            float64       `json:"live_min"`
	LiveMax            float64       `json:"live_max"`
	LiveP01            float64       `json:"live_p01"`
	LiveP05            float64       `json:"live_p05"`
	LiveP25            float64       `json:"live_p25"`
	LiveP50            float64       `json:"live_p50"`
	LiveP75            float64       `json:"live_p75"`
	LiveP95            float64       `json:"live_p95"`
	LiveP99            float64       `json:"live_p99"`
	MissingRate        float64       `json:"missing_rate"`
	UnseenCategoryRate float64       `json:"unseen_category_rate"`
	Severity           DriftSeverity `json:"severity"`
	Recommendation     string        `json:"recommendation"`
}

// DriftMeasurement represents a point-in-time evaluation of model drift across all features.
type DriftMeasurement struct {
	MeasurementID        string                        `json:"measurement_id"`
	Timestamp            time.Time                     `json:"timestamp"`
	ModelVersion         string                        `json:"model_version"`
	BaselineID           string                        `json:"baseline_id"`
	EvaluationWindow     int                           `json:"evaluation_window"`
	SampleCount          int                           `json:"sample_count"`
	OverallStatus        ModelDriftStatus              `json:"overall_status"`
	MaxPSI               float64                       `json:"max_psi"`
	MaxJSD               float64                       `json:"max_jsd"`
	MaxKL                float64                       `json:"max_kl"`
	DriftedFeatureCount  int                           `json:"drifted_feature_count"`
	CriticalFeatureCount int                           `json:"critical_feature_count"`
	FeatureResults       map[string]FeatureDriftResult `json:"feature_results"`
}

// DriftSummary provides a concise operational status payload for system status aggregation.
type DriftSummary struct {
	Status            ModelDriftStatus `json:"status"`
	MaxPSI            float64          `json:"max_psi"`
	MaxJSD            float64          `json:"max_jsd"`
	MaxKL             float64          `json:"max_kl"`
	DriftedFeatures   int              `json:"drifted_features"`
	CriticalFeatures  int              `json:"critical_features"`
	LastCalculatedAt  string           `json:"last_calculated_at"`
	BaselineID        string           `json:"baseline_id,omitempty"`
	BaselineSource    string           `json:"baseline_source,omitempty"`
}
