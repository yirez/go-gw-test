package main

import (
	g "go-gw-test/cmd/api_gw/internal/globals"
	"go-gw-test/cmd/api_gw/internal/repo"
	"go-gw-test/cmd/api_gw/internal/usecase"
	"go-gw-test/cmd/api_gw/internal/utils"
	"net/http"

	_ "go-gw-test/cmd/api_gw/docs"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// NewRouter builds the gorilla mux router for api_gw.
func NewRouter() http.Handler {
	rateLimiter := repo.NewRateLimiterRepo(g.Cfg.StandardConfigs.Clients.Redis)
	gatewayRepo := repo.NewGatewayRepo()
	authRepo := repo.NewAuthRepo(
		g.Cfg.StandardConfigs.AuthConfig.Endpoint,
		g.Cfg.StandardConfigs.AuthConfig.ServiceID,
		g.Cfg.StandardConfigs.AuthConfig.Secret,
		g.Cfg.StandardConfigs.Clients.Redis)
	authUseCase, err := usecase.NewAuthUseCase(authRepo, gatewayRepo, g.Cfg.EndpointConfiguration)
	if err != nil {
		zap.L().Fatal("init auth usecase", zap.Error(err))
	}

	gatewayUseCase, err := usecase.NewGatewayUseCase(rateLimiter, gatewayRepo, g.Cfg.EndpointConfiguration)
	if err != nil {
		zap.L().Fatal("init gateway usecase", zap.Error(err))
	}
	loggingUseCase := usecase.NewLoggingUseCaseImpl()

	router := mux.NewRouter()

	router.HandleFunc("/healthz", Health).Methods(http.MethodGet)
	router.HandleFunc("/readyz", Ready).Methods(http.MethodGet)
	router.HandleFunc("/metrics", Metrics).Methods(http.MethodGet)
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	router.PathPrefix("/api/v1/").HandlerFunc(gatewayUseCase.Proxy)

	router.NotFoundHandler = http.HandlerFunc(gatewayUseCase.NotFound)
	router.Use(gatewayUseCase.RequestIDMiddleware())
	router.Use(loggingUseCase.LoggingMiddleware())
	router.Use(authUseCase.TokenValidationMiddleware())

	return router
}

// Health returns a basic liveness response.
// @Summary Health check
// @Tags api-gw
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func Health(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready returns a basic readiness response.
// @Summary Readiness check
// @Tags api-gw
// @Produce json
// @Success 200 {object} map[string]string
// @Router /readyz [get]
func Ready(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// Metrics exposes a placeholder metrics endpoint.
// @Summary Metrics status
// @Tags api-gw
// @Produce json
// @Success 200 {object} map[string]string
// @Router /metrics [get]
func Metrics(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "metrics_not_implemented"})
}
