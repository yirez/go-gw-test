package repo

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-gw-test/cmd/api_gw/internal/types"
	"go-gw-test/cmd/api_gw/internal/utils"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// GatewayRepo defines routing and proxy helper operations.
type GatewayRepo interface {
	BuildRouteEntries(configs []types.EndpointConfig) ([]types.RouteEntry, error)
	MatchRoute(routes []types.RouteEntry, r *http.Request) (types.RouteEntry, bool)
	IsAllowedRoute(allowed []string, r *http.Request) bool
}

// GatewayRepoImpl implements GatewayRepo.
type GatewayRepoImpl struct{}

// NewGatewayRepo constructs a GatewayRepo implementation.
func NewGatewayRepo() *GatewayRepoImpl {
	return &GatewayRepoImpl{}
}

// BuildRouteEntries compiles endpoint config into route entries.
func (g *GatewayRepoImpl) BuildRouteEntries(configs []types.EndpointConfig) ([]types.RouteEntry, error) {
	routes := make([]types.RouteEntry, 0, len(configs))
	for _, cfg := range configs {
		pattern := normalizePattern(cfg.GwEndpoint)
		route := mux.NewRouter().NewRoute().Path(pattern)

		proxy, err := newReverseProxy(cfg.LiveEndpoint, cfg.LiveTimeoutSec)
		if err != nil {
			zap.L().Error("build reverse proxy", zap.String("live_endpoint", cfg.LiveEndpoint), zap.Error(err))
			return nil, err
		}

		routes = append(routes, types.RouteEntry{
			Config:  cfg,
			Route:   route,
			Proxy:   proxy,
			RateKey: sanitizeRateKey(cfg.GwEndpoint),
		})
	}

	return routes, nil
}

// MatchRoute finds the most specific matching configured route.
func (g *GatewayRepoImpl) MatchRoute(routes []types.RouteEntry, r *http.Request) (types.RouteEntry, bool) {
	var matched types.RouteEntry
	bestScore := -1
	match := &mux.RouteMatch{}

	for _, entry := range routes {
		if entry.Route.Match(r, match) {
			score := len(strings.Split(entry.Config.GwEndpoint, "/"))
			if score > bestScore {
				bestScore = score
				matched = entry
			}
		}
	}

	if bestScore < 0 {
		return types.RouteEntry{}, false
	}

	return matched, true
}

// IsAllowedRoute validates whether request path matches one of allowed route patterns.
func (g *GatewayRepoImpl) IsAllowedRoute(allowed []string, r *http.Request) bool {
	for _, pattern := range allowed {
		route := mux.NewRouter().NewRoute().Path(pattern)
		if route.Match(r, &mux.RouteMatch{}) {
			return true
		}
	}
	return false
}

func normalizePattern(pattern string) string {
	parts := strings.Split(pattern, "/")
	for i, part := range parts {
		if part == "??" {
			parts[i] = "{param" + strconv.Itoa(i) + "}"
		}
	}
	return strings.Join(parts, "/")
}

func sanitizeRateKey(pattern string) string {
	replacer := strings.NewReplacer("/", "-", "{", "", "}", "", "?", "", "*", "")
	return replacer.Replace(pattern)
}

func newReverseProxy(target string, timeoutSec int) (*httputil.ReverseProxy, error) {
	urlTarget, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(urlTarget)
	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		ResponseHeaderTimeout: time.Duration(timeoutSec) * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		zap.L().Error("proxy error", zap.Error(err))
		utils.WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "backend unavailable"})
	}

	return proxy, nil
}
