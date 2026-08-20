package features

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// VelocityStore provides sliding-window velocity metrics backed by Redis Sorted Sets.
type VelocityStore struct {
	client *redis.Client
}

// NewVelocityStore initializes a new VelocityStore with a Redis client.
func NewVelocityStore(client *redis.Client) *VelocityStore {
	return &VelocityStore{
		client: client,
	}
}

// VelocityMetrics contains aggregated counter metrics for risk evaluation.
type VelocityMetrics struct {
	TxnCountIP1h     int64 `json:"velocity_ip_1hr"`
	TxnCountToken24h int64 `json:"velocity_token_24hr"`
}

// RecordEvent records a transaction event for IP and Token velocity windows in Redis.
func (s *VelocityStore) RecordEvent(ctx context.Context, tenantID, ip, token string, amount int64) error {
	if s == nil || s.client == nil {
		return nil
	}

	now := time.Now()
	nowMs := float64(now.UnixMilli())
	nowNano := now.UnixNano()

	pipe := s.client.Pipeline()

	// 1. Record for IP (Key prefix: {tenant_id}:velocity:ip:{ip})
	if ip != "" {
		ipKey := fmt.Sprintf("%s:velocity:ip:%s", tenantID, ip)
		memberIP := fmt.Sprintf("%d:%d", nowNano, amount)

		pipe.ZAdd(ctx, ipKey, redis.Z{
			Score:  nowMs,
			Member: memberIP,
		})
		// Prune entries older than 2 hours to keep memory compact
		twoHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-2*time.Hour).UnixMilli()))
		pipe.ZRemRangeByScore(ctx, ipKey, "-inf", twoHoursAgoMs)
		// Set TTL for key eviction (3 hours)
		pipe.Expire(ctx, ipKey, 3*time.Hour)
	}

	// 2. Record for Token (Key prefix: {tenant_id}:velocity:token:{token})
	if token != "" {
		tokenKey := fmt.Sprintf("%s:velocity:token:%s", tenantID, token)
		memberToken := fmt.Sprintf("%d:%d", nowNano, amount)

		pipe.ZAdd(ctx, tokenKey, redis.Z{
			Score:  nowMs,
			Member: memberToken,
		})
		// Prune entries older than 25 hours
		twentyFiveHoursAgoMs := fmt.Sprintf("%f", float64(now.Add(-25*time.Hour).UnixMilli()))
		pipe.ZRemRangeByScore(ctx, tokenKey, "-inf", twentyFiveHoursAgoMs)
		// Set TTL for key eviction (26 hours)
		pipe.Expire(ctx, tokenKey, 26*time.Hour)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record velocity event in redis: %w", err)
	}

	return nil
}

// GetVelocityMetrics calculates transactions per IP (last 1 hour) and per Token (last 24 hours).
func (s *VelocityStore) GetVelocityMetrics(ctx context.Context, tenantID, ip, token string) (*VelocityMetrics, error) {
	if s == nil || s.client == nil {
		return &VelocityMetrics{
			TxnCountIP1h:     0,
			TxnCountToken24h: 0,
		}, nil
	}

	now := time.Now()
	nowMs := strconv.FormatInt(now.UnixMilli(), 10)
	oneHourAgoMs := strconv.FormatInt(now.Add(-1*time.Hour).UnixMilli(), 10)
	twentyFourHoursAgoMs := strconv.FormatInt(now.Add(-24*time.Hour).UnixMilli(), 10)

	pipe := s.client.Pipeline()

	var ipCountCmd *redis.IntCmd
	var tokenCountCmd *redis.IntCmd

	if ip != "" {
		ipKey := fmt.Sprintf("%s:velocity:ip:%s", tenantID, ip)
		ipCountCmd = pipe.ZCount(ctx, ipKey, oneHourAgoMs, nowMs)
	}

	if token != "" {
		tokenKey := fmt.Sprintf("%s:velocity:token:%s", tenantID, token)
		tokenCountCmd = pipe.ZCount(ctx, tokenKey, twentyFourHoursAgoMs, nowMs)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to fetch velocity metrics from redis: %w", err)
	}

	metrics := &VelocityMetrics{
		TxnCountIP1h:     0,
		TxnCountToken24h: 0,
	}

	if ipCountCmd != nil {
		metrics.TxnCountIP1h = ipCountCmd.Val()
	}

	if tokenCountCmd != nil {
		metrics.TxnCountToken24h = tokenCountCmd.Val()
	}

	return metrics, nil
}
