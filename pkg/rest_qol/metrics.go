package rest_qol

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPMetrics provides Prometheus instrumentation and an export handler.
type HTTPMetrics struct {
	requestTotal    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	handler         http.Handler
}

// NewHTTPMetrics builds service-scoped metrics registry and collectors.
func NewHTTPMetrics(serviceName string) *HTTPMetrics {
	registry := prometheus.NewRegistry()

	requestTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests.",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"method", "route", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request latency in seconds.",
			ConstLabels: prometheus.Labels{"service": serviceName},
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)

	registry.MustRegister(requestTotal, requestDuration)

	return &HTTPMetrics{
		requestTotal:    requestTotal,
		requestDuration: requestDuration,
		handler:         promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	}
}

// Handler returns the Prometheus exposition endpoint handler.
func (m *HTTPMetrics) Handler() http.Handler {
	return m.handler
}

// Middleware instruments request count and request duration.
func (m *HTTPMetrics) Middleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}
			startedAt := time.Now()

			next.ServeHTTP(recorder, r)

			route := routeTemplate(r)
			status := strconv.Itoa(recorder.status)
			duration := time.Since(startedAt).Seconds()

			m.requestTotal.WithLabelValues(r.Method, route, status).Inc()
			m.requestDuration.WithLabelValues(r.Method, route, status).Observe(duration)
		})
	}
}

func routeTemplate(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if route == nil {
		return r.URL.Path
	}
	pathTemplate, err := route.GetPathTemplate()
	if err != nil || pathTemplate == "" {
		return r.URL.Path
	}

	return pathTemplate
}
