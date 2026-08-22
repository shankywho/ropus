package features

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/redis/go-redis/v9"
)

// CardTestingSignal defines the risk severity classification for card testing bursts.
type CardTestingSignal string

const (
	CardTestingNormal     CardTestingSignal = "NORMAL"
	CardTestingLowSignal  CardTestingSignal = "LOW_SIGNAL"
	CardTestingSuspicious CardTestingSignal = "SUSPICIOUS"
	CardTestingHighSignal CardTestingSignal = "HIGH_SIGNAL"
)

// TokenFanOutSignal defines the risk severity classification for token fan-out across devices.
type TokenFanOutSignal string

const (
	TokenFanOutNormal     TokenFanOutSignal = "NORMAL"
	TokenFanOutSuspicious TokenFanOutSignal = "SUSPICIOUS"
	TokenFanOutHighSignal TokenFanOutSignal = "HIGH_SIGNAL"
)

// DefaultCardTestingThresholds defines initial conservative policy thresholds for card testing detection.
type CardTestingThresholds struct {
	LowUniqueTokens5m        int64
	SuspiciousUniqueTokens5m int64
	SuspiciousUniqueTokens1h int64
	HighUniqueTokens5m       int64
	HighUniqueTokens1h       int64
	HighTxAcrossTokens1h     int64
}

// GlobalCardTestingThresholds contains the initial configurable policy limits.
var GlobalCardTestingThresholds = CardTestingThresholds{
	LowUniqueTokens5m:        3,
	SuspiciousUniqueTokens5m: 5,
	SuspiciousUniqueTokens1h: 8,
	HighUniqueTokens5m:       10,
	HighUniqueTokens1h:       15,
	HighTxAcrossTokens1h:     20,
}

// PaymentTokenFeatures represents point-in-time device ↔ payment token features computed by Redis.
type PaymentTokenFeatures struct {
	// Device -> Payment Token Metrics
	DeviceUniqueTokens5m      int64             `json:"device_unique_tokens_5m"`
	DeviceUniqueTokens1h      int64             `json:"device_unique_tokens_1h"`
	DeviceUniqueTokens24h     int64             `json:"device_unique_tokens_24h"`
	DeviceTokenTxCount5m      int64             `json:"device_token_tx_count_5m"`
	DeviceTokenTxCount1h      int64             `json:"device_token_tx_count_1h"`
	DeviceTokenTxCount24h     int64             `json:"device_token_tx_count_24h"`
	DeviceTokenAmountSum24h   int64             `json:"device_token_amount_sum_24h"`
	CardTestingSignal         CardTestingSignal `json:"card_testing_signal"`

	// Payment Token -> Device Metrics
	TokenUniqueDevices1h      int64             `json:"token_unique_devices_1h"`
	TokenUniqueDevices24h     int64             `json:"token_unique_devices_24h"`
	TokenTxCount1h            int64             `json:"token_tx_count_1h"`
	TokenTxCount24h           int64             `json:"token_tx_count_24h"`
	TokenFanOutSignal         TokenFanOutSignal `json:"token_fan_out_signal"`

	// Pair Metrics
	DeviceTokenSeenBefore     int64             `json:"device_token_seen_before"` // 0 = New Link, 1 = Existing Link
	DeviceTokenTxCount        int64             `json:"device_token_tx_count"`

	// Degradation / Safety
	IsDegraded                bool              `json:"is_degraded"`
	DegradeReason             string            `json:"degrade_reason,omitempty"`
}

var rawPANRegex = regexp.MustCompile(`^[0-9]{13,19}$`)

// SanitizePaymentToken validates and sanitizes incoming payment token identifiers.
// Rejects empty tokens, oversized tokens (>256 chars), control characters, embedded null bytes,
// and raw unmasked primary account numbers (PANs) to prevent sensitive credential storage.
func SanitizePaymentToken(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	if len(trimmed) > 256 {
		return "", false
	}
	if !utf8.ValidString(trimmed) {
		return "", false
	}
	for _, r := range trimmed {
		if r == 0 || (unicode.IsControl(r) && r != '\t') {
			return "", false
		}
	}
	// Never accept or process raw card PANs
	if rawPANRegex.MatchString(trimmed) {
		return "", false
	}
	return trimmed, true
}

// HashPaymentToken produces a tenant-isolated cryptographic hash for a payment token identifier.
// Raw tokens are never used as Redis keys or logged in plain text.
func HashPaymentToken(tenantID, canonicalToken string) string {
	hasher := sha256.New()
	hasher.Write([]byte(tenantID))
	hasher.Write([]byte(":"))
	hasher.Write([]byte(canonicalToken))
	return hex.EncodeToString(hasher.Sum(nil))
}

