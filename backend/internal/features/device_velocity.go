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

// VelocitySignal defines the risk severity classification for transaction/amount bursts.
type VelocitySignal string

const (
	VelocityNormal     VelocitySignal = "NORMAL"
	VelocityLowSignal  VelocitySignal = "LOW_SIGNAL"
	VelocitySuspicious VelocitySignal = "SUSPICIOUS"
	VelocityHighSignal VelocitySignal = "HIGH_SIGNAL"
)

// DeviceVelocityThresholds defines configurable limits for multi-window burst detection.
type DeviceVelocityThresholds struct {
	Tx10sLow               int64
	Tx10sSuspicious        int64
	Tx10sHigh              int64
	Tx1mLow                int64
	Tx1mSuspicious         int64
	Tx1mHigh               int64
	Tx5mLow                int64
	Tx5mSuspicious         int64
	Tx5mHigh               int64
	Amount5mLow            int64 // Minor currency units
	Amount5mSuspicious     int64
	Amount5mHigh           int64
	AccelerationSuspicious float64
	AccelerationHigh       float64
}

// GlobalVelocityThresholds contains initial baseline policy limits.
var GlobalVelocityThresholds = DeviceVelocityThresholds{
	Tx10sLow:               2,
	Tx10sSuspicious:        4,
	Tx10sHigh:              8,
	Tx1mLow:                4,
	Tx1mSuspicious:         8,
	Tx1mHigh:               15,
	Tx5mLow:                8,
	Tx5mSuspicious:         15,
	Tx5mHigh:               30,
	Amount5mLow:            50000,  // e.g. 500.00
	Amount5mSuspicious:     200000, // e.g. 2000.00
	Amount5mHigh:           500000, // e.g. 5000.00
	AccelerationSuspicious: 4.0,
	AccelerationHigh:       8.0,
}

// DeviceVelocityFeatures holds point-in-time multi-window velocity and anomaly metrics.
type DeviceVelocityFeatures struct {
	// Transaction Counts across 7 windows
	DeviceTxCount10s               int64          `json:"device_tx_count_10s"`
	DeviceTxCount1m                int64          `json:"device_tx_count_1m"`
	DeviceTxCount5m                int64          `json:"device_tx_count_5m"`
	DeviceTxCount15m               int64          `json:"device_tx_count_15m"`
	DeviceTxCount1h                int64          `json:"device_tx_count_1h"`
	DeviceTxCount6h                int64          `json:"device_tx_count_6h"`
	DeviceTxCount24h               int64          `json:"device_tx_count_24h"`

	// Amount Velocity (Sums in minor units)
	DeviceAmountSum10s             int64          `json:"device_amount_sum_10s"`
	DeviceAmountSum1m              int64          `json:"device_amount_sum_1m"`
	DeviceAmountSum5m              int64          `json:"device_amount_sum_5m"`
	DeviceAmountSum15m             int64          `json:"device_amount_sum_15m"`
	DeviceAmountSum1h              int64          `json:"device_amount_sum_1h"`
	DeviceAmountSum6h              int64          `json:"device_amount_sum_6h"`
	DeviceAmountSum24h             int64          `json:"device_amount_sum_24h"`

	// Average Transaction Amounts
	DeviceAvgAmount1m              float64        `json:"device_avg_amount_1m"`
	DeviceAvgAmount5m              float64        `json:"device_avg_amount_5m"`
	DeviceAvgAmount1h              float64        `json:"device_avg_amount_1h"`
	DeviceAvgAmount24h             float64        `json:"device_avg_amount_24h"`

	// Maximum Transaction Amounts
	DeviceMaxAmount1h              int64          `json:"device_max_amount_1h"`
	DeviceMaxAmount24h             int64          `json:"device_max_amount_24h"`

	// Velocity Rates (Transactions per second)
	DeviceTxRate10s                float64        `json:"device_tx_rate_10s"`
	DeviceTxRate1m                 float64        `json:"device_tx_rate_1m"`
	DeviceTxRate5m                 float64        `json:"device_tx_rate_5m"`
	DeviceTxRate15m                float64        `json:"device_tx_rate_15m"`
	DeviceTxRate1h                 float64        `json:"device_tx_rate_1h"`

	// Velocity Acceleration Ratios
	TxAcceleration1m15m            float64        `json:"tx_acceleration_1m_15m"`
	TxAcceleration5m1h             float64        `json:"tx_acceleration_5m_1h"`
	TxAcceleration15m1h            float64        `json:"tx_acceleration_15m_1h"`
	AmountAcceleration5m1h         float64        `json:"amount_acceleration_5m_1h"`
	AmountAcceleration15m1h        float64        `json:"amount_acceleration_15m_1h"`

	// Amount Concentration Ratios [0.0 - 1.0]
	DeviceAmountConcentration5m1h  float64        `json:"device_amount_concentration_5m_1h"`
	DeviceAmountConcentration15m24h float64       `json:"device_amount_concentration_15m_24h"`

	// Anomaly Classification
	VelocitySignal                 VelocitySignal `json:"velocity_signal"`

	// Reliability / Graceful Degradation
	IsDegraded                     bool           `json:"is_degraded"`
	DegradeReason                  string         `json:"degrade_reason,omitempty"`
}

