package riskengine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"time"
)

// Canonical25FeatureNames lists the exact ordered 25 features for contract v2.5.
var Canonical25FeatureNames = []string{
	"amount",
	"ip_velocity_1h",
	"ip_velocity_24h",
	"token_velocity_24h",
	"is_new_device",
	"device_seen_before",
	"hour_of_day",
	"day_of_week",
	"product_cd",
	"card_type",
	"card_category",
	"email_domain",
	"dist1_missing",
	"device_type_mobile",
	"device_info_missing",
	"amount_to_mean_ratio",
	"country_match",
	"billing_shipping_match",
	"failed_attempts_1h",
	"avg_amount_7d",
	"max_amount_30d",
	"distinct_cards_1h",
	"distinct_ips_24h",
	"is_vpn_or_proxy",
	"chargeback_history_count",
}

// DatasetValidationResult encapsulates detailed results from dataset verification.
type DatasetValidationResult struct {
	DatasetID          string                  `json:"dataset_id"`
	FilePath           string                  `json:"file_path"`
	ChecksumSHA256     string                  `json:"checksum_sha256"`
	Timestamp          time.Time               `json:"timestamp"`
	SampleCount        uint32                  `json:"sample_count"`
	FeatureCount       uint16                  `json:"feature_count"`
	FeatureContract    string                  `json:"feature_contract"`
	PositiveFraudCount uint32                  `json:"positive_fraud_count"`
	FraudRate          float64                 `json:"fraud_rate"`
	MissingValueRate   float64                 `json:"missing_value_rate"`
	DataQualityScore   float64                 `json:"data_quality_score"`
	Passed             bool                    `json:"passed"`
	Violations         []string                `json:"violations,omitempty"`
	Metadata           TrainingDatasetMetadata `json:"metadata"`
}

// DatasetValidator performs strict structural and statistical validation on training datasets.
type DatasetValidator struct {
	minSampleQuorum    uint32
	maxMissingRate     float64
	minQualityScore    float64
	minPositiveSamples uint32
}

// NewDatasetValidator creates a validator with production safety thresholds.
func NewDatasetValidator(minSampleQuorum uint32) *DatasetValidator {
	if minSampleQuorum < 50 {
		minSampleQuorum = 50
	}
	return &DatasetValidator{
		minSampleQuorum:    minSampleQuorum,
		maxMissingRate:     0.15, // Max 15% missing values
		minQualityScore:    0.80, // Min 80% quality score
		minPositiveSamples: 1,    // At least 1 positive fraud sample required
	}
}

// ValidateDatasetMetadata checks training dataset metadata against safety policies.
func (v *DatasetValidator) ValidateDatasetMetadata(ctx context.Context, meta TrainingDatasetMetadata) (*DatasetValidationResult, error) {
	result := &DatasetValidationResult{
		DatasetID:       meta.DatasetID,
		Timestamp:       time.Now().UTC(),
		SampleCount:     meta.SampleCount,
		FeatureContract: meta.FeatureContract,
		Metadata:        meta,
		Violations:      make([]string, 0),
	}

	if meta.SampleCount < v.minSampleQuorum {
		result.Violations = append(result.Violations,
			fmt.Sprintf("Sample count %d is below minimum quorum threshold %d", meta.SampleCount, v.minSampleQuorum))
	}

	if meta.FeatureContract != MLFeatureContractV25 && meta.FeatureContract != "fraud-risk-25f-v2.5" && meta.FeatureContract != "v2.5" {
		result.Violations = append(result.Violations,
			fmt.Sprintf("Incompatible feature contract '%s'; expected canonical 25-feature contract v2.5", meta.FeatureContract))
	}

	if meta.MissingValueRate > v.maxMissingRate {
		result.Violations = append(result.Violations,
			fmt.Sprintf("Missing value rate %.4f exceeds maximum permitted threshold %.4f", meta.MissingValueRate, v.maxMissingRate))
	}

	if meta.DataQualityScore < v.minQualityScore {
		result.Violations = append(result.Violations,
			fmt.Sprintf("Data quality score %.4f is below minimum threshold %.4f", meta.DataQualityScore, v.minQualityScore))
	}

	if meta.ZeroLabelsDetected {
		result.Violations = append(result.Violations, "Dataset contains zero positive fraud labels; training cannot proceed")
	}

	result.Passed = len(result.Violations) == 0
	if !result.Passed {
		return result, fmt.Errorf("dataset validation failed (%d violations): %v", len(result.Violations), result.Violations)
	}

	return result, nil
}

// ValidateDatasetFile performs file integrity checks and calculates SHA-256 checksum.
func (v *DatasetValidator) ValidateDatasetFile(filePath string) (string, int64, error) {
	if filePath == "" {
		return "", 0, fmt.Errorf("dataset file path cannot be empty")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("dataset file not found or inaccessible at '%s': %w", filePath, err)
	}

	if info.IsDir() {
		return "", 0, fmt.Errorf("dataset path '%s' is a directory, expected file", filePath)
	}

	if info.Size() == 0 {
		return "", 0, fmt.Errorf("dataset file '%s' is empty (0 bytes)", filePath)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open dataset file '%s': %w", filePath, err)
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", 0, fmt.Errorf("failed to compute dataset checksum: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return checksum, info.Size(), nil
}

// ValidateDatasetSchemaJSON validates a dataset descriptor JSON if present.
func (v *DatasetValidator) ValidateDatasetSchemaJSON(jsonData []byte) (*DatasetValidationResult, error) {
	var descriptor struct {
		DatasetID          string   `json:"dataset_id"`
		SampleCount        uint32   `json:"sample_count"`
		FeatureContract    string   `json:"feature_contract"`
		FeatureNames       []string `json:"feature_names"`
		PositiveFraudCount uint32   `json:"positive_fraud_count"`
		MissingValueRate   float64  `json:"missing_value_rate"`
		DataQualityScore   float64  `json:"data_quality_score"`
	}

	if err := json.Unmarshal(jsonData, &descriptor); err != nil {
		return nil, fmt.Errorf("failed to parse dataset schema JSON: %w", err)
	}

	meta := TrainingDatasetMetadata{
		DatasetID:          descriptor.DatasetID,
		SampleCount:        descriptor.SampleCount,
		FeatureContract:    descriptor.FeatureContract,
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   descriptor.DataQualityScore,
		MissingValueRate:   descriptor.MissingValueRate,
		ZeroLabelsDetected: descriptor.PositiveFraudCount == 0,
	}

	res, err := v.ValidateDatasetMetadata(context.Background(), meta)
	if err != nil {
		return res, err
	}

	// Verify exact 25 feature columns if names provided
	if len(descriptor.FeatureNames) > 0 {
		if len(descriptor.FeatureNames) != 25 {
			res.Violations = append(res.Violations,
				fmt.Sprintf("Feature count %d != 25 expected features", len(descriptor.FeatureNames)))
			res.Passed = false
			return res, fmt.Errorf("feature count mismatch: %d != 25", len(descriptor.FeatureNames))
		}
	}

	return res, nil
}

// CheckNumericalSanity checks an array of float features for NaN or Inf.
func CheckNumericalSanity(values []float64) (hasNaN bool, hasInf bool) {
	for _, val := range values {
		if math.IsNaN(val) {
			hasNaN = true
		}
		if math.IsInf(val, 0) {
			hasInf = true
		}
	}
	return hasNaN, hasInf
}
