package repo

import (
	"context"
	"errors"
	"go-gw-test/cmd/auth_gw/internal/types"

	"gorm.io/gorm"
)

// AuthRepo defines persistence operations needed by auth_gw.
type AuthRepo interface {
	FindUserByUsername(ctx context.Context, username string) (types.UserRecord, error)
	FindServiceByID(ctx context.Context, serviceID string) (types.ServiceRecord, error)
	StoreToken(ctx context.Context, token types.TokenRecord) error
	FindToken(ctx context.Context, token string) (types.TokenRecord, error)
}

// AuthRepoImpl implements AuthRepo using GORM.
type AuthRepoImpl struct {
	db *gorm.DB
}

// NewAuthRepo constructs an AuthRepo implementation.
func NewAuthRepo(db *gorm.DB) *AuthRepoImpl {
	return &AuthRepoImpl{
		db: db,
	}
}

// FindUserByUsername loads a user record by username.
func (r *AuthRepoImpl) FindUserByUsername(ctx context.Context, username string) (types.UserRecord, error) {
	return types.UserRecord{}, errors.New("not implemented")
}

// FindServiceByID loads a service record by ID.
func (r *AuthRepoImpl) FindServiceByID(ctx context.Context, serviceID string) (types.ServiceRecord, error) {
	return types.ServiceRecord{}, errors.New("not implemented")
}

// StoreToken persists token metadata for validation.
func (r *AuthRepoImpl) StoreToken(ctx context.Context, token types.TokenRecord) error {
	return errors.New("not implemented")
}

// FindToken loads token metadata by token value.
func (r *AuthRepoImpl) FindToken(ctx context.Context, token string) (types.TokenRecord, error) {
	return types.TokenRecord{}, errors.New("not implemented")
}
