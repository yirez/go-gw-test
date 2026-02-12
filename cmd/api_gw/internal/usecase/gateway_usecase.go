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
	rateLimiter repo.RateLimiter
	gatewayRepo repo.GatewayRepo
	routes      []types.RouteEntry
}

// NewGatewayUseCase constructs a GatewayUseCase.
func NewGatewayUseCase(rateLimiter repo.RateLimiter, gatewayRepo repo.GatewayRepo, configs []types.EndpointConfig) (*GatewayUseCase, error) {
	routes, err := gatewayRepo.BuildRouteEntries(configs)
	if err != nil {
		return nil, err
	}

	return &GatewayUseCase{
		rateLimiter: rateLimiter,
		gatewayRepo: gatewayRepo,
		routes:      routes,
	}, nil
}

// Proxy handles the gateway proxy endpoint.
func (g *GatewayUseCase) Proxy(w http.ResponseWriter, r *http.Request) {
	entry, ok := g.gatewayRepo.MatchRoute(g.routes, r)
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
		count, _, err := g.rateLimiter.Increment(r.Context(), apiKey, entry.RateKey)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "rate limiter failed"})
			return
		}
		if int(count) > limit {
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