// PaymentTokenStore provides low-latency Redis queries and updates for payment token intelligence.
type PaymentTokenStore struct {
	client redis.Cmdable
}

// NewPaymentTokenStore constructs a new PaymentTokenStore.
func NewPaymentTokenStore(client redis.Cmdable) *PaymentTokenStore {
	return &PaymentTokenStore{client: client}
}

// GetPaymentTokenFeatures retrieves real-time point-in-time payment token features in a single pipelined Redis round-trip.
func (s *PaymentTokenStore) GetPaymentTokenFeatures(ctx context.Context, tenantID, deviceID, rawToken string) (*PaymentTokenFeatures, error) {
	return s.GetPaymentTokenFeaturesAtTime(ctx, tenantID, deviceID, rawToken, time.Now().UTC())
}

// GetPaymentTokenFeaturesAtTime executes the pipelined query relative to an injected reference timestamp.
func (s *PaymentTokenStore) GetPaymentTokenFeaturesAtTime(ctx context.Context, tenantID, deviceID, rawToken string, now time.Time) (*PaymentTokenFeatures, error) {
	defaultFeatures := &PaymentTokenFeatures{
		DeviceUniqueTokens5m:    0,
		DeviceUniqueTokens1h:    0,
		DeviceUniqueTokens24h:   0,
		DeviceTokenTxCount5m:    0,
		DeviceTokenTxCount1h:    0,
		DeviceTokenTxCount24h:   0,
		DeviceTokenAmountSum24h: 0,
		CardTestingSignal:       CardTestingNormal,
		TokenUniqueDevices1h:    0,
		TokenUniqueDevices24h:   0,
		TokenTxCount1h:          0,
		TokenTxCount24h:         0,
		TokenFanOutSignal:       TokenFanOutNormal,
		DeviceTokenSeenBefore:   0,
		DeviceTokenTxCount:      0,
		IsDegraded:              false,
	}

	if s == nil || s.client == nil || deviceID == "" || tenantID == "" {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "PAYMENT_TOKEN_FEATURE_STORE_UNAVAILABLE"
		return defaultFeatures, nil
	}

	cleanToken, hasToken := SanitizePaymentToken(rawToken)
	var tokenID string
	if hasToken {
		tokenID = HashPaymentToken(tenantID, cleanToken)
	}

	keyDevTok5m := fmt.Sprintf("%s:dev:tok5m:%s", tenantID, deviceID)
	keyDevTok1h := fmt.Sprintf("%s:dev:tok1h:%s", tenantID, deviceID)
	keyDevTok24h := fmt.Sprintf("%s:dev:tok24:%s", tenantID, deviceID)
	keyDevTx5m := fmt.Sprintf("%s:dev:tok_tx5m:%s", tenantID, deviceID)
	keyDevTx1h := fmt.Sprintf("%s:dev:tok_tx1h:%s", tenantID, deviceID)
	keyDevTx24h := fmt.Sprintf("%s:dev:tok_tx24h:%s", tenantID, deviceID)
	keyDevAmt24h := fmt.Sprintf("%s:dev:tok_amt24h:%s", tenantID, deviceID)

	var keyTokDev1h, keyTokDev24h, keyTokTx1h, keyTokTx24h, keyRelKnown string
	if hasToken {
		keyTokDev1h = fmt.Sprintf("%s:tok:dev1h:%s", tenantID, tokenID)
		keyTokDev24h = fmt.Sprintf("%s:tok:dev24:%s", tenantID, tokenID)
		keyTokTx1h = fmt.Sprintf("%s:tok:tx1h:%s", tenantID, tokenID)
		keyTokTx24h = fmt.Sprintf("%s:tok:tx24h:%s", tenantID, tokenID)
		keyRelKnown = fmt.Sprintf("%s:dev_tok:known:%s:%s", tenantID, deviceID, tokenID)
	}

	nowMs := strconv.FormatInt(now.UnixMilli(), 10)
	fiveMinAgoMs := strconv.FormatInt(now.Add(-5*time.Minute).UnixMilli(), 10)
	oneHourAgoMs := strconv.FormatInt(now.Add(-1*time.Hour).UnixMilli(), 10)
	twentyFourHoursAgoMs := strconv.FormatInt(now.Add(-24*time.Hour).UnixMilli(), 10)

	pipe := s.client.Pipeline()

	// 1. Device -> Unique Token Sets
	pipe.ZRemRangeByScore(ctx, keyDevTok5m, "-inf", fiveMinAgoMs)
	cmdDevTok5m := pipe.ZCard(ctx, keyDevTok5m)

	pipe.ZRemRangeByScore(ctx, keyDevTok1h, "-inf", oneHourAgoMs)
	cmdDevTok1h := pipe.ZCard(ctx, keyDevTok1h)

	pipe.ZRemRangeByScore(ctx, keyDevTok24h, "-inf", twentyFourHoursAgoMs)
	cmdDevTok24h := pipe.ZCard(ctx, keyDevTok24h)

	// 2. Device -> Transaction Count Windows
	pipe.ZRemRangeByScore(ctx, keyDevTx5m, "-inf", fiveMinAgoMs)
	cmdDevTx5m := pipe.ZCard(ctx, keyDevTx5m)

	pipe.ZRemRangeByScore(ctx, keyDevTx1h, "-inf", oneHourAgoMs)
	cmdDevTx1h := pipe.ZCard(ctx, keyDevTx1h)

	pipe.ZRemRangeByScore(ctx, keyDevTx24h, "-inf", twentyFourHoursAgoMs)
	cmdDevTx24h := pipe.ZCard(ctx, keyDevTx24h)

	// 3. Device -> 24h Amount Sum
	pipe.ZRemRangeByScore(ctx, keyDevAmt24h, "-inf", twentyFourHoursAgoMs)
	cmdDevAmt24h := pipe.ZRangeByScore(ctx, keyDevAmt24h, &redis.ZRangeBy{
		Min: twentyFourHoursAgoMs,
		Max: nowMs,
	})

	// 4. Token -> Device & Transaction Windows (if token present)
	var cmdTokDev1h, cmdTokDev24h, cmdTokTx1h, cmdTokTx24h, cmdRelKnown *redis.IntCmd
	if hasToken {
		pipe.ZRemRangeByScore(ctx, keyTokDev1h, "-inf", oneHourAgoMs)
		cmdTokDev1h = pipe.ZCard(ctx, keyTokDev1h)

		pipe.ZRemRangeByScore(ctx, keyTokDev24h, "-inf", twentyFourHoursAgoMs)
		cmdTokDev24h = pipe.ZCard(ctx, keyTokDev24h)

		pipe.ZRemRangeByScore(ctx, keyTokTx1h, "-inf", oneHourAgoMs)
		cmdTokTx1h = pipe.ZCard(ctx, keyTokTx1h)

		pipe.ZRemRangeByScore(ctx, keyTokTx24h, "-inf", twentyFourHoursAgoMs)
		cmdTokTx24h = pipe.ZCard(ctx, keyTokTx24h)

		cmdRelKnown = pipe.Exists(ctx, keyRelKnown)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "PAYMENT_TOKEN_FEATURE_STORE_UNAVAILABLE"
		return defaultFeatures, fmt.Errorf("failed to execute payment token feature pipeline: %w", err)
	}

	// Parse Device-centric metrics
	if cmdDevTok5m != nil {
		defaultFeatures.DeviceUniqueTokens5m = cmdDevTok5m.Val()
	}
	if cmdDevTok1h != nil {
		defaultFeatures.DeviceUniqueTokens1h = cmdDevTok1h.Val()
	}
	if cmdDevTok24h != nil {
		defaultFeatures.DeviceUniqueTokens24h = cmdDevTok24h.Val()
	}
	if cmdDevTx5m != nil {
		defaultFeatures.DeviceTokenTxCount5m = cmdDevTx5m.Val()
	}
	if cmdDevTx1h != nil {
		defaultFeatures.DeviceTokenTxCount1h = cmdDevTx1h.Val()
	}
	if cmdDevTx24h != nil {
		defaultFeatures.DeviceTokenTxCount24h = cmdDevTx24h.Val()
	}

	// Calculate 24h Amount Sum
	if cmdDevAmt24h != nil {
		var amtSum int64
		for _, member := range cmdDevAmt24h.Val() {
			idxColon := strings.LastIndex(member, ":")
			if idxColon != -1 && idxColon+1 < len(member) {
				amt, err := strconv.ParseInt(member[idxColon+1:], 10, 64)
				if err == nil && amt > 0 {
					amtSum += amt
				}
			}
		}
		defaultFeatures.DeviceTokenAmountSum24h = amtSum
	}

	// Parse Token-centric metrics
	if hasToken {
		if cmdTokDev1h != nil {
			defaultFeatures.TokenUniqueDevices1h = cmdTokDev1h.Val()
		}
		if cmdTokDev24h != nil {
			defaultFeatures.TokenUniqueDevices24h = cmdTokDev24h.Val()
		}
		if cmdTokTx1h != nil {
			defaultFeatures.TokenTxCount1h = cmdTokTx1h.Val()
		}
		if cmdTokTx24h != nil {
			defaultFeatures.TokenTxCount24h = cmdTokTx24h.Val()
		}
		if cmdRelKnown != nil && cmdRelKnown.Val() > 0 {
			defaultFeatures.DeviceTokenSeenBefore = 1
		} else {
			defaultFeatures.DeviceTokenSeenBefore = 0
		}
	}

	// Card-Testing Severity Classification
	t := GlobalCardTestingThresholds
	switch {
	case defaultFeatures.DeviceUniqueTokens5m >= t.HighUniqueTokens5m ||
		defaultFeatures.DeviceUniqueTokens1h >= t.HighUniqueTokens1h ||
		defaultFeatures.DeviceTokenTxCount1h >= t.HighTxAcrossTokens1h:
		defaultFeatures.CardTestingSignal = CardTestingHighSignal
	case defaultFeatures.DeviceUniqueTokens5m >= t.SuspiciousUniqueTokens5m ||
		defaultFeatures.DeviceUniqueTokens1h >= t.SuspiciousUniqueTokens1h:
		defaultFeatures.CardTestingSignal = CardTestingSuspicious
	case defaultFeatures.DeviceUniqueTokens5m >= t.LowUniqueTokens5m:
		defaultFeatures.CardTestingSignal = CardTestingLowSignal
	default:
		defaultFeatures.CardTestingSignal = CardTestingNormal
	}

	// Token Fan-Out Severity Classification
	switch {
	case defaultFeatures.TokenUniqueDevices1h >= 5 || defaultFeatures.TokenUniqueDevices24h >= 10:
		defaultFeatures.TokenFanOutSignal = TokenFanOutHighSignal
	case defaultFeatures.TokenUniqueDevices1h >= 3 || defaultFeatures.TokenUniqueDevices24h >= 5:
		defaultFeatures.TokenFanOutSignal = TokenFanOutSuspicious
	default:
		defaultFeatures.TokenFanOutSignal = TokenFanOutNormal
	}

	return defaultFeatures, nil
}

