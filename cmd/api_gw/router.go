package main

import (
	g "go-gw-test/cmd/api_gw/internal/globals"
	"go-gw-test/cmd/api_gw/internal/repo"
	"go-gw-test/cmd/api_gw/internal/usecase"
	"go-gw-test/pkg/rest_qol"
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

	router := mux.NewRouter()
	metrics := rest_qol.NewHTTPMetrics("api_gw")

	rest_qol.RegisterOperationalRoutes(router, httpSwagger.WrapHandler, metrics.Handler())

	router.PathPrefix("/api/v1/").HandlerFunc(gatewayUseCase.Proxy)

	router.NotFoundHandler = http.HandlerFunc(gatewayUseCase.NotFound)
	router.Use(rest_qol.RequestIDMiddleware("api-gw-"))
	router.Use(metrics.Middleware())
	router.Use(rest_qol.AccessLoggingMiddleware())
	router.Use(authUseCase.TokenValidationMiddleware())

	return router
}
