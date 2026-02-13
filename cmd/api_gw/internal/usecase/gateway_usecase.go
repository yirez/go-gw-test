package usecase

import (
	"net/http"

	"go-gw-test/cmd/api_gw/internal/repo"
	"go-gw-test/cmd/api_gw/internal/types"
	"go-gw-test/cmd/api_gw/internal/utils"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type contextKey string

const (
	ctxKeyAPIKey        contextKey = "api_key"
	ctxKeyTokenMetadata contextKey = "token_metadata"
)

// GatewayUseCase handles proxying logic for api_gw.
type GatewayUseCase struct {
	rr     repo.RateLimiterRepo
	gr     repo.GatewayRepo
	routes []types.RouteEntry
}

// NewGatewayUseCase constructs a GatewayUseCase.
func NewGatewayUseCase(rateLimiter repo.RateLimiterRepo, gatewayRepo repo.GatewayRepo, configs []types.EndpointConfig) (*GatewayUseCase, error) {
	routes, err := gatewayRepo.BuildRouteEntries(configs)
	if err != nil {
		return nil, err
	}

	return &GatewayUseCase{
		rr:     rateLimiter,
		gr:     gatewayRepo,
		routes: routes,
	}, nil
}

// Proxy handles the gateway proxy endpoint.
// @Summary Proxy request
// @Description Proxies API requests to configured backend based on endpoint rules.
// @Tags api-gw
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/{path} [get]
// @Router /api/v1/{path} [post]
// @Router /api/v1/{path} [put]
// @Router /api/v1/{path} [patch]
// @Router /api/v1/{path} [delete]
func (g *GatewayUseCase) Proxy(w http.ResponseWriter, r *http.Request) {
	entry, ok := g.gr.MatchRoute(g.routes, r)
	if !ok {
		utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "route not found"})
		return
	}

	apiKey, ok := r.Context().Value(ctxKeyAPIKey).(string)
	if !ok || apiKey == "" {
		zap.L().Error("proxy context missing api_key")
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	metadata, ok := r.Context().Value(ctxKeyTokenMetadata).(types.TokenMetadata)
	if !ok {
		zap.L().Error("proxy context missing token metadata", zap.String("api_key", apiKey))
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	limit := entry.Config.RateLimitReqPerSec
	if metadata.RateLimit > 0 && (limit == 0 || metadata.RateLimit < limit) {
		limit = metadata.RateLimit
	}

	if limit > 0 {
		count, _, err := g.rr.Increment(r.Context(), apiKey, entry.RateKey)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "rate limiter failed"})
			return
		}
		if int(count) > limit {
			zap.L().Warn("rate limit exceeded",
				zap.String("api_key", apiKey),
				zap.String("endpoint", entry.Config.GwEndpoint),
				zap.Int("limit", limit),
				zap.Int64("count", count),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			utils.WriteJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}
	}

	entry.Proxy.ServeHTTP(w, r)
}

// NotFound returns a JSON 404 response for unmatched routes.
func (g *GatewayUseCase) NotFound(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

// RequestIDMiddleware ensures a request id exists for the request lifecycle.
func (g *GatewayUseCase) RequestIDMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = utils.NewRequestID()
				r.Header.Set("X-Request-Id", requestID)
			}
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r)
		})
	}
}
