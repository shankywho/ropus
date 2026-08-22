package riskengine

import (
	"math"
	"time"

	"github.com/shankywho/ropus/backend/internal/features"
)

// Feature contract versions
const (
	MLFeatureContractV15 = "v1.5"
	MLFeatureContractV25 = "v2.5"
)

// MLFeatureDefinition encapsulates metadata, defaults, bounds, and provenance for an ML feature.
type MLFeatureDefinition struct {
	Index           int     `json:"index"`
	Name            string  `json:"name"`
	DataType        string  `json:"data_type"`
	DefaultValue    float64 `json:"default_value"`
	MinBound        float64 `json:"min_bound"`
	MaxBound        float64 `json:"max_bound"`
	SourceStore     string  `json:"source_store"`
	Description     string  `json:"description"`
	PointInTimeSafe bool    `json:"point_in_time_safe"`
}

// Canonical25FeatureDefinitions defines the exact schema and ordering of the 25-feature ML contract.
var Canonical25FeatureDefinitions = []MLFeatureDefinition{
	// -------------------------------------------------------------
	// Core 15 Legacy Features (V1.5 Contract, Indices 0-14)
	// -------------------------------------------------------------
	{
		Index:           0,
		Name:            "amount",
		DataType:        "float64",
		DefaultValue:    100.0,
		MinBound:        0.01,
		MaxBound:        1e9,
		SourceStore:     "transaction_request",
		Description:     "Transaction monetary amount in base currency units",
		PointInTimeSafe: true,
	},
	{
		Index:           1,
		Name:            "ip_velocity_1h",
		DataType:        "float64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        100000.0,
		SourceStore:     "redis:ip_velocity",
		Description:     "Count of prior transactions from the IP in the last 1 hour",
		PointInTimeSafe: true,
	},
	{
		Index:           2,
		Name:            "ip_velocity_24h",
		DataType:        "float64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        500000.0,
		SourceStore:     "redis:ip_velocity",
		Description:     "Count of prior transactions from the IP in the last 24 hours",
		PointInTimeSafe: true,
	},
	{
		Index:           3,
		Name:            "token_velocity_24h",
		DataType:        "float64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        100000.0,
		SourceStore:     "redis:token_velocity",
		Description:     "Count of prior transactions for payment token in last 24 hours",
		PointInTimeSafe: true,
	},
	{
		Index:           4,
		Name:            "device_seen_before",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1.0,
		SourceStore:     "redis:device_feature_store",
		Description:     "1 if device has recorded transactions prior to current, 0 if novel",
		PointInTimeSafe: true,
	},
	{
		Index:           5,
		Name:            "transaction_hour",
		DataType:        "int64",
		DefaultValue:    12.0,
		MinBound:        0.0,
		MaxBound:        23.0,
		SourceStore:     "transaction_context",
		Description:     "UTC hour of day (0-23)",
		PointInTimeSafe: true,
	},
	{
		Index:           6,
		Name:            "transaction_day",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        6.0,
		SourceStore:     "transaction_context",
		Description:     "UTC day of week (0=Monday to 6=Sunday)",
		PointInTimeSafe: true,
	},
	{
		Index:           7,
		Name:            "product_cd_encoded",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        100.0,
		SourceStore:     "transaction_request",
		Description:     "Categorical product code mapping",
		PointInTimeSafe: true,
	},
	{
		Index:           8,
		Name:            "card_type_encoded",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        100.0,
		SourceStore:     "transaction_request",
		Description:     "Categorical card network mapping",
		PointInTimeSafe: true,
	},
	{
		Index:           9,
		Name:            "card_category_encoded",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        100.0,
		SourceStore:     "transaction_request",
		Description:     "Categorical card credit/debit classification",
		PointInTimeSafe: true,
	},
	{
		Index:           10,
		Name:            "email_domain_risk",
		DataType:        "float64",
		DefaultValue:    0.035,
		MinBound:        0.0,
		MaxBound:        1.0,
		SourceStore:     "transaction_request",
		Description:     "Historical fraud risk score of email domain",
		PointInTimeSafe: true,
	},
	{
		Index:           11,
		Name:            "dist1_missing",
		DataType:        "int64",
		DefaultValue:    1.0,
		MinBound:        0.0,
		MaxBound:        1.0,
		SourceStore:     "transaction_request",
		Description:     "1 if distance to billing address is missing/unspecified, 0 otherwise",
		PointInTimeSafe: true,
	},
	{
		Index:           12,
		Name:            "device_type_mobile",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1.0,
		SourceStore:     "transaction_request",
		Description:     "1 if device is classified as mobile, 0 for desktop/other",
		PointInTimeSafe: true,
	},
	{
		Index:           13,
		Name:            "device_info_missing",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1.0,
		SourceStore:     "transaction_request",
		Description:     "1 if device metadata/fingerprint is missing, 0 if present",
		PointInTimeSafe: true,
	},
	{
		Index:           14,
		Name:            "amount_to_mean_ratio",
		DataType:        "float64",
		DefaultValue:    1.0,
		MinBound:        0.0,
		MaxBound:        1000.0,
		SourceStore:     "redis:device_velocity",
		Description:     "Ratio of current transaction amount to rolling 24h historical mean",
		PointInTimeSafe: true,
	},

	// -------------------------------------------------------------
	// 10 New Advanced Behavioral & Reputation Features (Indices 15-24)
	// -------------------------------------------------------------
	{
		Index:           15,
		Name:            "device_tx_count_5m",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        10000.0,
		SourceStore:     "redis:device_velocity (Phase 3.6)",
		Description:     "Point-in-time count of transactions on this device in the last 5 minutes",
		PointInTimeSafe: true,
	},
	{
		Index:           16,
		Name:            "device_tx_count_1h",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        50000.0,
		SourceStore:     "redis:device_velocity (Phase 3.6)",
		Description:     "Point-in-time count of transactions on this device in the last 1 hour",
		PointInTimeSafe: true,
	},
	{
		Index:           17,
		Name:            "device_amount_sum_24h",
		DataType:        "float64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1e9,
		SourceStore:     "redis:device_velocity (Phase 3.6)",
		Description:     "Point-in-time cumulative transaction amount on this device in 24 hours",
		PointInTimeSafe: true,
	},
	{
		Index:           18,
		Name:            "tx_acceleration_5m_1h",
		DataType:        "float64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1000.0,
		SourceStore:     "redis:device_velocity (Phase 3.6)",
		Description:     "Transaction velocity acceleration ratio (5m rate vs 1h rate)",
		PointInTimeSafe: true,
	},
	{
		Index:           19,
		Name:            "device_amount_concentration_5m_1h",
		DataType:        "float64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1.0,
		SourceStore:     "redis:device_velocity (Phase 3.6)",
		Description:     "Ratio of 5-minute transaction volume to 1-hour transaction volume in [0, 1]",
		PointInTimeSafe: true,
	},
	{
		Index:           20,
		Name:            "device_unique_tokens_1h",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1000.0,
		SourceStore:     "redis:payment_token (Phase 3.5)",
		Description:     "Count of distinct payment tokens observed on this device in the last 1 hour",
		PointInTimeSafe: true,
	},
	{
		Index:           21,
		Name:            "token_unique_devices_1h",
		DataType:        "int64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1000.0,
		SourceStore:     "redis:payment_token (Phase 3.5)",
		Description:     "Count of distinct devices observed using this payment token in the last 1 hour",
		PointInTimeSafe: true,
	},
	{
		Index:           22,
		Name:            "device_reputation_score",
		DataType:        "float64",
		DefaultValue:    0.50,
		MinBound:        0.0,
		MaxBound:        1.0,
		SourceStore:     "redis:device_reputation (Phase 3.7)",
		Description:     "Deterministic device trust/risk score in [0.0 (Trusted), 1.0 (High Risk)]",
		PointInTimeSafe: true,
	},
	{
		Index:           23,
		Name:            "device_fraud_rate",
		DataType:        "float64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1.0,
		SourceStore:     "redis:device_reputation (Phase 3.7)",
		Description:     "Historical confirmed fraud rate for this device in [0.0, 1.0]",
		PointInTimeSafe: true,
	},
	{
		Index:           24,
		Name:            "device_dispute_rate",
		DataType:        "float64",
		DefaultValue:    0.0,
		MinBound:        0.0,
		MaxBound:        1.0,
		SourceStore:     "redis:device_reputation (Phase 3.7)",
		Description:     "Historical dispute/chargeback rate for this device in [0.0, 1.0]",
		PointInTimeSafe: true,
	},
}

