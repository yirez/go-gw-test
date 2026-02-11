package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go-gw-test/cmd/auth_gw/internal/repo"
	"go-gw-test/cmd/auth_gw/internal/types"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase defines auth application behavior.
type AuthUseCase interface {
	Login(ctx context.Context, req types.LoginRequest) (types.LoginResponse, error)
	ServiceToken(ctx context.Context, req types.ServiceTokenRequest) (types.ServiceTokenResponse, error)
	Validate(ctx context.Context, req types.ValidateRequest) (types.ValidateResponse, error)
}

// AuthUseCaseImpl implements AuthUseCase.
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
func (u *AuthUseCaseImpl) Login(ctx context.Context, req types.LoginRequest) (types.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return types.LoginResponse{}, errors.New("invalid credentials")
	}

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

// ServiceToken authenticates a service and issues a token.
func (u *AuthUseCaseImpl) ServiceToken(ctx context.Context, req types.ServiceTokenRequest) (types.ServiceTokenResponse, error) {
	if req.ServiceID == "" || req.Secret == "" {
		return types.ServiceTokenResponse{}, errors.New("invalid credentials")
	}

	serviceID, err := parseServiceID(req.ServiceID)
	if err != nil {
		return types.ServiceTokenResponse{}, errors.New("invalid credentials")
	}

	service, err := u.repo.FindServiceByID(ctx, serviceID)
	if err != nil {
		return types.ServiceTokenResponse{}, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(service.SecretHash), []byte(req.Secret))
	if err != nil {
		zap.L().Error("compare service secret", zap.Error(err))
		return types.ServiceTokenResponse{}, errors.New("invalid credentials")
	}

	token, err := u.issueToken("service", fmt.Sprint(service.ID), service.Role)
	if err != nil {
		return types.ServiceTokenResponse{}, err
	}

	return types.ServiceTokenResponse{Token: token}, nil
}

// Validate validates a token and returns metadata for api_gw.
func (u *AuthUseCaseImpl) Validate(ctx context.Context, req types.ValidateRequest) (types.ValidateResponse, error) {
	if req.Token == "" {
		return types.ValidateResponse{}, errors.New("missing token")
	}

	parsed, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
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

	subject, ok := claims["sub"].(string)
	if !ok || subject == "" {
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
		APIKey:    subject,
		Role:      role,
		ExpiresAt: expiresAt,
	}, nil
}

func (u *AuthUseCaseImpl) issueToken(tokenType string, subject string, role string) (string, error) {
	expiresAt := time.Now().UTC().Add(u.tokenTTL)
	claims := jwt.MapClaims{
		"sub":        subject,
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
