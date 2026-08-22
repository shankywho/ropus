package riskengine

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalDistributedLock_Lifecycle(t *testing.T) {
	ctx := context.Background()
	lock := NewLocalDistributedLock()

	// 1. Acquire Lock
	lease, err := lock.Acquire(ctx, "model:promotion", "pod-backend-1", 100*time.Millisecond)
	require.NoError(t, err)
	assert.NotNil(t, lease)
	assert.Equal(t, "model:promotion", lease.ResourceKey)
	assert.Equal(t, "pod-backend-1", lease.OwnerID)
	assert.False(t, lease.IsExpired())

	// 2. Second Worker Attempts to Acquire Same Resource -> Fails
	_, err2 := lock.Acquire(ctx, "model:promotion", "pod-backend-2", 100*time.Millisecond)
	assert.ErrorIs(t, err2, ErrLockHeld)

	// 3. Extend Lease
	err = lock.Extend(ctx, lease, 200*time.Millisecond)
	require.NoError(t, err)

	// 4. Release Lock
	err = lock.Release(ctx, lease)
	require.NoError(t, err)

	// 5. Subsequent Worker Can Now Acquire
	lease2, err := lock.Acquire(ctx, "model:promotion", "pod-backend-2", 100*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, "pod-backend-2", lease2.OwnerID)
	_ = lock.Release(ctx, lease2)
}

func TestLocalDistributedLock_ConcurrentContention(t *testing.T) {
	ctx := context.Background()
	lock := NewLocalDistributedLock()

	numWorkers := 20
	var acquiredCount int64
	var rejectedCount int64
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			lease, err := lock.Acquire(ctx, "critical:recovery:leader", "pod-"+string(rune('A'+workerID)), 50*time.Millisecond)
			if err == nil {
				atomic.AddInt64(&acquiredCount, 1)
				time.Sleep(10 * time.Millisecond)
				_ = lock.Release(ctx, lease)
			} else {
				atomic.AddInt64(&rejectedCount, 1)
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int64(1), atomic.LoadInt64(&acquiredCount), "Exactly one worker should acquire the lock concurrently")
	assert.Equal(t, int64(numWorkers-1), atomic.LoadInt64(&rejectedCount))
}

func TestLocalDistributedLock_AutoExpiry(t *testing.T) {
	ctx := context.Background()
	lock := NewLocalDistributedLock()

	// Acquire short-lived lock
	lease, err := lock.Acquire(ctx, "resource:short", "pod-1", 10*time.Millisecond)
	require.NoError(t, err)
	assert.NotNil(t, lease)

	// Wait for expiration
	time.Sleep(25 * time.Millisecond)

	// Second worker acquires after expiry without explicit release
	lease2, err := lock.Acquire(ctx, "resource:short", "pod-2", 50*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, "pod-2", lease2.OwnerID)
	_ = lock.Release(ctx, lease2)
}

func TestRedisDistributedLock_OfflineClient(t *testing.T) {
	ctx := context.Background()
	redisLock := NewRedisDistributedLock(nil)

	_, err := redisLock.Acquire(ctx, "test:key", "pod-1", 1*time.Second)
	assert.ErrorIs(t, err, ErrLockStoreOffline)

	err = redisLock.Release(ctx, &LockLease{ResourceKey: "test:key"})
	assert.ErrorIs(t, err, ErrLockStoreOffline)

	err = redisLock.Extend(ctx, &LockLease{ResourceKey: "test:key"}, 1*time.Second)
	assert.ErrorIs(t, err, ErrLockStoreOffline)
}
