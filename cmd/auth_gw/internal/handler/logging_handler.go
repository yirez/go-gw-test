package handler

import (
	"go-gw-test/cmd/auth_gw/internal/usecase"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// LoggingHandler exposes HTTP handlers for logging.
type LoggingHandler struct {
	au usecase.AuthUseCase
}

// NewLoggingHandler constructs an LoggingHandler.
func NewLoggingHandler() *LoggingHandler {
	return &LoggingHandler{}
}

// LoggingMiddleware emits basic access logs for each request.
func (l *LoggingHandler) LoggingMiddleware() mux.MiddlewareFunc {
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
