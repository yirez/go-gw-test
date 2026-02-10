package usecase

import (
	"context"
	"errors"

	"go-gw-test/internal/auth_gw/repo"
	"go-gw-test/internal/auth_gw/types"
)

// AuthUsecase defines auth application behavior.
type AuthUsecase interface {
	Login(ctx context.Context, req types.LoginRequest) (types.LoginResponse, error)
	ServiceToken(ctx context.Context, req types.ServiceTokenRequest) (types.ServiceTokenResponse, error)
	Validate(ctx context.Context, req types.ValidateRequest) (types.ValidateResponse, error)
}

// AuthUsecaseImpl implements AuthUsecase.
type AuthUsecaseImpl struct {
	repo repo.AuthRepo
}

// NewAuthUsecase constructs an AuthUsecase implementation.
func NewAuthUsecase(authRepo repo.AuthRepo) *AuthUsecaseImpl {
	return &AuthUsecaseImpl{
		repo: authRepo,
	}
}

// Login authenticates a user and issues a token.
func (u *AuthUsecaseImpl) Login(ctx context.Context, req types.LoginRequest) (types.LoginResponse, error) {
	return types.LoginResponse{}, errors.New("not implemented")
}

// ServiceToken authenticates a service and issues a token.
func (u *AuthUsecaseImpl) ServiceToken(ctx context.Context, req types.ServiceTokenRequest) (types.ServiceTokenResponse, error) {
	return types.ServiceTokenResponse{}, errors.New("not implemented")
}

// Validate validates a token and returns metadata for api_gw.
func (u *AuthUsecaseImpl) Validate(ctx context.Context, req types.ValidateRequest) (types.ValidateResponse, error) {
	return types.ValidateResponse{}, errors.New("not implemented")
}