// DeviceVelocityStore manages multi-window device velocity in Redis.
type DeviceVelocityStore struct {
	client redis.Cmdable
}

// NewDeviceVelocityStore constructs a new DeviceVelocityStore.
func NewDeviceVelocityStore(client redis.Cmdable) *DeviceVelocityStore {
	return &DeviceVelocityStore{client: client}
}

// GetVelocityFeatures retrieves point-in-time multi-window velocity metrics in a single pipelined Redis call.
func (s *DeviceVelocityStore) GetVelocityFeatures(ctx context.Context, tenantID, deviceID string) (*DeviceVelocityFeatures, error) {
	return s.GetVelocityFeaturesAtTime(ctx, tenantID, deviceID, time.Now().UTC())
}

// GetVelocityFeaturesAtTime retrieves velocity features relative to an injected reference timestamp.
func (s *DeviceVelocityStore) GetVelocityFeaturesAtTime(ctx context.Context, tenantID, deviceID string, now time.Time) (*DeviceVelocityFeatures, error) {
	defaultFeatures := &DeviceVelocityFeatures{
		VelocitySignal: VelocityNormal,
		IsDegraded:     false,
	}

	if s == nil || s.client == nil || deviceID == "" || tenantID == "" {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "VELOCITY_FEATURE_STORE_UNAVAILABLE"
		return defaultFeatures, nil
	}

	keyEvents := fmt.Sprintf("%s:vel:dev:events:%s", tenantID, deviceID)

	nowMs := strconv.FormatInt(now.UnixMilli(), 10)
	twentyFourHoursAgoMs := strconv.FormatInt(now.Add(-24*time.Hour).UnixMilli(), 10)
	twentyFiveHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-25*time.Hour).UnixMilli()))

	pipe := s.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, keyEvents, "-inf", twentyFiveHoursAgoMs)
	cmdEvents := pipe.ZRangeByScore(ctx, keyEvents, &redis.ZRangeBy{
		Min: twentyFourHoursAgoMs,
		Max: nowMs,
	})

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "VELOCITY_FEATURE_STORE_UNAVAILABLE"
		return defaultFeatures, fmt.Errorf("failed to fetch velocity events from redis: %w", err)
	}

	events := cmdEvents.Val()
	nowMilli := now.UnixMilli()

	cutoff10s := nowMilli - 10*1000
	cutoff1m := nowMilli - 60*1000
	cutoff5m := nowMilli - 5*60*1000
	cutoff15m := nowMilli - 15*60*1000
	cutoff1h := nowMilli - 60*60*1000
	cutoff6h := nowMilli - 6*60*60*1000
	cutoff24h := nowMilli - 24*60*60*1000

	var (
		c10s, c1m, c5m, c15m, c1h, c6h, c24h int64
		s10s, s1m, s5m, s15m, s1h, s6h, s24h int64
		max1h, max24h                        int64
	)

	for _, member := range events {
		// Member format: "<nano>_<tx_id>:<amount>" or "<nano>_<tx_id>"
		var eventTimeMs int64
		var amount int64

		idxUnderscore := strings.Index(member, "_")
		if idxUnderscore != -1 {
			nanoStr := member[:idxUnderscore]
			if nanoVal, err := strconv.ParseInt(nanoStr, 10, 64); err == nil {
				eventTimeMs = nanoVal / 1e6
			}
		}

		idxColon := strings.LastIndex(member, ":")
		if idxColon != -1 && idxColon+1 < len(member) {
			if amt, err := strconv.ParseInt(member[idxColon+1:], 10, 64); err == nil {
				amount = amt
			}
		}

		if eventTimeMs < cutoff24h {
			continue
		}

		// 24h Window
		c24h++
		s24h += amount
		if amount > max24h {
			max24h = amount
		}

		// 6h Window
		if eventTimeMs >= cutoff6h {
			c6h++
			s6h += amount
		}

		// 1h Window
		if eventTimeMs >= cutoff1h {
			c1h++
			s1h += amount
			if amount > max1h {
				max1h = amount
			}
		}

		// 15m Window
		if eventTimeMs >= cutoff15m {
			c15m++
			s15m += amount
		}

		// 5m Window
		if eventTimeMs >= cutoff5m {
			c5m++
			s5m += amount
		}

		// 1m Window
		if eventTimeMs >= cutoff1m {
			c1m++
			s1m += amount
		}

		// 10s Window
		if eventTimeMs >= cutoff10s {
			c10s++
			s10s += amount
		}
	}

	// Populate Counts and Sums
	f := &DeviceVelocityFeatures{
		DeviceTxCount10s:   c10s,
		DeviceTxCount1m:    c1m,
		DeviceTxCount5m:    c5m,
		DeviceTxCount15m:   c15m,
		DeviceTxCount1h:    c1h,
		DeviceTxCount6h:    c6h,
		DeviceTxCount24h:   c24h,
		DeviceAmountSum10s: s10s,
		DeviceAmountSum1m:  s1m,
		DeviceAmountSum5m:  s5m,
		DeviceAmountSum15m: s15m,
		DeviceAmountSum1h:  s1h,
		DeviceAmountSum6h:  s6h,
		DeviceAmountSum24h: s24h,
		DeviceMaxAmount1h:  max1h,
		DeviceMaxAmount24h: max24h,
		IsDegraded:         false,
	}

	// Calculate Average Amounts
	if c1m > 0 {
		f.DeviceAvgAmount1m = float64(s1m) / float64(c1m)
	}
	if c5m > 0 {
		f.DeviceAvgAmount5m = float64(s5m) / float64(c5m)
	}
	if c1h > 0 {
		f.DeviceAvgAmount1h = float64(s1h) / float64(c1h)
	}
	if c24h > 0 {
		f.DeviceAvgAmount24h = float64(s24h) / float64(c24h)
	}

	// Calculate Transaction Rates (tx / sec)
	f.DeviceTxRate10s = float64(c10s) / 10.0
	f.DeviceTxRate1m = float64(c1m) / 60.0
	f.DeviceTxRate5m = float64(c5m) / 300.0
	f.DeviceTxRate15m = float64(c15m) / 900.0
	f.DeviceTxRate1h = float64(c1h) / 3600.0

	// Calculate Acceleration Ratios (clamped to 1000.0 with epsilon = 0.001)
	const (
		eps      = 0.001
		maxClamp = 1000.0
	)
	calcAccel := func(shortVal, longVal, ratio float64) float64 {
		expected := longVal / ratio
		if expected < eps {
			expected = eps
		}
		res := shortVal / expected
		if math.IsNaN(res) || math.IsInf(res, 0) || res > maxClamp {
			return maxClamp
		}
		if res < 0 {
			return 0
		}
		return math.Round(res*100) / 100
	}

	f.TxAcceleration1m15m = calcAccel(float64(c1m), float64(c15m), 15.0)
	f.TxAcceleration5m1h = calcAccel(float64(c5m), float64(c1h), 12.0)
	f.TxAcceleration15m1h = calcAccel(float64(c15m), float64(c1h), 4.0)

	f.AmountAcceleration5m1h = calcAccel(float64(s5m), float64(s1h), 12.0)
	f.AmountAcceleration15m1h = calcAccel(float64(s15m), float64(s1h), 4.0)

	// Calculate Amount Concentration Ratios [0.0 - 1.0]
	if s1h > 0 {
		f.DeviceAmountConcentration5m1h = math.Min(1.0, float64(s5m)/float64(s1h))
	}
	if s24h > 0 {
		f.DeviceAmountConcentration15m24h = math.Min(1.0, float64(s15m)/float64(s24h))
	}

	// Burst Severity Classification
	t := GlobalVelocityThresholds
	switch {
	case c10s >= t.Tx10sHigh || c1m >= t.Tx1mHigh || c5m >= t.Tx5mHigh || s5m >= t.Amount5mHigh || f.TxAcceleration5m1h >= t.AccelerationHigh:
		f.VelocitySignal = VelocityHighSignal
	case c10s >= t.Tx10sSuspicious || c1m >= t.Tx1mSuspicious || c5m >= t.Tx5mSuspicious || s5m >= t.Amount5mSuspicious || f.TxAcceleration5m1h >= t.AccelerationSuspicious:
		f.VelocitySignal = VelocitySuspicious
	case c10s >= t.Tx10sLow || c1m >= t.Tx1mLow || c5m >= t.Tx5mLow || s5m >= t.Amount5mLow:
		f.VelocitySignal = VelocityLowSignal
	default:
		f.VelocitySignal = VelocityNormal
	}

	return f, nil
}

