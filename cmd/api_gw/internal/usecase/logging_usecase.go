package usecase

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type LoggingUseCase interface {
	LoggingMiddleware() mux.MiddlewareFunc
}

type LoggingUseCaseImpl struct {
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func NewLoggingUseCaseImpl() *LoggingUseCaseImpl {
	return &LoggingUseCaseImpl{}
}

// WriteHeader captures status code before writing response headers.
func (s *statusRecorder) WriteHeader(statusCode int) {
	s.status = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}

// LoggingMiddleware emits basic access logs for each request.
func (u *LoggingUseCaseImpl) LoggingMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", recorder.status),
				zap.Duration("latency", time.Since(startedAt)),
				zap.String("request_id", r.Header.Get("X-Request-Id")),
			}

			if recorder.status >= http.StatusInternalServerError {
				zap.L().Error("request", fields...)
				return
			}
			if recorder.status >= http.StatusBadRequest {
				zap.L().Warn("request", fields...)
				return
			}

			zap.L().Info("request", fields...)
		})
	}
}
