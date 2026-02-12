package main

import (
	"context"
	g "go-gw-test/cmd/orders_gw/internal/globals"
	"go-gw-test/cmd/orders_gw/internal/repo"
	"net/http"

	"go-gw-test/cmd/orders_gw/internal/usecase"
	"go-gw-test/cmd/orders_gw/internal/utils"

	"github.com/gorilla/mux"
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

	ordersHandler := usecase.NewOrdersUseCase(ordersRepo)

	router := mux.NewRouter()

	router.HandleFunc("/healthz", Health).Methods(http.MethodGet)
	router.HandleFunc("/readyz", Ready).Methods(http.MethodGet)
	router.HandleFunc("/metrics", Metrics).Methods(http.MethodGet)

	router.HandleFunc("/api/v1/orders", ordersHandler.ListOrders).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/orders/{id}", ordersHandler.GetOrder).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/orders/{id}/items", ordersHandler.GetOrderItems).Methods(http.MethodGet)

	router.NotFoundHandler = http.HandlerFunc(ordersHandler.NotFound)
	router.Use(ordersHandler.LoggingMiddleware())

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
