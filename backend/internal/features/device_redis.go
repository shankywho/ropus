package features

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// DeviceFeatures holds real-time rolling metrics computed by the Redis feature store.
type DeviceFeatures struct {
	DeviceTxCount1m         int64  `json:"device_tx_count_1m"`
	DeviceTxCount1h         int64  `json:"device_tx_count_1h"`
	DeviceTxCount24h        int64  `json:"device_tx_count_24h"`
	DeviceAmountSum24h      int64  `json:"device_amount_sum_24h"`
	DeviceUniqueAccounts24h int64  `json:"device_unique_accounts_24h"`
	DeviceUniqueTokens24h   int64  `json:"device_unique_tokens_24h"`
	DeviceSeenBefore        int64  `json:"device_seen_before"` // 0 = New / First Seen, 1 = Known
	IsDegraded              bool   `json:"is_degraded"`
	DegradeReason           string `json:"degrade_reason,omitempty"`
}

// DeviceFeatureStore provides low-latency sliding-window device intelligence backed by Redis.
type DeviceFeatureStore struct {
	client redis.Cmdable
}

// NewDeviceFeatureStore constructs a new Redis-backed DeviceFeatureStore.
func NewDeviceFeatureStore(client redis.Cmdable) *DeviceFeatureStore {
	return &DeviceFeatureStore{client: client}
}

// GetDeviceFeatures retrieves all 7 device features in a single pipelined Redis round-trip.
// Point-in-Time Safety: This query returns the state BEFORE the current transaction is recorded.
func (s *DeviceFeatureStore) GetDeviceFeatures(ctx context.Context, tenantID, deviceID string) (*DeviceFeatures, error) {
	return s.GetDeviceFeaturesAtTime(ctx, tenantID, deviceID, time.Now().UTC())
}

// GetDeviceFeaturesAtTime executes the pipelined feature query relative to an injected reference timestamp.
func (s *DeviceFeatureStore) GetDeviceFeaturesAtTime(ctx context.Context, tenantID, deviceID string, now time.Time) (*DeviceFeatures, error) {
	defaultFeatures := &DeviceFeatures{
		DeviceTxCount1m:         0,
		DeviceTxCount1h:         0,
		DeviceTxCount24h:        0,
		DeviceAmountSum24h:      0,
		DeviceUniqueAccounts24h: 0,
		DeviceUniqueTokens24h:   0,
		DeviceSeenBefore:        0,
		IsDegraded:              false,
	}

	if s == nil || s.client == nil || deviceID == "" || tenantID == "" {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "DEVICE_FEATURE_STORE_UNAVAILABLE"
		return defaultFeatures, nil
	}

	key1m := fmt.Sprintf("%s:vel:dev:1m:%s", tenantID, deviceID)
	key1h := fmt.Sprintf("%s:vel:dev:1h:%s", tenantID, deviceID)
	key24h := fmt.Sprintf("%s:vel:dev:24h:%s", tenantID, deviceID)
	keyAcc := fmt.Sprintf("%s:dev:acc24:%s", tenantID, deviceID)
	keyTok := fmt.Sprintf("%s:dev:tok24:%s", tenantID, deviceID)
	keyKnown := fmt.Sprintf("%s:dev:known:%s", tenantID, deviceID)

	nowMs := strconv.FormatInt(now.UnixMilli(), 10)
	oneMinAgoMs := strconv.FormatInt(now.Add(-1*time.Minute).UnixMilli(), 10)
	oneHourAgoMs := strconv.FormatInt(now.Add(-1*time.Hour).UnixMilli(), 10)
	twentyFourHoursAgoMs := strconv.FormatInt(now.Add(-24*time.Hour).UnixMilli(), 10)

	pipe := s.client.Pipeline()

	// 1. Transaction counts across windows
	cmd1m := pipe.ZCount(ctx, key1m, oneMinAgoMs, nowMs)
	cmd1h := pipe.ZCount(ctx, key1h, oneHourAgoMs, nowMs)

	// 2. 24h events (to calculate both 24h count and 24h amount sum)
	cmd24h := pipe.ZRangeByScore(ctx, key24h, &redis.ZRangeBy{
		Min: twentyFourHoursAgoMs,
		Max: nowMs,
	})

	// 3. Prune and count distinct accounts in 24h window
	pipe.ZRemRangeByScore(ctx, keyAcc, "-inf", twentyFourHoursAgoMs)
	cmdAcc := pipe.ZCard(ctx, keyAcc)

	// 4. Prune and count distinct payment tokens in 24h window
	pipe.ZRemRangeByScore(ctx, keyTok, "-inf", twentyFourHoursAgoMs)
	cmdTok := pipe.ZCard(ctx, keyTok)

	// 5. Device novelty / first-seen flag
	cmdKnown := pipe.Exists(ctx, keyKnown)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "DEVICE_FEATURE_STORE_UNAVAILABLE"
		return defaultFeatures, fmt.Errorf("failed to fetch device features from redis pipeline: %w", err)
	}

	// Parse 1m and 1h transaction counts
	if cmd1m != nil {
		defaultFeatures.DeviceTxCount1m = cmd1m.Val()
	}
	if cmd1h != nil {
		defaultFeatures.DeviceTxCount1h = cmd1h.Val()
	}

	// Parse 24h transaction count and sum 24h amounts without float accumulation
	if cmd24h != nil {
		elements := cmd24h.Val()
		defaultFeatures.DeviceTxCount24h = int64(len(elements))
		var amountSum int64
		for _, elem := range elements {
			// Member format: "<timestamp_or_nonce>:<amount>"
			idx := strings.LastIndex(elem, ":")
			if idx != -1 && idx+1 < len(elem) {
				amtStr := elem[idx+1:]
				if amt, parseErr := strconv.ParseInt(amtStr, 10, 64); parseErr == nil {
					amountSum += amt
				}
			}
		}
		defaultFeatures.DeviceAmountSum24h = amountSum
	}

	// Parse distinct accounts and payment tokens
	if cmdAcc != nil {
		defaultFeatures.DeviceUniqueAccounts24h = cmdAcc.Val()
	}
	if cmdTok != nil {
		defaultFeatures.DeviceUniqueTokens24h = cmdTok.Val()
	}

	// Parse novelty
	if cmdKnown != nil && cmdKnown.Val() > 0 {
		defaultFeatures.DeviceSeenBefore = 1
	} else {
		defaultFeatures.DeviceSeenBefore = 0
	}

	return defaultFeatures, nil
}

