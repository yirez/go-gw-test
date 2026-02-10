package appservice

import (
	"fmt"
	"net/http"

	"go-gw-test/internal/auth_gw/handler"
	"go-gw-test/internal/auth_gw/repo"
	"go-gw-test/internal/auth_gw/router"
	"go-gw-test/internal/auth_gw/usecase"
	"go-gw-test/pkg/configuration_manager"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AppService wires dependencies and exposes HTTP router and server address.
type AppService struct {
	cfg    configuration_manager.StandardConfig
	db     *gorm.DB
	router http.Handler
}

// NewAppService builds an AppService with all auth_gw components.
func NewAppService(cfg configuration_manager.StandardConfig, db *gorm.DB) *AppService {
	authRepo := repo.NewAuthRepo(db)
	authUsecase := usecase.NewAuthUsecase(authRepo)
	authHandler := handler.NewAuthHandler(authUsecase)

	appRouter := router.NewRouter(authHandler)

	return &AppService{
		cfg:    cfg,
		db:     db,
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
