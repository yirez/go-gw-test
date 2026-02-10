package appservice

import (
	"fmt"
	"net/http"

	"go-gw-test/cmd/auth_gw/internal/handler"
	"go-gw-test/cmd/auth_gw/internal/repo"
	"go-gw-test/cmd/auth_gw/internal/router"
	"go-gw-test/cmd/auth_gw/internal/usecase"

	"go-gw-test/pkg/configuration_manager"

	"go.uber.org/zap"
)

// AppService wires dependencies and exposes HTTP router and server address.
type AppService struct {
	cfg    configuration_manager.StandardConfig
	router http.Handler
}

// NewAppService builds an AppService with all auth_gw components.
func NewAppService(cfg configuration_manager.StandardConfig) *AppService {
	authRepo := repo.NewAuthRepo(cfg.Clients.DB)
	authUsecase := usecase.NewAuthUseCase(authRepo)
	authHandler := handler.NewAuthHandler(authUsecase)

	appRouter := router.NewRouter(authHandler)

	return &AppService{
		cfg:    cfg,
		router: appRouter,
	}
}

// Router returns the root HTTP handler.
func (s *AppService) Router() http.Handler {
	return s.router
}

// Address returns the host:port address for the HTTP server.
func (s *AppService) Address() string {
	return fmt.Sprintf(":%d", s.cfg.Port)
}

// WrapError attaches standard fields to server-level errors.
func WrapError(err error) zap.Field {
	return zap.Error(err)
}