// RecordPaymentTokenTransaction records a point-in-time transaction event into Redis.
func (s *PaymentTokenStore) RecordPaymentTokenTransaction(ctx context.Context, tenantID, deviceID, rawToken, transactionID string, amount int64) error {
	return s.RecordPaymentTokenTransactionAtTime(ctx, tenantID, deviceID, rawToken, transactionID, amount, time.Now().UTC())
}

// RecordPaymentTokenTransactionAtTime records the transaction relative to an injected reference timestamp.
func (s *PaymentTokenStore) RecordPaymentTokenTransactionAtTime(ctx context.Context, tenantID, deviceID, rawToken, transactionID string, amount int64, now time.Time) error {
	if s == nil || s.client == nil || deviceID == "" || tenantID == "" {
		return nil
	}

	cleanToken, hasToken := SanitizePaymentToken(rawToken)
	if !hasToken {
		return nil
	}
	tokenID := HashPaymentToken(tenantID, cleanToken)

	nowMs := float64(now.UnixMilli())
	nowMilliInt := now.UnixMilli()

	keyDevTok5m := fmt.Sprintf("%s:dev:tok5m:%s", tenantID, deviceID)
	keyDevTok1h := fmt.Sprintf("%s:dev:tok1h:%s", tenantID, deviceID)
	keyDevTok24h := fmt.Sprintf("%s:dev:tok24:%s", tenantID, deviceID)
	keyDevTx5m := fmt.Sprintf("%s:dev:tok_tx5m:%s", tenantID, deviceID)
	keyDevTx1h := fmt.Sprintf("%s:dev:tok_tx1h:%s", tenantID, deviceID)
	keyDevTx24h := fmt.Sprintf("%s:dev:tok_tx24h:%s", tenantID, deviceID)
	keyDevAmt24h := fmt.Sprintf("%s:dev:tok_amt24h:%s", tenantID, deviceID)

	keyTokDev1h := fmt.Sprintf("%s:tok:dev1h:%s", tenantID, tokenID)
	keyTokDev24h := fmt.Sprintf("%s:tok:dev24:%s", tenantID, tokenID)
	keyTokTx1h := fmt.Sprintf("%s:tok:tx1h:%s", tenantID, tokenID)
	keyTokTx24h := fmt.Sprintf("%s:tok:tx24h:%s", tenantID, tokenID)
	keyRelKnown := fmt.Sprintf("%s:dev_tok:known:%s:%s", tenantID, deviceID, tokenID)

	pipe := s.client.Pipeline()

	tenMinAgoMs := fmt.Sprintf("%f", float64(now.Add(-10*time.Minute).UnixMilli()))
	twoHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-2*time.Hour).UnixMilli()))
	twentyFiveHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-25*time.Hour).UnixMilli()))

	// 1. Device -> Token Unique Sets
	pipe.ZAdd(ctx, keyDevTok5m, redis.Z{Score: nowMs, Member: tokenID})
	pipe.ZRemRangeByScore(ctx, keyDevTok5m, "-inf", tenMinAgoMs)
	pipe.Expire(ctx, keyDevTok5m, 15*time.Minute)

	pipe.ZAdd(ctx, keyDevTok1h, redis.Z{Score: nowMs, Member: tokenID})
	pipe.ZRemRangeByScore(ctx, keyDevTok1h, "-inf", twoHoursAgoMs)
	pipe.Expire(ctx, keyDevTok1h, 3*time.Hour)

	pipe.ZAdd(ctx, keyDevTok24h, redis.Z{Score: nowMs, Member: tokenID})
	pipe.ZRemRangeByScore(ctx, keyDevTok24h, "-inf", twentyFiveHoursAgoMs)
	pipe.Expire(ctx, keyDevTok24h, 26*time.Hour)

	// 2. Device -> Transaction Count Windows
	txMember := fmt.Sprintf("%d_%s", nowMilliInt, transactionID)
	pipe.ZAdd(ctx, keyDevTx5m, redis.Z{Score: nowMs, Member: txMember})
	pipe.ZRemRangeByScore(ctx, keyDevTx5m, "-inf", tenMinAgoMs)
	pipe.Expire(ctx, keyDevTx5m, 15*time.Minute)

	pipe.ZAdd(ctx, keyDevTx1h, redis.Z{Score: nowMs, Member: txMember})
	pipe.ZRemRangeByScore(ctx, keyDevTx1h, "-inf", twoHoursAgoMs)
	pipe.Expire(ctx, keyDevTx1h, 3*time.Hour)

	pipe.ZAdd(ctx, keyDevTx24h, redis.Z{Score: nowMs, Member: txMember})
	pipe.ZRemRangeByScore(ctx, keyDevTx24h, "-inf", twentyFiveHoursAgoMs)
	pipe.Expire(ctx, keyDevTx24h, 26*time.Hour)

	// 3. Device -> Amount Sum 24h
	amtMember := fmt.Sprintf("%d_%s:%d", nowMilliInt, transactionID, amount)
	pipe.ZAdd(ctx, keyDevAmt24h, redis.Z{Score: nowMs, Member: amtMember})
	pipe.ZRemRangeByScore(ctx, keyDevAmt24h, "-inf", twentyFiveHoursAgoMs)
	pipe.Expire(ctx, keyDevAmt24h, 26*time.Hour)

	// 4. Token -> Device Sets & Tx Windows
	pipe.ZAdd(ctx, keyTokDev1h, redis.Z{Score: nowMs, Member: deviceID})
	pipe.ZRemRangeByScore(ctx, keyTokDev1h, "-inf", twoHoursAgoMs)
	pipe.Expire(ctx, keyTokDev1h, 3*time.Hour)

	pipe.ZAdd(ctx, keyTokDev24h, redis.Z{Score: nowMs, Member: deviceID})
	pipe.ZRemRangeByScore(ctx, keyTokDev24h, "-inf", twentyFiveHoursAgoMs)
	pipe.Expire(ctx, keyTokDev24h, 26*time.Hour)

	pipe.ZAdd(ctx, keyTokTx1h, redis.Z{Score: nowMs, Member: txMember})
	pipe.ZRemRangeByScore(ctx, keyTokTx1h, "-inf", twoHoursAgoMs)
	pipe.Expire(ctx, keyTokTx1h, 3*time.Hour)

	pipe.ZAdd(ctx, keyTokTx24h, redis.Z{Score: nowMs, Member: txMember})
	pipe.ZRemRangeByScore(ctx, keyTokTx24h, "-inf", twentyFiveHoursAgoMs)
	pipe.Expire(ctx, keyTokTx24h, 26*time.Hour)

	// 5. Mark (device, token) pair known for 90 days
	pipe.Set(ctx, keyRelKnown, "1", 90*24*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record payment token transaction in redis: %w", err)
	}

	return nil
}
