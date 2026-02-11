package main

import (
	g "go-gw-test/cmd/auth_gw/internal/globals"
	"go-gw-test/cmd/auth_gw/internal/repo"
	"go-gw-test/cmd/auth_gw/internal/utils"
	"net/http"
	"time"

	"go-gw-test/cmd/auth_gw/internal/usecase"

	"github.com/gorilla/mux"
)

// NewRouter builds the gorilla mux router for auth_gw.
func NewRouter() http.Handler {
	authRepo := repo.NewAuthRepo(g.Cfg.StandardConfigs.Clients.DB)
	authUseCase := usecase.NewAuthUseCase(authRepo, g.Cfg.JwtSigningKey, time.Hour)
	loggingUseCase := usecase.NewLoggingUseCaseImpl()

	router := mux.NewRouter()

	router.HandleFunc("/healthz", Health).Methods(http.MethodGet)
	router.HandleFunc("/readyz", Ready).Methods(http.MethodGet)
	router.HandleFunc("/metrics", Metrics).Methods(http.MethodGet)

	router.HandleFunc("/auth/login", authUseCase.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/service-token", authUseCase.ServiceToken).Methods(http.MethodPost)
	router.HandleFunc("/auth/validate", authUseCase.Validate).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(authUseCase.NotFound)

	router.Use(loggingUseCase.LoggingMiddleware())
	router.Use(authUseCase.AuthMiddleware())

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
