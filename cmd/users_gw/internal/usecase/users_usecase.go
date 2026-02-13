package usecase

import (
	"errors"
	"net/http"
	"strconv"

	"go-gw-test/cmd/users_gw/internal/repo"
	"go-gw-test/cmd/users_gw/internal/types"
	"go-gw-test/cmd/users_gw/internal/utils"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UsersUseCaseImpl implements users endpoints and helpers.
type UsersUseCaseImpl struct {
	repo repo.UsersRepo
}

// NewUsersUseCase constructs a UsersUseCase implementation.
func NewUsersUseCase(usersRepo repo.UsersRepo) *UsersUseCaseImpl {
	return &UsersUseCaseImpl{
		repo: usersRepo,
	}
}

// ListUsers returns all users.
// @Summary List users
// @Tags users-gw
// @Produce json
// @Success 200 {object} types.UsersResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/users [get]
func (u *UsersUseCaseImpl) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := u.repo.ListUsers(ctx)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
		return
	}

	resp := mapUsersResponse(users)
	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetUser returns a user profile by ID.
// @Summary Get user
// @Tags users-gw
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} types.UserProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [get]
func (u *UsersUseCaseImpl) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idValue := mux.Vars(r)["id"]
	userID, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil {
		zap.L().Error("parse user id", zap.Error(err))
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	user, err := u.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}

		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	resp := types.UserProfileResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetContactInfo returns the stored contact info for a user.
// @Summary Get user contact info
// @Tags users-gw
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} types.ContactInfoResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id}/contact [get]
func (u *UsersUseCaseImpl) GetContactInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idValue := mux.Vars(r)["id"]
	userID, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil {
		zap.L().Error("parse user id", zap.Error(err))
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	_, err = u.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	info, err := u.repo.GetContactInfo(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "contact info not found"})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch contact info"})
		return
	}

	resp := types.ContactInfoResponse{
		UserID:       info.UserID,
		Email:        info.Email,
		Phone:        info.Phone,
		AddressLine1: info.AddressLine1,
		City:         info.City,
		Country:      info.Country,
	}
	utils.WriteJSON(w, http.StatusOK, resp)
}

// NotFound returns a JSON 404 response for unmatched routes.
func (u *UsersUseCaseImpl) NotFound(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

// mapUsersResponse maps entities into a list response.
func mapUsersResponse(users []types.UserProfile) types.UsersResponse {
	resp := types.UsersResponse{Users: make([]types.UserProfileResponse, 0, len(users))}
	for _, user := range users {
		resp.Users = append(resp.Users, types.UserProfileResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
		})
	}

	return resp
}