// RecordVelocityTransaction records a point-in-time transaction event into the device velocity ledger.
func (s *DeviceVelocityStore) RecordVelocityTransaction(ctx context.Context, tenantID, deviceID, transactionID string, amount int64) error {
	return s.RecordVelocityTransactionAtTime(ctx, tenantID, deviceID, transactionID, amount, time.Now().UTC())
}

// RecordVelocityTransactionAtTime records the transaction relative to an injected reference timestamp.
func (s *DeviceVelocityStore) RecordVelocityTransactionAtTime(ctx context.Context, tenantID, deviceID, transactionID string, amount int64, now time.Time) error {
	if s == nil || s.client == nil || deviceID == "" || tenantID == "" {
		return nil
	}

	nowMs := float64(now.UnixMilli())
	nowNano := now.UnixNano()

	keyEvents := fmt.Sprintf("%s:vel:dev:events:%s", tenantID, deviceID)
	eventMember := fmt.Sprintf("%d_%s:%d", nowNano, transactionID, amount)
	twentyFiveHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-25*time.Hour).UnixMilli()))

	pipe := s.client.Pipeline()
	pipe.ZAdd(ctx, keyEvents, redis.Z{Score: nowMs, Member: eventMember})
	pipe.ZRemRangeByScore(ctx, keyEvents, "-inf", twentyFiveHoursAgoMs)
	pipe.Expire(ctx, keyEvents, 26*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record velocity transaction in redis: %w", err)
	}

	return nil
}
