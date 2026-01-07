package limiter

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed scripts/limit.lua
var slidingWindowScript string

type SlidingWindowLimiter struct {
	client     *redis.Client
	windowSize time.Duration
}

func NewSlidingWindowLimiter(client *redis.Client, windowSize time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client:     client,
		windowSize: windowSize,
	}
}

func (s *SlidingWindowLimiter) Allow(ctx context.Context, userID string, limit int) (bool, error) {
	now := time.Now()

	currentWindow := now.Truncate(s.windowSize)
	previousWindow := currentWindow.Add(-s.windowSize)

	currKey := fmt.Sprintf("rate:%s:%d", userID, currentWindow.Unix())
	prevKey := fmt.Sprintf("rate:%s:%d", userID, previousWindow.Unix())

	timePassed := now.Sub(currentWindow)
	percentage := float64(timePassed) / float64(s.windowSize)
	prevWindowWeight := 1.0 - percentage

	res, err := s.client.Eval(ctx, slidingWindowScript,
		[]string{currKey, prevKey},
		limit,
		prevWindowWeight,
		s.windowSize.Seconds(),
	).Result()

	if err != nil {
		return false, fmt.Errorf("redis execution failed: %w", err)
	}

	return res.(int64) == 1, nil
}
