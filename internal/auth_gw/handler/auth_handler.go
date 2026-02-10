package handler

import (
	"encoding/json"
	"net/http"

	"go-gw-test/internal/auth_gw/types"
	"go-gw-test/internal/auth_gw/usecase"

	"go.uber.org/zap"
)

// AuthHandler exposes HTTP handlers for auth_gw endpoints.
type AuthHandler struct {
	usecase usecase.AuthUsecase
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		usecase: authUsecase,
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

	resp, err := h.usecase.Login(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse(err.Error()))
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

	resp, err := h.usecase.ServiceToken(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse(err.Error()))
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

	resp, err := h.usecase.Validate(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// NotFound returns a JSON 404 response for unmatched routes.
func (h *AuthHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotFound, errorResponse("not found"))
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
