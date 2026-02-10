package usecase

import (
	"context"
	"errors"
	"go-gw-test/cmd/auth_gw/internal/repo"
	"go-gw-test/cmd/auth_gw/internal/types"
)

// AuthUseCase defines auth application behavior.
type AuthUseCase interface {
	Login(ctx context.Context, req types.LoginRequest) (types.LoginResponse, error)
	ServiceToken(ctx context.Context, req types.ServiceTokenRequest) (types.ServiceTokenResponse, error)
	Validate(ctx context.Context, req types.ValidateRequest) (types.ValidateResponse, error)
}

// AuthUseCaseImpl implements AuthUseCase.
type AuthUseCaseImpl struct {
	repo repo.AuthRepo
}

// NewAuthUseCase constructs an AuthUseCase implementation.
func NewAuthUseCase(authRepo repo.AuthRepo) *AuthUseCaseImpl {
	return &AuthUseCaseImpl{
		repo: authRepo,
	}
}

// Login authenticates a user and issues a token.
func (u *AuthUseCaseImpl) Login(ctx context.Context, req types.LoginRequest) (types.LoginResponse, error) {
	return types.LoginResponse{}, errors.New("not implemented")
}

// ServiceToken authenticates a service and issues a token.
func (u *AuthUseCaseImpl) ServiceToken(ctx context.Context, req types.ServiceTokenRequest) (types.ServiceTokenResponse, error) {
	return types.ServiceTokenResponse{}, errors.New("not implemented")
}

// Validate validates a token and returns metadata for api_gw.
func (u *AuthUseCaseImpl) Validate(ctx context.Context, req types.ValidateRequest) (types.ValidateResponse, error) {
	return types.ValidateResponse{}, errors.New("not implemented")
}
