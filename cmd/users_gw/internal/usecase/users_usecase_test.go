package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-gw-test/cmd/users_gw/internal/types"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type fakeUsersRepo struct {
	users      []types.UserProfile
	listErr    error
	user       types.UserProfile
	userErr    error
	contact    types.UserContactInfo
	contactErr error
}

// ListUsers returns configured fake users.
func (f *fakeUsersRepo) ListUsers(ctx context.Context) ([]types.UserProfile, error) {
	return f.users, f.listErr
}

// FindUserByID returns configured fake user lookup.
func (f *fakeUsersRepo) FindUserByID(ctx context.Context, userID int64) (types.UserProfile, error) {
	return f.user, f.userErr
}

// GetContactInfo returns configured fake contact lookup.
func (f *fakeUsersRepo) GetContactInfo(ctx context.Context, userID int64) (types.UserContactInfo, error) {
	return f.contact, f.contactErr
}

// SeedIfEmpty is a no-op fake.
func (f *fakeUsersRepo) SeedIfEmpty(ctx context.Context) error {
	return nil
}

// TestUsersUseCaseListUsersSuccess verifies user listing response shape.
func TestUsersUseCaseListUsersSuccess(t *testing.T) {
	u := NewUsersUseCase(&fakeUsersRepo{
		users: []types.UserProfile{{ID: 1, Name: "Ayla", Email: "a@x", Phone: "1"}},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rr := httptest.NewRecorder()
	u.ListUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp types.UsersResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Users) != 1 {
		t.Fatalf("expected one user, got %d", len(resp.Users))
	}
}

// TestUsersUseCaseGetUserInvalidID verifies bad path id handling.
func TestUsersUseCaseGetUserInvalidID(t *testing.T) {
	u := NewUsersUseCase(&fakeUsersRepo{})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/users/not-int", nil), map[string]string{"id": "not-int"})
	rr := httptest.NewRecorder()

	u.GetUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

// TestUsersUseCaseGetUserNotFound verifies not found propagation.
func TestUsersUseCaseGetUserNotFound(t *testing.T) {
	u := NewUsersUseCase(&fakeUsersRepo{userErr: gorm.ErrRecordNotFound})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil), map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	u.GetUser(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

// TestUsersUseCaseGetContactInfoNotFound verifies contact-info 404 path.
func TestUsersUseCaseGetContactInfoNotFound(t *testing.T) {
	u := NewUsersUseCase(&fakeUsersRepo{
		user:       types.UserProfile{ID: 1},
		contactErr: gorm.ErrRecordNotFound,
	})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/users/1/contact", nil), map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	u.GetContactInfo(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

// TestUsersUseCaseGetContactInfoUserLookupError verifies user lookup failures map to 500.
func TestUsersUseCaseGetContactInfoUserLookupError(t *testing.T) {
	u := NewUsersUseCase(&fakeUsersRepo{
		userErr: errors.New("db fail"),
	})
	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/api/v1/users/1/contact", nil), map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	u.GetContactInfo(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
}
