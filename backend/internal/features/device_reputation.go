package features

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// OutcomeType defines the historical event outcome classification for a transaction.
type OutcomeType string

const (
	OutcomeSuccess     OutcomeType = "SUCCESS"
	OutcomeFailure     OutcomeType = "FAILURE"
	OutcomeDispute     OutcomeType = "DISPUTE"
	OutcomeFraud       OutcomeType = "FRAUD"
	OutcomeChargeback  OutcomeType = "CHARGEBACK"
	OutcomeRefund      OutcomeType = "REFUND"
	OutcomeTransaction OutcomeType = "TRANSACTION"
)

// DeviceReputationFeatures encapsulates point-in-time trust, dispute, and fraud history.
type DeviceReputationFeatures struct {
	DeviceTotalTransactions      int64   `json:"device_total_transactions"`
	DeviceSuccessfulTransactions int64   `json:"device_successful_transactions"`
	DeviceFailedTransactions     int64   `json:"device_failed_transactions"`
	DeviceDisputedTransactions   int64   `json:"device_disputed_transactions"`
	DeviceFraudTransactions      int64   `json:"device_fraud_transactions"`
	DeviceRefundedTransactions   int64   `json:"device_refunded_transactions"`
	DeviceChargebackCount        int64   `json:"device_chargeback_count"`

	DeviceDisputeRate            float64 `json:"device_dispute_rate"`
	DeviceFraudRate              float64 `json:"device_fraud_rate"`
	DeviceRefundRate             float64 `json:"device_refund_rate"`
	DeviceSuccessRate            float64 `json:"device_success_rate"`

	DeviceRecentDisputeCount     int64   `json:"device_recent_dispute_count"`     // 30-day window
	DeviceRecentFraudCount       int64   `json:"device_recent_fraud_count"`       // 30-day window
	DeviceRecentChargebackCount  int64   `json:"device_recent_chargeback_count"`  // 30-day window

	DeviceDaysSinceFirstSeen     float64 `json:"device_days_since_first_seen"`
	DeviceDaysSinceLastDispute   float64 `json:"device_days_since_last_dispute"` // -1.0 if never
	DeviceDaysSinceLastFraud     float64 `json:"device_days_since_last_fraud"`   // -1.0 if never

	DeviceReputationScore        float64 `json:"device_reputation_score"` // 0.0 (Trusted) to 1.0 (High Risk)
	IsDegraded                   bool    `json:"is_degraded"`
	DegradeReason                string  `json:"degrade_reason,omitempty"`
}

// DeviceReputationStore provides real-time Redis-backed and point-in-time reputation analysis.
type DeviceReputationStore struct {
	client redis.Cmdable
}

// NewDeviceReputationStore constructs a new DeviceReputationStore.
func NewDeviceReputationStore(client redis.Cmdable) *DeviceReputationStore {
	return &DeviceReputationStore{client: client}
}

// GetReputationFeatures evaluates the device reputation as of time.Now().
func (s *DeviceReputationStore) GetReputationFeatures(ctx context.Context, tenantID, deviceID string) (*DeviceReputationFeatures, error) {
	return s.GetReputationFeaturesAtTime(ctx, tenantID, deviceID, time.Now().UTC())
}

