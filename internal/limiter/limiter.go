package limiter

import "context"

// RateLimiter defines the behavior for different limiting algorithms
type RateLimiter interface {
	// Allow checks if a request is allowed.
	// Returns true if allowed, false if limit exceeded.
	// Returns error if the backend storage fails.
	Allow(ctx context.Context, userID string, limit int) (bool, error)
}
