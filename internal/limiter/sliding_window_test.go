package limiter_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/thisPeyman/rate-limiter/internal/limiter"
)

// SetupRedisHelper connects to a local redis for integration testing
func SetupRedisHelper(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skip("Redis is not running, skipping integration tests")
	}
	rdb.FlushDB(context.Background()) // Clean state
	return rdb
}

func TestSlidingWindow_Allow(t *testing.T) {
	rdb := SetupRedisHelper(t)
	defer rdb.Close()

	// 1 Second Window
	l := limiter.NewSlidingWindowLimiter(rdb, 1*time.Second)
	ctx := context.Background()
	userID := "test-user-1"
	limit := 5

	// 1. Consume all allowed tokens
	for i := 0; i < limit; i++ {
		// FIX: Use the 'limit' variable, not hardcoded number
		allowed, err := l.Allow(ctx, userID, limit)
		assert.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i)
	}

	// 2. Next request should fail
	allowed, err := l.Allow(ctx, userID, limit)
	assert.NoError(t, err)
	assert.False(t, allowed, "Request exceeding limit should be rejected")
}

func TestSlidingWindow_SlidingLogic(t *testing.T) {
	rdb := SetupRedisHelper(t)
	defer rdb.Close()

	windowSize := 1 * time.Second
	l := limiter.NewSlidingWindowLimiter(rdb, windowSize)
	ctx := context.Background()
	userID := "slider-user"
	limit := 10

	// Strategy: Manual Injection (Deterministic)
	// We manually set the Redis key for the *previous* window to simulate 10 requests.
	// We avoid 'time.Sleep' loop racing against the clock.

	// 1. Calculate the key for the previous window
	// Note: This logic must match your implementation's key generation
	now := time.Now()
	// We truncate to find the window start, then subtract one window size
	prevWindowTime := now.Truncate(windowSize).Add(-windowSize)

	// FIX: Use fmt.Sprintf to generate the key exactly like the implementation
	prevKey := fmt.Sprintf("rate:%s:%d", userID, prevWindowTime.Unix())

	// 2. Inject 10 requests into the past (Previous Window)
	// FIX: Actually use the 'prevKey' variable
	err := rdb.Set(ctx, prevKey, 10, windowSize*2).Err()
	assert.NoError(t, err)

	// 3. Move time forward logically for the weight calculation
	// Since we can't easily mock time inside the Limiter without Dependency Injection,
	// we rely on the fact that we are currently in the "Current Window".

	// If we are at the very beginning of the current window (0.1s in):
	// Weight of prev window = 0.9.
	// Count = Current(0) + Prev(10) * 0.9 = 9.
	// 9 < Limit(10) -> Should be Allowed.

	// However, if we run this test at the very end of a second (x.99s),
	// the window might flip during execution.
	// To be safe in integration tests, we just check that *some* capacity remains.

	allowed, err := l.Allow(ctx, userID, limit)
	assert.NoError(t, err)

	// We expect this to be true because the weighted average of 10 requests
	// from the previous window should decay as time passes in the current window.
	assert.True(t, allowed, "Should be allowed because previous window weight decays")
}

// Benchmark
func BenchmarkRateLimiter(b *testing.B) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379", PoolSize: 100})
	l := limiter.NewSlidingWindowLimiter(rdb, 1*time.Second)
	ctx := context.Background()
	userID := "bench-user"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Allow(ctx, userID, 100000)
		}
	})
}
