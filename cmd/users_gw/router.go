package main

import (
	"context"
	g "go-gw-test/cmd/users_gw/internal/globals"
	"go-gw-test/cmd/users_gw/internal/repo"
	"net/http"

	"go-gw-test/cmd/users_gw/internal/usecase"
	"go-gw-test/cmd/users_gw/internal/utils"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewRouter builds the gorilla mux router for users_gw.
func NewRouter() http.Handler {
	usersRepo := repo.NewUsersRepo(g.Cfg.StandardConfigs.Clients.DB)
	if g.Cfg.StandardConfigs.Env != "prod" {
		err := usersRepo.SeedIfEmpty(context.Background())
		if err != nil {
			zap.L().Error("seed users", zap.Error(err))
		}
	}

	usersUseCase := usecase.NewUsersUseCase(usersRepo)
	loggingUseCase := usecase.NewLoggingUseCaseImpl()

	router := mux.NewRouter()

	router.HandleFunc("/healthz", Health).Methods(http.MethodGet)
	router.HandleFunc("/readyz", Ready).Methods(http.MethodGet)
	router.HandleFunc("/metrics", Metrics).Methods(http.MethodGet)

	router.HandleFunc("/api/v1/users", usersUseCase.ListUsers).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/users/{id}", usersUseCase.GetUser).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/users/{id}/contact", usersUseCase.GetContactInfo).Methods(http.MethodGet)

	router.NotFoundHandler = http.HandlerFunc(usersUseCase.NotFound)
	router.Use(loggingUseCase.LoggingMiddleware())

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
