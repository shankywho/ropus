package riskengine_test

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/shankywho/ropus/backend/internal/features"
	"github.com/shankywho/ropus/backend/internal/riskengine"
)

func TestMLFeatures_CanonicalContract_Comprehensive(t *testing.T) {
	t.Run("1. Exact Feature Counts & Ordering", func(t *testing.T) {
		if len(riskengine.Canonical25FeatureDefinitions) != 25 {
			t.Fatalf("expected 25 canonical feature definitions, got %d", len(riskengine.Canonical25FeatureDefinitions))
		}

		for i, def := range riskengine.Canonical25FeatureDefinitions {
			if def.Index != i {
				t.Errorf("definition index mismatch: def.Index=%d vs slice index=%d", def.Index, i)
			}
			if def.Name == "" {
				t.Errorf("empty feature name at index %d", i)
			}
			if !def.PointInTimeSafe {
				t.Errorf("feature %s is not marked PointInTimeSafe", def.Name)
			}
		}
	})

	t.Run("2. Legacy 15 Feature Contract Frozen Compatibility", func(t *testing.T) {
		expectedLegacy15 := []string{
			"amount",
			"ip_velocity_1h",
			"ip_velocity_24h",
			"token_velocity_24h",
			"device_seen_before",
			"transaction_hour",
			"transaction_day",
			"product_cd_encoded",
			"card_type_encoded",
			"card_category_encoded",
			"email_domain_risk",
			"dist1_missing",
			"device_type_mobile",
			"device_info_missing",
			"amount_to_mean_ratio",
		}

		for i, name := range expectedLegacy15 {
			actualName := riskengine.Canonical25FeatureDefinitions[i].Name
			if actualName != name {
				t.Errorf("legacy feature order mismatch at index %d: expected %s, got %s", i, name, actualName)
			}
		}
	})

	t.Run("3. New 10 Behavioral & Reputation Features", func(t *testing.T) {
		expectedNew10 := []string{
			"device_tx_count_5m",
			"device_tx_count_1h",
			"device_amount_sum_24h",
			"tx_acceleration_5m_1h",
			"device_amount_concentration_5m_1h",
			"device_unique_tokens_1h",
			"token_unique_devices_1h",
			"device_reputation_score",
			"device_fraud_rate",
			"device_dispute_rate",
		}

		for i, name := range expectedNew10 {
			actualName := riskengine.Canonical25FeatureDefinitions[15+i].Name
			if actualName != name {
				t.Errorf("new feature order mismatch at index %d: expected %s, got %s", 15+i, name, actualName)
			}
		}
	})

	t.Run("4. Empty History Default Construction", func(t *testing.T) {
		now := time.Date(2026, 8, 21, 14, 30, 0, 0, time.UTC)
		v25 := riskengine.BuildCanonical25FeatureVector(
			1000,
			nil, // novel device
			nil,
			nil,
			nil,
			nil,
			nil,
			now,
		)

		if v25.Version != riskengine.MLFeatureContractV25 {
			t.Errorf("expected version %s, got %s", riskengine.MLFeatureContractV25, v25.Version)
		}
		if len(v25.Features) != 25 {
			t.Fatalf("expected 25 features, got %d", len(v25.Features))
		}

		// Check neutral reputation
		if repScore := v25.FeatureMap["device_reputation_score"]; repScore != 0.50 {
			t.Errorf("expected neutral reputation 0.50, got %f", repScore)
		}
		if devSeen := v25.FeatureMap["device_seen_before"]; devSeen != 0.0 {
			t.Errorf("expected device_seen_before 0.0, got %f", devSeen)
		}
		if devMissing := v25.FeatureMap["device_info_missing"]; devMissing != 1.0 {
			t.Errorf("expected device_info_missing 1.0, got %f", devMissing)
		}

		// Adapter to V15
		v15 := riskengine.ExtractLegacy15FeatureVector(v25)
		if v15.Version != riskengine.MLFeatureContractV15 {
			t.Errorf("expected legacy version %s, got %s", riskengine.MLFeatureContractV15, v15.Version)
		}
		if len(v15.Features) != 15 {
			t.Fatalf("expected 15 features in legacy vector, got %d", len(v15.Features))
		}
	})

	t.Run("5. NaN and Inf Sanitization & Clamping", func(t *testing.T) {
		defRatio := riskengine.Canonical25FeatureDefinitions[19] // device_amount_concentration_5m_1h [0, 1]

		// NaN -> Default
		sanNaN := riskengine.SanitizeFeatureValue(math.NaN(), defRatio)
		if sanNaN != defRatio.DefaultValue {
			t.Errorf("expected NaN sanitized to default %f, got %f", defRatio.DefaultValue, sanNaN)
		}

		// +Inf -> MaxBound
		sanPosInf := riskengine.SanitizeFeatureValue(math.Inf(1), defRatio)
		if sanPosInf != defRatio.MaxBound {
			t.Errorf("expected +Inf sanitized to MaxBound %f, got %f", defRatio.MaxBound, sanPosInf)
		}

		// -Inf -> MinBound
		sanNegInf := riskengine.SanitizeFeatureValue(math.Inf(-1), defRatio)
		if sanNegInf != defRatio.MinBound {
			t.Errorf("expected -Inf sanitized to MinBound %f, got %f", defRatio.MinBound, sanNegInf)
		}

		// Out-of-bounds high -> clamped to MaxBound
		sanHigh := riskengine.SanitizeFeatureValue(99.9, defRatio)
		if sanHigh != 1.0 {
			t.Errorf("expected 99.9 clamped to 1.0, got %f", sanHigh)
		}

		// Out-of-bounds low -> clamped to MinBound
		sanLow := riskengine.SanitizeFeatureValue(-5.0, defRatio)
		if sanLow != 0.0 {
			t.Errorf("expected -5.0 clamped to 0.0, got %f", sanLow)
		}
	})

	t.Run("6. Point-in-Time Behavioral Signal Assembly", func(t *testing.T) {
		now := time.Date(2026, 8, 21, 15, 0, 0, 0, time.UTC)
		devIdentity := &features.DeviceIdentity{
			DeviceID: "dev_test_pit_01",
			Status:   features.DeviceStatusValid,
			IsValid:  true,
		}
		devFeatures := &features.DeviceFeatures{
			DeviceSeenBefore: 1,
		}
		tokenFeatures := &features.PaymentTokenFeatures{
			DeviceUniqueTokens1h: 3,
			TokenUniqueDevices1h: 2,
		}
		velFeatures := &features.DeviceVelocityFeatures{
			DeviceTxCount5m:                  4,
			DeviceTxCount1h:                  12,
			DeviceAmountSum24h:               50000.0,
			TxAcceleration5m1h:               8.5,
			DeviceAmountConcentration5m1h:   0.65,
			DeviceAvgAmount24h:               2500.0,
		}
		repFeatures := &features.DeviceReputationFeatures{
			DeviceReputationScore: 0.15,
			DeviceFraudRate:       0.0,
			DeviceDisputeRate:     0.02,
		}
		velMetrics := &features.VelocityMetrics{
			TxnCountIP1h:     6,
			TxnCountToken24h: 3,
		}

		v25 := riskengine.BuildCanonical25FeatureVector(
			5000,
			devIdentity,
			devFeatures,
			tokenFeatures,
			velFeatures,
			repFeatures,
			velMetrics,
			now,
		)

		if v25.FeatureMap["amount"] != 5000.0 {
			t.Errorf("expected amount 5000.0, got %f", v25.FeatureMap["amount"])
		}
		if v25.FeatureMap["device_tx_count_5m"] != 4.0 {
			t.Errorf("expected device_tx_count_5m 4.0, got %f", v25.FeatureMap["device_tx_count_5m"])
		}
		if v25.FeatureMap["device_unique_tokens_1h"] != 3.0 {
			t.Errorf("expected device_unique_tokens_1h 3.0, got %f", v25.FeatureMap["device_unique_tokens_1h"])
		}
		if v25.FeatureMap["device_reputation_score"] != 0.15 {
			t.Errorf("expected reputation 0.15, got %f", v25.FeatureMap["device_reputation_score"])
		}
		if v25.FeatureMap["amount_to_mean_ratio"] != 2.0 { // 5000 / 2500
			t.Errorf("expected amount_to_mean_ratio 2.0, got %f", v25.FeatureMap["amount_to_mean_ratio"])
		}

		slice15 := riskengine.ToLegacy15FloatSlice(v25)
		if len(slice15) != 15 {
			t.Fatalf("expected 15 floats, got %d", len(slice15))
		}
		if slice15[0] != 5000.0 || slice15[1] != 6.0 || slice15[3] != 3.0 {
			t.Errorf("unexpected values in slice15: %v", slice15)
		}

		slice25 := riskengine.ToCanonical25FloatSlice(v25)
		if len(slice25) != 25 {
			t.Fatalf("expected 25 floats, got %d", len(slice25))
		}
		if slice25[15] != 4.0 || slice25[22] != 0.15 {
			t.Errorf("unexpected values in slice25: %v", slice25)
		}
	})

	t.Run("7. Concurrency: 100 Simultaneous Vector Constructions", func(t *testing.T) {
		var wg sync.WaitGroup
		numWorkers := 100

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				now := time.Now().UTC()
				v25 := riskengine.BuildCanonical25FeatureVector(
					int64(1000+idx*10),
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					now,
				)
				if len(v25.Features) != 25 {
					t.Errorf("worker %d got wrong feature count %d", idx, len(v25.Features))
				}
				v15 := riskengine.ExtractLegacy15FeatureVector(v25)
				if len(v15.Features) != 15 {
					t.Errorf("worker %d got wrong legacy count %d", idx, len(v15.Features))
				}
			}(i)
		}
		wg.Wait()
	})
}
