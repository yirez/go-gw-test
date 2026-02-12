package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-gw-test/cmd/api_gw/internal/types"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var errTokenNotFound = errors.New("token not found")
var errUnauthorized = errors.New("unauthorized")

// ErrTokenNotFound exposes the not-found sentinel error.
func ErrTokenNotFound() error {
	return errTokenNotFound
}

// AuthRepo defines auth_gw integration operations.
type AuthRepo interface {
	ExtractBearerTokenFromRequest(req *http.Request) (string, error)
	ValidateToken(ctx context.Context, token string) (types.ValidateResponse, error)

	GetToken(ctx context.Context, apiKey string) (types.TokenMetadata, error)
	TouchExpiry(ctx context.Context, apiKey string, expiresAt time.Time) error
}

// AuthRepoImpl implements AuthRepo against auth_gw HTTP endpoints.
type AuthRepoImpl struct {
	endpoint     string
	serviceID    string
	secret       string
	httpClient   *http.Client
	tokenMu      sync.Mutex
	serviceToken string
	redisClient  *redis.Client
}

// NewAuthRepo constructs an AuthRepo implementation.
func NewAuthRepo(endpoint string, serviceID string, secret string, redisClient *redis.Client) *AuthRepoImpl {
	return &AuthRepoImpl{
		endpoint:    endpoint,
		serviceID:   serviceID,
		secret:      secret,
		redisClient: redisClient,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}
}

// ValidateToken validates a client token by calling auth_gw.
func (r *AuthRepoImpl) ValidateToken(ctx context.Context, token string) (types.ValidateResponse, error) {
	serviceToken, err := r.getServiceToken(ctx)
	if err != nil {
		zap.L().Error("get service token", zap.Error(err))
		return types.ValidateResponse{}, err
	}

	resp, err := r.validateWithServiceToken(ctx, token, serviceToken)
	if err == nil {
		return resp, nil
	}
	if !errors.Is(err, errUnauthorized) {
		return types.ValidateResponse{}, err
	}

	if refreshErr := r.refreshServiceToken(ctx); refreshErr != nil {
		zap.L().Error("refresh service token", zap.Error(refreshErr))
		return types.ValidateResponse{}, refreshErr
	}

	serviceToken, err = r.getServiceToken(ctx)
	if err != nil {
		zap.L().Error("get refreshed service token", zap.Error(err))
		return types.ValidateResponse{}, err
	}

	return r.validateWithServiceToken(ctx, token, serviceToken)
}

func (r *AuthRepoImpl) validateWithServiceToken(ctx context.Context, token string, serviceToken string) (types.ValidateResponse, error) {
	payload, err := json.Marshal(types.ValidateRequest{Token: token})
	if err != nil {
		zap.L().Error("marshal validate payload", zap.Error(err))
		return types.ValidateResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/auth/validate", r.endpoint), bytes.NewReader(payload))
	if err != nil {
		zap.L().Error("build validate request", zap.Error(err))
		return types.ValidateResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", serviceToken))

	res, err := r.httpClient.Do(req)
	if err != nil {
		zap.L().Error("do validate request", zap.Error(err))
		return types.ValidateResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		zap.L().Warn("auth validate unauthorized")
		return types.ValidateResponse{}, errUnauthorized
	}
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("auth validate failed: %d", res.StatusCode)
		zap.L().Error("auth validate non-200", zap.Int("status_code", res.StatusCode), zap.Error(err))
		return types.ValidateResponse{}, err
	}

	var resp types.ValidateResponse
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		zap.L().Error("decode validate response", zap.Error(err))
		return types.ValidateResponse{}, err
	}

	return resp, nil
}

func (r *AuthRepoImpl) getServiceToken(ctx context.Context) (string, error) {
	r.tokenMu.Lock()
	defer r.tokenMu.Unlock()

	if r.serviceToken == "" {
		if err := r.refreshServiceTokenLocked(ctx); err != nil {
			return "", err
		}
	}

	return r.serviceToken, nil
}

func (r *AuthRepoImpl) refreshServiceToken(ctx context.Context) error {
	r.tokenMu.Lock()
	defer r.tokenMu.Unlock()
	return r.refreshServiceTokenLocked(ctx)
}