// MLFeature represents a single materialized, validated, and sanitized feature value.
type MLFeature struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// MLFeatureVector represents a strongly typed feature container.
type MLFeatureVector struct {
	Version        string             `json:"version"`
	Features       []MLFeature        `json:"features"`
	FeatureMap     map[string]float64 `json:"feature_map"`
	IsDegraded     bool               `json:"is_degraded"`
	DegradeReasons []string           `json:"degrade_reasons,omitempty"`
}

// SanitizeFeatureValue validates and clamps a feature value against its definition bounds.
func SanitizeFeatureValue(val float64, def MLFeatureDefinition) float64 {
	if math.IsNaN(val) {
		return def.DefaultValue
	}
	if math.IsInf(val, 1) {
		return def.MaxBound
	}
	if math.IsInf(val, -1) {
		return def.MinBound
	}
	if val < def.MinBound {
		return def.MinBound
	}
	if val > def.MaxBound {
		return def.MaxBound
	}
	return val
}

// BuildCanonical25FeatureVector constructs the point-in-time safe 25-feature vector.
func BuildCanonical25FeatureVector(
	amount int64,
	devIdentity *features.DeviceIdentity,
	devFeatures *features.DeviceFeatures,
	tokenFeatures *features.PaymentTokenFeatures,
	velFeatures *features.DeviceVelocityFeatures,
	repFeatures *features.DeviceReputationFeatures,
	velocityMetrics *features.VelocityMetrics,
	evalTime time.Time,
) *MLFeatureVector {
	if velocityMetrics == nil {
		velocityMetrics = &features.VelocityMetrics{}
	}
	if devFeatures == nil {
		devFeatures = &features.DeviceFeatures{}
	}
	if tokenFeatures == nil {
		tokenFeatures = &features.PaymentTokenFeatures{}
	}
	if velFeatures == nil {
		velFeatures = &features.DeviceVelocityFeatures{}
	}
	if repFeatures == nil {
		repFeatures = &features.DeviceReputationFeatures{DeviceReputationScore: 0.50}
	}

	degradeReasons := make([]string, 0)
	isDegraded := false

	if devFeatures.IsDegraded {
		isDegraded = true
		degradeReasons = append(degradeReasons, devFeatures.DegradeReason)
	}
	if tokenFeatures.IsDegraded {
		isDegraded = true
		degradeReasons = append(degradeReasons, tokenFeatures.DegradeReason)
	}
	if velFeatures.IsDegraded {
		isDegraded = true
		degradeReasons = append(degradeReasons, velFeatures.DegradeReason)
	}
	if repFeatures.IsDegraded {
		isDegraded = true
		degradeReasons = append(degradeReasons, repFeatures.DegradeReason)
	}

	// Calculate amount_to_mean_ratio
	amt := float64(amount)
	amtRatio := 1.0
	if velFeatures.DeviceAvgAmount24h > 0 {
		amtRatio = amt / velFeatures.DeviceAvgAmount24h
	} else if velFeatures.DeviceAmountSum24h > 0 {
		amtRatio = amt / (float64(velFeatures.DeviceAmountSum24h) + 1.0)
	}

	// Determine device seen before (Point-in-time safe: 0 if novel, 1 if seen before)
	devSeenBefore := 0.0
	if devIdentity != nil && devIdentity.IsValid && devFeatures.DeviceSeenBefore > 0 {
		devSeenBefore = 1.0
	}

	// Device info missing indicator
	devInfoMissing := 0.0
	if devIdentity == nil || !devIdentity.IsValid {
		devInfoMissing = 1.0
	}

	utcTime := evalTime.UTC()
	hourOfDay := float64(utcTime.Hour())
	dayOfWeek := float64(utcTime.Weekday())

	// Raw 25 feature values
	rawValues := map[string]float64{
		// Legacy 15 features
		"amount":                            amt,
		"ip_velocity_1h":                    float64(velocityMetrics.TxnCountIP1h),
		"ip_velocity_24h":                   float64(velocityMetrics.TxnCountToken24h), // proxy for 24h window
		"token_velocity_24h":                float64(velocityMetrics.TxnCountToken24h),
		"device_seen_before":                devSeenBefore,
		"transaction_hour":                  hourOfDay,
		"transaction_day":                   dayOfWeek,
		"product_cd_encoded":                0.0,
		"card_type_encoded":                 0.0,
		"card_category_encoded":             0.0,
		"email_domain_risk":                 0.035,
		"dist1_missing":                     1.0,
		"device_type_mobile":                0.0,
		"device_info_missing":               devInfoMissing,
		"amount_to_mean_ratio":              amtRatio,

		// 10 New features
		"device_tx_count_5m":                float64(velFeatures.DeviceTxCount5m),
		"device_tx_count_1h":                float64(velFeatures.DeviceTxCount1h),
		"device_amount_sum_24h":             float64(velFeatures.DeviceAmountSum24h),
		"tx_acceleration_5m_1h":             velFeatures.TxAcceleration5m1h,
		"device_amount_concentration_5m_1h": velFeatures.DeviceAmountConcentration5m1h,
		"device_unique_tokens_1h":           float64(tokenFeatures.DeviceUniqueTokens1h),
		"token_unique_devices_1h":           float64(tokenFeatures.TokenUniqueDevices1h),
		"device_reputation_score":           repFeatures.DeviceReputationScore,
		"device_fraud_rate":                 repFeatures.DeviceFraudRate,
		"device_dispute_rate":               repFeatures.DeviceDisputeRate,
	}

	featuresList := make([]MLFeature, len(Canonical25FeatureDefinitions))
	featureMap := make(map[string]float64, len(Canonical25FeatureDefinitions))

	for i, def := range Canonical25FeatureDefinitions {
		rawVal, exists := rawValues[def.Name]
		if !exists {
			rawVal = def.DefaultValue
		}
		sanitized := SanitizeFeatureValue(rawVal, def)
		featuresList[i] = MLFeature{
			Name:  def.Name,
			Value: sanitized,
		}
		featureMap[def.Name] = sanitized
	}

	return &MLFeatureVector{
		Version:        MLFeatureContractV25,
		Features:       featuresList,
		FeatureMap:     featureMap,
		IsDegraded:     isDegraded,
		DegradeReasons: degradeReasons,
	}
}

