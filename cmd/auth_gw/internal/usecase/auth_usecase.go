package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-gw-test/cmd/auth_gw/internal/utils"
	"go-gw-test/pkg/rest_qol"
	"net/http"
	"strconv"
	"time"

	"go-gw-test/cmd/auth_gw/internal/repo"
	"go-gw-test/cmd/auth_gw/internal/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCaseImpl implements auth flows and HTTP handlers.
type AuthUseCaseImpl struct {
	repo     repo.AuthRepo
	jwtKey   []byte
	tokenTTL time.Duration
}

// NewAuthUseCase constructs an AuthUseCase implementation.
func NewAuthUseCase(authRepo repo.AuthRepo, jwtKey []byte, tokenTTL time.Duration) *AuthUseCaseImpl {
	return &AuthUseCaseImpl{
		repo:     authRepo,
		jwtKey:   jwtKey,
		tokenTTL: tokenTTL,
	}
}

// Login authenticates a user and issues a token.
// @Summary Login
// @Description Authenticates user credentials and returns a signed JWT.
// @Tags auth-gw
// @Accept json
// @Produce json
// @Param request body types.LoginRequest true "Login payload"
// @Success 200 {object} types.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (u *AuthUseCaseImpl) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req types.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		zap.L().Error("decode login request", zap.Error(err))
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Username == "" || req.Password == "" {
		zap.L().Error("login request missing credentials")
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resp, err := u.loginCore(ctx, req)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// ServiceToken authenticates a service and issues a token.
// @Summary Service token
// @Description Authenticates service credentials and returns a signed JWT.
// @Tags auth-gw
// @Accept json
// @Produce json
// @Param request body types.ServiceTokenRequest true "Service token payload"
// @Success 200 {object} types.ServiceTokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/service-token [post]
func (u *AuthUseCaseImpl) ServiceToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req types.ServiceTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		zap.L().Error("decode service token request", zap.Error(err))
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.ServiceID == "" || req.Secret == "" {
		zap.L().Error("service token request missing credentials")
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	serviceID, err := parseServiceID(req.ServiceID)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	service, err := u.repo.FindServiceByID(ctx, serviceID)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(service.SecretHash), []byte(req.Secret))
	if err != nil {
		zap.L().Error("compare service secret failed", zap.Error(err))
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	token, err := u.issueToken("service", fmt.Sprint(service.ID), service.Role)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.ServiceTokenResponse{Token: token})
}

// Validate validates a token and returns metadata for api_gw.
// @Summary Validate token
// @Description Validates a JWT and returns api_key/role/expiry metadata.
// @Tags auth-gw
// @Accept json
// @Produce json
// @Param request body types.ValidateRequest true "Validate payload"
// @Success 200 {object} types.ValidateResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/validate [post]
func (u *AuthUseCaseImpl) Validate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req types.ValidateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		zap.L().Error("decode validate request", zap.Error(err))
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Token == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	resp, err := u.validateTokenCore(ctx, req.Token)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// NotFound returns a JSON 404 response for unmatched routes.
func (u *AuthUseCaseImpl) NotFound(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

// AuthMiddleware enforces bearer token validation on protected routes.
func (u *AuthUseCaseImpl) AuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicRoute(r.Method, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			token, err := rest_qol.BearerTokenFromRequest(r)
			if err != nil {
				zap.L().Error("auth middleware bearer token", zap.Error(err))
				utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			_, err = u.validateTokenCore(r.Context(), token)
			if err != nil {
				zap.L().Error("auth middleware validate token", zap.Error(err))
				utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// loginCore validates user credentials and issues user token.
func (u *AuthUseCaseImpl) loginCore(ctx context.Context, req types.LoginRequest) (types.LoginResponse, error) {
	user, err := u.repo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		return types.LoginResponse{}, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		zap.L().Error("compare user password", zap.Error(err))
		return types.LoginResponse{}, errors.New("invalid credentials")
	}

	token, err := u.issueToken("user", fmt.Sprint(user.ID), user.Role)
	if err != nil {
		return types.LoginResponse{}, err
	}

	return types.LoginResponse{Token: token}, nil
}

// validateTokenCore verifies JWT signature and extracts gateway metadata fields.
func (u *AuthUseCaseImpl) validateTokenCore(ctx context.Context, token string) (types.ValidateResponse, error) {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return u.jwtKey, nil
	})
	if err != nil {
		zap.L().Error("validate jwt", zap.Error(err))
		return types.ValidateResponse{}, errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return types.ValidateResponse{}, errors.New("invalid token")
	}

	apiKey, err := parseAPIKey(claims["jti"])
	if err != nil {
		return types.ValidateResponse{}, errors.New("invalid token")
	}

	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return types.ValidateResponse{}, errors.New("invalid token")
	}

	expiresAt, err := parseExpiry(claims["exp"])
	if err != nil {
		return types.ValidateResponse{}, errors.New("invalid token")
	}

	return types.ValidateResponse{
		APIKey:    apiKey,
		Role:      role,
		ExpiresAt: expiresAt,
	}, nil
}

// issueToken creates a signed JWT with a UUID api_key in jti claim.
func (u *AuthUseCaseImpl) issueToken(tokenType string, subject string, role string) (string, error) {
	expiresAt := time.Now().UTC().Add(u.tokenTTL)
	claims := jwt.MapClaims{
		"sub":        subject,
		"jti":        uuid.NewString(),
		"role":       role,
		"token_type": tokenType,
		"exp":        expiresAt.Unix(),
		"iat":        time.Now().UTC().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(u.jwtKey)
	if err != nil {
		zap.L().Error("sign jwt", zap.Error(err))
		return "", err
	}

	return signed, nil
}

// parseExpiry converts JWT exp claim into RFC3339 UTC format.
func parseExpiry(expClaim any) (string, error) {
	if expClaim == nil {
		return "", errors.New("missing exp")
	}

	switch value := expClaim.(type) {
	case float64:
		return time.Unix(int64(value), 0).UTC().Format(time.RFC3339), nil
	case int64:
		return time.Unix(value, 0).UTC().Format(time.RFC3339), nil
	case json.Number:
		parsed, err := value.Int64()
		if err != nil {
			return "", err
		}
		return time.Unix(parsed, 0).UTC().Format(time.RFC3339), nil
	default:
		return "", errors.New("invalid exp")
	}
}

// parseAPIKey validates jti claim as UUID and returns it.
func parseAPIKey(jtiClaim any) (string, error) {
	jti, ok := jtiClaim.(string)
	if !ok || jti == "" {
		err := errors.New("missing jti")
		zap.L().Error("parse api key from jti", zap.Error(err))
		return "", err
	}

	_, err := uuid.Parse(jti)
	if err != nil {
		zap.L().Error("parse api key uuid", zap.String("jti", jti), zap.Error(err))
		return "", err
	}

	return jti, nil
}

// parseServiceID parses service id from request payload.
func parseServiceID(serviceID string) (int64, error) {
	if serviceID == "" {
		return 0, errors.New("missing service id")
	}

	value, err := strconv.ParseInt(serviceID, 10, 64)
	if err != nil {
		zap.L().Error("parse service id", zap.Error(err))
		return 0, err
	}

	return value, nil
}

// isPublicRoute returns true for unprotected auth routes.
func isPublicRoute(method string, path string) bool {
	if method == http.MethodOptions {
		return true
	}

	switch path {
	case "/healthz", "/readyz", "/metrics", "/auth/login", "/auth/service-token":
		return true
	default:
		return false
	}
}