// GetReputationFeaturesAtTime evaluates the device reputation relative to an injected reference timestamp.
func (s *DeviceReputationStore) GetReputationFeaturesAtTime(ctx context.Context, tenantID, deviceID string, now time.Time) (*DeviceReputationFeatures, error) {
	defaultFeatures := &DeviceReputationFeatures{
		DeviceDisputeRate:          0.0,
		DeviceFraudRate:            0.0,
		DeviceRefundRate:           0.0,
		DeviceSuccessRate:          0.0,
		DeviceDaysSinceFirstSeen:   0.0,
		DeviceDaysSinceLastDispute: -1.0,
		DeviceDaysSinceLastFraud:   -1.0,
		DeviceReputationScore:      0.50, // Neutral default
		IsDegraded:                 false,
	}

	if s == nil || s.client == nil || deviceID == "" || tenantID == "" {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "DEVICE_REPUTATION_FEATURE_STORE_UNAVAILABLE"
		return defaultFeatures, nil
	}

	keyEvents := fmt.Sprintf("%s:rep:dev:events:%s", tenantID, deviceID)
	keyFirstSeen := fmt.Sprintf("%s:rep:dev:first_seen:%s", tenantID, deviceID)

	nowMs := strconv.FormatInt(now.UnixMilli(), 10)

	pipe := s.client.Pipeline()
	cmdFirstSeen := pipe.Get(ctx, keyFirstSeen)
	cmdEvents := pipe.ZRangeByScoreWithScores(ctx, keyEvents, &redis.ZRangeBy{
		Min: "-inf",
		Max: nowMs,
	})

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "DEVICE_REPUTATION_FEATURE_STORE_UNAVAILABLE"
		return defaultFeatures, fmt.Errorf("failed to fetch reputation events from redis: %w", err)
	}

	// 1. Calculate Days Since First Seen
	nowMilli := now.UnixMilli()
	if fsStr, err := cmdFirstSeen.Result(); err == nil && fsStr != "" {
		if fsMs, err := strconv.ParseInt(fsStr, 10, 64); err == nil && fsMs <= nowMilli {
			diffDays := float64(nowMilli-fsMs) / (1000.0 * 86400.0)
			defaultFeatures.DeviceDaysSinceFirstSeen = math.Round(diffDays*100) / 100
		}
	}

	events := cmdEvents.Val()
	if len(events) == 0 {
		return defaultFeatures, nil
	}

	cutoff30d := float64(now.Add(-30 * 24 * time.Hour).UnixMilli())

	var (
		totalTx, successTx, failedTx int64
		disputeCount, fraudCount     int64
		chargebackCount, refundCount int64
		recentDisputes, recentFraud  int64
		recentChargebacks            int64
		lastDisputeMs, lastFraudMs   int64 = -1, -1
	)

	// Idempotency Tracker for this point-in-time calculation
	seenTxOutcomes := make(map[string]bool)

	for _, z := range events {
		memberStr, ok := z.Member.(string)
		if !ok {
			continue
		}

		if seenTxOutcomes[memberStr] {
			continue
		}
		seenTxOutcomes[memberStr] = true

		eventTimeMs := int64(z.Score)

		// Member format: "<OUTCOME_TYPE>:<tx_id>"
		parts := strings.SplitN(memberStr, ":", 2)
		outcome := OutcomeType(parts[0])

		switch outcome {
		case OutcomeTransaction:
			totalTx++
		case OutcomeSuccess:
			totalTx++
			successTx++
		case OutcomeFailure:
			totalTx++
			failedTx++
		case OutcomeDispute:
			disputeCount++
			if z.Score >= cutoff30d {
				recentDisputes++
			}
			if eventTimeMs > lastDisputeMs {
				lastDisputeMs = eventTimeMs
			}
		case OutcomeFraud:
			fraudCount++
			if z.Score >= cutoff30d {
				recentFraud++
			}
			if eventTimeMs > lastFraudMs {
				lastFraudMs = eventTimeMs
			}
		case OutcomeChargeback:
			chargebackCount++
			disputeCount++
			if z.Score >= cutoff30d {
				recentChargebacks++
				recentDisputes++
			}
			if eventTimeMs > lastDisputeMs {
				lastDisputeMs = eventTimeMs
			}
		case OutcomeRefund:
			refundCount++
		}
	}

	f := &DeviceReputationFeatures{
		DeviceTotalTransactions:      totalTx,
		DeviceSuccessfulTransactions: successTx,
		DeviceFailedTransactions:     failedTx,
		DeviceDisputedTransactions:   disputeCount,
		DeviceFraudTransactions:      fraudCount,
		DeviceRefundedTransactions:   refundCount,
		DeviceChargebackCount:        chargebackCount,
		DeviceRecentDisputeCount:     recentDisputes,
		DeviceRecentFraudCount:       recentFraud,
		DeviceRecentChargebackCount:  recentChargebacks,
		DeviceDaysSinceFirstSeen:     defaultFeatures.DeviceDaysSinceFirstSeen,
		DeviceDaysSinceLastDispute:   -1.0,
		DeviceDaysSinceLastFraud:     -1.0,
		IsDegraded:                   false,
	}

	// Calculate Rates
	effectiveTotal := math.Max(1.0, float64(totalTx))
	f.DeviceDisputeRate = math.Round((float64(disputeCount)/effectiveTotal)*10000) / 10000
	f.DeviceFraudRate = math.Round((float64(fraudCount)/effectiveTotal)*10000) / 10000
	f.DeviceRefundRate = math.Round((float64(refundCount)/effectiveTotal)*10000) / 10000
	f.DeviceSuccessRate = math.Round((float64(successTx)/effectiveTotal)*10000) / 10000

	// Calculate Days Since Last Events
	if lastDisputeMs != -1 {
		diffDays := float64(nowMilli-lastDisputeMs) / (1000.0 * 86400.0)
		f.DeviceDaysSinceLastDispute = math.Max(0.0, math.Round(diffDays*100)/100)
	}
	if lastFraudMs != -1 {
		diffDays := float64(nowMilli-lastFraudMs) / (1000.0 * 86400.0)
		f.DeviceDaysSinceLastFraud = math.Max(0.0, math.Round(diffDays*100)/100)
	}

	// -------------------------------------------------------------
	// Deterministic Reputation Scoring Model
	// Score range: [0.0 (Very Trusted) - 1.0 (High Risk / Fraud)]
	// Neutral baseline: 0.50
	// -------------------------------------------------------------
	score := 0.50

	// 1. Trust Discounts
	if successTx > 0 {
		trustBenefit := math.Min(0.30, float64(successTx)*0.03) // Up to -0.30 at 10+ tx
		score -= trustBenefit
	}
	if f.DeviceDaysSinceFirstSeen > 0 {
		tenureBenefit := math.Min(0.15, f.DeviceDaysSinceFirstSeen*0.005) // Up to -0.15 at 30 days
		score -= tenureBenefit
	}

	// 2. Risk Penalties
	if fraudCount > 0 {
		score += 0.50 // Base fraud penalty
		if recentFraud > 0 {
			score += 0.30 // Recent fraud penalty (< 30 days)
		}
	}

	if disputeCount > 0 {
		disputePenalty := math.Min(0.50, float64(disputeCount)*0.25)
		score += disputePenalty
		if recentDisputes > 0 {
			score += math.Min(0.20, float64(recentDisputes)*0.10)
		}
	}

	if failedTx > 0 {
		failPenalty := math.Min(0.20, float64(failedTx)*0.05)
		score += failPenalty
	}

	if totalTx >= 3 && f.DeviceDisputeRate >= 0.10 {
		score += 0.20
	}

	// Clamp to [0.0, 1.0]
	if score < 0.0 {
		score = 0.0
	} else if score > 1.0 {
		score = 1.0
	}
	f.DeviceReputationScore = math.Round(score*1000) / 1000

	return f, nil
}

