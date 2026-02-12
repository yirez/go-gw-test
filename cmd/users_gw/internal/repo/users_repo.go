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
	GetContactInfo(ctx context.Context, userID int64) (types.UserContactInfo, error)
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

// GetContactInfo returns contact info for a user.
func (r *UsersRepoImpl) GetContactInfo(ctx context.Context, userID int64) (types.UserContactInfo, error) {
	var info types.UserContactInfo
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&info).Error
	if err != nil {
		zap.L().Error("find contact info", zap.Error(err))
		return types.UserContactInfo{}, err
	}

	return info, nil
}

// SeedIfEmpty inserts sample user profiles when no records exist.
func (r *UsersRepoImpl) SeedIfEmpty(ctx context.Context) error {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.UserProfile{}).Count(&count).Error
	if err != nil {
		zap.L().Error("count users", zap.Error(err))
		return err
	}

	if count == 0 {
		seedUsers := []types.UserProfile{
			{ID: 1, Name: "Ayla Demir", Email: "ayla@example.com", Phone: "+90-555-0101"},
			{ID: 2, Name: "Kerem Kaya", Email: "kerem@example.com", Phone: "+90-555-0102"},
			{ID: 3, Name: "Deniz Acar", Email: "deniz@example.com", Phone: "+90-555-0103"},
		}

		err = r.db.WithContext(ctx).Create(&seedUsers).Error
		if err != nil {
			zap.L().Error("seed users", zap.Error(err))
			return err
		}
	}

	var contactCount int64
	err = r.db.WithContext(ctx).Model(&types.UserContactInfo{}).Count(&contactCount).Error
	if err != nil {
		zap.L().Error("count contact info", zap.Error(err))
		return err
	}

	if contactCount > 0 {
		return nil
	}

	seedContacts := []types.UserContactInfo{
		{UserID: 1, Email: "ayla@example.com", Phone: "+90-555-0101", AddressLine1: "101 Elm St", City: "Istanbul", Country: "TR"},
		{UserID: 2, Email: "kerem@example.com", Phone: "+90-555-0102", AddressLine1: "22 Pine Ave", City: "Ankara", Country: "TR"},
		{UserID: 3, Email: "deniz@example.com", Phone: "+90-555-0103", AddressLine1: "8 Oak Blvd", City: "Izmir", Country: "TR"},
	}

	err = r.db.WithContext(ctx).Create(&seedContacts).Error
	if err != nil {
		zap.L().Error("seed contact info", zap.Error(err))
		return err
	}

	return nil
}
