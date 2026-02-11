package repo

import (
	"context"

	"go-gw-test/cmd/auth_gw/internal/types"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthRepo defines persistence operations needed by auth_gw.
type AuthRepo interface {
	FindUserByUsername(ctx context.Context, username string) (types.UserRecord, error)
	FindServiceByID(ctx context.Context, serviceID int64) (types.ServiceRecord, error)
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
	var record types.UserRecord
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&record).Error
	if err != nil {
		zap.L().Error("find user", zap.Error(err))
		return types.UserRecord{}, err
	}

	return record, nil
}

// FindServiceByID loads a service record by ID.
func (r *AuthRepoImpl) FindServiceByID(ctx context.Context, serviceID int64) (types.ServiceRecord, error) {
	var record types.ServiceRecord
	err := r.db.WithContext(ctx).Where("id = ?", serviceID).First(&record).Error
	if err != nil {
		zap.L().Error("find service", zap.Error(err))
		return types.ServiceRecord{}, err
	}

	return record, nil
}
