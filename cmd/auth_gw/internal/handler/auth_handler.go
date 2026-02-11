package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go-gw-test/cmd/auth_gw/internal/types"
	"go-gw-test/cmd/auth_gw/internal/usecase"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// AuthHandler exposes HTTP handlers for auth_gw endpoints.
type AuthHandler struct {
	au usecase.AuthUseCase
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(authUsecase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		au: authUsecase,
	}
}

// Health returns a basic liveness response.
func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready returns a basic readiness response.
func (h *AuthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// Metrics exposes a placeholder metrics endpoint.
func (h *AuthHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "metrics_not_implemented"})
}

// Login handles user login and token issuance.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req types.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logError("decode login request", err)
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	resp, err := h.au.Login(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ServiceToken handles service-to-service token requests.
func (h *AuthHandler) ServiceToken(w http.ResponseWriter, r *http.Request) {
	var req types.ServiceTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logError("decode service token request", err)
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	resp, err := h.au.ServiceToken(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Validate validates a token and returns its metadata.
func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	var req types.ValidateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logError("decode validate request", err)
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	resp, err := h.au.Validate(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// NotFound returns a JSON 404 response for unmatched routes.
func (h *AuthHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotFound, errorResponse("not found"))
}

// AuthMiddleware enforces bearer token validation on protected routes.
func (h *AuthHandler) AuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicRoute(r.Method, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			token, err := bearerToken(r)
			if err != nil {
				h.logError("auth middleware bearer token", err)
				writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
				return
			}

			_, err = h.au.Validate(r.Context(), types.ValidateRequest{Token: token})
			if err != nil {
				h.logError("auth middleware validate token", err)
				writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeJSON serializes a response with application/json content type.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// errorResponse builds a simple error payload.
func errorResponse(message string) map[string]string {
	return map[string]string{"error": message}
}

// logError writes an error log if a logger is available.
func (h *AuthHandler) logError(message string, err error) {
	zap.L().Error(message, zap.Error(err))
}

func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", fmt.Errorf("invalid authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", fmt.Errorf("empty bearer token")
	}

	return token, nil
}

func isPublicRoute(method string, path string) bool {
	if method == http.MethodOptions {
		return true
	}

	switch path {
	case "/healthz", "/readyz", "/metrics", "/auth/login", "/auth/service-token":
		return true
	default:
		return false
	}
}
