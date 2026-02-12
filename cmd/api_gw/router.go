package main

import (
	g "go-gw-test/cmd/api_gw/internal/globals"
	"go-gw-test/cmd/api_gw/internal/repo"
	"go-gw-test/cmd/api_gw/internal/usecase"
	"go-gw-test/cmd/api_gw/internal/utils"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewRouter builds the gorilla mux router for api_gw.
func NewRouter() http.Handler {
	rateLimiter := repo.NewRateLimiter(g.Cfg.StandardConfigs.Clients.Redis)
	gatewayRepo := repo.NewGatewayRepo()
	authRepo := repo.NewAuthRepo(g.Cfg.Auth.Endpoint, g.Cfg.Auth.ServiceID, g.Cfg.Auth.Secret, g.Cfg.StandardConfigs.Clients.Redis)
	authUseCase := usecase.NewAuthUseCase(authRepo, gatewayRepo)

	gatewayUseCase, err := usecase.NewGatewayUseCase(rateLimiter, gatewayRepo, g.Cfg.EndpointConfiguration)
	if err != nil {
		zap.L().Fatal("init gateway usecase", zap.Error(err))
	}
	loggingUseCase := usecase.NewLoggingUseCaseImpl()

	router := mux.NewRouter()

	router.HandleFunc("/healthz", Health).Methods(http.MethodGet)
	router.HandleFunc("/readyz", Ready).Methods(http.MethodGet)
	router.HandleFunc("/metrics", Metrics).Methods(http.MethodGet)

	router.PathPrefix("/api/v1/").HandlerFunc(gatewayUseCase.Proxy)

	router.NotFoundHandler = http.HandlerFunc(gatewayUseCase.NotFound)
	router.Use(gatewayUseCase.RequestIDMiddleware())
	router.Use(loggingUseCase.LoggingMiddleware())
	router.Use(authUseCase.TokenValidationMiddleware())

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
