package features

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/redis/go-redis/v9"
)

// MultiAccountSignal defines the risk severity classification for account-device clustering.
type MultiAccountSignal string

const (
	MultiAccountNormal     MultiAccountSignal = "NORMAL"
	MultiAccountLowSignal  MultiAccountSignal = "LOW_SIGNAL"
	MultiAccountSuspicious MultiAccountSignal = "SUSPICIOUS"
	MultiAccountHighSignal MultiAccountSignal = "HIGH_SIGNAL"
)

// AccountDeviceGraphFeatures holds real-time relationship graph signals computed by Redis.
type AccountDeviceGraphFeatures struct {
	// Device-centric features
	DeviceUniqueAccounts1h         int64              `json:"device_unique_accounts_1h"`
	DeviceUniqueAccounts24h        int64              `json:"device_unique_accounts_24h"`
	DeviceAccountSwitches1h        int64              `json:"device_account_switches_1h"`
	DeviceAccountSwitches24h       int64              `json:"device_account_switches_24h"`
	DeviceNewAccountOnKnownDevice  int64              `json:"device_new_account_on_known_device"` // 1 if known device + brand new account
	MultiAccountSignal             MultiAccountSignal `json:"multi_account_signal"`

	// Account-centric features
	AccountUniqueDevices1h         int64              `json:"account_unique_devices_1h"`
	AccountUniqueDevices24h        int64              `json:"account_unique_devices_24h"`
	AccountNewDevice1h             int64              `json:"account_new_device_1h"`
	AccountDeviceSwitches24h       int64              `json:"account_device_switches_24h"`

	// Specific relationship-centric features
	DeviceAccountSeenBefore        int64              `json:"device_account_seen_before"` // 0 = New Link, 1 = Existing Link
	DeviceAccountAgeDays           float64            `json:"device_account_age_days"`
	DeviceAccountTxCount           int64              `json:"device_account_tx_count"`

	// Reliability / Degradation
	IsDegraded                     bool               `json:"is_degraded"`
	DegradeReason                  string             `json:"degrade_reason,omitempty"`
}

// SanitizeAccountID validates and sanitizes client-provided account identifiers.
func SanitizeAccountID(raw string) (string, bool) {
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
	return trimmed, true
}

// AccountDeviceGraphStore provides low-latency Redis graph queries and updates.
type AccountDeviceGraphStore struct {
	client redis.Cmdable
}

// NewAccountDeviceGraphStore constructs a new AccountDeviceGraphStore.
func NewAccountDeviceGraphStore(client redis.Cmdable) *AccountDeviceGraphStore {
	return &AccountDeviceGraphStore{client: client}
}

// GetGraphFeatures retrieves real-time point-in-time account/device graph features in a single pipelined Redis round trip.
func (s *AccountDeviceGraphStore) GetGraphFeatures(ctx context.Context, tenantID, deviceID, accountID string) (*AccountDeviceGraphFeatures, error) {
	return s.GetGraphFeaturesAtTime(ctx, tenantID, deviceID, accountID, time.Now().UTC())
}