// RecordDeviceTransaction updates rolling sliding windows and distinct entity sets in a single pipeline.
func (s *DeviceFeatureStore) RecordDeviceTransaction(ctx context.Context, tenantID, deviceID, transactionID string, amount int64, accountID, paymentToken string) error {
	return s.RecordDeviceTransactionAtTime(ctx, tenantID, deviceID, transactionID, amount, accountID, paymentToken, time.Now().UTC())
}

// RecordDeviceTransactionAtTime updates Redis relative to an injected reference timestamp.
func (s *DeviceFeatureStore) RecordDeviceTransactionAtTime(ctx context.Context, tenantID, deviceID, transactionID string, amount int64, accountID, paymentToken string, now time.Time) error {
	if s == nil || s.client == nil || deviceID == "" || tenantID == "" {
		return nil
	}

	nowMs := float64(now.UnixMilli())
	nowNano := now.UnixNano()

	key1m := fmt.Sprintf("%s:vel:dev:1m:%s", tenantID, deviceID)
	key1h := fmt.Sprintf("%s:vel:dev:1h:%s", tenantID, deviceID)
	key24h := fmt.Sprintf("%s:vel:dev:24h:%s", tenantID, deviceID)
	keyAcc := fmt.Sprintf("%s:dev:acc24:%s", tenantID, deviceID)
	keyTok := fmt.Sprintf("%s:dev:tok24:%s", tenantID, deviceID)
	keyKnown := fmt.Sprintf("%s:dev:known:%s", tenantID, deviceID)

	eventMember := fmt.Sprintf("%d_%s", nowNano, transactionID)
	amountMember := fmt.Sprintf("%d_%s:%d", nowNano, transactionID, amount)

	pipe := s.client.Pipeline()

	// 1. 1-Minute Window (Score = nowMs, Member = eventMember)
	pipe.ZAdd(ctx, key1m, redis.Z{Score: nowMs, Member: eventMember})
	twoMinAgoMs := fmt.Sprintf("%f", float64(now.Add(-2*time.Minute).UnixMilli()))
	pipe.ZRemRangeByScore(ctx, key1m, "-inf", twoMinAgoMs)
	pipe.Expire(ctx, key1m, 5*time.Minute)

	// 2. 1-Hour Window
	pipe.ZAdd(ctx, key1h, redis.Z{Score: nowMs, Member: eventMember})
	twoHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-2*time.Hour).UnixMilli()))
	pipe.ZRemRangeByScore(ctx, key1h, "-inf", twoHoursAgoMs)
	pipe.Expire(ctx, key1h, 3*time.Hour)

	// 3. 24-Hour Window (stores amount for rolling sum)
	pipe.ZAdd(ctx, key24h, redis.Z{Score: nowMs, Member: amountMember})
	twentyFiveHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-25*time.Hour).UnixMilli()))
	pipe.ZRemRangeByScore(ctx, key24h, "-inf", twentyFiveHoursAgoMs)
	pipe.Expire(ctx, key24h, 26*time.Hour)

	// 4. Distinct Accounts (Score = latest seen timestamp, Member = accountID)
	if accountID != "" {
		pipe.ZAdd(ctx, keyAcc, redis.Z{Score: nowMs, Member: accountID})
		pipe.ZRemRangeByScore(ctx, keyAcc, "-inf", twentyFiveHoursAgoMs)
		pipe.Expire(ctx, keyAcc, 26*time.Hour)
	}

	// 5. Distinct Payment Tokens (Score = latest seen timestamp, Member = tokenID)
	if cleanToken, ok := SanitizePaymentToken(paymentToken); ok {
		tokenID := HashPaymentToken(tenantID, cleanToken)
		pipe.ZAdd(ctx, keyTok, redis.Z{Score: nowMs, Member: tokenID})
		pipe.ZRemRangeByScore(ctx, keyTok, "-inf", twentyFiveHoursAgoMs)
		pipe.Expire(ctx, keyTok, 26*time.Hour)
	}

	// 6. Atomically mark device as known with 90-day rolling TTL
	pipe.Set(ctx, keyKnown, "1", 90*24*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record device transaction in redis: %w", err)
	}

	return nil
}
