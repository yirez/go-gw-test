package main

import (
	"context"
	g "go-gw-test/cmd/users_gw/internal/globals"
	"go-gw-test/cmd/users_gw/internal/repo"
	"go-gw-test/pkg/rest_qol"
	"net/http"

	_ "go-gw-test/cmd/users_gw/docs"
	"go-gw-test/cmd/users_gw/internal/usecase"
	"go-gw-test/cmd/users_gw/internal/utils"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
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
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	router.HandleFunc("/api/v1/users", usersUseCase.ListUsers).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/users/{id}", usersUseCase.GetUser).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/users/{id}/contact", usersUseCase.GetContactInfo).Methods(http.MethodGet)

	router.NotFoundHandler = http.HandlerFunc(usersUseCase.NotFound)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rest_qol.EnsureRequestID(w, r)
			next.ServeHTTP(w, r)
		})
	})
	router.Use(loggingUseCase.LoggingMiddleware())

	return router
}

// Health returns a basic liveness response.
// @Summary Health check
// @Tags users-gw
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func Health(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready returns a basic readiness response.
// @Summary Readiness check
// @Tags users-gw
// @Produce json
// @Success 200 {object} map[string]string
// @Router /readyz [get]
func Ready(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// Metrics exposes a placeholder metrics endpoint.
// @Summary Metrics status
// @Tags users-gw
// @Produce json
// @Success 200 {object} map[string]string
// @Router /metrics [get]
func Metrics(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "metrics_not_implemented"})
}