// GetGraphFeaturesAtTime executes the pipelined query relative to an injected reference timestamp.
func (s *AccountDeviceGraphStore) GetGraphFeaturesAtTime(ctx context.Context, tenantID, deviceID, accountID string, now time.Time) (*AccountDeviceGraphFeatures, error) {
	defaultFeatures := &AccountDeviceGraphFeatures{
		DeviceUniqueAccounts1h:        0,
		DeviceUniqueAccounts24h:       0,
		DeviceAccountSwitches1h:       0,
		DeviceAccountSwitches24h:      0,
		DeviceNewAccountOnKnownDevice: 0,
		MultiAccountSignal:            MultiAccountNormal,
		AccountUniqueDevices1h:        0,
		AccountUniqueDevices24h:       0,
		AccountNewDevice1h:            0,
		AccountDeviceSwitches24h:      0,
		DeviceAccountSeenBefore:       0,
		DeviceAccountAgeDays:          0,
		DeviceAccountTxCount:          0,
		IsDegraded:                    false,
	}

	if s == nil || s.client == nil || deviceID == "" || tenantID == "" {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "ACCOUNT_DEVICE_GRAPH_UNAVAILABLE"
		return defaultFeatures, nil
	}

	cleanAccount, hasAccount := SanitizeAccountID(accountID)

	keyDevAcc1h := fmt.Sprintf("%s:dev:acc1h:%s", tenantID, deviceID)
	keyDevAcc24h := fmt.Sprintf("%s:dev:acc24:%s", tenantID, deviceID)
	keyDevSeq := fmt.Sprintf("%s:dev:acc_seq:%s", tenantID, deviceID)
	keyDevKnown := fmt.Sprintf("%s:dev:known:%s", tenantID, deviceID)

	var keyAccDev1h, keyAccDev24h, keyRelKnown string
	if hasAccount {
		keyAccDev1h = fmt.Sprintf("%s:acc:dev1h:%s", tenantID, cleanAccount)
		keyAccDev24h = fmt.Sprintf("%s:acc:dev24:%s", tenantID, cleanAccount)
		keyRelKnown = fmt.Sprintf("%s:dev_acc:known:%s:%s", tenantID, deviceID, cleanAccount)
	}

	nowMs := strconv.FormatInt(now.UnixMilli(), 10)
	oneHourAgoMs := strconv.FormatInt(now.Add(-1*time.Hour).UnixMilli(), 10)
	twentyFourHoursAgoMs := strconv.FormatInt(now.Add(-24*time.Hour).UnixMilli(), 10)

	pipe := s.client.Pipeline()

	// 1. Prune and count distinct accounts on device
	pipe.ZRemRangeByScore(ctx, keyDevAcc1h, "-inf", oneHourAgoMs)
	cmdDevAcc1h := pipe.ZCard(ctx, keyDevAcc1h)

	pipe.ZRemRangeByScore(ctx, keyDevAcc24h, "-inf", twentyFourHoursAgoMs)
	cmdDevAcc24h := pipe.ZCard(ctx, keyDevAcc24h)

	// 2. Fetch sequential account events on device for switch calculation
	cmdSeq := pipe.ZRangeByScore(ctx, keyDevSeq, &redis.ZRangeBy{
		Min: twentyFourHoursAgoMs,
		Max: nowMs,
	})

	// 3. Known device check (to distinguish new account on established device vs brand new device)
	cmdDevKnown := pipe.Exists(ctx, keyDevKnown)

	// 4. Account-centric device counts (if account provided)
	var cmdAccDev1h, cmdAccDev24h *redis.IntCmd
	var cmdRelKnown *redis.IntCmd
	if hasAccount {
		pipe.ZRemRangeByScore(ctx, keyAccDev1h, "-inf", oneHourAgoMs)
		cmdAccDev1h = pipe.ZCard(ctx, keyAccDev1h)

		pipe.ZRemRangeByScore(ctx, keyAccDev24h, "-inf", twentyFourHoursAgoMs)
		cmdAccDev24h = pipe.ZCard(ctx, keyAccDev24h)

		cmdRelKnown = pipe.Exists(ctx, keyRelKnown)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		defaultFeatures.IsDegraded = true
		defaultFeatures.DegradeReason = "ACCOUNT_DEVICE_GRAPH_UNAVAILABLE"
		return defaultFeatures, fmt.Errorf("failed to execute graph feature pipeline: %w", err)
	}

	// Parse Device-centric counts
	if cmdDevAcc1h != nil {
		defaultFeatures.DeviceUniqueAccounts1h = cmdDevAcc1h.Val()
	}
	if cmdDevAcc24h != nil {
		defaultFeatures.DeviceUniqueAccounts24h = cmdDevAcc24h.Val()
	}

	// Parse Account Switches from sequence
	if cmdSeq != nil {
		events := cmdSeq.Val()
		var lastAccount string
		var switches1h, switches24h int64
		oneHourAgoMilli := now.Add(-1 * time.Hour).UnixMilli()

		for _, item := range events {
			// Format: "<timestamp_ms>_<nonce>:<account_id>"
			idxColon := strings.Index(item, ":")
			if idxColon == -1 {
				continue
			}
			meta := item[:idxColon]
			acc := item[idxColon+1:]

			var eventTimeMs int64
			idxUnderscore := strings.Index(meta, "_")
			if idxUnderscore != -1 {
				eventTimeMs, _ = strconv.ParseInt(meta[:idxUnderscore], 10, 64)
			}

			if lastAccount != "" && acc != lastAccount {
				switches24h++
				if eventTimeMs >= oneHourAgoMilli {
					switches1h++
				}
			}
			lastAccount = acc
		}
		defaultFeatures.DeviceAccountSwitches1h = switches1h
		defaultFeatures.DeviceAccountSwitches24h = switches24h
	}

	deviceIsKnown := (cmdDevKnown != nil && cmdDevKnown.Val() > 0)

	// Parse Account-centric counts and relationship status
	if hasAccount {
		if cmdAccDev1h != nil {
			defaultFeatures.AccountUniqueDevices1h = cmdAccDev1h.Val()
		}
		if cmdAccDev24h != nil {
			defaultFeatures.AccountUniqueDevices24h = cmdAccDev24h.Val()
		}
		if defaultFeatures.AccountUniqueDevices1h == 0 {
			defaultFeatures.AccountNewDevice1h = 1
		}

		if cmdRelKnown != nil && cmdRelKnown.Val() > 0 {
			defaultFeatures.DeviceAccountSeenBefore = 1
		} else {
			defaultFeatures.DeviceAccountSeenBefore = 0
		}

		// New Account on Established Device: Device is known, but this specific account is brand new to it
		if deviceIsKnown && defaultFeatures.DeviceAccountSeenBefore == 0 {
			defaultFeatures.DeviceNewAccountOnKnownDevice = 1
		}
	}

	// Multi-Accounting Risk Signal Classification
	switch {
	case defaultFeatures.DeviceUniqueAccounts24h >= 10 || defaultFeatures.DeviceAccountSwitches1h >= 5:
		defaultFeatures.MultiAccountSignal = MultiAccountHighSignal
	case defaultFeatures.DeviceUniqueAccounts24h >= 5 || defaultFeatures.DeviceAccountSwitches1h >= 3:
		defaultFeatures.MultiAccountSignal = MultiAccountSuspicious
	case defaultFeatures.DeviceUniqueAccounts24h >= 2 || defaultFeatures.AccountUniqueDevices24h >= 2:
		defaultFeatures.MultiAccountSignal = MultiAccountLowSignal
	default:
		defaultFeatures.MultiAccountSignal = MultiAccountNormal
	}

	return defaultFeatures, nil
}