// ExtractLegacy15FeatureVector projects the canonical 25-feature vector into the legacy 15-feature contract.
func ExtractLegacy15FeatureVector(v25 *MLFeatureVector) *MLFeatureVector {
	if v25 == nil {
		return &MLFeatureVector{
			Version:    MLFeatureContractV15,
			Features:   make([]MLFeature, 0),
			FeatureMap: make(map[string]float64),
		}
	}

	legacyFeatures := make([]MLFeature, 15)
	legacyMap := make(map[string]float64, 15)

	for i := 0; i < 15; i++ {
		def := Canonical25FeatureDefinitions[i]
		val, exists := v25.FeatureMap[def.Name]
		if !exists {
			val = def.DefaultValue
		}
		legacyFeatures[i] = MLFeature{
			Name:  def.Name,
			Value: val,
		}
		legacyMap[def.Name] = val
	}

	return &MLFeatureVector{
		Version:        MLFeatureContractV15,
		Features:       legacyFeatures,
		FeatureMap:     legacyMap,
		IsDegraded:     v25.IsDegraded,
		DegradeReasons: v25.DegradeReasons,
	}
}

// ToLegacy15FloatSlice extracts the exact 15 floats in canonical order.
func ToLegacy15FloatSlice(v25 *MLFeatureVector) []float64 {
	slice := make([]float64, 15)
	if v25 == nil {
		for i := 0; i < 15; i++ {
			slice[i] = Canonical25FeatureDefinitions[i].DefaultValue
		}
		return slice
	}

	for i := 0; i < 15; i++ {
		def := Canonical25FeatureDefinitions[i]
		if val, ok := v25.FeatureMap[def.Name]; ok {
			slice[i] = val
		} else {
			slice[i] = def.DefaultValue
		}
	}
	return slice
}

// ToCanonical25FloatSlice extracts all 25 floats in canonical order.
func ToCanonical25FloatSlice(v25 *MLFeatureVector) []float64 {
	slice := make([]float64, 25)
	if v25 == nil {
		for i := 0; i < 25; i++ {
			slice[i] = Canonical25FeatureDefinitions[i].DefaultValue
		}
		return slice
	}

	for i := 0; i < 25; i++ {
		def := Canonical25FeatureDefinitions[i]
		if val, ok := v25.FeatureMap[def.Name]; ok {
			slice[i] = val
		} else {
			slice[i] = def.DefaultValue
		}
	}
	return slice
}
