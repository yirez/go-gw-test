package repo

import (
	"context"
	"testing"
	"time"

	"go-gw-test/cmd/api_gw/internal/types"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// TestAuthRepoSetAndGetTokenMeta verifies Redis token metadata persistence.
func TestAuthRepoSetAndGetTokenMeta(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	authRepo := NewAuthRepo("http://auth:8084", "1", "123", client)

	expiresAt := time.Now().UTC().Add(1 * time.Hour).Truncate(time.Second)
	in := types.TokenMetadata{
		APIKey:        "550e8400-e29b-41d4-a716-446655440000",
		RateLimit:     7,
		ExpiresAt:     expiresAt,
		AllowedRoutes: []string{"/api/v1/users/*"},
	}

	if err = authRepo.SetToken(context.Background(), in); err != nil {
		t.Fatalf("set token: %v", err)
	}

	out, err := authRepo.GetTokenMetaFromRedis(context.Background(), in.APIKey)
	if err != nil {
		t.Fatalf("get token: %v", err)
	}

	if out.APIKey != in.APIKey {
		t.Fatalf("api key mismatch: got %s want %s", out.APIKey, in.APIKey)
	}
	if out.RateLimit != in.RateLimit {
		t.Fatalf("rate limit mismatch: got %d want %d", out.RateLimit, in.RateLimit)
	}
	if !out.ExpiresAt.Equal(in.ExpiresAt) {
		t.Fatalf("expiresAt mismatch: got %s want %s", out.ExpiresAt, in.ExpiresAt)
	}
	if len(out.AllowedRoutes) != 1 || out.AllowedRoutes[0] != "/api/v1/users/*" {
		t.Fatalf("allowed routes mismatch: got %#v", out.AllowedRoutes)
	}
}
