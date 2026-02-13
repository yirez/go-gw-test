package rest_qol

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterOperationalRoutes registers common liveness/readiness/metrics routes and optional swagger UI route.
func RegisterOperationalRoutes(router *mux.Router, swaggerHandler http.Handler, metricsHandler http.Handler) {
	router.HandleFunc("/healthz", HealthHandler).Methods(http.MethodGet)
	router.HandleFunc("/readyz", ReadyHandler).Methods(http.MethodGet)
	if metricsHandler != nil {
		router.Path("/metrics").Methods(http.MethodGet).Handler(metricsHandler)
	} else {
		router.HandleFunc("/metrics", MetricsMissingHandler).Methods(http.MethodGet)
	}
	if swaggerHandler != nil {
		router.PathPrefix("/swagger/").Handler(swaggerHandler)
	}
}

// HealthHandler returns a basic liveness response.
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ReadyHandler returns a basic readiness response.
func ReadyHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// MetricsMissingHandler returns a placeholder metrics status response.
func MetricsMissingHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "metrics_not_implemented"})
}

// writeJSON writes JSON payload with status code.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
