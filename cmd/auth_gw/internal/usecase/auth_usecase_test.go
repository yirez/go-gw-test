package usecase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-gw-test/cmd/auth_gw/internal/types"

	"golang.org/x/crypto/bcrypt"
)

type fakeAuthRepo struct {
	user    types.UserRecord
	userErr error

	service    types.ServiceRecord
	serviceErr error
}

// FindUserByUsername returns configured fake user data.
func (f *fakeAuthRepo) FindUserByUsername(ctx context.Context, username string) (types.UserRecord, error) {
	return f.user, f.userErr
}

// FindServiceByID returns configured fake service data.
func (f *fakeAuthRepo) FindServiceByID(ctx context.Context, serviceID int64) (types.ServiceRecord, error) {
	return f.service, f.serviceErr
}

// TestAuthUseCaseLoginSuccess verifies login returns token with valid credentials.
func TestAuthUseCaseLoginSuccess(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate password hash: %v", err)
	}

	u := NewAuthUseCase(&fakeAuthRepo{
		user: types.UserRecord{
			ID:           1,
			Username:     "user_all",
			PasswordHash: string(hash),
			Role:         "user_all",
		},
	}, []byte("test-secret"), time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"username":"user_all","password":"123"}`))
	rr := httptest.NewRecorder()

	u.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp types.LoginResponse
	if err = json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected non-empty token")
	}
}

// TestAuthUseCaseLoginInvalidBody verifies malformed requests are rejected.
func TestAuthUseCaseLoginInvalidBody(t *testing.T) {
	u := NewAuthUseCase(&fakeAuthRepo{}, []byte("test-secret"), time.Hour)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{`))
	rr := httptest.NewRecorder()

	u.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

// TestAuthUseCaseValidateSuccess verifies validate endpoint returns api key metadata.
func TestAuthUseCaseValidateSuccess(t *testing.T) {
	u := NewAuthUseCase(&fakeAuthRepo{}, []byte("test-secret"), time.Hour)
	token, err := u.issueToken("user", "1", "user_all")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/validate", strings.NewReader(`{"token":"`+token+`"}`))
	rr := httptest.NewRecorder()

	u.Validate(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp types.ValidateResponse
	if err = json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.APIKey == "" || resp.Role == "" || resp.ExpiresAt == "" {
		t.Fatalf("expected complete validate response, got %#v", resp)
	}
}

// TestAuthUseCaseAuthMiddleware verifies protected routes require bearer token.
func TestAuthUseCaseAuthMiddleware(t *testing.T) {
	u := NewAuthUseCase(&fakeAuthRepo{}, []byte("test-secret"), time.Hour)
	mw := u.AuthMiddleware()

	protectedReq := httptest.NewRequest(http.MethodGet, "/auth/validate", nil)
	protectedRR := httptest.NewRecorder()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(protectedRR, protectedReq)

	if protectedRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected protected route to be unauthorized without token, got %d", protectedRR.Code)
	}

	publicReq := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	publicRR := httptest.NewRecorder()
	handler.ServeHTTP(publicRR, publicReq)

	if publicRR.Code != http.StatusNoContent {
		t.Fatalf("expected public route to pass middleware, got %d", publicRR.Code)
	}
}
