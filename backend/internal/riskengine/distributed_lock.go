package riskengine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrLockHeld         = errors.New("distributed lock is already acquired by another worker")
	ErrLockExpired      = errors.New("lock lease has expired or been revoked")
	ErrInvalidLease     = errors.New("lock lease is invalid or token mismatch")
	ErrLockStoreOffline = errors.New("distributed lock store is offline")
)

// LockLease represents a securely held distributed lock lease with a fencing token.
type LockLease struct {
	ResourceKey string    `json:"resource_key"`
	Token       string    `json:"token"`
	OwnerID     string    `json:"owner_id"`
	AcquiredAt  time.Time `json:"acquired_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// IsExpired checks if the lease has passed its expiration deadline.
func (l *LockLease) IsExpired() bool {
	return time.Now().UTC().After(l.ExpiresAt)
}

// DistributedLock defines the contract for cluster-wide atomic coordination.
type DistributedLock interface {
	Acquire(ctx context.Context, resourceKey, ownerID string, ttl time.Duration) (*LockLease, error)
	Release(ctx context.Context, lease *LockLease) error
	Extend(ctx context.Context, lease *LockLease, extraTTL time.Duration) error
}

// ---------------------------------------------------------------------------
// 1. In-Memory Local Lock (for standalone development & test environments)
// ---------------------------------------------------------------------------
type localLockEntry struct {
	token     string
	ownerID   string
	expiresAt time.Time
}

// LocalDistributedLock provides thread-safe in-memory locking with lease expiry.
type LocalDistributedLock struct {
	mu            sync.Mutex
	locks         map[string]localLockEntry
	acquiredCount int64
	releasedCount int64
	conflictCount int64
}

// NewLocalDistributedLock initializes an in-memory distributed lock simulator.
func NewLocalDistributedLock() *LocalDistributedLock {
	return &LocalDistributedLock{
		locks: make(map[string]localLockEntry),
	}
}

func (l *LocalDistributedLock) Acquire(ctx context.Context, resourceKey, ownerID string, ttl time.Duration) (*LockLease, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()
	if entry, exists := l.locks[resourceKey]; exists {
		if now.Before(entry.expiresAt) {
			atomic.AddInt64(&l.conflictCount, 1)
			return nil, ErrLockHeld
		}
		// Expired lock: treat as available
		delete(l.locks, resourceKey)
	}

	token := generateLockToken()
	expiresAt := now.Add(ttl)
	l.locks[resourceKey] = localLockEntry{
		token:     token,
		ownerID:   ownerID,
		expiresAt: expiresAt,
	}

	atomic.AddInt64(&l.acquiredCount, 1)
	return &LockLease{
		ResourceKey: resourceKey,
		Token:       token,
		OwnerID:     ownerID,
		AcquiredAt:  now,
		ExpiresAt:   expiresAt,
	}, nil
}

func (l *LocalDistributedLock) Release(ctx context.Context, lease *LockLease) error {
	if lease == nil {
		return ErrInvalidLease
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry, exists := l.locks[lease.ResourceKey]
	if !exists || entry.token != lease.Token {
		return ErrInvalidLease
	}

	delete(l.locks, lease.ResourceKey)
	atomic.AddInt64(&l.releasedCount, 1)
	return nil
}

func (l *LocalDistributedLock) Extend(ctx context.Context, lease *LockLease, extraTTL time.Duration) error {
	if lease == nil {
		return ErrInvalidLease
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry, exists := l.locks[lease.ResourceKey]
	if !exists || entry.token != lease.Token {
		return ErrInvalidLease
	}

	now := time.Now().UTC()
	if now.After(entry.expiresAt) {
		return ErrLockExpired
	}

	entry.expiresAt = entry.expiresAt.Add(extraTTL)
	l.locks[lease.ResourceKey] = entry
	lease.ExpiresAt = entry.expiresAt
	return nil
}

// Stats returns lock coordination metrics.
func (l *LocalDistributedLock) Stats() (acquired, released, conflicts int64) {
	return atomic.LoadInt64(&l.acquiredCount), atomic.LoadInt64(&l.releasedCount), atomic.LoadInt64(&l.conflictCount)
}

// ---------------------------------------------------------------------------
// 2. Redis Distributed Lock (for clustered multi-pod deployments)
// ---------------------------------------------------------------------------
type RedisDistributedLock struct {
	client *redis.Client
}

// NewRedisDistributedLock initializes a Redis-backed distributed lock manager.
func NewRedisDistributedLock(client *redis.Client) *RedisDistributedLock {
	return &RedisDistributedLock{client: client}
}

func (r *RedisDistributedLock) Acquire(ctx context.Context, resourceKey, ownerID string, ttl time.Duration) (*LockLease, error) {
	if r.client == nil {
		return nil, ErrLockStoreOffline
	}

	token := generateLockToken()
	now := time.Now().UTC()
	key := fmt.Sprintf("lock:%s", resourceKey)

	ok, err := r.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lock error: %w", err)
	}
	if !ok {
		return nil, ErrLockHeld
	}

	return &LockLease{
		ResourceKey: resourceKey,
		Token:       token,
		OwnerID:     ownerID,
		AcquiredAt:  now,
		ExpiresAt:   now.Add(ttl),
	}, nil
}

// Lua script for atomic release ensuring we only delete our own token
const releaseLuaScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end`

func (r *RedisDistributedLock) Release(ctx context.Context, lease *LockLease) error {
	if r.client == nil {
		return ErrLockStoreOffline
	}
	if lease == nil {
		return ErrInvalidLease
	}

	key := fmt.Sprintf("lock:%s", lease.ResourceKey)
	res, err := r.client.Eval(ctx, releaseLuaScript, []string{key}, lease.Token).Result()
	if err != nil {
		return fmt.Errorf("redis unlock error: %w", err)
	}

	if count, ok := res.(int64); !ok || count == 0 {
		return ErrInvalidLease
	}
	return nil
}

// Lua script for atomic extend
const extendLuaScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("pexpire", KEYS[1], ARGV[2])
else
    return 0
end`

func (r *RedisDistributedLock) Extend(ctx context.Context, lease *LockLease, extraTTL time.Duration) error {
	if r.client == nil {
		return ErrLockStoreOffline
	}
	if lease == nil {
		return ErrInvalidLease
	}

	key := fmt.Sprintf("lock:%s", lease.ResourceKey)
	extraMillis := extraTTL.Milliseconds()
	res, err := r.client.Eval(ctx, extendLuaScript, []string{key}, lease.Token, extraMillis).Result()
	if err != nil {
		return fmt.Errorf("redis extend error: %w", err)
	}

	if count, ok := res.(int64); !ok || count == 0 {
		return ErrLockExpired
	}

	lease.ExpiresAt = lease.ExpiresAt.Add(extraTTL)
	return nil
}

func generateLockToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
