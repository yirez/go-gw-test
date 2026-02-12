package repo

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// TestRateLimiterIncrement verifies per-second key increment behavior.
func TestRateLimiterIncrement(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	limiter := NewRateLimiterRepo(client)

	count1, _, err := limiter.Increment(context.Background(), "api-key-1", "users")
	if err != nil {
		t.Fatalf("increment #1: %v", err)
	}
	count2, _, err := limiter.Increment(context.Background(), "api-key-1", "users")
	if err != nil {
		t.Fatalf("increment #2: %v", err)
	}

	if count1 != 1 {
		t.Fatalf("expected first count=1, got %d", count1)
	}
	if count2 != 2 {
		t.Fatalf("expected second count=2, got %d", count2)
	}
}