// RecordGraphTransaction records point-in-time account-device graph events into Redis.
func (s *AccountDeviceGraphStore) RecordGraphTransaction(ctx context.Context, tenantID, deviceID, accountID, transactionID string) error {
	return s.RecordGraphTransactionAtTime(ctx, tenantID, deviceID, accountID, transactionID, time.Now().UTC())
}

// RecordGraphTransactionAtTime records graph events relative to an injected reference timestamp.
func (s *AccountDeviceGraphStore) RecordGraphTransactionAtTime(ctx context.Context, tenantID, deviceID, accountID, transactionID string, now time.Time) error {
	if s == nil || s.client == nil || deviceID == "" || tenantID == "" {
		return nil
	}

	cleanAccount, hasAccount := SanitizeAccountID(accountID)

	nowMs := float64(now.UnixMilli())
	nowMilliInt := now.UnixMilli()

	keyDevAcc1h := fmt.Sprintf("%s:dev:acc1h:%s", tenantID, deviceID)
	keyDevAcc24h := fmt.Sprintf("%s:dev:acc24:%s", tenantID, deviceID)
	keyDevSeq := fmt.Sprintf("%s:dev:acc_seq:%s", tenantID, deviceID)

	pipe := s.client.Pipeline()

	twoHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-2*time.Hour).UnixMilli()))
	twentyFiveHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-25*time.Hour).UnixMilli()))

	if hasAccount {
		keyAccDev1h := fmt.Sprintf("%s:acc:dev1h:%s", tenantID, cleanAccount)
		keyAccDev24h := fmt.Sprintf("%s:acc:dev24:%s", tenantID, cleanAccount)
		keyRelKnown := fmt.Sprintf("%s:dev_acc:known:%s:%s", tenantID, deviceID, cleanAccount)

		// 1. Device -> Account Sets
		pipe.ZAdd(ctx, keyDevAcc1h, redis.Z{Score: nowMs, Member: cleanAccount})
		pipe.ZRemRangeByScore(ctx, keyDevAcc1h, "-inf", twoHoursAgoMs)
		pipe.Expire(ctx, keyDevAcc1h, 3*time.Hour)

		pipe.ZAdd(ctx, keyDevAcc24h, redis.Z{Score: nowMs, Member: cleanAccount})
		pipe.ZRemRangeByScore(ctx, keyDevAcc24h, "-inf", twentyFiveHoursAgoMs)
		pipe.Expire(ctx, keyDevAcc24h, 26*time.Hour)

		// 2. Account -> Device Sets
		pipe.ZAdd(ctx, keyAccDev1h, redis.Z{Score: nowMs, Member: deviceID})
		pipe.ZRemRangeByScore(ctx, keyAccDev1h, "-inf", twoHoursAgoMs)
		pipe.Expire(ctx, keyAccDev1h, 3*time.Hour)

		pipe.ZAdd(ctx, keyAccDev24h, redis.Z{Score: nowMs, Member: deviceID})
		pipe.ZRemRangeByScore(ctx, keyAccDev24h, "-inf", twentyFiveHoursAgoMs)
		pipe.Expire(ctx, keyAccDev24h, 26*time.Hour)

		// 3. Sequential Account Transitions on Device
		seqMember := fmt.Sprintf("%d_%s:%s", nowMilliInt, transactionID, cleanAccount)
		pipe.ZAdd(ctx, keyDevSeq, redis.Z{Score: nowMs, Member: seqMember})
		pipe.ZRemRangeByScore(ctx, keyDevSeq, "-inf", twentyFiveHoursAgoMs)
		pipe.Expire(ctx, keyDevSeq, 26*time.Hour)

		// 4. Mark specific (device, account) pair as known for 90 days
		pipe.Set(ctx, keyRelKnown, "1", 90*24*time.Hour)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record graph transaction in redis: %w", err)
	}

	return nil
}
