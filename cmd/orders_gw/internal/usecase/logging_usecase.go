package usecase

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type LoggingUseCase interface {
	LoggingMiddleware() mux.MiddlewareFunc
}

type LoggingUseCaseImpl struct {
}

func NewLoggingUseCaseImpl() *LoggingUseCaseImpl {
	return &LoggingUseCaseImpl{}
}

// LoggingMiddleware emits basic access logs for each request.
func (u *LoggingUseCaseImpl) LoggingMiddleware() mux.MiddlewareFunc {
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
