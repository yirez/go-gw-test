package usecase

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-gw-test/cmd/api_gw/internal/repo"
	"go-gw-test/cmd/api_gw/internal/types"
)

type fakeAuthRepo struct {
	validateResp types.ValidateResponse
	validateErr  error
	metaResp     types.TokenMetadata
	metaErr      error
	setErr       error
	touchErr     error
	setCalled    bool
}

func (f *fakeAuthRepo) ValidateToken(ctx context.Context, token string) (types.ValidateResponse, error) {
	return f.validateResp, f.validateErr
}

func (f *fakeAuthRepo) GetTokenMetaFromRedis(ctx context.Context, apiKey string) (types.TokenMetadata, error) {
	return f.metaResp, f.metaErr
}

func (f *fakeAuthRepo) SetToken(ctx context.Context, metadata types.TokenMetadata) error {
	f.setCalled = true
	f.metaResp = metadata
	return f.setErr
}

func (f *fakeAuthRepo) TouchExpiry(ctx context.Context, apiKey string, expiresAt time.Time) error {
	return f.touchErr
}

// TestTokenValidationMiddlewareUnauthorizedWithoutHeader verifies missing bearer token handling.
func TestTokenValidationMiddlewareUnauthorizedWithoutHeader(t *testing.T) {
	authRepo := &fakeAuthRepo{}
	gatewayRepo := repo.NewGatewayRepo()
	useCase, err := NewAuthUseCase(authRepo, gatewayRepo, []types.EndpointConfig{
		{GwEndpoint: "/api/v1/users/*", LiveEndpoint: "http://users:8087", AllowedRole: []string{"user_users"}},
	})
	if err != nil {
		t.Fatalf("new auth usecase: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil)
	rr := httptest.NewRecorder()

	handler := useCase.TokenValidationMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

// TestTokenValidationMiddlewareInitializesMissingRedisMetadata verifies lazy token metadata bootstrap.
func TestTokenValidationMiddlewareInitializesMissingRedisMetadata(t *testing.T) {
	expiresAt := time.Now().UTC().Add(30 * time.Minute).Truncate(time.Second)
	authRepo := &fakeAuthRepo{
		validateResp: types.ValidateResponse{
			APIKey:    "550e8400-e29b-41d4-a716-446655440000",
			Role:      "user_users",
			ExpiresAt: expiresAt.Format(time.RFC3339),
		},
		metaErr: repo.ErrTokenNotFound(),
	}

	gatewayRepo := repo.NewGatewayRepo()
	useCase, err := NewAuthUseCase(authRepo, gatewayRepo, []types.EndpointConfig{
		{
			GwEndpoint:         "/api/v1/users/*",
			LiveEndpoint:       "http://users:8087",
			AllowedRole:        []string{"user_all", "user_users"},
			RateLimitReqPerSec: 5,
		},
	})
	if err != nil {
		t.Fatalf("new auth usecase: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	nextCalled := false
	handler := useCase.TokenValidationMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		metadata, ok := r.Context().Value(ctxKeyTokenMetadata).(types.TokenMetadata)
		if !ok {
			t.Fatalf("token metadata missing in context")
		}
		if metadata.APIKey == "" {
			t.Fatalf("api key missing in token metadata")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Fatalf("expected next handler to be called")
	}
	if !authRepo.setCalled {
		t.Fatalf("expected SetToken to be called for missing metadata")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rr.Code)
	}
}

// TestTokenValidationMiddlewareForbiddenRole verifies endpoint role guard.
func TestTokenValidationMiddlewareForbiddenRole(t *testing.T) {
	expiresAt := time.Now().UTC().Add(30 * time.Minute).Truncate(time.Second)
	authRepo := &fakeAuthRepo{
		validateResp: types.ValidateResponse{
			APIKey:    "550e8400-e29b-41d4-a716-446655440000",
			Role:      "user_orders",
			ExpiresAt: expiresAt.Format(time.RFC3339),
		},
		metaResp: types.TokenMetadata{
			APIKey:        "550e8400-e29b-41d4-a716-446655440000",
			RateLimit:     5,
			ExpiresAt:     expiresAt,
			AllowedRoutes: []string{"/api/v1/users/*"},
		},
	}

	gatewayRepo := repo.NewGatewayRepo()
	useCase, err := NewAuthUseCase(authRepo, gatewayRepo, []types.EndpointConfig{
		{
			GwEndpoint:         "/api/v1/users/*",
			LiveEndpoint:       "http://users:8087",
			AllowedRole:        []string{"user_users"},
			RateLimitReqPerSec: 5,
		},
	})
	if err != nil {
		t.Fatalf("new auth usecase: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	handler := useCase.TokenValidationMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rr.Code)
	}
}

// TestTokenValidationMiddlewarePropagatesRepoErrors verifies generic redis errors map to 500.
func TestTokenValidationMiddlewarePropagatesRepoErrors(t *testing.T) {
	expiresAt := time.Now().UTC().Add(30 * time.Minute).Truncate(time.Second)
	authRepo := &fakeAuthRepo{
		validateResp: types.ValidateResponse{
			APIKey:    "550e8400-e29b-41d4-a716-446655440000",
			Role:      "user_users",
			ExpiresAt: expiresAt.Format(time.RFC3339),
		},
		metaErr: errors.New("redis down"),
	}

	gatewayRepo := repo.NewGatewayRepo()
	useCase, err := NewAuthUseCase(authRepo, gatewayRepo, []types.EndpointConfig{
		{GwEndpoint: "/api/v1/users/*", LiveEndpoint: "http://users:8087", AllowedRole: []string{"user_users"}},
	})
	if err != nil {
		t.Fatalf("new auth usecase: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	handler := useCase.TokenValidationMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}
