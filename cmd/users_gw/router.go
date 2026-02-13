package main

import (
	"context"
	g "go-gw-test/cmd/users_gw/internal/globals"
	"go-gw-test/cmd/users_gw/internal/repo"
	"go-gw-test/pkg/rest_qol"
	"net/http"

	_ "go-gw-test/cmd/users_gw/docs"
	"go-gw-test/cmd/users_gw/internal/usecase"

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

	router := mux.NewRouter()

	rest_qol.RegisterOperationalRoutes(router, httpSwagger.WrapHandler)

	router.HandleFunc("/api/v1/users", usersUseCase.ListUsers).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/users/{id}", usersUseCase.GetUser).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/users/{id}/contact", usersUseCase.GetContactInfo).Methods(http.MethodGet)

	router.NotFoundHandler = http.HandlerFunc(usersUseCase.NotFound)
	router.Use(rest_qol.RequestIDMiddleware("direct-users-gw-"))
	router.Use(rest_qol.AccessLoggingMiddleware())

	return router
}
