package usecase

import (
	"context"
	"errors"
	"go-gw-test/cmd/api_gw/internal/repo"
	"go-gw-test/cmd/api_gw/internal/utils"
	"net/http"
	"strings"
	"time"

	"go-gw-test/cmd/api_gw/internal/types"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// AuthUseCaseImpl calls auth_gw for token validation.
type AuthUseCaseImpl struct {
	ar     repo.AuthRepo
	gr     repo.GatewayRepo
	routes []types.RouteEntry
}

type AuthUseCase interface {
	ValidateToken(ctx context.Context, token string) (types.ValidateResponse, error)
	TokenValidationMiddleware() mux.MiddlewareFunc
}

// NewAuthUseCase constructs an AuthUseCaseImpl.
func NewAuthUseCase(ar repo.AuthRepo, gr repo.GatewayRepo, endpointConfigs []types.EndpointConfig) (AuthUseCase, error) {
	routes, err := gr.BuildRouteEntries(endpointConfigs)
	if err != nil {
		return nil, err
	}

	return &AuthUseCaseImpl{
		ar:     ar,
		gr:     gr,
		routes: routes,
	}, nil
}

// ValidateToken validates a token via auth_gw and returns metadata.
func (u *AuthUseCaseImpl) ValidateToken(ctx context.Context, token string) (types.ValidateResponse, error) {
	return u.ar.ValidateToken(ctx, token)
}

// TokenValidationMiddleware validates incoming bearer tokens for proxy routes.
func (u *AuthUseCaseImpl) TokenValidationMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/v1/") {
				next.ServeHTTP(w, r)
				return
			}

			clientToken, err := u.ar.ExtractBearerTokenFromRequest(r)
			if err != nil {
				zap.L().Warn("missing bearer token", zap.Error(err))
				utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			validateResp, err := u.ValidateToken(r.Context(), clientToken)
			if err != nil {
				zap.L().Warn("token validation failed", zap.Error(err))
				utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			apiKey := validateResp.APIKey
			if apiKey == "" {
				zap.L().Warn("token validation returned empty api_key")
				utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			expiresAt, err := time.Parse(time.RFC3339, validateResp.ExpiresAt)
			if err != nil {
				zap.L().Warn("token validation returned invalid expires_at", zap.String("expires_at", validateResp.ExpiresAt), zap.Error(err))
				utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			metadata, err := u.ar.GetToken(r.Context(), apiKey)
			if err != nil {
				if errors.Is(err, repo.ErrTokenNotFound()) {
					utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
					return
				}
				utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "token lookup failed"})
				return
			}

			if err = u.ar.TouchExpiry(r.Context(), apiKey, expiresAt); err != nil {
				utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "token expiry sync failed"})
				return
			}

			now := time.Now().UTC()
			if now.After(metadata.ExpiresAt) || now.After(expiresAt) {
				utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "token expired"})
				return
			}

			if !u.gr.IsAllowedRoute(metadata.AllowedRoutes, r) {
				utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}

			entry, ok := u.gr.MatchRoute(u.routes, r)
			if !ok {
				utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "route not found"})
				return
			}

			if !u.gr.IsRoleAllowed(entry.Config.AllowedRole, validateResp.Role) {
				utils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyAPIKey, apiKey)
			ctx = context.WithValue(ctx, ctxKeyTokenMetadata, metadata)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
