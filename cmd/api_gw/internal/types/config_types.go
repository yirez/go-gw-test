package types

import (
	"net/http/httputil"
	"time"

	cmt "go-gw-test/pkg/configuration_manager/types"
)

// AppConfig wraps api_gw configuration and standard configs.
type AppConfig struct {
	StandardConfigs       cmt.StandardConfig
	EndpointConfiguration []EndpointConfig
}

// EndpointConfig defines gateway routing rules.
type EndpointConfig struct {
	LiveEndpoint       string   `mapstructure:"live_endpoint"`
	LiveTimeoutSec     int      `mapstructure:"live_timeout_sec"`
	GwEndpoint         string   `mapstructure:"gw_endpoint"`
	RateLimitReqPerSec int      `mapstructure:"rate_limit_req_per_sec"`
	AllowedRole        []string `mapstructure:"allowed_role"`
}

// TokenMetadata represents token data stored in Redis.
type TokenMetadata struct {
	APIKey        string
	RateLimit     int
	ExpiresAt     time.Time
	AllowedRoutes []string
}

// ServiceTokenRequest captures service-to-service login payload.
type ServiceTokenRequest struct {
	ServiceID string `json:"service_id"`
	Secret    string `json:"secret"`
}

// ServiceTokenResponse captures service token response.
type ServiceTokenResponse struct {
	Token string `json:"token"`
}

// ValidateRequest captures token validation payload.
type ValidateRequest struct {
	Token string `json:"token"`
}

// ValidateResponse represents auth_gw validate response payload.
type ValidateResponse struct {
	APIKey    string `json:"api_key"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
}

// RouteEntry holds compiled routing data.
type RouteEntry struct {
	Config  EndpointConfig
	Proxy   *httputil.ReverseProxy
	RateKey string
}
