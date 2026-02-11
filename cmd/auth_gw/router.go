package main

import (
	"go-gw-test/cmd/auth_gw/internal/utils"
	"net/http"

	"go-gw-test/cmd/auth_gw/internal/usecase"

	"github.com/gorilla/mux"
)

// NewRouter builds the gorilla mux router for auth_gw.
func NewRouter(authHandler *usecase.AuthUseCaseImpl) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/healthz", Health).Methods(http.MethodGet)
	router.HandleFunc("/readyz", Ready).Methods(http.MethodGet)
	router.HandleFunc("/metrics", Metrics).Methods(http.MethodGet)

	router.HandleFunc("/auth/login", authHandler.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/service-token", authHandler.ServiceToken).Methods(http.MethodPost)
	router.HandleFunc("/auth/validate", authHandler.Validate).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(authHandler.NotFound)

	router.Use(authHandler.LoggingMiddleware())
	router.Use(authHandler.AuthMiddleware())

	return router
}

// Health returns a basic liveness response.
func Health(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready returns a basic readiness response.
func Ready(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// Metrics exposes a placeholder metrics endpoint.
func Metrics(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "metrics_not_implemented"})
}
