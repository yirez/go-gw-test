package main

import (
	g "go-gw-test/cmd/auth_gw/internal/globals"
	"go-gw-test/cmd/auth_gw/internal/repo"
	"go-gw-test/pkg/rest_qol"
	"net/http"
	"time"

	_ "go-gw-test/cmd/auth_gw/docs"
	"go-gw-test/cmd/auth_gw/internal/usecase"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter builds the gorilla mux router for auth_gw.
func NewRouter() http.Handler {
	authRepo := repo.NewAuthRepo(g.Cfg.StandardConfigs.Clients.DB)
	authUseCase := usecase.NewAuthUseCase(authRepo, g.Cfg.JwtSigningKey, time.Hour)

	router := mux.NewRouter()

	rest_qol.RegisterOperationalRoutes(router, httpSwagger.WrapHandler)

	router.HandleFunc("/auth/login", authUseCase.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/service-token", authUseCase.ServiceToken).Methods(http.MethodPost)
	router.HandleFunc("/auth/validate", authUseCase.Validate).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(authUseCase.NotFound)

	router.Use(rest_qol.RequestIDMiddleware("direct-auth-gw-"))
	router.Use(rest_qol.AccessLoggingMiddleware())
	router.Use(authUseCase.AuthMiddleware())

	return router
}