// refreshServiceTokenLocked must run with tokenMu held.
func (r *AuthRepoImpl) refreshServiceTokenLocked(ctx context.Context) error {
	payload, err := json.Marshal(types.ServiceTokenRequest{ServiceID: r.serviceID, Secret: r.secret})
	if err != nil {
		zap.L().Error("marshal service-token payload", zap.Error(err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/auth/service-token", r.endpoint), bytes.NewReader(payload))
	if err != nil {
		zap.L().Error("build service-token request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := r.httpClient.Do(req)
	if err != nil {
		zap.L().Error("do service-token request", zap.Error(err))
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("auth service-token failed: %d", res.StatusCode)
		zap.L().Error("auth service-token non-200", zap.Int("status_code", res.StatusCode), zap.Error(err))
		return err
	}

	var resp types.ServiceTokenResponse
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		zap.L().Error("decode service-token response", zap.Error(err))
		return err
	}
	if resp.Token == "" {
		err = fmt.Errorf("auth service-token empty")
		zap.L().Error("service-token empty", zap.Error(err))
		return err
	}

	r.serviceToken = resp.Token
	return nil
}

func (r *AuthRepoImpl) ExtractBearerTokenFromRequest(req *http.Request) (string, error) {
	header := req.Header.Get("Authorization")
	if header == "" {
		return "", errors.New("authorization header missing")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("invalid authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", errors.New("empty bearer token")
	}

	return token, nil
}

// GetToken fetches token metadata from Redis.
func (r *AuthRepoImpl) GetToken(ctx context.Context, apiKey string) (types.TokenMetadata, error) {
	key := tokenKey(apiKey)
	values, err := r.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		zap.L().Error("redis hgetall token metadata", zap.String("key", key), zap.Error(err))
		return types.TokenMetadata{}, err
	}
	if len(values) == 0 {
		zap.L().Info("token metadata not found", zap.String("key", key))
		return types.TokenMetadata{}, errTokenNotFound
	}

	rateLimit := parseInt(values["rate_limit"])
	expiresAt, err := time.Parse(time.RFC3339, values["expires_at"])
	if err != nil {
		zap.L().Error("parse token expires_at", zap.String("key", key), zap.String("expires_at", values["expires_at"]), zap.Error(err))
		return types.TokenMetadata{}, fmt.Errorf("invalid expires_at")
	}

	allowedRoutes := parseAllowedRoutes(values["allowed_routes"])
	if len(allowedRoutes) == 0 {
		zap.L().Error("token allowed_routes missing or invalid", zap.String("key", key))
		return types.TokenMetadata{}, fmt.Errorf("allowed_routes missing")
	}

	apiKeyValue := values["api_key"]
	if apiKeyValue == "" {
		apiKeyValue = apiKey
	}

	return types.TokenMetadata{
		APIKey:        apiKeyValue,
		RateLimit:     rateLimit,
		ExpiresAt:     expiresAt,
		AllowedRoutes: allowedRoutes,
	}, nil
}

// TouchExpiry ensures the Redis key expires at the provided timestamp.
func (r *AuthRepoImpl) TouchExpiry(ctx context.Context, apiKey string, expiresAt time.Time) error {
	key := tokenKey(apiKey)
	err := r.redisClient.ExpireAt(ctx, key, expiresAt).Err()
	if err != nil {
		zap.L().Error("redis expireat token metadata", zap.String("key", key), zap.Time("expires_at", expiresAt), zap.Error(err))
		return err
	}

	return nil
}

func tokenKey(apiKey string) string {
	return fmt.Sprintf("token:%s", apiKey)
}

func parseAllowedRoutes(raw string) []string {
	if raw == "" {
		zap.L().Error("allowed_routes empty")
		return nil
	}

	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "[") {
		var routes []string
		if err := json.Unmarshal([]byte(trimmed), &routes); err == nil {
			return routes
		}
		zap.L().Warn("failed to unmarshal allowed_routes json, trying csv parse", zap.String("allowed_routes", trimmed))
	}

	parts := strings.Split(trimmed, ",")
	routes := make([]string, 0, len(parts))
	for _, part := range parts {
		route := strings.TrimSpace(part)
		if route != "" {
			routes = append(routes, route)
		}
	}

	return routes
}

func parseInt(value string) int {
	if value == "" {
		zap.L().Warn("rate_limit missing; defaulting to zero")
		return 0
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		zap.L().Warn("invalid rate_limit; defaulting to zero", zap.String("rate_limit", value), zap.Error(err))
		return 0
	}

	return parsed
}
