package router

import (
	"go-gw-test/cmd/auth_gw/internal/handler"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// NewRouter builds the gorilla mux router for auth_gw.
func NewRouter(authHandler *handler.AuthHandler) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/healthz", authHandler.Health).Methods(http.MethodGet)
	router.HandleFunc("/readyz", authHandler.Ready).Methods(http.MethodGet)
	router.HandleFunc("/metrics", authHandler.Metrics).Methods(http.MethodGet)

	router.HandleFunc("/auth/login", authHandler.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/service-token", authHandler.ServiceToken).Methods(http.MethodPost)
	router.HandleFunc("/auth/validate", authHandler.Validate).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(authHandler.NotFound)

	router.Use(loggingMiddleware())

	return router
}

// loggingMiddleware emits basic access logs for each request.
func loggingMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			zap.L().Info("request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)
			next.ServeHTTP(w, r)
		})
	}
}
