package main

import (
	"context"
	g "go-gw-test/cmd/orders_gw/internal/globals"
	"go-gw-test/cmd/orders_gw/internal/repo"
	"go-gw-test/pkg/rest_qol"
	"net/http"

	_ "go-gw-test/cmd/orders_gw/docs"
	"go-gw-test/cmd/orders_gw/internal/usecase"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// NewRouter builds the gorilla mux router for orders_gw.
func NewRouter() http.Handler {
	ordersRepo := repo.NewOrdersRepo(g.Cfg.StandardConfigs.Clients.DB)
	if g.Cfg.StandardConfigs.Env != "prod" {
		err := ordersRepo.SeedIfEmpty(context.Background())
		if err != nil {
			zap.L().Error("seed orders", zap.Error(err))
		}
	}

	ordersUseCase := usecase.NewOrdersUseCase(ordersRepo)

	router := mux.NewRouter()
	metrics := rest_qol.NewHTTPMetrics("orders_gw")

	rest_qol.RegisterOperationalRoutes(router, httpSwagger.WrapHandler, metrics.Handler())

	router.HandleFunc("/api/v1/orders", ordersUseCase.ListOrders).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/orders/{id}", ordersUseCase.GetOrder).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/orders/{id}/items", ordersUseCase.GetOrderItems).Methods(http.MethodGet)

	router.NotFoundHandler = http.HandlerFunc(ordersUseCase.NotFound)
	router.Use(rest_qol.RequestIDMiddleware("direct-orders-gw-"))
	router.Use(metrics.Middleware())
	router.Use(rest_qol.AccessLoggingMiddleware())

	return router
}
