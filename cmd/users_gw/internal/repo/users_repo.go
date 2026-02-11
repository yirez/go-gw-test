package repo

import (
	"context"

	"go-gw-test/cmd/users_gw/internal/types"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UsersRepo defines persistence operations for users_gw.
type UsersRepo interface {
	ListUsers(ctx context.Context) ([]types.UserProfile, error)
	FindUserByID(ctx context.Context, userID int64) (types.UserProfile, error)
	SeedIfEmpty(ctx context.Context) error
}

// UsersRepoImpl implements UsersRepo using GORM.
type UsersRepoImpl struct {
	db *gorm.DB
}

// NewUsersRepo constructs a UsersRepo implementation.
func NewUsersRepo(db *gorm.DB) *UsersRepoImpl {
	return &UsersRepoImpl{
		db: db,
	}
}

// ListUsers returns all user profiles.
func (r *UsersRepoImpl) ListUsers(ctx context.Context) ([]types.UserProfile, error) {
	var users []types.UserProfile
	err := r.db.WithContext(ctx).Find(&users).Error
	if err != nil {
		zap.L().Error("list users", zap.Error(err))
		return nil, err
	}

	return users, nil
}

// FindUserByID returns a user profile by ID.
func (r *UsersRepoImpl) FindUserByID(ctx context.Context, userID int64) (types.UserProfile, error) {
	var user types.UserProfile
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		zap.L().Error("find user", zap.Error(err))
		return types.UserProfile{}, err
	}

	return user, nil
}

// SeedIfEmpty inserts sample user profiles when no records exist.
func (r *UsersRepoImpl) SeedIfEmpty(ctx context.Context) error {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.UserProfile{}).Count(&count).Error
	if err != nil {
		zap.L().Error("count users", zap.Error(err))
		return err
	}

	if count > 0 {
		return nil
	}

	seed := []types.UserProfile{
		{ID: 1, Name: "Ayla Demir", Email: "ayla@example.com", Phone: "+90-555-0101"},
		{ID: 2, Name: "Kerem Kaya", Email: "kerem@example.com", Phone: "+90-555-0102"},
		{ID: 3, Name: "Deniz Acar", Email: "deniz@example.com", Phone: "+90-555-0103"},
	}

	err = r.db.WithContext(ctx).Create(&seed).Error
	if err != nil {
		zap.L().Error("seed users", zap.Error(err))
		return err
	}

	return nil
}
