package router

import (
	"go-gw-test/cmd/auth_gw/internal/handler"
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter builds the gorilla mux router for auth_gw.
func NewRouter(authHandler *handler.AuthHandler, loggingHandler *handler.LoggingHandler) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/healthz", authHandler.Health).Methods(http.MethodGet)
	router.HandleFunc("/readyz", authHandler.Ready).Methods(http.MethodGet)
	router.HandleFunc("/metrics", authHandler.Metrics).Methods(http.MethodGet)

	router.HandleFunc("/auth/login", authHandler.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/service-token", authHandler.ServiceToken).Methods(http.MethodPost)
	router.HandleFunc("/auth/validate", authHandler.Validate).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(authHandler.NotFound)

	router.Use(loggingHandler.LoggingMiddleware())
	router.Use(authHandler.AuthMiddleware())

	return router
}