// RecordOutcome records an outcome event (e.g. SUCCESS, FAILURE, DISPUTE, FRAUD) idempotently for a device.
func (s *DeviceReputationStore) RecordOutcome(ctx context.Context, tenantID, deviceID, transactionID string, outcome OutcomeType) error {
	return s.RecordOutcomeAtTime(ctx, tenantID, deviceID, transactionID, outcome, time.Now().UTC())
}

// RecordOutcomeAtTime records an outcome event relative to an injected timestamp.
func (s *DeviceReputationStore) RecordOutcomeAtTime(ctx context.Context, tenantID, deviceID, transactionID string, outcome OutcomeType, now time.Time) error {
	if s == nil || s.client == nil || deviceID == "" || tenantID == "" || transactionID == "" {
		return nil
	}

	keyEvents := fmt.Sprintf("%s:rep:dev:events:%s", tenantID, deviceID)
	keyFirstSeen := fmt.Sprintf("%s:rep:dev:first_seen:%s", tenantID, deviceID)

	nowMs := float64(now.UnixMilli())
	member := fmt.Sprintf("%s:%s", outcome, transactionID)

	pipe := s.client.Pipeline()
	// Set first seen if not already set (NX = only if not exists)
	pipe.SetNX(ctx, keyFirstSeen, strconv.FormatInt(now.UnixMilli(), 10), 365*24*time.Hour)

	// Add to outcome ZSET with 90-day retention
	pipe.ZAdd(ctx, keyEvents, redis.Z{Score: nowMs, Member: member})
	ninetyDaysAgoMs := fmt.Sprintf("%f", float64(now.Add(-90*24*time.Hour).UnixMilli()))
	pipe.ZRemRangeByScore(ctx, keyEvents, "-inf", ninetyDaysAgoMs)
	pipe.Expire(ctx, keyEvents, 95*24*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record reputation outcome in redis: %w", err)
	}

	return nil
}
