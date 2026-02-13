package main

import (
	g "go-gw-test/cmd/auth_gw/internal/globals"
	"go-gw-test/cmd/auth_gw/internal/repo"
	"go-gw-test/cmd/auth_gw/internal/utils"
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
	loggingUseCase := usecase.NewLoggingUseCaseImpl()

	router := mux.NewRouter()

	router.HandleFunc("/healthz", Health).Methods(http.MethodGet)
	router.HandleFunc("/readyz", Ready).Methods(http.MethodGet)
	router.HandleFunc("/metrics", Metrics).Methods(http.MethodGet)
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	router.HandleFunc("/auth/login", authUseCase.Login).Methods(http.MethodPost)
	router.HandleFunc("/auth/service-token", authUseCase.ServiceToken).Methods(http.MethodPost)
	router.HandleFunc("/auth/validate", authUseCase.Validate).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(authUseCase.NotFound)

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rest_qol.EnsureRequestID(w, r)
			next.ServeHTTP(w, r)
		})
	})
	router.Use(loggingUseCase.LoggingMiddleware())
	router.Use(authUseCase.AuthMiddleware())

	return router
}

// Health returns a basic liveness response.
// @Summary Health check
// @Tags auth-gw
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func Health(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready returns a basic readiness response.
// @Summary Readiness check
// @Tags auth-gw
// @Produce json
// @Success 200 {object} map[string]string
// @Router /readyz [get]
func Ready(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// Metrics exposes a placeholder metrics endpoint.
// @Summary Metrics status
// @Tags auth-gw
// @Produce json
// @Success 200 {object} map[string]string
// @Router /metrics [get]
func Metrics(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "metrics_not_implemented"})
}
