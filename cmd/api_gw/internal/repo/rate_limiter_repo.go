package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiterRepo defines redis operations for rate limiting.
type RateLimiterRepo interface {
	Increment(ctx context.Context, apiKey string, endpointKey string) (int64, time.Time, error)
}

// RateLimiterRepoImpl implements RateLimiterRepo using Redis.
type RateLimiterRepoImpl struct {
	client *redis.Client
}

// NewRateLimiterRepo constructs a RateLimiterRepo implementation.
func NewRateLimiterRepo(client *redis.Client) *RateLimiterRepoImpl {
	return &RateLimiterRepoImpl{client: client}
}

// Increment increments the rate counter for the current second.
func (r *RateLimiterRepoImpl) Increment(ctx context.Context, apiKey string, endpointKey string) (int64, time.Time, error) {
	now := time.Now().UTC()
	window := now.Unix() // unix second, since our rate limit is per second, this works.
	key := fmt.Sprintf("rl:%s:%s:%d", apiKey, endpointKey, window)

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 2*time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		zap.L().Error("rate limiter exec", zap.Error(err))
		return 0, time.Time{}, err
	}

	return incr.Val(), time.Unix(window+1, 0).UTC(), nil
}
