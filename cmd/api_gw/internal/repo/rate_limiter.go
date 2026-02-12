package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter defines redis operations for rate limiting.
type RateLimiter interface {
	Increment(ctx context.Context, apiKey string, endpointKey string) (int64, time.Time, error)
}

// RateLimiterImpl implements RateLimiter using Redis.
type RateLimiterImpl struct {
	client *redis.Client
}

// NewRateLimiter constructs a RateLimiter implementation.
func NewRateLimiter(client *redis.Client) *RateLimiterImpl {
	return &RateLimiterImpl{client: client}
}

// Increment increments the rate counter for the current second.
func (r *RateLimiterImpl) Increment(ctx context.Context, apiKey string, endpointKey string) (int64, time.Time, error) {
	now := time.Now().UTC()
	window := now.Unix()
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
